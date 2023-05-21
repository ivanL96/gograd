package main

import (
	"gograd/tensor"
	"testing"
)

func TestIndex(t *testing.T) {
	a := tensor.Range[int32](8).Reshape(2, 2, 2)
	sub1 := a.Index(1)
	assertEqualSlices(t, sub1.Data(), []int32{4, 5, 6, 7})
	assertEqualSlices(t, sub1.Shape(), tensor.Shape{2, 2})

	sub2 := sub1.Index(1)
	assertEqualSlices(t, sub2.Data(), []int32{6, 7})
	assertEqualSlices(t, sub2.Shape(), tensor.Shape{2})

	b := tensor.InitTensor([]int32{0, 1, 2, 3}, tensor.Shape{2, 2})
	bsub := b.Index(-2)
	assertEqualSlices(t, bsub.Data(), []int32{0, 1})
	assertEqualSlices(t, bsub.Shape(), tensor.Shape{2})
}

func TestTransposeAndIndex(t *testing.T) {

	a := tensor.Range[int32](8).Reshape(4, 2)
	a.ToString()
	a = a.Transpose().Index(1)
	a.ToString()
	assertEqualSlices(t, a.Data(), []int32{1, 3, 5, 7})
	assertEqualSlices(t, a.Shape(), tensor.Shape{4})

	b := tensor.Range[int32](8).Reshape(4, 1, 1, 2)
	b.ToString()
	b = b.Transpose().Index(1)
	assertEqualSlices(t, b.Data(), []int32{1, 3, 5, 7})
	assertEqualSlices(t, b.Shape(), tensor.Shape{1, 1, 4})

	c := tensor.Range[int32](3*4*2).Reshape(3, 4, 2)
	c = c.Transpose().Index(0)
	assertEqualSlices(t,
		c.Data(), []int32{0, 8, 16, 2, 10, 18, 4, 12, 20, 6, 14, 22})
	assertEqualSlices(t, c.Shape(), tensor.Shape{4, 3})

	d := tensor.Range[int32](8).Reshape(4, 2, 1)
	d = d.Transpose().Index(0)
	d.ToString()
	assertEqualSlices(t, d.Data(), []int32{0, 2, 4, 6, 1, 3, 5, 7})
	assertEqualSlices(t, d.Shape(), tensor.Shape{2, 4})

	e := tensor.InitTensor([]int32{3}, tensor.Shape{1, 1, 1, 1})
	e = e.Transpose().Index(0)
	assertEqualSlices(t, e.Data(), []int32{3})
	assertEqualSlices(t, e.Shape(), tensor.Shape{1, 1, 1})
}