package tuner

import (
	"bufio"
	"fmt"
	"math"
	"math/bits"
	"os"
	"strings"
	"sync"

	"github.com/gabtar/aconcagua/internal/engine"
)

const ScalingFactor = 0.0084 // Best Scaling factor found for zurichess training dataset

// DatasetEntry is an struct conatining a single training example
type DatasetEntry struct {
	Fen     string
	Result  float64
	Weights []PositionWeight
	Phase   int
}

// NewDataset returns a new preallocated dataset
func NewDataset(size int) (dataset []DatasetEntry) {
	dataset = make([]DatasetEntry, size)
	for i := range size {
		dataset[i] = DatasetEntry{
			Fen:     "",
			Result:  0.0,
			Weights: make([]PositionWeight, 0, 810),
			Phase:   0,
		}
	}

	return
}

// LoadDataSet loads a dataset from a file
func LoadDataSet(filename string, size int) (dataset []DatasetEntry) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	// Preallocate memory to load entries faster
	dataset = NewDataset(size)

	scanner := bufio.NewScanner(file)

	// Increase buffer size for faster scanning
	buf := make([]byte, 0, 1024*1024) // 1MB buffer
	scanner.Buffer(buf, 1024*1024)

	pos := engine.NewPosition()
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		// parts := strings.Split(line, "\"") for zurichess dataset
		parts := strings.Split(line, "[") // for lichess-big3-resolved dataset
		resultString := map[string]float64{
			"1-0":     1.0,
			"0-1":     0.0,
			"1/2-1/2": 0.5,
		}

		fen := parts[0]
		weights := generatePositionWeights(fen)
		pos.LoadFromFenString(fen)
		phase := getMiddleGamePhase(pos)
		result := resultString[parts[1]]
		dataset = append(dataset, DatasetEntry{Fen: fen, Result: result, Weights: weights, Phase: phase})

		count++
		if count >= size {
			break
		}

		// Progress indicator
		if count%100000 == 0 {
			fmt.Printf("Loaded %d positions...\n", count)
		}
	}

	return dataset
}

// GetEvaluationParams returns the current evaluation params
func GetEvaluationParams() (params [810]float64) {
	intParams := [810]int{}

	// Psqt params
	for piece := range 6 {
		copy(intParams[piece*64:(piece+1)*64], engine.MiddlegamePSQT[piece][0:64])
		copy(intParams[(piece+6)*64:(piece+7)*64], engine.EndgamePSQT[piece][0:64])
	}

	// Piece values params
	copy(intParams[768:774], engine.MiddlegamePieceValue[:])
	copy(intParams[774:780], engine.EndgamePieceValue[:])

	// Mobility params
	intParams[780] = engine.QueenMobilityBonusMg
	intParams[781] = engine.QueenMobilityBonusEg

	intParams[782] = engine.RookMobilityBonusMg
	intParams[783] = engine.RookMobilityBonusEg

	intParams[784] = engine.BishopMobilityBonusMg
	intParams[785] = engine.BishopMobilityBonusEg

	intParams[786] = engine.KnightMobilityBonusMg
	intParams[787] = engine.KnightMobilityBonusEg

	// Pawn Structure params
	intParams[788] = engine.DoubledPawnPenaltyMg
	intParams[789] = engine.DoubledPawnPenaltyEg

	intParams[790] = engine.IsolatedPawnPenaltyMg
	intParams[791] = engine.IsolatedPawnPenaltyEg

	intParams[792] = engine.BackwardPawnPenaltyMg
	intParams[793] = engine.BackwardPawnPenaltyEg

	for i := range 8 {
		intParams[794+i] = engine.PassedPawnsBonusMg[i]
		intParams[802+i] = engine.PassedPawnsBonusEg[i]
	}

	for i := range 810 {
		params[i] = float64(intParams[i])
	}
	return
}

// saveParams sotres the best params found in a file
func saveParams(bestParams [810]float64, iteration int) {
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
func paramsToPrettyFormat(bestParams [810]float64) (psqt string) {
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

	// Pawn Structure Weights
	psqt += fmt.Sprintf("DoubledPawnPenaltyMg: %d\n", int(bestParams[788]))
	psqt += fmt.Sprintf("DoubledPawnPenaltyEg: %d\n", int(bestParams[789]))
	psqt += fmt.Sprintf("IsolatedPawnPenaltyMg: %d\n", int(bestParams[790]))
	psqt += fmt.Sprintf("IsolatedPawnPenaltyEg: %d\n", int(bestParams[791]))
	psqt += fmt.Sprintf("BackwardPawnPenaltyMg: %d\n", int(bestParams[792]))
	psqt += fmt.Sprintf("BackwardPawnPenaltyEg: %d\n", int(bestParams[793]))

	// Passed Pawns
	psqt += "PassedPawnsBonusMg: "
	for i := range 8 {
		psqt += fmt.Sprintf("%d, ", int(bestParams[794+i]))
	}
	psqt += "\n"
	psqt += "PassedPawnsBonusEg: "
	for i := range 8 {
		psqt += fmt.Sprintf("%d, ", int(bestParams[802+i]))
	}
	psqt += "\n"

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
func MeanSquareError(scalingFactor float64, params *[810]float64, dataset *[]DatasetEntry) float64 {
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
func worker(scalingFactor float64, params *[810]float64, jobs <-chan WorkerJob, results chan<- WorkerResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		score := evaluatePosition(*params, job.entry.Weights)

		sigmoid := 1 / (1 + math.Exp(-scalingFactor*score))
		errorValue := math.Pow(job.entry.Result-sigmoid, 2)

		results <- WorkerResult{error: errorValue, index: job.index}
	}
}

// PositionWeight is a struct that represents a single score attribute of a position
// The paramIndex corresponds with the index of the params array we are trying to optimize
// The product of the param value and the weigth value represents the final score
type PositionWeight struct {
	paramIndex int16
	weight     int16
}

// evaluatePosition returns the static evaluation of a position based on the weights and current params
func evaluatePosition(params [810]float64, weights []PositionWeight) (evaluation float64) {
	eval := 0.0
	for _, attr := range weights {
		eval += params[attr.paramIndex] * float64(attr.weight)
	}
	evaluation = eval / 62.0
	return
}

// generatePositionWeights returns all the position weights of a position
func generatePositionWeights(fen string) (weights []PositionWeight) {
	pos := engine.NewPosition()
	pos.LoadFromFenString(fen)
	phase := getMiddleGamePhase(pos)
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
	pos := engine.NewPosition()
	pos.LoadFromFenString(fen)

	for piece, bb := range pos.Bitboards {
		colorModifier := 1 - int(piece/6)*2
		for bb > 0 {
			sq := engine.Bsf(bb.NextBit())
			if colorModifier == 1 {
				sq = sq ^ 56 // white pieces uses mirror square index in psqt
			}

			// Piece value Mg
			weights = append(weights, PositionWeight{
				paramIndex: int16(768 + piece%6),
				weight:     int16(colorModifier * phase),
			})
			// Piece value Eg
			weights = append(weights, PositionWeight{
				paramIndex: int16(768 + piece%6 + 6),
				weight:     int16(colorModifier * (62 - phase)),
			})
			// PSQT value Mg
			weights = append(weights, PositionWeight{
				paramIndex: int16((piece%6)*64 + sq),
				weight:     int16(colorModifier * phase),
			})
			// PSQT value Eg
			weights = append(weights, PositionWeight{
				paramIndex: int16(384 + (piece%6)*64 + sq),
				weight:     int16(colorModifier * (62 - phase)),
			})
		}
	}

	return
}

// generateMobilityWeights returns the position weithts of the mobility of the position
func generateMobilityWeights(fen string, phase int) (weights []PositionWeight) {
	pos := engine.NewPosition()
	pos.LoadFromFenString(fen)
	var mobilityBase = [4]int{engine.QueenMobilityBase, engine.RookMobilityBase, engine.BishopMobilityBase, engine.KnightMobilityBase}
	var pieces = [8]int{engine.WhiteQueen, engine.WhiteRook, engine.WhiteBishop, engine.WhiteKnight, engine.BlackQueen, engine.BlackRook, engine.BlackBishop, engine.BlackKnight}

	blocks := ^pos.EmptySquares()
	enemyPawnsAttacks := [2]engine.Bitboard{
		engine.Attacks(engine.BlackPawn, pos.Bitboards[engine.BlackPawn], blocks),
		engine.Attacks(engine.WhitePawn, pos.Bitboards[engine.WhitePawn], blocks),
	}
	whiteBB := pos.Bitboards[1:5]
	blackBB := pos.Bitboards[7:11]
	bitboards := append(whiteBB, blackBB...)

	for piece, bb := range bitboards {
		colorModifier := 1 - int(piece/4)*2
		for bb > 0 {
			fromBB := bb.NextBit()
			attacks := engine.Attacks(pieces[piece], fromBB, blocks)
			safeSquares := bits.OnesCount64(uint64(attacks & ^enemyPawnsAttacks[piece/4])) - mobilityBase[piece%4]

			// mg
			weights = append(weights, PositionWeight{
				paramIndex: int16(780 + piece%4*2),
				weight:     int16(colorModifier * safeSquares * phase),
			})

			// eg
			weights = append(weights, PositionWeight{
				paramIndex: int16(780 + piece%4*2 + 1),
				weight:     int16(colorModifier * safeSquares * (62 - phase)),
			})
		}
	}
	return
}

// generateDoubledPawnsWeights returns the position weights of the doubled pawns
func generateDoubledPawnsWeights(fen string, phase int) (weights []PositionWeight) {
	pos := engine.NewPosition()
	pos.LoadFromFenString(fen)
	wDoubled := bits.OnesCount64(uint64(engine.DoubledPawns(pos, engine.White)))
	bDoubled := bits.OnesCount64(uint64(engine.DoubledPawns(pos, engine.Black)))
	sideModifier := [2]int{1, -1}
	doubledPawns := [2]int{wDoubled, bDoubled}

	for i := range 2 {
		weights = append(weights, PositionWeight{
			paramIndex: int16(788),
			weight:     int16(sideModifier[i] * doubledPawns[i] * phase),
		})

		weights = append(weights, PositionWeight{
			paramIndex: int16(789),
			weight:     int16(sideModifier[i] * doubledPawns[i] * (62 - phase)),
		})
	}

	return
}

// generateIsolatedPawnsWeights returns the position weights of the isolated pawns
func generateIsolatedPawnsWeights(fen string, phase int) (weights []PositionWeight) {
	pos := engine.NewPosition()
	pos.LoadFromFenString(fen)
	wIsolated := bits.OnesCount64(uint64(engine.IsolatedPawns(pos, engine.White)))
	bIsolated := bits.OnesCount64(uint64(engine.IsolatedPawns(pos, engine.Black)))
	sideModifier := [2]int{1, -1}
	isolatedPawns := [2]int{wIsolated, bIsolated}

	for i := range 2 {
		weights = append(weights, PositionWeight{
			paramIndex: int16(790),
			weight:     int16(sideModifier[i] * isolatedPawns[i] * phase),
		})

		weights = append(weights, PositionWeight{
			paramIndex: int16(791),
			weight:     int16(sideModifier[i] * isolatedPawns[i] * (62 - phase)),
		})
	}

	return
}

// generateBackwardsPawnsWeights returns the position weights of the backwards pawns
func generateBackwardsPawnsWeights(fen string, phase int) (weights []PositionWeight) {
	pos := engine.NewPosition()
	pos.LoadFromFenString(fen)
	whitePawnsAttacks := engine.Attacks(engine.WhitePawn, pos.Bitboards[engine.WhitePawn], pos.EmptySquares())
	blackPawnsAttacks := engine.Attacks(engine.BlackPawn, pos.Bitboards[engine.BlackPawn], pos.EmptySquares())

	wBackwards := bits.OnesCount64(uint64(engine.BackwardPawns(pos.Bitboards[engine.WhitePawn], blackPawnsAttacks, engine.White)))
	bBackwards := bits.OnesCount64(uint64(engine.BackwardPawns(pos.Bitboards[engine.BlackPawn], whitePawnsAttacks, engine.Black)))
	sideModifier := [2]int{1, -1}
	backwardsPawns := [2]int{wBackwards, bBackwards}

	for i := range 2 {
		weights = append(weights, PositionWeight{
			paramIndex: int16(792),
			weight:     int16(sideModifier[i] * backwardsPawns[i] * phase),
		})

		weights = append(weights, PositionWeight{
			paramIndex: int16(793),
			weight:     int16(sideModifier[i] * backwardsPawns[i] * (62 - phase)),
		})
	}

	return
}

// generatePassedPawnsWeights returns the position weights of the passed pawns
func generatePassedPawnsWeights(fen string, phase int) (weights []PositionWeight) {
	pos := engine.NewPosition()
	pos.LoadFromFenString(fen)
	wPassedPawns := engine.PassedPawns(pos.Bitboards[engine.WhitePawn], pos.Bitboards[engine.BlackPawn], engine.White)
	bPassedPawns := engine.PassedPawns(pos.Bitboards[engine.BlackPawn], pos.Bitboards[engine.WhitePawn], engine.Black)

	for wPassedPawns > 0 {
		fromBB := wPassedPawns.NextBit()
		sq := engine.Bsf(fromBB)
		rank := sq / 8

		weights = append(weights, PositionWeight{
			paramIndex: int16(794 + rank),
			weight:     int16(phase),
		})
		weights = append(weights, PositionWeight{
			paramIndex: int16(802 + rank),
			weight:     int16(62 - phase),
		})
	}

	for bPassedPawns > 0 {
		fromBB := bPassedPawns.NextBit()
		sq := engine.Bsf(fromBB)
		rank := 7 - sq/8

		weights = append(weights, PositionWeight{
			paramIndex: int16(794 + rank),
			weight:     int16(-1 * phase),
		})
		weights = append(weights, PositionWeight{
			paramIndex: int16(802 + rank),
			weight:     int16(-1 * (62 - phase)),
		})
	}

	return
}

// getMiddleGamePhase returns the value of the middle game phase of a position
func getMiddleGamePhase(pos *engine.Position) (mgPhase int) {
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
func FindOptimalScalingFactor(dataset []DatasetEntry, params [810]float64) float64 {
	bestK := 0.0
	bestError := math.Inf(1)

	for k := 0.0001; k <= 0.1; k += 0.0001 {
		totalError := 0.0

		for _, entry := range dataset {
			eval := evaluatePosition(params, entry.Weights)
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
