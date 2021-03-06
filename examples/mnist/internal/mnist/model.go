// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mnist

import (
	"github.com/nlpodyssey/spago/pkg/mat/rand"
	"github.com/nlpodyssey/spago/pkg/ml/ag"
	"github.com/nlpodyssey/spago/pkg/ml/initializers"
	"github.com/nlpodyssey/spago/pkg/ml/nn"
	"github.com/nlpodyssey/spago/pkg/ml/nn/cnn"
	"github.com/nlpodyssey/spago/pkg/ml/nn/convolution"
	"github.com/nlpodyssey/spago/pkg/ml/nn/perceptron"
	"github.com/nlpodyssey/spago/pkg/ml/nn/stack"
)

// NewMLP returns a new multi-layer perceptron initialized to zeros.
func NewMLP(in, hidden, out int, hiddenAct, outputAct ag.OpName) *stack.Model {
	return stack.New(
		perceptron.New(in, hidden, hiddenAct),
		perceptron.New(hidden, out, outputAct),
	)
}

// InitRandom initializes the model using the Xavier (Glorot) method.
func InitMLP(model *stack.Model, rndGen *rand.LockedRand) {
	for i, layer := range model.Layers {
		var gain float64
		if i == len(model.Layers)-1 { // last layer
			gain = initializers.Gain(ag.OpSoftmax)
		} else {
			gain = initializers.Gain(layer.(*perceptron.Model).Activation)
		}
		layer.ForEachParam(func(param *nn.Param) {
			if param.Type() == nn.Weights {
				initializers.XavierUniform(param.Value(), gain, rndGen)
			}
		})
	}
}

// NewCNN returns a new CNN initialized to zeros.
func NewCNN(
	kernelSizeX, kernelSizeY int,
	inputChannels, outputChannels int,
	maxPoolingRows, maxPoolingCols int,
	hidden, out int,
	hiddenAct, outputActivation ag.OpName,
) *cnn.Model {
	return cnn.NewModel(
		convolution.New(kernelSizeX, kernelSizeY, 1, 1, inputChannels, outputChannels, hiddenAct),
		maxPoolingRows, maxPoolingCols,
		perceptron.New(hidden, out, outputActivation),
	)
}

// InitCNN initializes the model using the Xavier (Glorot) method.
func InitCNN(model *cnn.Model, rndGen *rand.LockedRand) {
	for i := 0; i < len(model.Convolution.K); i++ {
		initializers.XavierUniform(model.Convolution.K[i].Value(), initializers.Gain(model.Convolution.Activation), rndGen)
		initializers.XavierUniform(model.Convolution.B[i].Value(), initializers.Gain(model.Convolution.Activation), rndGen)
	}
	model.FinalLayer.ForEachParam(func(param *nn.Param) {
		if param.Type() == nn.Weights {
			initializers.XavierUniform(param.Value(), initializers.Gain(ag.OpSoftmax), rndGen)
		}
	})
}
