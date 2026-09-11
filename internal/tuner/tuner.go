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
const TuneableParams = 987

// GetEvaluationParams returns the current evaluation params
func GetEvaluationParams() (params [TuneableParams]float64) {
	intParams := [TuneableParams]int{}

	// Psqt params
	for piece := range 6 {
		copy(intParams[piece*64:(piece+1)*64], engine.MiddlegamePSQT[piece][0:64])
		copy(intParams[(piece+6)*64:(piece+7)*64], engine.EndgamePSQT[piece][0:64])
	}

	// Piece values params
	copy(intParams[768:774], engine.MiddlegamePieceValue[:])
	copy(intParams[774:780], engine.EndgamePieceValue[:])

	// Convert to float
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
}

// generatePieceScoreWeights returns the weights of the pieces socre in the board
func generatePieceScoreWeights(pos *engine.Position, phase int, weights *[]PositionWeight) {
	mgPhase := min(phase, engine.MaxPhaseValue)
	egPhase := engine.MaxPhaseValue - phase

	for piece, bb := range pos.Pieces {
		side := engine.Color(piece / 6)
		for bb > 0 {
			sq := engine.Bsf(bb.NextBit())
			if side.Modifier() == 1 {
				sq = sq ^ 56 // white pieces uses mirror square index in psqt
			}

			*weights = append(*weights,
				PositionWeight{paramIndex: int16(768 + piece%6), weight: int16(side.Modifier() * mgPhase)},
				PositionWeight{paramIndex: int16(768 + piece%6 + 6), weight: int16(side.Modifier() * egPhase)},
				PositionWeight{paramIndex: int16((piece%6)*64 + sq), weight: int16(side.Modifier() * mgPhase)},
				PositionWeight{paramIndex: int16(384 + (piece%6)*64 + sq), weight: int16(side.Modifier() * egPhase)},
			)
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
