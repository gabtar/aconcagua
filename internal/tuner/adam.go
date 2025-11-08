package tuner

import (
	"fmt"
	"math"
)

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

// Update updates the parameters
func (adam *AdamOptimizer) Update(params *[810]float64, gradients []float64) {
	adam.t++

	for i := range params {
		adam.m[i] = adam.beta1*adam.m[i] + (1-adam.beta1)*gradients[i]
		adam.v[i] = adam.beta2*adam.v[i] + (1-adam.beta2)*gradients[i]*gradients[i]

		mHat := adam.m[i] / (1 - math.Pow(adam.beta1, float64(adam.t)))
		vHat := adam.v[i] / (1 - math.Pow(adam.beta2, float64(adam.t)))

		params[i] -= adam.learningRate * mHat / (math.Sqrt(vHat) + adam.epsilon)
	}
}

// ComputeGradients computes the gradients of the loss with respect to the parameters
func ComputeGradients(entry DatasetEntry, params [810]float64, K float64) []float64 {
	gradients := make([]float64, len(params))

	eval := evaluatePosition(params, entry.Weights)
	predicted := 1.0 / (1.0 + math.Exp(-K*eval))
	actual := entry.Result

	error := predicted - actual
	lossGradient := 2 * error * K * predicted * (1 - predicted)

	for _, attr := range entry.Weights {
		if attr.paramIndex >= 0 && attr.paramIndex < len(gradients) {
			evalGradient := float64(attr.weight) / 62.0
			gradients[attr.paramIndex] += lossGradient * evalGradient
		}
	}

	return gradients
}

func AdamTuner(params [810]float64, dataset []DatasetEntry, K float64, epochs int) {
	adam := NewAdamOptimizer(len(params), 0.1)

	fmt.Printf("Starting Adam optimization with %d parameters, %d positions, K=%.6f\n",
		len(params), len(dataset), K)

	for epoch := 1; epoch <= epochs; epoch++ {
		totalGradients := make([]float64, len(params))
		totalLoss := 0.0

		for _, entry := range dataset {
			gradients := ComputeGradients(entry, params, K)

			for i := range totalGradients {
				totalGradients[i] += gradients[i]
			}

			eval := evaluatePosition(params, entry.Weights)
			predicted := 1.0 / (1.0 + math.Exp(-K*eval))
			error := predicted - entry.Result
			totalLoss += error * error
		}

		for i := range totalGradients {
			totalGradients[i] /= float64(len(dataset))
		}

		adam.Update(&params, totalGradients)

		mse := totalLoss / float64(len(dataset))
		fmt.Printf("Epoch %3d: MSE = %.8f, LR = %.6f\n", epoch, mse, adam.learningRate)

		if epoch > 0 && epoch%50 == 0 {
			adam.learningRate *= 0.9
		}

		if epoch > 10 && mse < 0.001 || epoch == epochs {
			fmt.Printf("Converged at epoch %d\n", epoch)
			saveParams(params, epoch)
			break
		}
	}
}
