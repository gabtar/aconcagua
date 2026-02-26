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

	"github.com/gabtar/aconcagua/internal/engine"
)

// const ScalingFactor = 0.0084 // Best Scaling factor found for zurichess training dataset

// ScalingFactor is the scaling factor for the training dataset
const ScalingFactor = 0.008000000000000007 // lichess-big3-resolved

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
			Weights: make([]PositionWeight, 0, 200),
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
	start := time.Now()
	for scanner.Scan() {
		line := scanner.Text()
		// parts := strings.Split(line, "\"") for zurichess dataset
		// resultString := map[string]float64{
		// 	"1-0":     1.0,
		// 	"0-1":     0.0,
		// 	"1/2-1/2": 0.5,
		// }
		parts := strings.Split(line, "[") // for lichess-big3-resolved dataset
		resultString := map[string]float64{
			"1.0]": 1.0,
			"0.0]": 0.0,
			"0.5]": 0.5,
		}

		fen := parts[0]
		pos.LoadFromFenString(fen)
		phase := getMiddleGamePhase(pos)
		result := resultString[parts[1]]

		generatePositionWeights(pos, phase, &dataset[count].Weights)
		dataset[count].Fen = fen
		dataset[count].Result = result
		dataset[count].Phase = phase

		count++
		if count >= size {
			break
		}

		if count%100000 == 0 {
			elapsed := time.Since(start)
			fmt.Printf("Loaded %d entries in %s\n", count, elapsed)
		}
	}

	return dataset
}

// GetEvaluationParams returns the current evaluation params
func GetEvaluationParams() (params [1000]float64) {
	intParams := [1000]int{}

	// Psqt params
	for piece := range 6 {
		copy(intParams[piece*64:(piece+1)*64], engine.MiddlegamePSQT[piece][0:64])
		copy(intParams[(piece+6)*64:(piece+7)*64], engine.EndgamePSQT[piece][0:64])
	}

	// Piece values params
	copy(intParams[768:774], engine.MiddlegamePieceValue[:])
	copy(intParams[774:780], engine.EndgamePieceValue[:])

	// Mobility params
	copy(intParams[780:808], engine.QueenMobilityMg[:])
	copy(intParams[808:836], engine.QueenMobilityEg[:])
	copy(intParams[836:851], engine.RookMobilityMg[:])
	copy(intParams[851:866], engine.RookMobilityEg[:])
	copy(intParams[866:880], engine.BishopMobilityMg[:])
	copy(intParams[880:894], engine.BishopMobilityEg[:])
	copy(intParams[894:903], engine.KnightMobilityMg[:])
	copy(intParams[903:912], engine.KnightMobilityEg[:])

	// Pawn Structure params
	intParams[912] = engine.DoubledPawnPenaltyMg
	intParams[913] = engine.DoubledPawnPenaltyEg

	intParams[914] = engine.IsolatedPawnPenaltyMg
	intParams[915] = engine.IsolatedPawnPenaltyEg

	intParams[916] = engine.BackwardPawnPenaltyMg
	intParams[917] = engine.BackwardPawnPenaltyEg

	// Passed Pawns
	copy(intParams[918:926], engine.PassedPawnsBonusMg[:])
	copy(intParams[926:934], engine.PassedPawnsBonusEg[:])

	// Material adjustments
	intParams[934] = engine.BishopPairBonusMg
	intParams[935] = engine.BishopPairBonusEg
	intParams[936] = engine.RookOnOpenFileMg
	intParams[937] = engine.RookOnSemiOpenFileMg
	intParams[938] = engine.KnightOutpostBonusMg
	intParams[939] = engine.KnightOutpostBonusEg
	intParams[940] = engine.BishopOutpostBonusMg
	intParams[941] = engine.BishopOutpostBonusEg

	// King Safety - Open Files
	intParams[942] = engine.KingOnOpenFilePenaltyMg
	intParams[943] = engine.KingOnSemiOpenFilePenaltyMg
	intParams[944] = engine.KingNearOpenFilePenaltyMg
	intParams[945] = engine.KingNearSemiOpenFilePenaltyMg

	intParams[946] = engine.PawnShieldBonusMg[0]
	intParams[947] = engine.PawnShieldBonusMg[1]
	intParams[948] = engine.PawnShieldBonusMg[2]

	// King Safety Table
	copy(intParams[949:999], engine.KingSafetyTable[:])

	// Tempo
	intParams[999] = engine.TempoBonus

	for i := range 1000 {
		params[i] = float64(intParams[i])
	}
	return
}

// saveParams sotres the best params found in a file
func saveParams(bestParams [1000]float64, iteration int) {
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
func paramsToPrettyFormat(bestParams [1000]float64) (psqt string) {
	intParams := [1000]int{}
	for i := range 1000 {
		intParams[i] = int(bestParams[i])
	}

	piece := []string{"King", "Queen", "Rook", "Bishop", "Knight", "Pawn"}
	psqt = "MiddlegamePSQT: \n"
	for i := range 6 {
		psqt += fmt.Sprintf("// %s\n", piece[i])
		psqt += "{\n"
		for j := range 8 {
			for k := range 8 {
				psqt += fmt.Sprintf("%d, ", intParams[i*64+j*8+k])
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
				psqt += fmt.Sprintf("%d, ", intParams[(i+6)*64+j*8+k])
			}
			psqt += "\n"
		}
		psqt += "},\n"
	}

	// Pieces Values
	psqt += fmt.Sprintf("MiddlegamePieceValue:  %#v\n", intParams[768:774])
	psqt += fmt.Sprintf("EndgamePieceValue:  %#v\n", intParams[774:780])

	// Mobility Params
	psqt += fmt.Sprintf("QueenMobilityMg: %#v\n", intParams[780:808])
	psqt += fmt.Sprintf("QueenMobilityEg: %#v\n", intParams[808:836])
	psqt += fmt.Sprintf("RookMobilityMg: %#v\n", intParams[836:851])
	psqt += fmt.Sprintf("RookMobilityEg: %#v\n", intParams[851:866])
	psqt += fmt.Sprintf("BishopMobilityMg: %#v\n", intParams[866:880])
	psqt += fmt.Sprintf("BishopMobilityEg: %#v\n", intParams[880:894])
	psqt += fmt.Sprintf("KnightMobilityMg: %#v\n", intParams[894:903])
	psqt += fmt.Sprintf("KnightMobilityEg: %#v\n", intParams[903:912])

	// Pawn Structure Weights
	psqt += fmt.Sprintf("DoubledPawnPenaltyMg: %d\n", intParams[912])
	psqt += fmt.Sprintf("DoubledPawnPenaltyEg: %d\n", intParams[913])
	psqt += fmt.Sprintf("IsolatedPawnPenaltyMg: %d\n", intParams[914])
	psqt += fmt.Sprintf("IsolatedPawnPenaltyEg: %d\n", intParams[915])
	psqt += fmt.Sprintf("BackwardPawnPenaltyMg: %d\n", intParams[916])
	psqt += fmt.Sprintf("BackwardPawnPenaltyEg: %d\n", intParams[917])

	// Passed Pawns
	psqt += fmt.Sprintf("PassedPawnsPenaltyMg: %#v\n", intParams[918:926])
	psqt += fmt.Sprintf("PassedPawnsPenaltyEg: %#v\n", intParams[926:934])

	// Material adjustments
	psqt += fmt.Sprintf("BishopPairBonusMg: %d\n", intParams[934])
	psqt += fmt.Sprintf("BishopPairBonusEg: %d\n", intParams[935])
	psqt += fmt.Sprintf("RookOnOpenFileMg: %d\n", intParams[936])
	psqt += fmt.Sprintf("RookOnSemiOpenFileMg: %d\n", intParams[937])

	psqt += fmt.Sprintf("KnightOutpostBonusMg: %d\n", intParams[938])
	psqt += fmt.Sprintf("KnightOutpostBonusEg: %d\n", intParams[939])
	psqt += fmt.Sprintf("BishopOutpostBonusMg: %d\n", intParams[940])
	psqt += fmt.Sprintf("BishopOutpostBonusEg: %d\n", intParams[941])

	// King Safety Open Files
	psqt += fmt.Sprintf("KingOnOpenFilePenaltyMg: %d\n", intParams[942])
	psqt += fmt.Sprintf("KingOnSemiOpenFilePenaltyMg: %d\n", intParams[943])
	psqt += fmt.Sprintf("KingNearOpenFilePenaltyMg: %d\n", intParams[944])
	psqt += fmt.Sprintf("KingNearSemiOpenFilePenaltyMg: %d\n", intParams[945])

	psqt += fmt.Sprintf("PawnShieldBonusMg: %#v\n", intParams[946:949])

	// King Safety Table
	psqt += "KingSafetyTable: \n"
	for i := range 5 {
		psqt += "\t"
		for j := range 10 {
			psqt += fmt.Sprintf("%d, ", intParams[949+i*10+j])
		}
		psqt += "\n"
	}

	// Tempo
	psqt += fmt.Sprintf("TempoBonus: %d\n", intParams[999])

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
func MeanSquareError(scalingFactor float64, params *[1000]float64, dataset *[]DatasetEntry) float64 {
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
func worker(scalingFactor float64, params *[1000]float64, jobs <-chan WorkerJob, results chan<- WorkerResult, wg *sync.WaitGroup) {
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
func evaluatePosition(params [1000]float64, weights []PositionWeight) (evaluation float64) {
	eval := 0.0
	for _, attr := range weights {
		eval += params[attr.paramIndex] * float64(attr.weight)
	}
	intEval := int(math.Round(eval))
	evaluation = float64(intEval / 62)
	return
}

// generatePositionWeights returns all the position weights of a position
func generatePositionWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	generatePieceScoreWeights(pos, phase, weights)
	generateMobilityWeights(pos, phase, weights)
	generateDoubledPawnsWeights(pos, phase, weights)
	generateIsolatedPawnsWeights(pos, phase, weights)
	generateBackwardsPawnsWeights(pos, phase, weights)
	generatePassedPawnsWeights(pos, phase, weights)
	generateMaterialAdjustmentsWeights(pos, phase, weights)
	generateKingSafetyWeights(pos, phase, weights)

	// Tempo Bonus weight
	sideModifier := 1 - int(pos.Turn)*2
	*weights = append(*weights,
		PositionWeight{paramIndex: 999, weight: int16(sideModifier * phase)},
		PositionWeight{paramIndex: 999, weight: int16(sideModifier * (62 - phase))},
	)
}

// generatePieceScoreWeights returns the weights of the pieces socre in the board
func generatePieceScoreWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	for piece, bb := range pos.Bitboards {
		colorModifier := 1 - int(piece/6)*2
		for bb > 0 {
			sq := engine.Bsf(bb.NextBit())
			if colorModifier == 1 {
				sq = sq ^ 56 // white pieces uses mirror square index in psqt
			}

			*weights = append(*weights,
				PositionWeight{paramIndex: int16(768 + piece%6), weight: int16(colorModifier * phase)},
				PositionWeight{paramIndex: int16(768 + piece%6 + 6), weight: int16(colorModifier * (62 - phase))},
				PositionWeight{paramIndex: int16((piece%6)*64 + sq), weight: int16(colorModifier * phase)},
				PositionWeight{paramIndex: int16(384 + (piece%6)*64 + sq), weight: int16(colorModifier * (62 - phase))},
			)
		}
	}
}

// generateMobilityWeights returns the position weithts of the mobility of the position
func generateMobilityWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	var pieces = [8]int{engine.WhiteQueen, engine.WhiteRook, engine.WhiteBishop, engine.WhiteKnight, engine.BlackQueen, engine.BlackRook, engine.BlackBishop, engine.BlackKnight}

	// Start indexes for mobility arrays (Queen, rook, bishop, knight)
	mgIndexes := [4]int{780, 836, 866, 894}
	egIndexes := [4]int{808, 851, 880, 903}

	blocks := ^pos.EmptySquares()
	enemyPawnsAttacks := [2]engine.Bitboard{
		engine.Attacks(engine.BlackPawn, pos.Bitboards[engine.BlackPawn], blocks),
		engine.Attacks(engine.WhitePawn, pos.Bitboards[engine.WhitePawn], blocks),
	}
	whiteBB := pos.Bitboards[1:5]
	blackBB := pos.Bitboards[7:11]

	// NOTE: Need to make a new slice, otherwise we will mutate the Bitboards array in Position struct
	bitboards := make([]engine.Bitboard, 0, 8)
	bitboards = append(bitboards, whiteBB...)
	bitboards = append(bitboards, blackBB...)

	for piece, bb := range bitboards {
		colorModifier := 1 - int(piece/4)*2
		for bb > 0 {
			fromBB := bb.NextBit()
			attacks := engine.Attacks(pieces[piece], fromBB, blocks)
			safeSquares := bits.OnesCount64(uint64(attacks & ^enemyPawnsAttacks[piece/4]))

			*weights = append(*weights,
				PositionWeight{paramIndex: int16(mgIndexes[piece%4] + safeSquares), weight: int16(colorModifier * phase)},
				PositionWeight{paramIndex: int16(egIndexes[piece%4] + safeSquares), weight: int16(colorModifier * (62 - phase))},
			)
		}
	}
}

// generateDoubledPawnsWeights returns the position weights of the doubled pawns
func generateDoubledPawnsWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	wDoubled := bits.OnesCount64(uint64(engine.DoubledPawns(pos, engine.White)))
	bDoubled := bits.OnesCount64(uint64(engine.DoubledPawns(pos, engine.Black)))

	*weights = append(*weights,
		PositionWeight{paramIndex: 912, weight: int16(wDoubled * phase)},
		PositionWeight{paramIndex: 913, weight: int16(wDoubled * (62 - phase))},
		PositionWeight{paramIndex: 912, weight: int16(-bDoubled * phase)},
		PositionWeight{paramIndex: 913, weight: int16(-bDoubled * (62 - phase))},
	)
}

// generateIsolatedPawnsWeights returns the position weights of the isolated pawns
func generateIsolatedPawnsWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	wIsolated := bits.OnesCount64(uint64(engine.IsolatedPawns(pos, engine.White)))
	bIsolated := bits.OnesCount64(uint64(engine.IsolatedPawns(pos, engine.Black)))

	*weights = append(*weights,
		PositionWeight{paramIndex: 914, weight: int16(wIsolated * phase)},
		PositionWeight{paramIndex: 915, weight: int16(wIsolated * (62 - phase))},
		PositionWeight{paramIndex: 914, weight: int16(-bIsolated * phase)},
		PositionWeight{paramIndex: 915, weight: int16(-bIsolated * (62 - phase))},
	)
}

// generateBackwardsPawnsWeights returns the position weights of the backwards pawns
func generateBackwardsPawnsWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	whitePawnsAttacks := engine.Attacks(engine.WhitePawn, pos.Bitboards[engine.WhitePawn], pos.EmptySquares())
	blackPawnsAttacks := engine.Attacks(engine.BlackPawn, pos.Bitboards[engine.BlackPawn], pos.EmptySquares())

	wBackwards := bits.OnesCount64(uint64(engine.BackwardPawns(pos.Bitboards[engine.WhitePawn], blackPawnsAttacks, engine.White)))
	bBackwards := bits.OnesCount64(uint64(engine.BackwardPawns(pos.Bitboards[engine.BlackPawn], whitePawnsAttacks, engine.Black)))

	*weights = append(*weights,
		PositionWeight{paramIndex: 916, weight: int16(wBackwards * phase)},
		PositionWeight{paramIndex: 917, weight: int16(wBackwards * (62 - phase))},
		PositionWeight{paramIndex: 916, weight: int16(-bBackwards * phase)},
		PositionWeight{paramIndex: 917, weight: int16(-bBackwards * (62 - phase))},
	)
}

// generatePassedPawnsWeights returns the position weights of the passed pawns
func generatePassedPawnsWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	wPassedPawns := engine.PassedPawns(pos.Bitboards[engine.WhitePawn], pos.Bitboards[engine.BlackPawn], engine.White)
	bPassedPawns := engine.PassedPawns(pos.Bitboards[engine.BlackPawn], pos.Bitboards[engine.WhitePawn], engine.Black)

	for wPassedPawns > 0 {
		fromBB := wPassedPawns.NextBit()
		sq := engine.Bsf(fromBB)
		rank := sq / 8

		*weights = append(*weights,
			PositionWeight{paramIndex: int16(918 + rank), weight: int16(phase)},
			PositionWeight{paramIndex: int16(926 + rank), weight: int16(62 - phase)},
		)
	}

	for bPassedPawns > 0 {
		fromBB := bPassedPawns.NextBit()
		sq := engine.Bsf(fromBB)
		rank := 7 - sq/8

		*weights = append(*weights,
			PositionWeight{paramIndex: int16(918 + rank), weight: int16(-phase)},
			PositionWeight{paramIndex: int16(926 + rank), weight: int16(-(62 - phase))},
		)
	}
}

// generateMaterialAdjustmentsWeights returns the position weights of the material adjustments
func generateMaterialAdjustmentsWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	// Bishop pairs bonuses
	whiteBishopCount := bits.OnesCount64(uint64(pos.Bitboards[engine.WhiteBishop]))
	if whiteBishopCount >= 2 {
		*weights = append(*weights,
			PositionWeight{paramIndex: 934, weight: int16(phase)},
			PositionWeight{paramIndex: 935, weight: int16(62 - phase)},
		)
	}
	blackBishopCount := bits.OnesCount64(uint64(pos.Bitboards[engine.BlackBishop]))
	if blackBishopCount >= 2 {
		*weights = append(*weights,
			PositionWeight{paramIndex: 934, weight: int16(-phase)},
			PositionWeight{paramIndex: 935, weight: int16(-(62 - phase))},
		)
	}

	generateOutpostWeights(pos, phase, weights)

	// Rook on Open/SemiOpenFile
	generateRookOnOpenFileWeights(pos, phase, weights, engine.White)
	generateRookOnOpenFileWeights(pos, phase, weights, engine.Black)
}

// generateRookOnOpenFileWeights returns the position weights of the rook on open file
func generateRookOnOpenFileWeights(pos *engine.Position, phase int, weights *[]PositionWeight, side engine.Color) {
	sideModifier := 1 - int(side)*2
	alliedPawns := pos.Bitboards[engine.Pawn+int(side)*6]
	enemyPawns := pos.Bitboards[engine.Pawn+int(side.Opponent())*6]

	rooks := pos.Bitboards[engine.Rook+int(side)*6]

	for rooks > 0 {
		fromBB := rooks.NextBit()
		sq := engine.Bsf(fromBB)

		if (alliedPawns|enemyPawns)&engine.Files[sq%8] == 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: int16(936), weight: int16(sideModifier * phase)},
			)
		}
		if alliedPawns&engine.Files[sq%8] == 0 && enemyPawns&engine.Files[sq%8] > 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: int16(937), weight: int16(sideModifier * phase)},
			)
		}
	}
}

// generateOutpostWeights returns the position weights of the outpost
func generateOutpostWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	outpostSquares := [2]engine.Bitboard{
		engine.OutpostSquares(pos.Bitboards[engine.WhitePawn], pos.Bitboards[engine.BlackPawn], engine.White),
		engine.OutpostSquares(pos.Bitboards[engine.BlackPawn], pos.Bitboards[engine.WhitePawn], engine.Black),
	}

	// Knights Outposts
	wKnights := pos.Bitboards[engine.WhiteKnight]
	for wKnights > 0 {
		fromBB := wKnights.NextBit()

		if outpostSquares[engine.White]&fromBB > 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: 938, weight: int16(phase)},
				PositionWeight{paramIndex: 939, weight: int16(62 - phase)},
			)
		}
	}

	bKnights := pos.Bitboards[engine.BlackKnight]
	for bKnights > 0 {
		fromBB := bKnights.NextBit()

		if outpostSquares[engine.Black]&fromBB > 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: 938, weight: int16(-phase)},
				PositionWeight{paramIndex: 939, weight: int16(-(62 - phase))},
			)
		}
	}

	// Bishops Outpost
	wBishops := pos.Bitboards[engine.WhiteBishop]
	for wBishops > 0 {
		fromBB := wBishops.NextBit()

		if outpostSquares[engine.White]&fromBB > 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: 940, weight: int16(phase)},
				PositionWeight{paramIndex: 941, weight: int16(62 - phase)},
			)
		}
	}

	bBishops := pos.Bitboards[engine.BlackBishop]
	for bBishops > 0 {
		fromBB := bBishops.NextBit()

		if outpostSquares[engine.Black]&fromBB > 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: 940, weight: int16(-phase)},
				PositionWeight{paramIndex: 941, weight: int16(-(62 - phase))},
			)
		}
	}
}

// generateKingSafetyWeights returns the position weights of the king safety
func generateKingSafetyWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	generateKingOnOpenFileWeights(pos, phase, weights, engine.White)
	generateKingOnOpenFileWeights(pos, phase, weights, engine.Black)
	generatePawnShieldWeights(pos, phase, weights)
	generateSafetyWeights(pos, phase, weights)
}

// generateKingOnOpenFileWeights returns the position weights of the king on open file
func generateKingOnOpenFileWeights(pos *engine.Position, phase int, weights *[]PositionWeight, side engine.Color) {
	sideModifier := 1 - int(side)*2
	alliedPawns := pos.Bitboards[engine.Pawn+int(side)*6]
	enemyPawns := pos.Bitboards[engine.Pawn+int(side.Opponent())*6]
	kingFile := engine.Bsf(pos.KingPosition(side)) % 8

	for file := kingFile - 1; file <= kingFile+1; file++ {
		if file < 0 || file > 7 {
			continue
		}

		if (alliedPawns|enemyPawns)&engine.Files[file] == 0 {
			if kingFile == file {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(942), weight: int16(sideModifier * phase)},
				)
			} else {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(944), weight: int16(sideModifier * phase)},
				)
			}
		} else if alliedPawns&engine.Files[file] == 0 && enemyPawns&engine.Files[file] > 0 {
			if kingFile == file {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(943), weight: int16(sideModifier * phase)},
				)
			} else {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(945), weight: int16(sideModifier * phase)},
				)
			}
		}
	}
}

// generatePawnShieldWeights returns the position weights of the pawn shield
func generatePawnShieldWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	for color := engine.White; color <= engine.Black; color++ {
		king := pos.KingPosition(engine.Color(color))
		kingFile := engine.Bsf(king) % 8
		kingRank := engine.Bsf(king) / 8
		alliedPawns := pos.Bitboards[engine.Pawn+int(color)*6]

		for file := kingFile - 1; file <= kingFile+1; file++ {
			if file < 0 || file > 7 {
				continue
			}
			pawnsInFile := alliedPawns & engine.Files[file]
			rankIncrement := 1
			if color == engine.Black {
				rankIncrement = -1
			}
			for i := range 3 {
				rank := (kingRank + rankIncrement) + i*rankIncrement
				if rank < 0 || rank > 7 {
					break
				}
				if pawnsInFile&engine.Ranks[rank] > 0 {
					*weights = append(*weights,
						PositionWeight{paramIndex: int16(946 + i), weight: int16((1 - int(color)*2) * phase)},
					)
					break
				}
			}
		}
	}
}

// generateSafetyWeights returns the position weights of the king safety attacks weights
func generateSafetyWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	blocks := ^pos.EmptySquares()

	for color := engine.White; color <= engine.Black; color++ {
		c := engine.Color(color)
		enemyKing := pos.KingPosition(c.Opponent())
		kingZone := engine.KingZone[engine.Bsf(enemyKing)]

		attackersCount := 0
		attackWeight := 0

		pieces := []struct {
			pieceType int
			weight    int
		}{
			{engine.Knight, engine.KnightAttackWeight},
			{engine.Bishop, engine.BishopAttackWeight},
			{engine.Rook, engine.RookAttackWeight},
			{engine.Queen, engine.QueenAttackWeight},
		}

		colorModifier := 1
		if color == engine.Black {
			colorModifier = -1
		}

		for _, p := range pieces {
			piece := p.pieceType + int(color)*6
			bb := pos.Bitboards[piece]

			for bb > 0 {
				fromBB := bb.NextBit()
				attacks := engine.Attacks(piece, fromBB, blocks)

				if attacks&kingZone != 0 {
					attackersCount++
					attackCount := bits.OnesCount64(uint64(attacks & kingZone))
					attackWeight += attackCount * p.weight
				}
			}
		}

		// Safety is only applied if there are at least 2 attackers
		if attackersCount >= 2 {
			safetyIndex := min(attackWeight, 49)
			*weights = append(*weights,
				PositionWeight{
					paramIndex: int16(949 + safetyIndex),
					weight:     int16(colorModifier * phase),
				},
			)
		}
	}
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
func FindOptimalScalingFactor(dataset []DatasetEntry, params [1000]float64) float64 {
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
