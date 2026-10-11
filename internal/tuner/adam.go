package tuner

import (
	"bufio"
	"fmt"
	"math"
	"os"

	"github.com/gabtar/aconcagua/internal/engine"
)

// DefaultBatchSize is the default samples to load in a batch for a tuning session
const DefaultBatchSize = 128000

// AdamOptimizer implements the Adam optimization algorithm
type AdamOptimizer struct {
	m, v         []float64
	beta1, beta2 float64
	learningRate float64
	epsilon      float64
	t            int
}

// NewAdamOptimizer creates a new Adam optimizer
func NewAdamOptimizer(numParams int, lr float64) *AdamOptimizer {
	return &AdamOptimizer{
		m:            make([]float64, numParams),
		v:            make([]float64, numParams),
		beta1:        0.9,
		beta2:        0.999,
		learningRate: lr,
		epsilon:      1e-8,
		t:            0,
	}
}

// gradients is a global variable to calculate the gradients during the tuning
var gradients = [TuneableParams]float64{}

// Update updates the parameters
func (adam *AdamOptimizer) Update(params *[TuneableParams]float64, gradients *[]float64) {
	adam.t++

	for i := range params {
		adam.m[i] = adam.beta1*adam.m[i] + (1-adam.beta1)*(*gradients)[i]
		adam.v[i] = adam.beta2*adam.v[i] + (1-adam.beta2)*(*gradients)[i]*(*gradients)[i]

		mHat := adam.m[i] / (1 - math.Pow(adam.beta1, float64(adam.t)))
		vHat := adam.v[i] / (1 - math.Pow(adam.beta2, float64(adam.t)))

		params[i] -= adam.learningRate * mHat / (math.Sqrt(vHat) + adam.epsilon)
	}
}

// ComputeGradients computes the gradients of the loss with respect to the parameters
func ComputeGradients(entry *DatasetEntry, params *[TuneableParams]float64, K float64) [TuneableParams]float64 {
	// clear gradients
	for i := range len(gradients) {
		gradients[i] = 0.0
	}

	eval := evaluatePosition(params, &entry.Weights)
	predicted := 1.0 / (1.0 + math.Exp(-K*eval))
	actual := entry.Result

	error := predicted - actual
	lossGradient := 2 * error * K * predicted * (1 - predicted)

	for _, attr := range entry.Weights {
		if attr.paramIndex >= 0 && int(attr.paramIndex) < len(gradients) {
			evalGradient := float64(attr.weight) / float64(engine.MaxPhaseValue)
			gradients[attr.paramIndex] += lossGradient * evalGradient
		}
	}

	return gradients
}

func AdamTuner(filename string, entries int, K float64, epochs int) {
	params := getEvaluationParams()
	adam := NewAdamOptimizer(len(params), 0.1)
	// Use ceiling division to include the last partial batch
	totalBatches := (entries + DefaultBatchSize - 1) / DefaultBatchSize
	data := NewDataset(DefaultBatchSize)

	file, err, reset, close := openFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	scanner := bufio.NewScanner(file)

	fmt.Printf("Starting Adam optimization with %d parameters, %d positions, K=%.6f\n",
		len(params), entries, K)

	for epoch := 1; epoch <= epochs; epoch++ {
		reset()
		scanner = bufio.NewScanner(file)
		totalGradients := make([]float64, len(params))
		totalLoss := 0.0

		for batch := 0; batch < totalBatches; batch++ {
			// Calculate actual size for this batch (last batch may be smaller)
			batchSize := DefaultBatchSize
			if batch == totalBatches-1 {
				batchSize = entries - DefaultBatchSize*batch
				if batchSize <= 0 {
					batchSize = DefaultBatchSize
				}
			}

			data.Load(scanner, batchSize)

			for i := range batchSize {
				gradients := ComputeGradients(&data[i], &params, K)

				for j := range totalGradients {
					totalGradients[j] += gradients[j]
				}

				eval := evaluatePosition(&params, &data[i].Weights)
				predicted := 1.0 / (1.0 + math.Exp(-K*eval))
				error := predicted - data[i].Result
				totalLoss += error * error
			}
		}

		// Average gradients over all entries
		for i := range totalGradients {
			totalGradients[i] /= float64(entries)
		}

		adam.Update(&params, &totalGradients)

		mse := totalLoss / float64(entries)
		fmt.Printf("Epoch %3d: MSE = %.8f, LR = %.6f\n", epoch, mse, adam.learningRate)

		if epoch > 0 && epoch%50 == 0 {
			adam.learningRate *= 0.9
		}

		if epoch > 10 && mse < 0.001 || epoch == epochs {
			fmt.Printf("Converged at epoch %d\n", epoch)
			saveParams(params, epoch)
			close()
			break
		}
	}
}

// openFile opens a file and returns the file, an error, a reset funcion and a close function
func openFile(path string) (file *os.File, err error, reset func(), close func()) {
	file, err = os.Open(path)
	if err != nil {
		fmt.Println(err)
	}

	reset = func() {
		_, err := file.Seek(0, 0)
		if err != nil {
			fmt.Println(err)
		}
	}

	close = func() {
		file.Close()
	}

	return
}
