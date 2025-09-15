package tunner

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/gabtar/aconcagua/aconcagua"
)

var PawnHashTable = aconcagua.NewPawnHashTable(8)

// DataEntry is an struct conatining a single training example
type DataEntry struct {
	Fen    string
	Result float64
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
		result := resultString[parts[1][0:len(parts[1])-3]]
		dataset = append(dataset, DataEntry{Fen: fen, Result: result})
	}

	return dataset
}

// setEvaluationParams sets the current evaluation params to use in the Evaluation function
func setEvaluationParams(params [768]int) {
	// TODO: Just pass the index and value modified
	// Convert the index according to the piece / square
	// Be aware of the square mirror in the pieces square tables

	for piece := range 6 {
		aconcagua.MiddlegamePSQT[piece] = [64]int(params[piece*64 : (piece+1)*64])
		aconcagua.EndgamePSQT[piece] = [64]int(params[(piece+6)*64 : (piece+7)*64])
	}
	aconcagua.GeneratePiecesScoreTables()
}

func SetParam(index int, value int) {
	aconcagua.SetPsqt(index, value)
}

// getEvaluationParams returns the current evaluation params
func getEvaluationParams() (params [768]int) {
	for piece := range 6 {
		copy(params[piece*64:(piece+1)*64], aconcagua.MiddlegamePSQT[piece][0:64])
		copy(params[(piece+6)*64:(piece+7)*64], aconcagua.EndgamePSQT[piece][0:64])
	}
	return params
}

// Tunner is finds a set of parameters that minimize the mean square error
func Tunner(scalingFactor float64, dataset []DataEntry, iteration int) {
	paramAdjustValue := 1 // increment/decrement params by this value
	bestParams := getEvaluationParams()
	bestErr := MeanSquareError(scalingFactor, dataset)
	improved := true

	for improved {
		improved = false

		paramsTunned := 0
		for i := range len(bestParams) {
			newParams := bestParams
			newParams[i] += paramAdjustValue
			setEvaluationParams(newParams)
			// SetParam(i, newParams[i])
			err := MeanSquareError(scalingFactor, dataset)

			if err < bestErr {
				paramsTunned++
				bestParams[i] = newParams[i]
				improved = true
				bestErr = err
				continue
			} else {
				newParams[i] -= 2 * paramAdjustValue
				setEvaluationParams(newParams)
				// SetParam(i, newParams[i])
				err = MeanSquareError(scalingFactor, dataset)
				if err < bestErr {
					paramsTunned++
					bestParams[i] = newParams[i]
					bestErr = err
					improved = true
					continue
				}
			}

			// SetParam(i, bestParams[i])
			setEvaluationParams(bestParams)
		}

		// store psqt
		savePSQT(bestParams, iteration)
		fmt.Println("Iteration #", iteration, " Mean Square Error: ", MeanSquareError(scalingFactor, dataset), "Params Tunned: ", paramsTunned)
		iteration++
	}
}

// savePSQT sotres the best params found in a file
func savePSQT(bestParams [768]int, iteration int) {
	dir := "PSQT"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	} else {
		fmt.Println("PSQT directory already exists")
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
	entry DataEntry
	index int
}

// WorkerResult represents the result of a calculation
type WorkerResult struct {
	error float64
	index int
}

// MeanSquareError returns the mean square error using parallel processing
func MeanSquareError(scalingFactor float64, dataset []DataEntry) float64 {
	const numWorkers = 4 // limit to 3 threads/jobs at a time

	jobs := make(chan WorkerJob, len(dataset))
	results := make(chan WorkerResult, len(dataset))
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Add(1)
		go worker(scalingFactor, jobs, results, &wg)
	}

	go func() {
		for i, entry := range dataset {
			jobs <- WorkerJob{entry: entry, index: i}
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

	return totalError / float64(len(dataset))
}

// worker processes jobs from the jobs channel
func worker(scalingFactor float64, jobs <-chan WorkerJob, results chan<- WorkerResult, wg *sync.WaitGroup) {
	defer wg.Done()

	pos := aconcagua.InitialPosition()

	for job := range jobs {
		pos = aconcagua.NewPositionFromFen(job.entry.Fen)
		score := pos.Evaluate(PawnHashTable)
		sigmoid := 1 / (1 + math.Pow(10, -scalingFactor*float64(score)/400))
		errorValue := math.Pow(job.entry.Result-sigmoid, 2)

		results <- WorkerResult{error: errorValue, index: job.index}
	}
}

// TODO: Replace evaluation function w/ products of positionAttributes
// Each position will have a different set of attributes representing mobility, piece value, psqt values, and some position features
// So it will be cached for all iterations and will calculate the static evaluation much faster, to get the mean square error
// My eval is Eval = psqt + piece value + mobility + pawn strucutre
// so each position can be = (h1_psqt(pawn)mg + pawn value mg) + queenMobSc * freeSquares + ...... and so on
// calcualte for both game phases

// PositionAttribute is a struct that represents a single attribute of a position
// The paramIndex corresponds with the index of the params array we are trying to optimize
// If the param index is -1, then the bias is a fixed value/attribute of the position
// The bias contains the a fixed value/attribute of the position
// For example if we are not tuning mobility, the mobility attributes only have a bias with a fixed value for each position
type PositionAttribute struct {
	paramIndex int
	bias       int
}

// evaluatePosition returns the static evaluation of a position based on the attributes and current params
func evaluatePosition(params [768]int, attributes []PositionAttribute, phase int) (value int) {
	for attr := range len(attributes) {
		if attributes[attr].paramIndex >= 0 {
			value += params[attributes[attr].paramIndex] + attributes[attr].bias
		} else {
			value += attributes[attr].bias
		}
	}
	// TODO: side to move relative...
	return
}

// func generatePSQTAttributes(fen string) []PositionAttribute {
//
// }

//
// func generateMobilityAttributes(fen string) []PositionAttribute {
// }
//
// func generatePawnStructureAttributes(fen string) []PositionAttribute {
// }
