package tuner

import (
	"bufio"
	"fmt"
	"math"
	"math/bits"
	"os"
	"strings"
	"sync"

	"github.com/gabtar/aconcagua/aconcagua"
)

const ScalingFactor = 0.0084 // Best Scaling factor found for zurichess training dataset

// DatasetEntry is an struct conatining a single training example
type DatasetEntry struct {
	Fen     string
	Result  float64
	Weights []PositionWeight
	Phase   int
}

// LoadDataSet loads a dataset from a file
func LoadDataSet(filename string) (dataset []DatasetEntry) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\"")
		resultString := map[string]float64{
			"1-0":     1.0,
			"0-1":     0.0,
			"1/2-1/2": 0.5,
		}

		fen := parts[0]
		weigths := generatePositionWeights(fen)
		phase := GetMiddleGamePhase(aconcagua.NewPositionFromFen(fen))
		result := resultString[parts[1]]
		dataset = append(dataset, DatasetEntry{Fen: fen, Result: result, Weights: weigths, Phase: phase})
	}

	return dataset
}

// GetEvaluationParams returns the current evaluation params
func GetEvaluationParams() (params [788]float64) {
	intParams := [788]int{}

	// Psqt weights
	for piece := range 6 {
		copy(intParams[piece*64:(piece+1)*64], aconcagua.MiddlegamePSQT[piece][0:64])
		copy(intParams[(piece+6)*64:(piece+7)*64], aconcagua.EndgamePSQT[piece][0:64])
	}

	// Piece values weights
	copy(intParams[768:774], aconcagua.MiddlegamePieceValue[:])
	copy(intParams[774:780], aconcagua.EndgamePieceValue[:])

	// Mobility weigths
	intParams[780] = aconcagua.QueenMobilityBonusMg
	intParams[781] = aconcagua.QueenMobilityBonusEg

	intParams[782] = aconcagua.RookMobilityBonusMg
	intParams[783] = aconcagua.RookMobilityBonusEg

	intParams[784] = aconcagua.BishopMobilityBonusMg
	intParams[785] = aconcagua.BishopMobilityBonusEg

	intParams[786] = aconcagua.KnightMobilityBonusMg
	intParams[787] = aconcagua.KnightMobilityBonusEg

	// TODO:
	// Pawn Structure Weights
	// 6 + 16 of passed pawns

	for i := range 788 {
		params[i] = float64(intParams[i])
	}
	return
}

// saveParams sotres the best params found in a file
func saveParams(bestParams [788]float64, iteration int) {
	dir := "tuner/params"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	filename := fmt.Sprintf("%s/params_%d.txt", dir, iteration)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	_, err = file.WriteString(paramsToPrettyFormat(bestParams))
	if err != nil {
		fmt.Println(err)
	}
}

// paramsToPrettyFormat returns the params as a string
func paramsToPrettyFormat(bestParams [788]float64) (psqt string) {
	piece := []string{"King", "Queen", "Rook", "Bishop", "Knight", "Pawn"}
	psqt = "MiddlegamePSQT: \n"
	for i := range 6 {
		psqt += fmt.Sprintf("// %s\n", piece[i])
		psqt += "{\n"
		for j := range 8 {
			for k := range 8 {
				psqt += fmt.Sprintf("%d, ", int(bestParams[i*64+j*8+k]))
			}
			psqt += "\n"
		}
		psqt += "},\n"
	}
	psqt += "EndgamePSQT: \n"
	for i := range 6 {
		psqt += fmt.Sprintf("// %s\n", piece[i])
		psqt += "{\n"
		for j := range 8 {
			for k := range 8 {
				psqt += fmt.Sprintf("%d, ", int(bestParams[(i+6)*64+j*8+k]))
			}
			psqt += "\n"
		}
		psqt += "},\n"
	}

	// Pieces Values
	psqt += "MiddlegamePieceValue: "
	for i := range 6 {
		psqt += fmt.Sprintf("%d, ", int(bestParams[768+i]))
	}
	psqt += "\n"
	psqt += "EndgamePieceValue: "
	for i := range 6 {
		psqt += fmt.Sprintf("%d, ", int(bestParams[774+i]))
	}
	psqt += "\n"

	// Mobility Params
	for i := range 4 {
		psqt += fmt.Sprintf("MobilityBonusMg%s: %d\n", piece[i+1], int(bestParams[780+i*2]))
		psqt += fmt.Sprintf("MobilityBonusEg%s: %d\n", piece[i+1], int(bestParams[780+i*2+1]))
	}

	return psqt
}

// WorkerJob represents a single calculation job
type WorkerJob struct {
	entry *DatasetEntry
	index int
}

// WorkerResult represents the result of a calculation
type WorkerResult struct {
	error float64
	index int
}

// MeanSquareError returns the mean square error using parallel processing
func MeanSquareError(scalingFactor float64, params *[788]float64, dataset *[]DatasetEntry) float64 {
	const numWorkers = 4
	entries := len(*dataset)

	jobs := make(chan WorkerJob, entries)
	results := make(chan WorkerResult, entries)
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Add(1)
		go worker(scalingFactor, params, jobs, results, &wg)
	}

	go func() {
		for i := range *dataset {
			jobs <- WorkerJob{entry: &(*dataset)[i], index: i}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	totalError := 0.0
	for result := range results {
		totalError += result.error
	}

	return totalError / float64(entries)
}

// worker processes jobs from the jobs channel
func worker(scalingFactor float64, params *[788]float64, jobs <-chan WorkerJob, results chan<- WorkerResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		score := EvaluatePosition(*params, job.entry.Weights)

		sigmoid := 1 / (1 + math.Exp(-scalingFactor*score))
		errorValue := math.Pow(job.entry.Result-sigmoid, 2)

		results <- WorkerResult{error: errorValue, index: job.index}
	}
}

// PositionWeight is a struct that represents a single score attribute of a position
// The paramIndex corresponds with the index of the params array we are trying to optimize
// If the param index is -1, then the weigth is a fixed value/attribute of the position
// The product of the param value and the weigth value represents the final score
type PositionWeight struct {
	paramIndex int
	weight     int
}

// EvaluatePosition returns the static evaluation of a position based on the weights and current params
func EvaluatePosition(params [788]float64, weights []PositionWeight) (evaluation float64) {
	eval := 0.0
	for _, attr := range weights {
		if attr.paramIndex >= 0 {
			eval += params[attr.paramIndex] * float64(attr.weight)
		} else {
			eval += float64(attr.weight)
		}
	}
	evaluation = eval / 62.0
	return
}

// generatePositionWeights returns all the position weights of a position
func generatePositionWeights(fen string) (weights []PositionWeight) {
	pos := aconcagua.NewPositionFromFen(fen)
	phase := GetMiddleGamePhase(pos)
	weights = generatePieceScoreWeights(fen, phase)
	weights = append(weights, generateMobilityWeights(fen, phase)...)
	weights = append(weights, generateDoubledPawnsWeights(fen, phase)...)
	weights = append(weights, generateIsolatedPawnsWeights(fen, phase)...)
	weights = append(weights, generateBackwardsPawnsWeights(fen, phase)...)
	weights = append(weights, generatePassedPawnsWeights(fen, phase)...)

	return
}

// generatePieceScoreWeights returns the weights of the pieces socre in the board
func generatePieceScoreWeights(fen string, phase int) (weights []PositionWeight) {
	pos := aconcagua.NewPositionFromFen(fen)

	for piece, bb := range pos.Bitboards {
		colorModifier := 1 - int(piece/6)*2
		for bb > 0 {
			sq := aconcagua.Bsf(bb.NextBit())
			if colorModifier == 1 {
				sq = sq ^ 56 // white pieces uses mirror square index in psqt
			}

			// Piece value Mg
			weights = append(weights, PositionWeight{
				paramIndex: 768 + piece%6,
				weight:     colorModifier * phase,
			})
			// Piece value Eg
			weights = append(weights, PositionWeight{
				paramIndex: 768 + piece%6 + 6,
				weight:     colorModifier * (62 - phase),
			})
			// PSQT value Mg
			weights = append(weights, PositionWeight{
				paramIndex: (piece%6)*64 + sq,
				weight:     colorModifier * phase,
			})
			// PSQT value Eg
			weights = append(weights, PositionWeight{
				paramIndex: 384 + (piece%6)*64 + sq,
				weight:     colorModifier * (62 - phase),
			})
		}
	}

	return
}

// generateMobilityWeights returns the position weithts of the mobility of the position
func generateMobilityWeights(fen string, phase int) (weights []PositionWeight) {
	pos := aconcagua.NewPositionFromFen(fen)
	var mobilityBase = [4]int{aconcagua.QueenMobilityBase, aconcagua.RookMobilityBase, aconcagua.BishopMobilityBase, aconcagua.KnightMobilityBase}
	var pieces = [8]int{aconcagua.WhiteQueen, aconcagua.WhiteRook, aconcagua.WhiteBishop, aconcagua.WhiteKnight, aconcagua.BlackQueen, aconcagua.BlackRook, aconcagua.BlackBishop, aconcagua.BlackKnight}

	blocks := ^pos.EmptySquares()
	enemyPawnsAttacks := [2]aconcagua.Bitboard{
		aconcagua.Attacks(aconcagua.BlackPawn, pos.Bitboards[aconcagua.BlackPawn], blocks),
		aconcagua.Attacks(aconcagua.WhitePawn, pos.Bitboards[aconcagua.WhitePawn], blocks),
	}
	whiteBB := pos.Bitboards[1:5]
	blackBB := pos.Bitboards[7:11]
	bitboards := append(whiteBB, blackBB...)

	for piece, bb := range bitboards {
		colorModifier := 1 - int(piece/4)*2
		for bb > 0 {
			fromBB := bb.NextBit()
			attacks := aconcagua.Attacks(pieces[piece], fromBB, blocks)
			safeSquares := bits.OnesCount64(uint64(attacks & ^enemyPawnsAttacks[piece/4])) - mobilityBase[piece%4]

			// mg
			weights = append(weights, PositionWeight{
				paramIndex: 780 + piece%4*2,
				weight:     colorModifier * safeSquares * phase,
			})

			// eg
			weights = append(weights, PositionWeight{
				paramIndex: 780 + piece%4*2 + 1,
				weight:     colorModifier * safeSquares * (62 - phase),
			})
		}
	}
	return
}

// generateDoubledPawnsWeights returns the position weights of the doubled pawns
func generateDoubledPawnsWeights(fen string, phase int) (weights []PositionWeight) {
	pos := aconcagua.NewPositionFromFen(fen)
	wDoubled := bits.OnesCount64(uint64(aconcagua.DoubledPawns(pos, aconcagua.White)))
	bDoubled := bits.OnesCount64(uint64(aconcagua.DoubledPawns(pos, aconcagua.Black)))
	penalties := [2]int{aconcagua.DoubledPawnPenaltyMg, aconcagua.DoubledPawnPenaltyEg}
	sideModifier := [2]int{1, -1}
	doubledPawns := [2]int{wDoubled, bDoubled}

	for i := range 2 {
		weights = append(weights, PositionWeight{
			paramIndex: -1,
			weight:     sideModifier[i] * (penalties[0]*doubledPawns[i]*phase + penalties[1]*doubledPawns[i]*(62-phase)),
		})
	}

	return
}

// generateIsolatedPawnsWeights returns the position weights of the isolated pawns
func generateIsolatedPawnsWeights(fen string, phase int) (weights []PositionWeight) {
	pos := aconcagua.NewPositionFromFen(fen)
	wIsolated := bits.OnesCount64(uint64(aconcagua.IsolatedPawns(pos, aconcagua.White)))
	bIsolated := bits.OnesCount64(uint64(aconcagua.IsolatedPawns(pos, aconcagua.Black)))
	penalties := [2]int{aconcagua.IsolatedPawnPenaltyMg, aconcagua.IsolatedPawnPenaltyEg}
	sideModifier := [2]int{1, -1}
	isolatedPawns := [2]int{wIsolated, bIsolated}

	for i := range 2 {
		weights = append(weights, PositionWeight{
			paramIndex: -1,
			weight:     sideModifier[i] * (penalties[0]*isolatedPawns[i]*phase + penalties[1]*isolatedPawns[i]*(62-phase)),
		})
	}

	return
}

// generateBackwardsPawnsWeights returns the position weights of the backwards pawns
func generateBackwardsPawnsWeights(fen string, phase int) (weights []PositionWeight) {
	pos := aconcagua.NewPositionFromFen(fen)
	whitePawnsAttacks := aconcagua.Attacks(aconcagua.WhitePawn, pos.Bitboards[aconcagua.WhitePawn], pos.EmptySquares())
	blackPawnsAttacks := aconcagua.Attacks(aconcagua.BlackPawn, pos.Bitboards[aconcagua.BlackPawn], pos.EmptySquares())

	wBackwards := bits.OnesCount64(uint64(aconcagua.BackwardPawns(pos.Bitboards[aconcagua.WhitePawn], blackPawnsAttacks, aconcagua.White)))
	bBackwards := bits.OnesCount64(uint64(aconcagua.BackwardPawns(pos.Bitboards[aconcagua.BlackPawn], whitePawnsAttacks, aconcagua.Black)))
	penalties := [2]int{aconcagua.BackwardPawnPenaltyMg, aconcagua.BackwardPawnPenaltyEg}
	sideModifier := [2]int{1, -1}
	backwardsPawns := [2]int{wBackwards, bBackwards}

	for i := range 2 {
		weights = append(weights, PositionWeight{
			paramIndex: -1,
			weight:     sideModifier[i] * (penalties[0]*backwardsPawns[i]*phase + penalties[1]*backwardsPawns[i]*(62-phase)),
		})
	}

	return
}

// generatePassedPawnsWeights returns the position weights of the passed pawns
func generatePassedPawnsWeights(fen string, phase int) (weights []PositionWeight) {
	pos := aconcagua.NewPositionFromFen(fen)
	passedPawnBonus := [8]int{0, 0, 10, 20, 30, 40, 50, 0}
	wPassedPawns := aconcagua.PassedPawns(pos.Bitboards[aconcagua.WhitePawn], pos.Bitboards[aconcagua.BlackPawn], aconcagua.White)
	bPassedPawns := aconcagua.PassedPawns(pos.Bitboards[aconcagua.BlackPawn], pos.Bitboards[aconcagua.WhitePawn], aconcagua.Black)

	bonus := 0
	for wPassedPawns > 0 {
		fromBB := wPassedPawns.NextBit()
		sq := aconcagua.Bsf(fromBB)
		rank := sq / 8
		bonus += passedPawnBonus[rank]
	}
	weights = append(weights, PositionWeight{
		paramIndex: -1,
		weight:     bonus*phase + bonus*2*(62-phase),
	})

	bonus = 0
	for bPassedPawns > 0 {
		fromBB := bPassedPawns.NextBit()
		sq := aconcagua.Bsf(fromBB)
		rank := 7 - sq/8
		bonus += passedPawnBonus[rank]
	}
	weights = append(weights, PositionWeight{
		paramIndex: -1,
		weight:     -bonus*phase - bonus*2*(62-phase),
	})

	return
}

// GetMiddleGamePhase returns the value of the middle game phase of a position
func GetMiddleGamePhase(pos *aconcagua.Position) (mgPhase int) {
	phaseInc := [6]int{0, 9, 5, 3, 3, 0}
	for p, bb := range pos.Bitboards {
		for bb > 0 {
			bb.NextBit()
			mgPhase += phaseInc[p%6]
		}
	}
	mgPhase = min(mgPhase, 62)
	return
}

// FindOptimalScalingFactor returns the scaling factor that minimizes the mean square error
func FindOptimalScalingFactor(dataset []DatasetEntry, params [788]float64) float64 {
	bestK := 0.0
	bestError := math.Inf(1)

	for k := 0.0001; k <= 0.1; k += 0.0001 {
		totalError := 0.0

		for _, entry := range dataset {
			eval := EvaluatePosition(params, entry.Weights)
			predicted := 1.0 / (1.0 + math.Exp(-k*eval))
			actual := entry.Result
			error := predicted - actual
			totalError += error * error
		}

		mse := totalError / float64(len(dataset))
		if mse < bestError {
			bestError = mse
			bestK = k
		}
	}

	return bestK
}
