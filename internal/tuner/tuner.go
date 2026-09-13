package tuner

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gabtar/aconcagua/internal/engine"
)

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
		parts := strings.Split(line, "[") // for lichess-big3-resolved dataset
		resultString := map[string]float64{
			"1.0]": 1.0,
			"0.0]": 0.0,
			"0.5]": 0.5,
		}

		fen := parts[0]
		pos.LoadFromFenString(fen)
		phase := engine.GetEvalPhase(pos)
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

// Number of total tuneable params
const TuneableParams = 928

// GetEvaluationParams returns a flat array with the current evaluation params
func GetEvaluationParams() (params [TuneableParams]float64) {
	intParams := [TuneableParams]int{}

	// Psqt params. Flat original Psqt array into a single array
	// The internal order of the psqt coefficients inside the flat array is:
	// KingMg(0-63), KingEg(64-127), QueenMg(127-191), ....
	for p, psqt := range engine.Psqt {
		for sq, score := range psqt {
			mgIndex := 64*2*p + sq
			egIndex := 64*(2*p+1) + sq
			intParams[mgIndex], intParams[egIndex] = score.Get()
		}
	}

	// Pieces Values. Indexes goes from 768-7779
	for p, score := range engine.PiecesValues {
		mgIndex := 768 + p*2
		egIndex := mgIndex + 1
		intParams[mgIndex], intParams[egIndex] = score.Get()
	}

	// Mobility. Index range 780-911
	// Queen. 780-835
	for i, sc := range engine.QueenMobility {
		mobIndex := 780 + 2*i
		intParams[mobIndex], intParams[mobIndex+1] = sc.Get()
	}

	// Rook. 836-865
	for i, sc := range engine.RookMobility {
		mobIndex := 836 + 2*i
		intParams[mobIndex], intParams[mobIndex+1] = sc.Get()
	}

	// Bishop. 866-893
	for i, sc := range engine.BishopMobility {
		mobIndex := 866 + 2*i
		intParams[mobIndex], intParams[mobIndex+1] = sc.Get()
	}

	// Knight. 894-911
	for i, sc := range engine.KnightMobility {
		mobIndex := 894 + 2*i
		intParams[mobIndex], intParams[mobIndex+1] = sc.Get()
	}

	// Material adjustment params. 912-927
	intParams[912], intParams[913] = engine.BishopPairBonus.Get()
	intParams[914], intParams[915] = engine.RookOnOpenFileBonus.Get()
	intParams[916], intParams[917] = engine.RookOnSemiOpenFileBonus.Get()
	intParams[918], intParams[919] = engine.RookOnSeventhRankBonus.Get()
	intParams[920], intParams[921] = engine.QueenOnSeventhRankBonus.Get()
	intParams[922], intParams[923] = engine.KnightOutpostBonus.Get()
	intParams[924], intParams[925] = engine.ConnectedKnightBonus.Get()
	intParams[926], intParams[927] = engine.BishopOutpostBonus.Get()

	// Convert to float params
	for i := range TuneableParams {
		params[i] = float64(intParams[i])
	}

	return
}

// saveParams sotres the best params found in a file
func saveParams(bestParams [TuneableParams]float64, iteration int) {
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
func paramsToPrettyFormat(bestParams [TuneableParams]float64) (psqt string) {
	intParams := [TuneableParams]int{}
	for i := range TuneableParams {
		intParams[i] = int(bestParams[i])
	}

	// Psqt
	piece := []string{"King", "Queen", "Rook", "Bishop", "Knight", "Pawn"}
	psqt = "// Pieces Square Tables\n"
	psqt += "var Psqt = [6][64]Score{\n"
	for p := range 6 {
		psqt += fmt.Sprintf("  // %s\n  {\n", piece[p])
		for rank := range 8 {
			psqt += "    "
			for file := range 8 {
				sq := 8*rank + file
				mgIndex := 64*2*p + sq
				egIndex := 64*(2*p+1) + sq
				psqt += fmt.Sprintf("S(%d, %d), ", intParams[mgIndex], intParams[egIndex])
			}
			psqt = psqt[:len(psqt)-1] // remove last space " "
			psqt += "\n"
		}
		psqt += "  },\n"
	}
	psqt += "}\n\n"

	// Pieces Values
	psqt += "// Pieces Values\n"
	psqt += "var PiecesValues = [6]Score{"
	for p := range 6 {
		mgIndex := 768 + p*2
		egIndex := mgIndex + 1
		psqt += fmt.Sprintf("S(%d, %d), ", intParams[mgIndex], intParams[egIndex])
	}
	psqt = psqt[:len(psqt)-2] // remove last ", "
	psqt += "}\n\n"

	// Mobility
	mobility := []struct {
		name     string
		size     int
		startIdx int
	}{
		{"QueenMobility", 28, 780},
		{"RookMobility", 15, 836},
		{"BishopMobility", 14, 866},
		{"KnightMobility", 9, 894},
	}
	psqt += "// Mobility Arrays\n"
	for _, m := range mobility {
		psqt += fmt.Sprintf("%s = [%d]Score{\n  ", m.name, m.size)
		for i := range m.size {
			mgIndex := m.startIdx + 2*i
			egIndex := mgIndex + 1
			psqt += fmt.Sprintf("S(%d, %d), ", intParams[mgIndex], intParams[egIndex])
			if (i+1)%6 == 0 && i != m.size-1 {
				psqt += "\n  "
			}
		}
		psqt = psqt[:len(psqt)-1] // remove last " "
		psqt += "\n}\n\n"
	}

	// Material Adjustments
	adjustments := []struct {
		name     string
		startIdx int
	}{
		{"BishopPairBonus", 912},
		{"RookOnOpenFileBonus", 914},
		{"RookOnSemiOpenFileBonus", 916},
		{"RookOnSeventhRankBonus", 918},
		{"QueenOnSeventhRankBonus", 920},
		{"KnightOutpostBonus", 922},
		{"ConnectedKnightBonus", 924},
		{"BishopOutpostBonus", 926},
	}
	psqt += "// Material Adjustments\n"
	for _, a := range adjustments {
		psqt += fmt.Sprintf("%-24s= S(%d, %d)\n", a.name, intParams[a.startIdx], intParams[a.startIdx+1])
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
func MeanSquareError(scalingFactor float64, params *[TuneableParams]float64, dataset *[]DatasetEntry) float64 {
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
func worker(scalingFactor float64, params *[TuneableParams]float64, jobs <-chan WorkerJob, results chan<- WorkerResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		score := evaluatePosition(params, &job.entry.Weights)

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
func evaluatePosition(params *[TuneableParams]float64, weights *[]PositionWeight) (evaluation float64) {
	eval := 0.0
	for i := range len(*weights) {
		idx := (*weights)[i].paramIndex
		weight := (*weights)[i].weight

		eval += (*params)[idx] * float64(weight)
	}
	evaluation = eval / 24
	return
}

// generatePositionWeights returns all the position weights of a position
func generatePositionWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	generatePieceScoreWeights(pos, phase, weights)
	generateMobilityWeights(pos, phase, weights)
	generateMaterialAdjustmentWeights(pos, phase, weights)
}

// generatePieceScoreWeights generates the weights of the pieces socre in the position
func generatePieceScoreWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	mgPhase := min(phase, engine.MaxPhaseValue)
	egPhase := engine.MaxPhaseValue - phase

	for piece, bb := range pos.Pieces {
		side := engine.SideOf(piece)
		for bb > 0 {
			sq := engine.Bsf(bb.NextBit())
			if side == engine.White {
				sq = sq ^ 56 // white pieces uses mirror square index in psqt
			}
			role := engine.RoleOf(piece)
			mgPsqtIndex := int16(64*2*role + sq)
			egPsqtIndex := int16(64*(2*role+1) + sq)

			*weights = append(*weights,
				// Piece Value
				PositionWeight{paramIndex: int16(768 + 2*role), weight: int16(side.Modifier() * mgPhase)},
				PositionWeight{paramIndex: int16(768 + 2*role + 1), weight: int16(side.Modifier() * egPhase)},
				// Psqt
				PositionWeight{paramIndex: mgPsqtIndex, weight: int16(side.Modifier() * mgPhase)},
				PositionWeight{paramIndex: egPsqtIndex, weight: int16(side.Modifier() * egPhase)},
			)
		}
	}
}

// generateMobilityWeights generates the mobility weights of the position
func generateMobilityWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	mgPhase := min(phase, engine.MaxPhaseValue)
	egPhase := engine.MaxPhaseValue - phase
	startIndex := [4]int16{780, 836, 866, 894}
	pieces := [4]engine.Bitboard{
		pos.Pieces[engine.WhiteQueen] | pos.Pieces[engine.BlackQueen],
		pos.Pieces[engine.WhiteRook] | pos.Pieces[engine.BlackRook],
		pos.Pieces[engine.WhiteBishop] | pos.Pieces[engine.BlackBishop],
		pos.Pieces[engine.WhiteKnight] | pos.Pieces[engine.BlackKnight],
	}
	roles := [4]int{engine.Queen, engine.Rook, engine.Bishop, engine.Knight}
	attackedByPawns := [2]engine.Bitboard{
		engine.Attacks(engine.WhitePawn, pos.Pieces[engine.WhitePawn], pos.Sides[engine.All]),
		engine.Attacks(engine.BlackPawn, pos.Pieces[engine.BlackPawn], pos.Sides[engine.All]),
	}

	for p, bb := range pieces {
		for bb > 0 {
			nextPiece := bb.NextBit()
			side := engine.Color(engine.White)
			if nextPiece&pos.Sides[engine.Black] > 0 {
				side = engine.Black
			}
			safeSquares := (engine.Attacks(engine.PieceOf(roles[p], side), nextPiece, pos.Sides[engine.All]) & ^attackedByPawns[side.Opponent()]).Count()
			mgIdx := startIndex[p] + 2*int16(safeSquares)
			egIdx := mgIdx + 1

			*weights = append(*weights,
				PositionWeight{paramIndex: mgIdx, weight: int16(side.Modifier() * mgPhase)},
				PositionWeight{paramIndex: egIdx, weight: int16(side.Modifier() * egPhase)},
			)
		}
	}
}

// generateMaterialAdjustmentWeights generates the PositionWeights for material adjustments params in the position
func generateMaterialAdjustmentWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	mgPhase := min(phase, engine.MaxPhaseValue)
	egPhase := engine.MaxPhaseValue - phase

	// Bishop Pair Bonus
	for side := engine.Color(engine.White); side <= engine.Black; side++ {
		if pos.Pieces[engine.PieceOf(engine.Bishop, side)].Count() >= 2 {
			*weights = append(*weights,
				PositionWeight{paramIndex: 912, weight: int16(side.Modifier() * mgPhase)},
				PositionWeight{paramIndex: 913, weight: int16(side.Modifier() * egPhase)},
			)
		}
	}

	// Rooks on open files / seventh rank / queens on seventh rank
	// Outposts / Connected Knights
	for side := engine.Color(engine.White); side <= engine.Black; side++ {
		opponent := side.Opponent()
		pawns := [2]engine.Bitboard{
			pos.Pieces[engine.WhitePawn],
			pos.Pieces[engine.BlackPawn],
		}
		outposts := [2]engine.Bitboard{
			engine.OutpostSquares(pawns[engine.White], pawns[engine.Black], engine.White),
			engine.OutpostSquares(pawns[engine.Black], pawns[engine.White], engine.Black),
		}

		rooks := pos.Pieces[engine.PieceOf(engine.Rook, side)]
		enemyKing := pos.Pieces[engine.PieceOf(engine.King, opponent)]
		for rooks > 0 {
			nextRook := rooks.NextBit()
			from := engine.Bsf(nextRook)
			file := from % 8
			rank := from / 8
			if side == engine.White {
				rank = 7 - rank
			}
			// Open files
			if (pawns[side]|pawns[opponent])&engine.Files[file] == 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 914, weight: int16(side.Modifier() * mgPhase)},
					PositionWeight{paramIndex: 915, weight: int16(side.Modifier() * egPhase)},
				)
			} else if pawns[side]&engine.Files[file] == 0 && pawns[opponent]&engine.Files[file] > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 916, weight: int16(side.Modifier() * mgPhase)},
					PositionWeight{paramIndex: 917, weight: int16(side.Modifier() * egPhase)},
				)
			}
			// Seventh rank
			relativeKingRank := engine.Bsf(enemyKing) / 8
			if side == engine.White {
				relativeKingRank = 7 - relativeKingRank
			}
			if relativeKingRank == 7 && rank == 6 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 918, weight: int16(side.Modifier() * mgPhase)},
					PositionWeight{paramIndex: 919, weight: int16(side.Modifier() * egPhase)},
				)
			}
		}

		// Queen on seventh
		queens := pos.Pieces[engine.PieceOf(engine.Queen, side)]
		for queens > 0 {
			nextQueen := queens.NextBit()
			rank := engine.Bsf(nextQueen) / 8
			if side == engine.White {
				rank = 7 - rank
			}
			relativeKingRank := engine.Bsf(enemyKing) / 8
			if side == engine.White {
				relativeKingRank = 7 - relativeKingRank
			}

			if relativeKingRank == 7 && rank == 6 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 920, weight: int16(side.Modifier() * mgPhase)},
					PositionWeight{paramIndex: 921, weight: int16(side.Modifier() * egPhase)},
				)
			}
		}

		// Knights Outposts / Connected knights
		knights := pos.Pieces[engine.PieceOf(engine.Knight, side)]
		for knights > 0 {
			nextKnight := knights.NextBit()
			if nextKnight&outposts[side] > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 922, weight: int16(side.Modifier() * mgPhase)},
					PositionWeight{paramIndex: 923, weight: int16(side.Modifier() * egPhase)},
				)
			}

			knightAttacks := engine.Attacks(engine.PieceOf(engine.Knight, side), nextKnight, pos.Sides[engine.All])
			if knightAttacks&pos.Pieces[engine.PieceOf(engine.Knight, side)] > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 924, weight: int16(side.Modifier() * mgPhase)},
					PositionWeight{paramIndex: 925, weight: int16(side.Modifier() * egPhase)},
				)
			}
		}

		// Bishop outposts
		bishops := pos.Pieces[engine.PieceOf(engine.Bishop, side)]
		for bishops > 0 {
			nextBishop := bishops.NextBit()
			if nextBishop&outposts[side] > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 926, weight: int16(side.Modifier() * mgPhase)},
					PositionWeight{paramIndex: 927, weight: int16(side.Modifier() * egPhase)},
				)
			}
		}
	}
}

// FindOptimalScalingFactor returns the scaling factor that minimizes the mean square error
func FindOptimalScalingFactor(dataset []DatasetEntry, params [TuneableParams]float64) float64 {
	bestK := 0.0
	bestError := math.Inf(1)

	for k := 0.0001; k <= 0.1; k += 0.0001 {
		totalError := 0.0

		for _, entry := range dataset {
			eval := evaluatePosition(&params, &entry.Weights)
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
