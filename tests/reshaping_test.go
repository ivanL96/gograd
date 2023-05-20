package main

import (
	"gograd/tensor"
	"testing"
)

// RESHAPING
func TestBroadcast(t *testing.T) {
	a := tensor.InitEmptyTensor[int32](3, 2)
	br_a := a.Broadcast(3, 1, 1)
	assertEqualSlices(t, br_a.Shape(), tensor.Shape{3, 3, 2})

	b := tensor.InitEmptyTensor[int32](1, 1)
	br_b := b.Broadcast(6)
	assertEqualSlices(t, br_b.Shape(), tensor.Shape{1, 6})

	c := tensor.InitEmptyTensor[int32](1)
	br_c := c.Broadcast(3)
	assertEqualSlices(t, br_c.Shape(), tensor.Shape{3})
}

func TestFlatten(t *testing.T) {
	super_nested_arr := tensor.InitTensor([]int32{1, 2, 3, 4, 5, 6, 7, 8, 9}, tensor.Shape{3, 1, 3, 1, 1})
	assertEqualSlices(t, super_nested_arr.Flatten().Shape(), tensor.Shape{9})
}

func TestReshape(t *testing.T) {
	a := tensor.InitTensor([]int32{1, 2, 3, 4, 5, 6}, tensor.Shape{2, 3})
	assertEqualSlices(t, a.Reshape(3, 2).Shape(), tensor.Shape{3, 2})
	assertEqualSlices(t, a.Reshape(1, 1, 1, 3, 2).Shape(), tensor.Shape{1, 1, 1, 3, 2})
}

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
