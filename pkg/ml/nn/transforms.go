// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nn

import (
	"github.com/nlpodyssey/spago/pkg/mat"
	"github.com/nlpodyssey/spago/pkg/ml/ag"
	"math"
	"sync"
)

// Linear performs a linear transformation of the type Wx.
func Linear(g *ag.Graph, w, x ag.Node) ag.Node {
	return g.Mul(w, x)
}

// Affine performs an affine transformation over an arbitrary (odd) number of nodes held in the input.
// The first node is the “bias”, which is added to the output as-is.
// The remaining nodes of the form "Wx" are multiplied together in pairs, then added.
// The pairs except the first whose "x" is nil are not considered.
// y = b + W1x1 + W2x2 + ... + WnXn
func Affine(g *ag.Graph, xs ...ag.Node) ag.Node {
	if len(xs)%2 == 0 {
		panic("nn: the number of arguments of the affine transformation should be odd")
	}

	// Optimize bounds checks
	x := xs[2]
	w := xs[1]
	y := g.Add(xs[0], Linear(g, w, x)) // b + Wx

	for i := 3; i < len(xs)-1; i += 2 {
		w := xs[i]
		x := xs[i+1]
		if x != nil {
			y = g.Add(y, Linear(g, w, x))
		}
	}
	return y
}

// BiLinear performs a bilinear transformation of the type (x_1 W x_2)
func BiLinear(g *ag.Graph, w, x1, x2 ag.Node) ag.Node {
	return g.Mul(g.Mul(g.T(x1), w), x2)
}

// BiAffine performs a biaffine transformation.
func BiAffine(g *ag.Graph, w, u, v, b, x1, x2 ag.Node) ag.Node {
	return g.Add(g.Add(g.Add(BiLinear(g, w, x1, x2), g.Mul(g.T(u), x1)), g.Mul(g.T(v), x2)), b)
}

// Conv2D performs a 2D convolution.
func Conv2D(g *ag.Graph, w, x ag.Node, xStride, yStride int) ag.Node {
	var dimx, dimy int
	if (x.Value().Rows()-w.Value().Rows())%xStride != 0 {
		panic("Incompatible stride value for rows")
	}
	if (x.Value().Columns()-w.Value().Columns())%yStride != 0 {
		panic("Incompatible stride value for columns")
	}
	dimx = (x.Value().Rows()-w.Value().Rows())/xStride + 1
	dimy = (x.Value().Columns()-w.Value().Columns())/yStride + 1

	var outList []ag.Node
	for i := 0; i < dimx; i++ {
		for j := 0; j < dimy; j++ {
			var view = g.View(x, i*xStride, j*yStride, w.Value().Rows(), w.Value().Columns())
			var dotProduct = g.Dot(view, w)
			outList = append(outList, dotProduct)
		}
	}

	return g.Reshape(g.Concat(outList...), dimx, dimy)
}

// ScaledDotProductAttention is a self-attention mechanism relating different positions of a single sequence in order to compute a representation of the same sequence.
// This method requires that the query, the key and the value vectors have already been obtained from the input sequence.
// The scaled factor is the square root of the dimension of the key vectors.
func ScaledDotProductAttention(g *ag.Graph, qs, ks, vs []ag.Node, scaledFactor float64) (context []ag.Node, probs []mat.Matrix) {
	context = make([]ag.Node, len(qs))
	probs = make([]mat.Matrix, len(qs))
	keys := g.Stack(ks...)
	values := g.T(g.Stack(vs...))
	divTerm := g.NewScalar(scaledFactor)
	for i, q := range qs {
		attScores := g.DivScalar(g.Mul(keys, q), divTerm)
		attProbs := g.Softmax(attScores)
		context[i] = g.Mul(values, attProbs)
		probs[i] = attProbs.Value()
	}
	return
}

// ScaledDotProductAttentionConcurrent does the same thing as ScaledDotProductAttention but processes input concurrently.
func ScaledDotProductAttentionConcurrent(g *ag.Graph, qs, ks, vs []ag.Node, scaledFactor float64) (context []ag.Node, probs []mat.Matrix) {
	context = make([]ag.Node, len(qs))
	probs = make([]mat.Matrix, len(qs))
	keys := g.Stack(ks...)
	values := g.T(g.Stack(vs...))
	divTerm := g.NewScalar(scaledFactor)
	var wg sync.WaitGroup
	wg.Add(len(qs))
	for i, q := range qs {
		go func(i int, q ag.Node) {
			defer wg.Done()
			attScores := g.DivScalar(g.Mul(keys, q), divTerm)
			attProbs := g.Softmax(attScores)
			context[i] = g.Mul(values, attProbs)
			probs[i] = attProbs.Value()
		}(i, q)
	}
	wg.Wait()
	return
}

// Separate returns a matrix of Node(s) represented as a slice of slice containing the elements extracted from the input.
// The dimensions of the resulting matrix are the same of the input.
func Separate(g *ag.Graph, x ag.Node) [][]ag.Node {
	rows, cols := x.Value().Dims()
	ys := make([][]ag.Node, rows)
	for i := range ys {
		row := make([]ag.Node, cols)
		for j := range row {
			row[j] = g.At(x, i, j)
		}
		ys[i] = row
	}
	return ys
}

// SeparateVec returns a slice of Node(s) containing the elements extracted from the input.
// The size of the vector equals the number of input elements.
// You can think of this method as the inverse of the ag.Concat operator.
func SeparateVec(g *ag.Graph, x ag.Node) []ag.Node {
	size := x.Value().Size()
	ys := make([]ag.Node, size)
	for i := 0; i < size; i++ {
		ys[i] = g.AtVec(x, i)
	}
	return ys
}

// TODO: optimize, this is extremely inefficient!
func SplitVec(g *ag.Graph, x ag.Node, chunks int) []ag.Node {
	size := int(math.Ceil(float64(x.Value().Size()) / float64(chunks)))
	lastSize := x.Value().Size() % chunks
	ys := make([]ag.Node, chunks)
	for c := 0; c < chunks; c++ {
		length := 0
		if c == chunks-1 && lastSize > 0 {
			length = lastSize
		} else {
			length = size
		}
		tmp := make([]ag.Node, length)
		for i := 0; i < length; i++ {
			tmp[i] = g.AtVec(x, i+c*size)
		}
		ys[c] = g.Concat(tmp...)
	}
	return ys
}
