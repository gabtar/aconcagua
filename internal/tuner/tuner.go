package tuner

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/gabtar/aconcagua/internal/engine"
)

// ScalingFactor is the scaling factor for the training dataset
const ScalingFactor = 0.00820

// DatasetEntry is an struct conatining a single training example
type DatasetEntry struct {
	Fen     string
	Result  float64
	Weights []PositionWeight
	Phase   int
}

// Dataset is a set of DatasetEntry
type Dataset []DatasetEntry

// NewDataset returns a new preallocated dataset
func NewDataset(size int) (dataset Dataset) {
	dataset = make([]DatasetEntry, size)
	for i := range size {
		dataset[i] = DatasetEntry{
			Fen:     "",
			Result:  0.0,
			Weights: make([]PositionWeight, 0, 100),
			Phase:   0,
		}
	}
	return
}

// clear clears the dataset
func (dt *Dataset) clear() {
	for i := range len(*dt) {
		(*dt)[i] = DatasetEntry{
			Fen:     "",
			Result:  0.0,
			Weights: make([]PositionWeight, 0, 100),
			Phase:   0,
		}
	}
}

// Load loads a up to 'size' entries in the dataset
func (dt *Dataset) Load(scanner *bufio.Scanner, size int) {
	dt.clear()
	pos := engine.NewPosition()

	for i := 0; i < size && scanner.Scan(); i++ {
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

		generatePositionWeights(pos, phase, &(*dt)[i].Weights)
		(*dt)[i].Fen = fen
		(*dt)[i].Result = result
		(*dt)[i].Phase = phase
	}
	return
}

// Number of total tuneable params
const TuneableParams = 1130

// getEvaluationParams returns a flat array with the current evaluation params
func getEvaluationParams() (params [TuneableParams]float64) {
	intParams := [TuneableParams]int{}

	// Psqt params. Flat original Psqt array into a single array
	// The internal order of the psqt coefficients inside the flat array is:
	// KingMg(0-63), KingEg(64-127), QueenMg(128-192), ....
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

	// Pawn structure params. 928-953
	intParams[928], intParams[929] = engine.DoubledPawnPenalty.Get()
	intParams[930], intParams[931] = engine.IsolatedPawnPenalty.Get()
	intParams[932], intParams[933] = engine.BackwardPawnPenalty.Get()
	intParams[934], intParams[935] = engine.DefendedPawnBonus.Get()
	intParams[936], intParams[937] = engine.ConnectedPawnBonus.Get()
	for rank, score := range engine.PassedPawnsBonus {
		mgIndex := 938 + 2*rank
		intParams[mgIndex], intParams[mgIndex+1] = score.Get()
	}

	// Candidate passed pawns params. 954-985
	for flag, ranks := range engine.CandidatePassedPawnsBonus {
		for rank, score := range ranks {
			mgIndex := 954 + flag*16 + rank*2
			intParams[mgIndex], intParams[mgIndex+1] = score.Get()
		}
	}

	// King pawn shield params. 986-1017
	for sameFile, dists := range engine.PawnShield {
		for dist, score := range dists {
			mgIndex := 986 + sameFile*16 + dist*2
			intParams[mgIndex], intParams[mgIndex+1] = score.Get()
		}
	}

	// King pawn storm params. 1018-1081
	for sameFile, blockedRow := range engine.PawnStorm {
		for blocked, dists := range blockedRow {
			for dist, score := range dists {
				mgIndex := 1018 + sameFile*32 + blocked*16 + dist*2
				intParams[mgIndex], intParams[mgIndex+1] = score.Get()
			}
		}
	}

	// King on open files params. 1082-1085
	for sameFile, score := range engine.KingOnOpenFiles {
		mgIndex := 1082 + sameFile*2
		intParams[mgIndex], intParams[mgIndex+1] = score.Get()
	}

	// King attacks Weights. 1086-1093
	intParams[1086], intParams[1087] = engine.QueenAttackWeight.Get()
	intParams[1088], intParams[1089] = engine.RookAttackWeight.Get()
	intParams[1090], intParams[1091] = engine.BishopAttackWeight.Get()
	intParams[1092], intParams[1093] = engine.KnightAttackWeight.Get()

	// King Zone Defense. 1094-1095
	intParams[1094], intParams[1095] = engine.KingZoneDefenseBonus.Get()

	// Safe Checks + EnemyQueen. 1096-1105
	intParams[1096], intParams[1097] = engine.SafeQueenCheck.Get()
	intParams[1098], intParams[1099] = engine.SafeRookCheck.Get()
	intParams[1100], intParams[1101] = engine.SafeBishopCheck.Get()
	intParams[1102], intParams[1103] = engine.SafeKnightCheck.Get()
	intParams[1104], intParams[1105] = engine.EnemyQueen.Get()

	// Threats. 1106-1123
	intParams[1106], intParams[1107] = engine.MinorAttackedByPawnThreat.Get()
	intParams[1108], intParams[1109] = engine.MajorAttackedByPawnThreat.Get()
	intParams[1110], intParams[1111] = engine.RookAttackedByMinorThreat.Get()
	intParams[1112], intParams[1113] = engine.QueenAttackedByMinorThreat.Get()
	intParams[1114], intParams[1115] = engine.HangingPawnThreat.Get()

	intParams[1116], intParams[1117] = engine.PinnedPieceThreat[0].Get()
	intParams[1118], intParams[1119] = engine.PinnedPieceThreat[1].Get()
	intParams[1120], intParams[1121] = engine.PinnedPieceThreat[2].Get()
	intParams[1122], intParams[1123] = engine.PinnedPieceThreat[3].Get()

	// Passed Pawns. 1124-1129
	intParams[1124], intParams[1125] = engine.PassedKingDistance.Get()
	intParams[1126], intParams[1127] = engine.PassedEnemyKingDistance.Get()
	intParams[1128], intParams[1129] = engine.PassedPawnProtected.Get()

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
	psqt += "}\n"

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
	psqt += "\n// Mobility Arrays\n"
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
		psqt += "\n}\n"
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
	psqt += "\n// Material Adjustments\n"
	for _, a := range adjustments {
		psqt += fmt.Sprintf("%-24s= S(%d, %d)\n", a.name, intParams[a.startIdx], intParams[a.startIdx+1])
	}

	// Pawn Structure
	pawnStructures := []struct {
		name     string
		startIdx int
	}{
		{"DoubledPawnPenalty", 928},
		{"IsolatedPawnPenalty", 930},
		{"BackwardPawnPenalty", 932},
		{"DefendedPawnBonus", 934},
		{"ConnectedPawnBonus", 936},
	}
	psqt += "\n// Pawn Structure\n"
	for _, p := range pawnStructures {
		psqt += fmt.Sprintf("%-24s= S(%d, %d)\n", p.name, intParams[p.startIdx], intParams[p.startIdx+1])
	}
	psqt += "PassedPawnsBonus    = [8]Score{"
	for rank := range 8 {
		mgIndex := 938 + 2*rank
		psqt += fmt.Sprintf("S(%d, %d)", intParams[mgIndex], intParams[mgIndex+1])
		if rank < 7 {
			psqt += ", "
		}
	}
	psqt += "}\n"

	// Candidate Passed Pawns
	psqt += "CandidatePassedPawnsBonus = [2][8]Score{\n"
	for flag := range 2 {
		row := "  {"
		for rank := range 8 {
			mgIndex := 954 + flag*16 + rank*2
			row += fmt.Sprintf("S(%d, %d)", intParams[mgIndex], intParams[mgIndex+1])
			if rank < 7 {
				row += ", "
			}
		}
		row += "},"
		psqt += row + "\n"
	}
	psqt += "}\n"

	// King Safety
	psqt += "\n// King Safety\n"
	psqt += "PawnShield          = [2][8]Score{\n"
	for sameFile := range 2 {
		row := "  {"
		for dist := range 8 {
			mgIndex := 986 + sameFile*16 + dist*2
			row += fmt.Sprintf("S(%d, %d)", intParams[mgIndex], intParams[mgIndex+1])
			if dist < 7 {
				row += ", "
			}
		}
		row += "},"
		psqt += row + "\n"
	}
	psqt += "}\n"
	psqt += "PawnStorm           = [2][2][8]Score{\n"
	for sameFile := range 2 {
		psqt += "  {\n"
		for blocked := range 2 {
			row := "    {"
			for dist := range 8 {
				mgIndex := 1018 + sameFile*32 + blocked*16 + dist*2
				row += fmt.Sprintf("S(%d, %d)", intParams[mgIndex], intParams[mgIndex+1])
				if dist < 7 {
					row += ", "
				}
			}
			row += "},"
			psqt += row + "\n"
		}
		psqt += "  },\n"
	}
	psqt += "}\n"
	psqt += "KingOnOpenFiles     = [2]Score{"
	for sameFile := range 2 {
		mgIndex := 1082 + sameFile*2
		psqt += fmt.Sprintf("S(%d, %d)", intParams[mgIndex], intParams[mgIndex+1])
		if sameFile < 1 {
			psqt += ", "
		}
	}
	psqt += "}\n"

	// King Attacks
	psqt += "\n// King Attacks\n"
	kingAttacks := []struct {
		name     string
		startIdx int
	}{
		{"KnightAttackWeight", 1092},
		{"BishopAttackWeight", 1090},
		{"RookAttackWeight", 1088},
		{"QueenAttackWeight", 1086},
		{"KingZoneDefenseBonus", 1094},
		{"SafeQueenCheck", 1096},
		{"SafeRookCheck", 1098},
		{"SafeBishopCheck", 1100},
		{"SafeKnightCheck", 1102},
		{"EnemyQueen", 1104},
	}
	for _, ka := range kingAttacks {
		psqt += fmt.Sprintf("%-24s= S(%d, %d)\n", ka.name, intParams[ka.startIdx], intParams[ka.startIdx+1])
	}

	// Threats
	psqt += "\n// Threats\n"
	threats := []struct {
		name     string
		startIdx int
	}{
		{"MinorAttackedByPawnThreat", 1106},
		{"MajorAttackedByPawnThreat", 1108},
		{"RookAttackedByMinorThreat", 1110},
		{"QueenAttackedByMinorThreat", 1112},
		{"HangingPawnThreat", 1114},
	}
	for _, th := range threats {
		psqt += fmt.Sprintf("%-24s= S(%d, %d)\n", th.name, intParams[th.startIdx], intParams[th.startIdx+1])
	}

	psqt += "\nPinnedPieceThreat     = [4]Score{"
	for i := range 4 {
		mgIdx := 1116 + i*2
		psqt += fmt.Sprintf("S(%d, %d),", intParams[mgIdx], intParams[mgIdx+1])
	}
	psqt = psqt[:len(psqt)-1]
	psqt += "}"

	psqt += "\nPassed Pawns\n"
	psqt += fmt.Sprintf("PassedKingDistance = S(%d, %d)\n", intParams[1124], intParams[1125])
	psqt += fmt.Sprintf("PassedEnemyKingDistance = S(%d, %d)\n", intParams[1126], intParams[1127])
	psqt += fmt.Sprintf("PassedPawnProtected = S(%d, %d)\n", intParams[1128], intParams[1129])

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
	weight     float64
}

// evaluatePosition returns the static evaluation of a position based on the weights and current params
func evaluatePosition(params *[TuneableParams]float64, weights *[]PositionWeight) (evaluation float64) {
	eval := 0.0
	for i := range len(*weights) {
		idx := (*weights)[i].paramIndex
		weight := (*weights)[i].weight

		eval += (*params)[idx] * weight
	}
	evaluation = eval / float64(engine.MaxPhaseValue)
	return
}

// generatePositionWeights returns all the position weights of a position
func generatePositionWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	mgPhase := float64(min(phase, engine.MaxPhaseValue))
	egPhase := float64(engine.MaxPhaseValue - mgPhase)
	scaleFactor := float64(engine.ScaleFactor(pos))

	generatePieceScoreWeights(pos, mgPhase, egPhase, scaleFactor, weights)
	generateMobilityWeights(pos, mgPhase, egPhase, scaleFactor, weights)
	generateMaterialAdjustmentWeights(pos, mgPhase, egPhase, scaleFactor, weights)
	generatePawnStructureWeights(pos, mgPhase, egPhase, scaleFactor, weights)
	generateKingSafetyWeights(pos, mgPhase, egPhase, scaleFactor, weights)
	generateKingAttacksWeights(pos, mgPhase, egPhase, scaleFactor, weights)
	generateThreatsWeights(pos, mgPhase, egPhase, scaleFactor, weights)
	generatePassedPawnsWeights(pos, mgPhase, egPhase, scaleFactor, weights)
}

// generatePieceScoreWeights generates the weights of the pieces socre in the position
func generatePieceScoreWeights(pos *engine.Position, mgPhase, egPhase, sf float64, weights *[]PositionWeight) {
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
			sideModifier := float64(side.Modifier())

			*weights = append(*weights,
				// Piece Value
				PositionWeight{paramIndex: int16(768 + 2*role), weight: sideModifier * mgPhase},
				PositionWeight{paramIndex: int16(768 + 2*role + 1), weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				// Psqt
				PositionWeight{paramIndex: mgPsqtIndex, weight: sideModifier * mgPhase},
				PositionWeight{paramIndex: egPsqtIndex, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
			)
		}
	}
}

// generateMobilityWeights generates the mobility weights of the position
func generateMobilityWeights(pos *engine.Position, mgPhase, egPhase, sf float64, weights *[]PositionWeight) {
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
	pinnedPieces := pos.PinnedPieces(engine.White) | pos.PinnedPieces(engine.Black)

	for p, bb := range pieces {
		for bb > 0 {
			nextPiece := bb.NextBit()
			role := roles[p]
			side := engine.Color(engine.White)
			if nextPiece&pos.Sides[engine.Black] > 0 {
				side = engine.Black
			}
			// sideModifier must be derived AFTER the black correction above,
			// otherwise every black Q/R/B/N term is scored from white's side.
			sideModifier := float64(side.Modifier())
			safeSquares := (engine.Attacks(engine.PieceOf(role, side), nextPiece, pos.Sides[engine.All]) & ^attackedByPawns[side.Opponent()]).Count()
			mgIdx := startIndex[p] + 2*int16(safeSquares)
			egIdx := mgIdx + 1

			*weights = append(*weights,
				PositionWeight{paramIndex: mgIdx, weight: sideModifier * mgPhase},
				PositionWeight{paramIndex: egIdx, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
			)

			// PinnedPieceThreat
			if pinnedPieces&nextPiece > 0 {
				mgIdx := int16(1116 + 2*(role-1))
				*weights = append(*weights,
					PositionWeight{paramIndex: mgIdx, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: mgIdx + 1, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}
		}
	}
}

// generateMaterialAdjustmentWeights generates the PositionWeights for material adjustments params in the position
func generateMaterialAdjustmentWeights(pos *engine.Position, mgPhase, egPhase, sf float64, weights *[]PositionWeight) {
	// Bishop Pair Bonus
	for side := engine.Color(engine.White); side <= engine.Black; side++ {
		if pos.Pieces[engine.PieceOf(engine.Bishop, side)].Count() >= 2 {
			sideModifier := float64(side.Modifier())
			*weights = append(*weights,
				PositionWeight{paramIndex: 912, weight: sideModifier * mgPhase},
				PositionWeight{paramIndex: 913, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
			)
		}
	}

	// Rooks on open files / seventh rank / queens on seventh rank
	// Outposts / Connected Knights
	for side := engine.Color(engine.White); side <= engine.Black; side++ {
		sideModifier := float64(side.Modifier())
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
			// Open files
			if (pawns[side]|pawns[opponent])&engine.Files[file] == 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 914, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: 915, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			} else if pawns[side]&engine.Files[file] == 0 && pawns[opponent]&engine.Files[file] > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 916, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: 917, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}
			// Seventh rank
			kingRank := engine.RankRelativeToSide(engine.Bsf(enemyKing), side)
			rookRank := engine.RankRelativeToSide(from, side)
			if kingRank == 7 && rookRank == 6 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 918, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: 919, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}
		}

		// Queen on seventh
		queens := pos.Pieces[engine.PieceOf(engine.Queen, side)]
		for queens > 0 {
			from := engine.Bsf(queens.NextBit())

			kingRank := engine.RankRelativeToSide(engine.Bsf(enemyKing), side)
			queenRank := engine.RankRelativeToSide(from, side)
			if kingRank == 7 && queenRank == 6 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 920, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: 921, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}
		}

		// Knights Outposts / Connected knights
		knights := pos.Pieces[engine.PieceOf(engine.Knight, side)]
		for knights > 0 {
			nextKnight := knights.NextBit()
			if nextKnight&outposts[side] > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 922, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: 923, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}

			knightAttacks := engine.Attacks(engine.PieceOf(engine.Knight, side), nextKnight, pos.Sides[engine.All])
			if knightAttacks&pos.Pieces[engine.PieceOf(engine.Knight, side)] > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 924, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: 925, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}
		}

		// Bishop outposts
		bishops := pos.Pieces[engine.PieceOf(engine.Bishop, side)]
		for bishops > 0 {
			nextBishop := bishops.NextBit()
			if nextBishop&outposts[side] > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 926, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: 927, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}
		}
	}
}

// generatePawnStructureWeights generates the PositionWeights for pawn structure params in the position
func generatePawnStructureWeights(pos *engine.Position, mgPhase, egPhase, sf float64, weights *[]PositionWeight) {
	pawns := [2]engine.Bitboard{
		pos.Pieces[engine.WhitePawn],
		pos.Pieces[engine.BlackPawn],
	}
	attackedByPawns := [2]engine.Bitboard{
		engine.Attacks(engine.WhitePawn, pawns[engine.White], pos.Sides[engine.All]),
		engine.Attacks(engine.BlackPawn, pawns[engine.Black], pos.Sides[engine.All]),
	}
	backwards := [2]engine.Bitboard{
		engine.BackwardPawns(pawns[engine.White], attackedByPawns[engine.Black], engine.White),
		engine.BackwardPawns(pawns[engine.Black], attackedByPawns[engine.White], engine.Black),
	}
	passed := [2]engine.Bitboard{
		engine.PassedPawns(pawns[engine.White], pawns[engine.Black], engine.White),
		engine.PassedPawns(pawns[engine.Black], pawns[engine.White], engine.Black),
	}

	for side := engine.Color(engine.White); side <= engine.Black; side++ {
		sideModifier := float64(side.Modifier())
		sidePawns := pawns[side]
		for sidePawns > 0 {
			nextPawn := sidePawns.NextBit()
			from := engine.Bsf(nextPawn)
			file := from % 8

			// Doubled
			pawnsInFile := pawns[side] & engine.Files[file]
			if pawnsInFile.Count() > 1 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 928, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: 929, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}

			// Isolated
			if engine.IsolatedAdjacentFilesMask[file]&pawns[side] == 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: 930, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: 931, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}

			// Backward
			isBackward := backwards[side]&nextPawn > 0
			if isBackward {
				*weights = append(*weights,
					PositionWeight{paramIndex: 932, weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: 933, weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}

			// Passed
			if passed[side]&nextPawn > 0 {
				rank := engine.RankRelativeToSide(from, side)
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(938 + 2*rank), weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: int16(939 + 2*rank), weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			} else {
				candidateFlag, rank := engine.CandidatePassedPawn(nextPawn, pawns[side], pawns[side.Opponent()], side)
				if candidateFlag >= 0 {
					*weights = append(*weights,
						PositionWeight{paramIndex: int16(954 + candidateFlag*16 + rank*2), weight: sideModifier * mgPhase},
						PositionWeight{paramIndex: int16(954 + candidateFlag*16 + rank*2 + 1), weight: sideModifier * sf / engine.ScaleNormal * egPhase},
					)
				}
			}

			// Defended pawn. A pawn defended by allied pawns
			defenders := (engine.Attacks(engine.PieceOf(engine.Pawn, side.Opponent()), nextPawn, pos.Sides[engine.All]) & pawns[side]).Count()
			if defenders > 0 {
				defenders := float64(defenders)
				*weights = append(*weights,
					PositionWeight{paramIndex: 934, weight: sideModifier * mgPhase * defenders},
					PositionWeight{paramIndex: 935, weight: sideModifier * sf / engine.ScaleNormal * egPhase * defenders},
				)
			}

			// Connected pawn. Allied pawns on adjacent files, and not backward
			connected := (engine.IsolatedAdjacentFilesMask[file] & pawns[side]).Count()
			if !isBackward && connected > 0 {
				connected := float64(connected)
				*weights = append(*weights,
					PositionWeight{paramIndex: 936, weight: sideModifier * mgPhase * connected},
					PositionWeight{paramIndex: 937, weight: sideModifier * sf / engine.ScaleNormal * egPhase * connected},
				)
			}
		}
	}
}

// generateKingSafetyWeights generates the PositionWeights for the king safety
// pawn shield and pawn storm params in the position
func generateKingSafetyWeights(pos *engine.Position, mgPhase, egPhase, sf float64, weights *[]PositionWeight) {
	for side := engine.Color(engine.White); side <= engine.Black; side++ {
		sideModifier := float64(side.Modifier())
		opponent := side.Opponent()
		king := pos.Pieces[engine.PieceOf(engine.King, side)]
		if king == 0 {
			continue
		}
		from := engine.Bsf(king)
		kingFile, kingRank := from%8, from/8

		// Squares in front of the king on the king file and the two adjacent files
		frontMask := engine.KingFrontMask[side][from]

		for file := max(0, kingFile-1); file <= min(7, kingFile+1); file++ {
			shielders := pos.Pieces[engine.PieceOf(engine.Pawn, side)] & engine.Files[file] & frontMask
			stormers := pos.Pieces[engine.PieceOf(engine.Pawn, opponent)] & engine.Files[file] & frontMask

			sameFile := 0
			if file == kingFile {
				sameFile = 1
			}

			// Shield. The nearest allied pawn ahead of the king on each file
			shieldRank := 8
			if shielders > 0 {
				shieldRank = engine.NearestFromSide(shielders, side) / 8
				dist := abs(kingRank - shieldRank)
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(986 + sameFile*16 + dist*2), weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: int16(986 + sameFile*16 + dist*2 + 1), weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}

			// Storm. The most advanced enemy pawn ahead of the king on each file
			if stormers > 0 {
				stormRank := engine.NearestFromSide(stormers, opponent) / 8
				blocked := 0
				if shieldRank != 8 && abs(shieldRank-stormRank) == 1 {
					blocked = 1
				}
				dist := abs(kingRank - stormRank)
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(1018 + sameFile*32 + blocked*16 + dist*2), weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: int16(1018 + sameFile*32 + blocked*16 + dist*2 + 1), weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}

			// King on open/near open files
			if (shielders | stormers) == 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(1082 + sameFile*2), weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: int16(1082 + sameFile*2 + 1), weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}
		}
	}
}

// generateKingAttacksWeights generates the PositionWeights for King attacks
func generateKingAttacksWeights(pos *engine.Position, mgPhase, egPhase, sf float64, weights *[]PositionWeight) {
	blocks := pos.Sides[engine.All]

	attackedBy := [2][6]engine.Bitboard{}
	for piece, bb := range pos.Pieces {
		side := engine.SideOf(piece)
		role := engine.RoleOf(piece)
		for bb > 0 {
			nextPiece := bb.NextBit()
			attackedBy[side][role] |= engine.Attacks(piece, nextPiece, pos.Sides[engine.All])
		}
	}

	for side := engine.Color(engine.White); side <= engine.Black; side++ {
		sideModifier := float64(side.Modifier())
		opponent := side.Opponent()
		enemyKing := pos.KingPosition(opponent)
		enemyKingZone := engine.KingZone[opponent][engine.Bsf(enemyKing)]
		enemyPawns := pos.Pieces[engine.PieceOf(engine.Pawn, opponent)]
		enemyDefendedSquares := float64((engine.Attacks(engine.PieceOf(engine.Pawn, opponent), enemyPawns, blocks) & enemyKingZone).Count())
		tempWeights := []PositionWeight{}
		attackersCount := 0

		for piece := engine.Queen; piece <= engine.Knight; piece++ {
			pieceBB := pos.Pieces[engine.PieceOf(piece, side)]
			for pieceBB > 0 {
				fromBB := pieceBB.NextBit()
				attacks := engine.Attacks(piece, fromBB, blocks)
				if attacks&enemyKingZone > 0 {
					attackersCount++
					tempWeights = append(tempWeights,
						PositionWeight{paramIndex: int16(1086 + 2*(piece-1)), weight: sideModifier * mgPhase},
						PositionWeight{paramIndex: int16(1087 + 2*(piece-1)), weight: sideModifier * sf / engine.ScaleNormal * egPhase},
					)
				}
			}
		}

		// Apply safety only if condition is met
		if attackersCount > (1 - pos.Pieces[engine.PieceOf(engine.Queen, side)].Count()) {
			*weights = append(*weights, tempWeights...)

			// Zone Defense Weights
			*weights = append(*weights,
				PositionWeight{paramIndex: int16(1094), weight: -sideModifier * enemyDefendedSquares * mgPhase},
				PositionWeight{paramIndex: int16(1095), weight: -sideModifier * enemyDefendedSquares * sf / engine.ScaleNormal * egPhase},
			)

			king := pos.KingPosition(opponent)
			defendedByPawns := attackedBy[opponent][engine.Pawn]
			knightChecksThreats := engine.Attacks(engine.WhiteKnight, king, pos.Sides[engine.All])
			bishopChecksThreats := engine.Attacks(engine.WhiteBishop, king, pos.Sides[engine.All])
			rookChecksThreats := engine.Attacks(engine.WhiteRook, king, pos.Sides[engine.All])
			queenChecksThreats := engine.Attacks(engine.WhiteQueen, king, pos.Sides[engine.All])

			knightChecks := float64((knightChecksThreats & ^defendedByPawns & attackedBy[side][engine.Knight]).Count())
			bishopChecks := float64((bishopChecksThreats & ^defendedByPawns & attackedBy[side][engine.Bishop]).Count())
			rookChecks := float64((rookChecksThreats & ^defendedByPawns & attackedBy[side][engine.Rook]).Count())
			queenChecks := float64((queenChecksThreats & ^defendedByPawns & attackedBy[side][engine.Queen]).Count())
			enemyQueens := float64(pos.Pieces[engine.PieceOf(engine.Queen, side)].Count())

			if queenChecks > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(1096), weight: sideModifier * queenChecks * mgPhase},
					PositionWeight{paramIndex: int16(1097), weight: sideModifier * queenChecks * sf / engine.ScaleNormal * egPhase},
				)
			}

			if rookChecks > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(1098), weight: sideModifier * rookChecks * mgPhase},
					PositionWeight{paramIndex: int16(1099), weight: sideModifier * rookChecks * sf / engine.ScaleNormal * egPhase},
				)
			}

			if bishopChecks > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(1100), weight: sideModifier * bishopChecks * mgPhase},
					PositionWeight{paramIndex: int16(1101), weight: sideModifier * bishopChecks * sf / engine.ScaleNormal * egPhase},
				)
			}

			if knightChecks > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(1102), weight: sideModifier * knightChecks * mgPhase},
					PositionWeight{paramIndex: int16(1103), weight: sideModifier * knightChecks * sf / engine.ScaleNormal * egPhase},
				)
			}

			if enemyQueens > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(1104), weight: sideModifier * enemyQueens * mgPhase},
					PositionWeight{paramIndex: int16(1105), weight: sideModifier * enemyQueens * sf / engine.ScaleNormal * egPhase},
				)
			}
		}
	}
}

// generateThreatsWeights generates the weights for threats in the position
func generateThreatsWeights(pos *engine.Position, mgPhase, egPhase, sf float64, weights *[]PositionWeight) {
	attackedBy := [2][6]engine.Bitboard{}
	blocks := pos.Sides[engine.All]
	for piece, pieceBB := range pos.Pieces {
		side := engine.SideOf(piece)
		role := engine.RoleOf(piece)
		for pieceBB > 0 {
			nextPiece := pieceBB.NextBit()
			attackedBy[side][role] |= engine.Attacks(piece, nextPiece, blocks)
		}
	}

	for side := engine.Color(engine.White); side <= engine.Black; side++ {
		opponent := side.Opponent()
		sideModifier := float64(side.Modifier())

		attackedByMinors := attackedBy[opponent][engine.Knight] | attackedBy[opponent][engine.Bishop]
		attackedByMajors := attackedBy[opponent][engine.Rook] | attackedBy[opponent][engine.Queen]
		attackedByPawns := attackedBy[opponent][engine.Pawn]
		attackedByEnemy := attackedByMinors | attackedByMajors | attackedByPawns
		defendedByAllied := attackedBy[side][engine.Pawn] | attackedBy[side][engine.Knight] |
			attackedBy[side][engine.Bishop] | attackedBy[side][engine.Rook] |
			attackedBy[side][engine.Queen] | attackedBy[side][engine.King]

		minors := pos.Pieces[engine.PieceOf(engine.Knight, side)] | pos.Pieces[engine.PieceOf(engine.Bishop, side)]
		rooks := pos.Pieces[engine.PieceOf(engine.Rook, side)]
		queens := pos.Pieces[engine.PieceOf(engine.Queen, side)]

		miniorsAttackedByPawns := float64((minors & attackedByPawns).Count())
		if miniorsAttackedByPawns > 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: int16(1106), weight: sideModifier * miniorsAttackedByPawns * mgPhase},
				PositionWeight{paramIndex: int16(1107), weight: sideModifier * miniorsAttackedByPawns * sf / engine.ScaleNormal * egPhase},
			)
		}

		majorsAttackedByPawns := float64(((rooks | queens) & attackedByPawns).Count())
		if majorsAttackedByPawns > 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: int16(1108), weight: sideModifier * majorsAttackedByPawns * mgPhase},
				PositionWeight{paramIndex: int16(1109), weight: sideModifier * majorsAttackedByPawns * sf / engine.ScaleNormal * egPhase},
			)
		}

		rooksAttackedByMinors := float64((rooks & attackedByMinors).Count())
		if rooksAttackedByMinors > 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: int16(1110), weight: sideModifier * rooksAttackedByMinors * mgPhase},
				PositionWeight{paramIndex: int16(1111), weight: sideModifier * rooksAttackedByMinors * sf / engine.ScaleNormal * egPhase},
			)
		}

		queensAttackedByMinors := float64((queens & attackedByMinors).Count())
		if queensAttackedByMinors > 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: int16(1112), weight: sideModifier * queensAttackedByMinors * mgPhase},
				PositionWeight{paramIndex: int16(1113), weight: sideModifier * queensAttackedByMinors * sf / engine.ScaleNormal * egPhase},
			)
		}

		hangingPawns := float64((pos.Pieces[engine.PieceOf(engine.Pawn, side)] & attackedByEnemy & ^defendedByAllied).Count())
		if hangingPawns > 0 {
			*weights = append(*weights,
				PositionWeight{paramIndex: int16(1114), weight: sideModifier * hangingPawns * mgPhase},
				PositionWeight{paramIndex: int16(1115), weight: sideModifier * hangingPawns * sf / engine.ScaleNormal * egPhase},
			)
		}
	}
}

// generatePassedPawnsWeights returns the weights for the passed pawns
func generatePassedPawnsWeights(pos *engine.Position, mgPhase, egPhase, sf float64, weights *[]PositionWeight) {
	for side := engine.Color(engine.White); side <= engine.Black; side++ {
		opponent := side.Opponent()
		sideModifier := float64(side.Modifier())
		alliedPawns := pos.Pieces[engine.PieceOf(engine.Pawn, side)]
		enemyPawns := pos.Pieces[engine.PieceOf(engine.Pawn, opponent)]
		passed := engine.PassedPawns(alliedPawns, enemyPawns, side)
		protected := engine.Attacks(engine.PieceOf(engine.Pawn, opponent), alliedPawns, pos.Sides[engine.All])

		for passed > 0 {
			nextPassed := passed.NextBit()
			from := engine.Bsf(nextPassed)
			alliedKingSq := engine.Bsf(pos.KingPosition(side))
			enemyKingSq := engine.Bsf(pos.KingPosition(opponent))

			*weights = append(*weights,
				PositionWeight{paramIndex: int16(1124), weight: sideModifier * float64(engine.ManhattanDistance(from, alliedKingSq)) * mgPhase},
				PositionWeight{paramIndex: int16(1125), weight: sideModifier * float64(engine.ManhattanDistance(from, alliedKingSq)) * sf / engine.ScaleNormal * egPhase},
			)

			*weights = append(*weights,
				PositionWeight{paramIndex: int16(1126), weight: sideModifier * float64(engine.ManhattanDistance(from, enemyKingSq)) * mgPhase},
				PositionWeight{paramIndex: int16(1127), weight: sideModifier * float64(engine.ManhattanDistance(from, enemyKingSq)) * sf / engine.ScaleNormal * egPhase},
			)

			if protected&alliedPawns > 0 {
				*weights = append(*weights,
					PositionWeight{paramIndex: int16(1128), weight: sideModifier * mgPhase},
					PositionWeight{paramIndex: int16(1129), weight: sideModifier * sf / engine.ScaleNormal * egPhase},
				)
			}
		}
	}
}

// fillUp flood a bitboard toward rank 8, replicating engine.fillUp
func fillUp(b engine.Bitboard) engine.Bitboard {
	b |= b << 8
	b |= b << 16
	b |= b << 32
	return b
}

// fillDown flood a bitboard toward rank 1, replicating engine.fillDown
func fillDown(b engine.Bitboard) engine.Bitboard {
	b |= b >> 8
	b |= b >> 16
	b |= b >> 32
	return b
}

// bitboardFromIndex returns the bitboard of the square index, or 0 if out of bounds,
// replicating engine.bitboardFromIndex
func bitboardFromIndex(sq int) engine.Bitboard {
	if sq > 63 || sq < 0 {
		return 0
	}
	return engine.Bitboards[sq]
}

// abs returns the absolute value of number
func abs(number int) int {
	if number < 0 {
		return -number
	}
	return number
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
