package tensor

import (
	types "flamego/tensor/types"
	"reflect"
)

// fieldalignment -fix flamego/tensor
type Tensor[T types.TensorType] struct {
	data_buff []T
	shape     types.Shape
	strides   []int
	dim_order []uint16
	flags     uint8
}

type TensorList[T types.TensorType] []*Tensor[T]

func (tensor *Tensor[T]) Shape() types.Shape {
	return tensor.shape
}

func (tensor *Tensor[T]) Strides() []int {
	return tensor.strides
}

func (tensor *Tensor[T]) Order() []uint16 {
	return tensor.dim_order
}

// accessing internal data struct will automatically disable all optimization flags for this Tensor.
func (tensor *Tensor[T]) Data() []T {
	tensor.ResetFlags()
	return tensor.data_buff
}

func (tensor *Tensor[T]) data() []T {
	return tensor.data_buff
}

func (tensor *Tensor[T]) DType() reflect.Type {
	return getTypeArray(tensor.data())
}

// tensor helper flags
const (
	SameValuesFlag uint8 = 1 << iota
)

func (tensor *Tensor[T]) SetFlag(flag uint8) {
	tensor.flags |= flag
}
func (tensor *Tensor[T]) ClearFlag(flag uint8) {
	tensor.flags &^= flag
}
func (tensor *Tensor[T]) ToggleFlag(flag uint8) {
	tensor.flags ^= flag
}
func (tensor *Tensor[T]) HasFlag(flag uint8) bool {
	return tensor.flags&flag != 0
}
func (tensor *Tensor[T]) ResetFlags() {
	tensor.flags = 0
}

// func (tensor *Tensor[T]) Flags() {
// 	for i, flag := range []uint8{SameValuesFlag} {
// 		tensor.hasFlag(flag)
// 	}
// }
