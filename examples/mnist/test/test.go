// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/nlpodyssey/spago/examples/mnist/internal/mnist"
	"github.com/nlpodyssey/spago/examples/mnist/third_party/GoMNIST"
	"github.com/nlpodyssey/spago/pkg/ml/ag"
	"github.com/nlpodyssey/spago/pkg/utils"
	"os"
)

func main() {
	modelPath := os.Args[1]

	var datasetPath string
	if len(os.Args) > 2 {
		datasetPath = os.Args[2]
	} else {
		// assuming default path
		datasetPath = "examples/mnist/third_party/GoMNIST/data"
	}

	_, testSet, err := GoMNIST.Load(datasetPath)
	if err != nil {
		panic("Error reading MNIST data.")
	}

	// new model initialized with zeros
	model := mnist.NewMLP(
		784, // input
		100, // hidden
		10,  // output
		ag.OpReLU,
		ag.OpSoftmax,
	)

	err = utils.DeserializeFromFile(modelPath, model)
	if err != nil {
		panic("mnist: error during model deserialization.")
	}

	precision := mnist.NewEvaluator(model).Evaluate(mnist.Dataset{
		Set:              testSet,
		FeaturesAsVector: true,
	}).Precision()
	fmt.Printf("Accuracy: %.2f\n", 100*precision)
}
