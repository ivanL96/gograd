package tensor

import (
	"errors"
	"fmt"
	"gograd/tensor/internal"
	types "gograd/tensor/types"
	"math"
	"runtime"
	"sync"
)

var numCPU int = runtime.NumCPU()

// set of primitive common tensor methods
//tensor initialization-----------------------------------------------------

func makeTensor[T types.TensorType](dataPtr *[]T, shape types.Shape, copy bool) *Tensor[T] {
	var shapeProd types.Dim = 1
	for _, dim := range shape {
		shapeProd *= dim
	}

	var tensor Tensor[T]

	if shapeProd == 0 {
		tensor.Err = errors.New("shape cannot have zero dim")
		return &tensor
	}

	var data []T
	switch dataPtr {
	case nil:
		// if nil ptr create an empty slice with size of 'shapeProd'
		data = make([]T, shapeProd)
	default:
		if copy {
			// copies data
			data = append([]T(nil), (*dataPtr)...)
		} else {
			data = *dataPtr
		}
	}
	if len(shape) == 0 || int(shapeProd) != len(data) {
		tensor.Err = fmt.Errorf("makeTensor: Value length %v cannot have shape %v", len(data), shape)
		return &tensor
	}

	tensor.shape = append(types.Shape(nil), shape...)
	tensor.strides = shape.GetStrides()
	tensor.data_buff = data
	tensor.dim_order = shape.InitDimOrder()
	return &tensor
}

// inits a tensor with data
func CreateTensor[T types.TensorType](value []T, shape types.Shape) *Tensor[T] {
	return makeTensor(&value, shape, true)
}

func CreateTensorNoCopy[T types.TensorType](value []T, shape types.Shape) *Tensor[T] {
	return makeTensor(&value, shape, false)
}

// inits an empty tensor with specific shape
func CreateEmptyTensor[T types.TensorType](shape ...types.Dim) *Tensor[T] {
	return makeTensor[T](nil, shape, true)
}

func Ones[T types.TensorType](shape ...types.Dim) *Tensor[T] {
	return CreateEmptyTensor[T](shape...).Fill(1)
}

func Zeros[T types.TensorType](shape ...types.Dim) *Tensor[T] {
	return CreateEmptyTensor[T](shape...)
}

// A tensor with Shape 1
func Scalar[T types.TensorType](value T) *Tensor[T] {
	return &Tensor[T]{
		shape:     types.Shape{1},
		strides:   []int{1},
		data_buff: []T{value},
		dim_order: []uint16{0},
	}
}

// -----------------------------------------------------

// Naive impl with copying the data & tensor
//
// Example:
//
// AsType(int32, float64)(int_tensor) ==> float64 tensor
func AsType[OLD_T, NEW_T types.TensorType](tensor *Tensor[OLD_T]) *Tensor[NEW_T] {
	if tensor.Err != nil {
		ret := Scalar[NEW_T](0)
		ret.Err = tensor.Err
		return ret
	}

	out_data := make([]NEW_T, len(tensor.data()))
	out_tensor := CreateTensorNoCopy(out_data, tensor.shape)

	var wg sync.WaitGroup
	wg.Add(numCPU)

	chunk_size := (len(out_data) + numCPU - 1) / numCPU

	for i := 0; i < numCPU; i++ {
		start := i * chunk_size
		end := (i + 1) * chunk_size
		if end > len(out_data) {
			end = len(out_data)
		}

		go func(start, end int) {
			defer wg.Done()
			if start >= end {
				return
			}
			var out_batch []NEW_T
			var in_batch []OLD_T
			switch end {
			case len(out_data):
				out_batch = out_data[start:end]
				in_batch = tensor.data()[start:end]
			default:
				out_batch = out_data[start:]
				in_batch = tensor.data()[start:]
			}
			internal.Convert_type_loop[OLD_T, NEW_T](in_batch, out_batch)
		}(start, end)
	}
	wg.Wait()
	return out_tensor
}

func (tensor *Tensor[T]) Copy() *Tensor[T] {
	if tensor.Err != nil {
		return tensor
	}
	newData := make([]T, len(tensor.data()))
	copy(newData, tensor.data())
	newTensor := CreateTensorNoCopy(newData, tensor.shape)
	newTensor.strides = tensor.strides
	newTensor.dim_order = tensor.dim_order
	return newTensor
}

// Creates a tensor with data ranged from 'start' to 'end'
// limits: min 1 and max 3 arguments. Start, End, Step
func Range[T types.TensorType](limits ...int) *Tensor[T] {
	if len(limits) == 0 {
		panic("range requires at least one argument")
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
	tensor := CreateEmptyTensor[T](types.Dim(length))
	for i := 0; i < length; i++ {
		tensor.data()[i] = T(start)
		start += step
	}
	return tensor
}

// Fills tensor with same value
func (tensor *Tensor[T]) Fill(value T) *Tensor[T] {
	if tensor.Err != nil {
		return tensor
	}
	var wg sync.WaitGroup
	wg.Add(numCPU)

	data := tensor.data()
	chunk_size := (len(data) + numCPU - 1) / numCPU

	for i := 0; i < numCPU; i++ {
		start := i * chunk_size
		end := (i + 1) * chunk_size
		if end > len(data) {
			end = len(data)
		}

		go func(start, end int) {
			defer wg.Done()
			if start >= end {
				return
			}
			var batch []T
			switch end {
			case len(data):
				batch = data[start:]
			default:
				batch = data[start:end]
			}
			internal.Fill_data_loop(batch, value)
		}(start, end)
	}
	wg.Wait()
	return tensor
}

// creates an array (2d tensor) with ones on the diagonal and zeros elsewhere
func Eye[T types.TensorType](x, y types.Dim) *Tensor[T] {
	eye := CreateEmptyTensor[T](types.Shape{x, y}...)
	for i := 0; i < int(x); i++ {
		for j := 0; j < int(y); j++ {
			if i == j {
				fidx := get_flat_idx_fast(eye.strides, i, j)
				eye.data()[fidx] = 1
			}
		}
	}
	return eye
}

func (tensor *Tensor[T]) DiagFlat() *Tensor[T] {
	if tensor.Err != nil {
		return tensor
	}
	n := types.Dim(tensor.Size())
	diag := CreateEmptyTensor[T](types.Shape{n, n}...)
	for i := 0; i < int(n); i++ {
		for j := 0; j < int(n); j++ {
			if i == j {
				fidx := get_flat_idx_fast(diag.strides, i, j)
				diag.data()[fidx] = tensor.data()[i]
			}
		}
	}
	return diag
}

// sets new value of the same shape
func (tensor *Tensor[T]) SetData(value []T) *Tensor[T] {
	if tensor.Err != nil {
		return tensor
	}
	length := uint(len(value))
	var prod uint = 1
	for _, dim := range tensor.shape {
		prod = uint(dim) * prod
	}
	if prod != length {
		tensor.Err = fmt.Errorf(
			"Shape %v cannot fit the number of elements %v. Change the shape first",
			tensor.shape, length)
		return tensor
	}
	// TODO avoid data_buff
	tensor.data_buff = value
	return tensor
}

// set scalar to specific index
func (tensor *Tensor[T]) Set(indexes []int, value T) {
	if tensor.Err != nil {
		return
	}
	flatIndex, err := tensor.getFlatIndex(indexes...)
	if err != nil {
		tensor.Err = err
		return
	}
	tensor.data()[flatIndex] = value
}

func (tensor *Tensor[T]) CreateIterator() *TensorIterator {
	return CreateIterator(len(tensor.data()), tensor.shape)
}

// Compares shapes and data:
// iterates over two tensors and compares elementwise
func (tensor *Tensor[T]) IsEqual(other *Tensor[T]) (bool, error) {
	if tensor.Err != nil {
		return false, tensor.Err
	}
	if other.Err != nil {
		return false, other.Err
	}
	if !tensor.shape.Equals(other.shape) {
		return false, nil
	}
	if tensor.IsContiguous() && other.IsContiguous() {
		return EqualSlices[T](tensor.data(), other.data()), nil
	}

	it := tensor.CreateIterator()
	for it.Iterate() {
		idx := it.Next()
		if tensor.Get_fast(idx...) != other.Get_fast(idx...) {
			return false, nil
		}
	}
	return true, nil
}

func (tensor *Tensor[T]) IsAllClose(tensor_or_scalar *Tensor[T], tol float64) (bool, error) {
	if tensor.Err != nil {
		return false, tensor.Err
	}
	if tensor_or_scalar.Err != nil {
		return false, tensor_or_scalar.Err
	}
	if tensor_or_scalar.Shape().IsScalarLike() {
		other_val := tensor_or_scalar.Item()
		for _, val := range tensor.data() {
			if math.Abs(float64(val-other_val)) > tol {
				return false, nil
			}
		}
	} else if tensor_or_scalar.Shape().Equals(tensor.Shape()) {
		it := tensor.CreateIterator()
		for it.Iterate() {
			idx := it.Next()
			a := tensor.Get_fast(idx...)
			b := tensor_or_scalar.Get_fast(idx...)
			if math.Abs(float64(a-b)) > tol {
				return false, nil
			}
		}
	} else {
		return false, errors.New("other argument must be either tensor with the same shape or scalar")
	}
	return true, nil
}

func (tensor *Tensor[T]) HasNaN() (bool, error) {
	if tensor.Err != nil {
		return false, tensor.Err
	}
	data := tensor.data()
	for i := 0; i < len(data); i++ {
		if math.IsNaN(float64(data[i])) {
			return true, nil
		}
	}
	return false, nil
}
