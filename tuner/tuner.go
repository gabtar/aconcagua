package tuner

import (
	"bufio"
	"fmt"
	"math"
	"math/bits"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gabtar/aconcagua/aconcagua"
)

// DataEntry is an struct conatining a single training example
type DataEntry struct {
	Fen         string
	Result      float64
	Attributes  []PositionAttribute
	Phase       int
	WhiteToMove bool
}

// LoadDataSet loads a dataset from a file
func LoadDataSet(filename string) (dataset []DataEntry) {
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
		whiteToMove := aconcagua.NewPositionFromFen(fen).Turn == aconcagua.White
		attributes := GeneratePositionAttributes(fen)
		phase := GetPhase(aconcagua.NewPositionFromFen(fen))
		result := resultString[parts[1][0:len(parts[1])-3]]
		dataset = append(dataset, DataEntry{Fen: fen, Result: result, Attributes: attributes, Phase: phase, WhiteToMove: whiteToMove})
	}

	return dataset
}

// GetEvaluationParams returns the current evaluation params
func GetEvaluationParams() (params [768]int) {
	for piece := range 6 {
		copy(params[piece*64:(piece+1)*64], aconcagua.MiddlegamePSQT[piece][0:64])
		copy(params[(piece+6)*64:(piece+7)*64], aconcagua.EndgamePSQT[piece][0:64])
	}
	return params
}

// Tuner finds a set of parameters that minimize the mean square error
func Tuner(scalingFactor float64, dataset []DataEntry, iteration int) {
	paramAdjustValue := 1 // increment/decrement params by this value
	bestParams := GetEvaluationParams()
	bestErr := MeanSquareError(scalingFactor, &bestParams, &dataset)
	improved := true
	startTime := time.Now()

	for improved {
		iterationStartTime := time.Now()
		improved = false

		paramsTuned := 0
		for i := range len(bestParams) {
			newParams := bestParams
			newParams[i] += paramAdjustValue
			err := MeanSquareError(scalingFactor, &newParams, &dataset)

			if err < bestErr {
				paramsTuned++
				bestParams[i] = newParams[i]
				improved = true
				bestErr = err
				continue
			} else {
				newParams[i] -= 2 * paramAdjustValue
				err = MeanSquareError(scalingFactor, &newParams, &dataset)
				if err < bestErr {
					paramsTuned++
					bestParams[i] = newParams[i]
					bestErr = err
					improved = true
					continue
				}
			}
		}

		savePSQT(bestParams, iteration)
		sessionTime := time.Since(startTime)
		iterationTime := time.Since(iterationStartTime)
		fmt.Printf("Iter #%d | MeanSqErr: %.8f | Params tuned: %d | Iteration Time: %.2f mins | Session Time: %.2f hours\n", iteration, bestErr, paramsTuned, iterationTime.Minutes(), sessionTime.Hours())
		iteration++
	}
}

// savePSQT sotres the best params found in a file
func savePSQT(bestParams [768]int, iteration int) {
	dir := "PSQT"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}

	filename := fmt.Sprintf("%s/PSQT_%d.txt", dir, iteration)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	_, err = file.WriteString(PSQTtoString(bestParams))
	if err != nil {
		fmt.Println(err)
	}
}

// PSQTtoString returns the PSQT as a string
func PSQTtoString(bestParams [768]int) (psqt string) {
	piece := []string{"King", "Queen", "Rook", "Bishop", "Knight", "Pawn"}
	psqt = "MiddlegamePSQT: \n"
	for i := range 6 {
		psqt += fmt.Sprintf("// %s\n", piece[i])
		psqt += "{\n"
		for j := range 8 {
			for k := range 8 {
				psqt += fmt.Sprintf("%d, ", bestParams[i*64+j*8+k])
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
				psqt += fmt.Sprintf("%d, ", bestParams[(i+6)*64+j*8+k])
			}
			psqt += "\n"
		}
		psqt += "},\n"
	}

	return psqt
}

// TODO:: Some developers use a gradient descent to find the local minima. Its faster than this routine, but i need to find out how to implement it
// Code from CPW of a local optimization routine
// C++ optimization code

// vector<int> localOptimize(const vector<int>& initialGuess) {
//    const int nParams = initialGuess.size();
//    double bestE = E(initialGuess);
//    vector<int> bestParValues = initialGuess;
//    bool improved = true;
//    while ( improved ) {
//       improved = false;
//       for (int pi = 0; pi < nParams; pi++) {
//          vector<int> newParValues = bestParValues;
//          newParValues[pi] += 1;
//          double newE = E(newParValues);
//          if (newE < bestE) {
//             bestE = newE;
//             bestParValues = newParValues;
//             improved = true;
//          } else {
//             newParValues[pi] -= 2;
//             newE = E(newParValues);
//             if (newE < bestE) {
//                bestE = newE;
//                bestParValues = newParValues;
//                improved = true;
//             }
//          }
//       }
//    }
//    return bestParValues;
// }
//

// WorkerJob represents a single calculation job
type WorkerJob struct {
	entry *DataEntry
	index int
}

// WorkerResult represents the result of a calculation
type WorkerResult struct {
	error float64
	index int
}

// MeanSquareError returns the mean square error using parallel processing
func MeanSquareError(scalingFactor float64, params *[768]int, dataset *[]DataEntry) float64 {
	const numWorkers = 4 // limit to 3 threads/jobs at a time
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
func worker(scalingFactor float64, params *[768]int, jobs <-chan WorkerJob, results chan<- WorkerResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		score := EvaluatePosition(*params, job.entry.Attributes, job.entry.Phase, job.entry.WhiteToMove)

		sigmoid := 1 / (1 + math.Exp(-scalingFactor*float64(score)))
		errorValue := math.Pow(job.entry.Result-sigmoid, 2)

		results <- WorkerResult{error: errorValue, index: job.index}
	}
}

// PositionAttribute is a struct that represents a single attribute of a position
// The paramIndex corresponds with the index of the params array we are trying to optimize
// If the param index is -1, then the bias is a fixed value/attribute of the position
// The bias contains the a fixed value/attribute of the position
// For example if we are not tuning mobility, the mobility attributes only have a bias with a fixed value for each position
type PositionAttribute struct {
	paramIndex int
	bias       int
	color      int // 0 white - 1 black
}

// EvaluatePosition returns the static evaluation of a position based on the attributes and current params
func EvaluatePosition(params [768]int, attributes []PositionAttribute, phase int, whiteToMove bool) (value int) {
	psqtScore := 0
	staticValues := 0
	for _, attr := range attributes {
		if attr.paramIndex >= 0 {
			psqtScore += attr.color * (params[attr.paramIndex]*phase + (params[attr.paramIndex+384])*(62-phase) + attr.bias)
		} else {
			staticValues += attr.bias * attr.color
		}
	}
	value = (psqtScore + staticValues) / 62
	if !whiteToMove {
		value = -value
	}
	return
}

// GeneratePositionAttributes returns all the position attributes of a position
func GeneratePositionAttributes(fen string) (attributes []PositionAttribute) {
	pos := aconcagua.NewPositionFromFen(fen)
	phase := GetPhase(pos)
	attributes = generatePieceScoreAttributes(fen, phase)
	attributes = append(attributes, generateMobilityAttributes(fen, phase)...)
	attributes = append(attributes, generateDoubledPawnsAttributes(fen, phase)...)
	attributes = append(attributes, generateIsolatedPawnsAttributes(fen, phase)...)
	attributes = append(attributes, generateBackwardsPawnsAttributes(fen, phase)...)
	attributes = append(attributes, generatePassedPawnsAttributes(fen, phase)...)

	return
}

// generatePieceScoreAttributes returns the attributes of the piece socre (piece value + psqt)
func generatePieceScoreAttributes(fen string, phase int) (attributes []PositionAttribute) {
	// TODO: split into psqt and piece value attributes
	pos := aconcagua.NewPositionFromFen(fen)

	for piece, bb := range pos.Bitboards {
		colorModifier := 1 - int(piece/6)*2
		for bb > 0 {
			sq := aconcagua.Bsf(bb.NextBit())
			if colorModifier == 1 {
				sq = sq ^ 56 // white pieces uses mirror square index in psqt
			}

			attributes = append(attributes, PositionAttribute{
				paramIndex: (piece%6)*64 + sq,
				bias:       aconcagua.MiddlegamePieceValue[piece%6]*phase + aconcagua.EndgamePieceValue[piece%6]*(62-phase),
				color:      colorModifier,
			})
		}
	}

	return
}

// generateMobilityAttributes returns the attributes of the mobility of the position
func generateMobilityAttributes(fen string, phase int) (attributes []PositionAttribute) {
	pos := aconcagua.NewPositionFromFen(fen)
	var mgMobilityBonus = [4]int{aconcagua.QueenMobilityBonusMg, aconcagua.RookMobilityBonusMg, aconcagua.BishopMobilityBonusMg, aconcagua.KnightMobilityBonusMg}
	var egMobilityBonus = [4]int{aconcagua.QueenMobilityBonusEg, aconcagua.RookMobilityBonusEg, aconcagua.BishopMobilityBonusEg, aconcagua.KnightMobilityBonusEg}
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

			attributes = append(attributes, PositionAttribute{
				paramIndex: -1,
				bias:       safeSquares*mgMobilityBonus[piece%4]*phase + safeSquares*egMobilityBonus[piece%4]*(62-phase),
				color:      colorModifier,
			})
		}
	}
	return
}

// generateDoubledPawnsAttributes returns the attributes of the doubled pawns
func generateDoubledPawnsAttributes(fen string, phase int) (attributes []PositionAttribute) {
	pos := aconcagua.NewPositionFromFen(fen)
	wDoubled := bits.OnesCount64(uint64(aconcagua.DoubledPawns(pos, aconcagua.White)))
	bDoubled := bits.OnesCount64(uint64(aconcagua.DoubledPawns(pos, aconcagua.Black)))
	penalties := [2]int{aconcagua.DoubledPawnPenaltyMg, aconcagua.DoubledPawnPenaltyEg}
	sideModifier := [2]int{1, -1}
	doubledPawns := [2]int{wDoubled, bDoubled}

	for i := range 2 {
		attributes = append(attributes, PositionAttribute{
			paramIndex: -1,
			bias:       penalties[0]*doubledPawns[i]*phase + penalties[1]*doubledPawns[i]*(62-phase),
			color:      sideModifier[i],
		})
	}

	return
}

// generateIsolatedPawnsAttributes returns the attributes of the isolated pawns
func generateIsolatedPawnsAttributes(fen string, phase int) (attributes []PositionAttribute) {
	pos := aconcagua.NewPositionFromFen(fen)
	wIsolated := bits.OnesCount64(uint64(aconcagua.IsolatedPawns(pos, aconcagua.White)))
	bIsolated := bits.OnesCount64(uint64(aconcagua.IsolatedPawns(pos, aconcagua.Black)))
	penalties := [2]int{aconcagua.IsolatedPawnPenaltyMg, aconcagua.IsolatedPawnPenaltyEg}
	sideModifier := [2]int{1, -1}
	isolatedPawns := [2]int{wIsolated, bIsolated}

	for i := range 2 {
		attributes = append(attributes, PositionAttribute{
			paramIndex: -1,
			bias:       penalties[0]*isolatedPawns[i]*phase + penalties[1]*isolatedPawns[i]*(62-phase),
			color:      sideModifier[i],
		})
	}

	return
}

// generateBackwardsPawnsAttributes returns the attributes of the backwards pawns
func generateBackwardsPawnsAttributes(fen string, phase int) (attributes []PositionAttribute) {
	pos := aconcagua.NewPositionFromFen(fen)
	whitePawnsAttacks := aconcagua.Attacks(aconcagua.WhitePawn, pos.Bitboards[aconcagua.WhitePawn], pos.EmptySquares())
	blackPawnsAttacks := aconcagua.Attacks(aconcagua.BlackPawn, pos.Bitboards[aconcagua.BlackPawn], pos.EmptySquares())

	wBackwards := bits.OnesCount64(uint64(aconcagua.BackwardPawns(pos.Bitboards[aconcagua.WhitePawn], blackPawnsAttacks, aconcagua.White)))
	bBackwards := bits.OnesCount64(uint64(aconcagua.BackwardPawns(pos.Bitboards[aconcagua.BlackPawn], whitePawnsAttacks, aconcagua.Black)))
	penalties := [2]int{aconcagua.BackwardPawnPenaltyMg, aconcagua.BackwardPawnPenaltyEg}
	sideModifier := [2]int{1, -1}
	backwardsPawns := [2]int{wBackwards, bBackwards}

	for i := range 2 {
		attributes = append(attributes, PositionAttribute{
			paramIndex: -1,
			bias:       penalties[0]*backwardsPawns[i]*phase + penalties[1]*backwardsPawns[i]*(62-phase),
			color:      sideModifier[i],
		})
	}

	return
}

// generatePassedPawnsAttributes returns the attributes of the passed pawns
func generatePassedPawnsAttributes(fen string, phase int) (attributes []PositionAttribute) {
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
	attributes = append(attributes, PositionAttribute{
		paramIndex: -1,
		bias:       bonus*phase + bonus*2*(62-phase),
		color:      1,
	})

	bonus = 0
	for bPassedPawns > 0 {
		fromBB := bPassedPawns.NextBit()
		sq := aconcagua.Bsf(fromBB)
		rank := 7 - sq/8
		bonus += passedPawnBonus[rank]
	}
	attributes = append(attributes, PositionAttribute{
		paramIndex: -1,
		bias:       bonus*phase + bonus*2*(62-phase),
		color:      -1,
	})

	return
}

func GetPhase(pos *aconcagua.Position) (phase int) {
	phaseInc := [6]int{0, 9, 5, 3, 3, 0}
	for p, bb := range pos.Bitboards {
		for bb > 0 {
			bb.NextBit()
			phase += phaseInc[p%6]
		}
	}
	phase = min(phase, 62)
	return
}
