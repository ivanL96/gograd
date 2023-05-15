package tensor

import (
	"fmt"
)

func make_tensor[T TensorType](data_p *[]T, shape Shape) *Tensor[T] {
	var shapeProd Dim = 1
	for _, dim := range shape {
		shapeProd *= dim
	}
	var data []T
	if data_p == nil {
		data = make([]T, shapeProd)
	} else {
		data = append([]T(nil), (*data_p)...)
	}
	if len(shape) == 0 || int(shapeProd) != len(data) {
		panic(fmt.Sprintf("Value length %v cannot have shape %v", len(data), shape))
	}
	dim_order := make([]int, len(shape))
	for i := range dim_order {
		dim_order[i] = i
	}
	return &Tensor[T]{
		shape:     append(Shape(nil), shape...),
		strides:   getStrides(shape),
		data:      data,
		dtype:     getTypeArray(data),
		shapeProd: shapeProd,
		dim_order: dim_order,
	}
}

func InitTensor[T TensorType](value []T, shape Shape) *Tensor[T] {
	// inits a tensor with data
	return make_tensor(&value, shape)
}

func InitEmptyTensor[T TensorType](shape ...Dim) *Tensor[T] {
	return make_tensor[T](nil, shape)
}

func (tensor *Tensor[T]) Set(value []T) *Tensor[T] {
	// sets new value of the same shape
	length := uint(len(value))
	var prod uint = 1
	for _, dim := range tensor.shape {
		prod = uint(dim) * prod
	}
	if prod != length {
		msg := fmt.Sprintf(
			"Shape %v cannot fit the number of elements %v. Change the shape first",
			tensor.shape, length)
		panic(msg)
	}
	tensor.data = value
	return tensor
}

func AsType[OLDT TensorType, NEWT TensorType](tensor *Tensor[OLDT]) *Tensor[NEWT] {
	// naive impl with copying the data & tensor
	// example:
	// AsType(int32, float64)(tensor) ==> float64 tensor
	data := make([]NEWT, len(tensor.data))
	for i, val := range tensor.data {
		data[i] = NEWT(val)
	}
	new_tensor := InitTensor(data, tensor.shape)
	return new_tensor
}

func (tensor *Tensor[T]) Copy() *Tensor[T] {
	new_data := make([]T, len(tensor.data))
	copy(new_data, tensor.data)
	new_copy := InitTensor(new_data, tensor.shape)
	return new_copy
}

func (tensor *Tensor[T]) IsEqual(other_tensor *Tensor[T]) bool {
	// iterates over two tensors and compares elementwise
	if !Equal_1D_slices(tensor.shape, other_tensor.shape) {
		return false
	}
	if !Equal_1D_slices(tensor.dim_order, other_tensor.dim_order) {
		return false
	}
	if !Equal_1D_slices(tensor.data, other_tensor.data) {
		return false
	}
	return true
}

func (tensor *Tensor[T]) Fill(fill_value T) *Tensor[T] {
	for i := range tensor.data {
		tensor.data[i] = fill_value
	}
	return tensor
}

func Range[T TensorType](limits ...int) *Tensor[T] {
	// Created a tensor with data ranged from 'start' to 'end'
	// limits: min 1 and max 3 arguments. Start, End, Step
	if len(limits) == 0 {
		panic("Range requires at least one argument")
	}
	start, end, step := 0, 0, 1
	if len(limits) == 1 {
		end = limits[0]
	}
	if len(limits) >= 2 {
		start = limits[0]
		end = limits[1]
	}
	if len(limits) == 3 {
		step = limits[2]
	}
	length := ((end - start) + step - 1) / step
	tensor := InitEmptyTensor[T](Dim(length))
	for i := 0; i < length; i++ {
		tensor.data[i] = T(start)
		start += step
	}
	return tensor
}
