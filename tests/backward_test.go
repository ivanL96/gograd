package main

import (
	"gograd/grad"
	"gograd/tensor"
	types "gograd/tensor/types"
	"math"
	"testing"
)

// go test -run TestNewPerson -v
// go test '-run=^TestMyTest$'
func TestGradAdd(t *testing.T) {
	a := grad.Variable[float32](tensor.CreateTensor(
		[]float32{4}, types.Shape{1, 1}))
	b := grad.Variable[float32](tensor.CreateTensor(
		[]float32{5}, types.Shape{1, 1}))
	z := a.Add(b).MustAssert()
	assertEqualSlices(t, z.Value.Data(), []float32{9})
	z.Backward(nil)
	assertEqualSlices(t, a.Grad.Data(), []float32{1})
	assertEqualSlices(t, b.Grad.Data(), []float32{1})

	_add := func(x float64) float64 {
		return x + 5
	}
	deriv := grad.NumericDeriv(grad.EPSILON, 4, _add)
	assert(t, math.Abs(1-deriv) <= 0.0001)
	a.Value.MustAssert()
	b.Value.MustAssert()
	z.Value.MustAssert()
}

func TestGradSub(t *testing.T) {
	a := grad.Variable[float32](tensor.CreateTensor(
		[]float32{4}, types.Shape{1, 1}))
	b := grad.Variable[float32](tensor.CreateTensor(
		[]float32{5}, types.Shape{1, 1}))
	z := a.Sub(b).MustAssert()
	assertEqualSlices(t, z.Value.Data(), []float32{-1})
	z.Backward(nil)
	assertEqualSlices(t, a.Grad.Data(), []float32{1})
	assertEqualSlices(t, b.Grad.Data(), []float32{-1})
	a.Value.MustAssert()
	b.Value.MustAssert()
	z.Value.MustAssert()

	_sub := func(x float64) float64 {
		return x - 5
	}
	deriv := grad.NumericDeriv(grad.EPSILON, 4, _sub)
	assert(t, math.Abs(1-deriv) <= 0.0001)
}

func TestGradMul(t *testing.T) {
	a := grad.Variable[float32](tensor.CreateTensor(
		[]float32{4}, types.Shape{1, 1}))
	b := grad.Variable[float32](tensor.CreateTensor(
		[]float32{5}, types.Shape{1, 1}))
	z := a.Mul(b).MustAssert()
	assertEqualSlices(t, z.Value.Data(), []float32{20})
	z.Backward(nil)
	assertEqualSlices(t, a.Grad.Data(), []float32{5})
	assertEqualSlices(t, b.Grad.Data(), []float32{4})
	a.Value.MustAssert()
	b.Value.MustAssert()
	z.Value.MustAssert()

	_mul := func(x float64) float64 {
		return x * 5
	}
	deriv := grad.NumericDeriv(grad.EPSILON, 4, _mul)
	assert(t, math.Abs(5-deriv) <= 0.0001)

	_mul2 := func(x float64) float64 {
		return x * 4
	}
	deriv = grad.NumericDeriv(grad.EPSILON, 5, _mul2)
	assert(t, math.Abs(4-deriv) <= 0.0001)
}

func TestGradMatMul(t *testing.T) {
	a := grad.Variable(tensor.Range[float32](5).Reshape(1, 5))
	b := grad.Variable(tensor.Range[float32](5).Reshape(5, 1))
	z := a.MatMul(b).MustAssert()
	z.Backward(nil)
	assertEqualSlices(t, z.Value.Shape(), types.Shape{1, 1})
	assertEqualSlices(t, z.Value.Data(), []float32{30})
	assertEqualSlices(t, a.Grad.Data(), []float32{0, 1, 2, 3, 4})
	assertEqualSlices(t, a.Grad.Shape(), types.Shape{1, 5})
	assertEqualSlices(t, b.Grad.Data(), []float32{0, 1, 2, 3, 4})
	assertEqualSlices(t, b.Grad.Shape(), types.Shape{5, 1})
}

func TestGradMatMulMean(t *testing.T) {
	a := grad.Variable(tensor.Range[float32](8).Reshape(4, 2))
	b := grad.Variable(tensor.Range[float32](10).Reshape(2, 5))
	z := a.MatMul(b).Mean().MustAssert()
	z.Backward(nil)
	assertEqualSlices(t, z.Value.Shape(), types.Shape{1})
	assertEqualSlices(t, z.Value.Data(), []float32{34})
	assertEqualSlices(t, a.Grad.Data(), []float32{
		0.5, 1.75, 0.50, 1.75, 0.5, 1.75, 0.5, 1.75})
	assertEqualSlices(t, a.Grad.Shape(), types.Shape{4, 2})
	is_close, err := b.Grad.IsAllClose(
		tensor.CreateTensor([]float32{0.6, 0.6, 0.6, 0.6, 0.6, 0.8, 0.8, 0.8, 0.8, 0.8}, types.Shape{2, 5}),
		0.00001,
	)
	assert(t, is_close)
	assert(t, err == nil)
	assertEqualSlices(t, b.Grad.Shape(), types.Shape{2, 5})
}

func TestGradMatMulMSE(t *testing.T) {
	x := grad.Constant[float32](tensor.Range[float32](10).Div(tensor.Scalar[float32](10)).Reshape(10, 1))
	y := grad.Constant[float32](tensor.Range[float32](10).Reshape(10, 1))
	w := grad.Variable[float32](tensor.Range[float32](1).Reshape(1, 1))
	b := grad.Variable[float32](tensor.Range[float32](1).Reshape(1, 1))
	yhat := x.MatMul(w).Add(b)
	loss := yhat.MSE(y).MustAssert()
	loss.Backward(nil)
	assertEqualSlices(t, loss.Value.Data(), []float32{28.5})
	assertEqualSlices(t, w.Grad.Data(), []float32{-5.7})
	assertEqualSlices(t, b.Grad.Data(), []float32{-9})
}

func TestGradReluMean(t *testing.T) {
	a := grad.Variable[float32](
		tensor.CreateTensor([]float32{-1., 0., 4., -2.}, types.Shape{4}),
	)
	b := a.Relu().MustAssert()
	c := b.Mean().MustAssert()
	c.Backward(nil)
	assertEqualSlices(t, a.Grad.Data(), []float32{0, 0, 0.25, 0})
	assertEqualSlices(t, a.Grad.Shape(), types.Shape{4})
}

// func TestGradCrossEntropy(t *testing.T) {
// 	xdata := []float32{
// 		.2, .8,
// 		.4, .6,
// 		.1, .9}
// 	x := grad.VarFrom(xdata, types.Shape{3, 2})
// 	y := grad.VarFrom([]float32{1, 0, 1}, types.Shape{3})
// 	out := x.SoftmaxCrossEntropy(y)
// 	fmt.Println(out.MustAssert().ToString())
// 	fmt.Println(out.Mean().MustAssert().ToString())
// }

func TestGradStep(t *testing.T) {
	rng := tensor.NewRNG(1)
	a := rng.RandomFloat32(10)
	grad := rng.RandomFloat32(10)
	lr := tensor.Scalar[float32](0.001)
	for i := 0; i < 10; i++ { // to imitate repetitive param update
		a.GradientStep(grad, lr).MustAssert()
	}
	// fmt.Println(a.ToString())
	out := tensor.CreateTensor([]float32{
		0.59861338, 0.93110406, 0.65791476, 0.43333712, 0.42039126,
		0.67995483, 0.06498063, 0.15495403, 0.09599983, 0.29790273,
	}, types.Shape{10})

	is_close, err := a.IsAllClose(out, 0.001)
	assert(t, is_close)
	assert(t, err == nil)
}
