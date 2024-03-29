package tensor

import (
	types "gograd/tensor/types"
	"reflect"
)

// fieldalignment -fix gograd/tensor
type Tensor[T types.TensorType] struct {
	Err       error
	data_buff []T
	shape     types.Shape
	strides   []int
	dim_order []uint16
}

func (tensor *Tensor[T]) Shape() types.Shape {
	return tensor.shape
}

func (tensor *Tensor[T]) Size() uint32 {
	var res types.Dim = 1
	for _, d := range tensor.shape {
		res *= d
	}
	return uint32(res)
}

func (tensor *Tensor[T]) Strides() []int {
	return tensor.strides
}

func (tensor *Tensor[T]) Order() []uint16 {
	return tensor.dim_order
}

func (tensor *Tensor[T]) Data() []T {
	return tensor.data()
}

func (tensor *Tensor[T]) data() []T {
	return tensor.data_buff
}

func (tensor *Tensor[T]) Item() T {
	data := tensor.data()
	if len(data) > 1 {
		panic("cannot use Item() on non-scalar tensors")
	}
	return data[0]
}

func (tensor *Tensor[T]) DType() reflect.Type {
	return getTypeArray(tensor.data())
}

func (tensor *Tensor[T]) MustAssert() *Tensor[T] {
	if tensor.Err != nil {
		panic(tensor.Err)
	}
	return tensor
}

func MustAssertAll[T types.TensorType](tensors ...*Tensor[T]) {
	for _, t := range tensors {
		t.MustAssert()
	}
}

// returns first tensor that has an error
func AnyErrors[T types.TensorType](tensors ...*Tensor[T]) *Tensor[T] {
	for _, t := range tensors {
		if t.Err != nil {
			return t
		}
	}
	return nil
}
