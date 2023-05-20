package tensor

import (
	"reflect"

	"golang.org/x/exp/constraints"
)

type Any interface{}

type Float interface {
	float32 | float64
}

type TensorType interface {
	constraints.Float | constraints.Integer
}

type Dim uint
type Shape []Dim

type Tensor[T TensorType] struct {
	dtype     reflect.Type
	data      []T
	shape     Shape
	strides   []int
	dim_order []int
}

func (tensor *Tensor[T]) Shape() Shape {
	return tensor.shape
}

func (tensor *Tensor[T]) Strides() []int {
	return tensor.strides
}

func (tensor *Tensor[T]) Order() []int {
	return tensor.dim_order
}

func (tensor *Tensor[T]) Data() []T {
	return tensor.data
}

func (tensor *Tensor[T]) DType() reflect.Type {
	return tensor.dtype
}
