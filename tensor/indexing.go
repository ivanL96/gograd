package tensor

import (
	"errors"
	"fmt"
	"gograd/tensor/internal"
	"strconv"
	"strings"
)

func (tensor *Tensor[T]) getFlatIndex(indices ...int) (int, error) {
	flatIndex := 0
	for i, ind := range indices {
		dim := int(tensor.shape[i])
		// resolve negative indexes
		if ind < 0 {
			norm_ind := dim + ind
			if norm_ind < 0 {
				return 0, fmt.Errorf("index %v is out of bounds", ind)
			}
			ind = norm_ind
		}
		// bound check
		if ind >= dim {
			return 0, fmt.Errorf("index %v is out of bounds for dim %v", ind, dim)
		}
		flatIndex += tensor.strides[i] * ind
	}
	return flatIndex, nil
}

func get_flat_idx_fast(strides []int, indices ...int) int {
	switch len(indices) {
	case 1:
		return strides[0] * indices[0]
	case 2:
		return strides[0]*indices[0] + strides[1]*indices[1]
	default:
		flatIndex := 0
		for i, ind := range indices {
			flatIndex += strides[i] * ind
		}
		return flatIndex
	}
}

// faster Get() without bounds checking. Does not support negative indexing
func (tensor *Tensor[T]) Get_fast(indices ...int) T {
	return tensor.data()[get_flat_idx_fast(tensor.strides, indices...)]
}

func (tensor *Tensor[T]) Get(indices ...int) (T, error) {
	if tensor.Err != nil {
		return 0, tensor.Err
	}
	if len(indices) != len(tensor.shape) {
		return 0, fmt.Errorf(
			"incorrect number of indices. Must be %v got %v", len(tensor.shape), len(indices))
	}
	flatIndex, err := tensor.getFlatIndex(indices...)
	if err != nil {
		return 0, err
	}
	return tensor.data()[flatIndex], nil
}

// returns sub data for given indices.
func (tensor *Tensor[T]) Index(indices ...int) *Tensor[T] {
	if tensor.Err != nil {
		return tensor
	}
	n_indices := len(indices)
	n_dims := len(tensor.shape)
	if n_indices == 0 {
		tensor.Err = errors.New("at leat one index is required in Index()")
		return tensor
	}
	if n_indices > n_dims {
		tensor.Err = errors.New("too many indices")
		return tensor
	}

	// index of the first elem in the sub tensor
	if n_indices == n_dims {
		flatIndex, err := tensor.getFlatIndex(indices...)
		if err != nil {
			tensor.Err = err
			return tensor
		}
		return Scalar[T](tensor.data()[flatIndex])
	}

	tensor = tensor.AsContiguous()
	innerShape := tensor.shape[n_indices:]
	flatIndex, err := tensor.getFlatIndex(indices...)
	if err != nil {
		tensor.Err = err
		return tensor
	}
	// if data layout is contiguous we can just take a slice start:end from data
	endFlatIndex := flatIndex + tensor.strides[n_indices-1]
	subData := tensor.data()[flatIndex:endFlatIndex]
	return CreateTensor(subData, innerShape)
}

// IdxRange is used to create a slice along specific axis.
// Setting start & end is needed to apply slicing boundaries.

type idxRange struct {
	start int
	end   int
}

// Idx() is an utility function is used for taking a specific index
func I(val int) *idxRange {
	return &idxRange{val, val}
}

// Axis() is used for taking entire axis-wide slice
func Axis() *idxRange {
	return &idxRange{0, -1}
}

// TODO Implement index slices
// func ISlc(start, end uint) *idxRange {
// 	return &idxRange{int(start), int(end)}
// }

func parse_indexes(expr string) ([]*idxRange, error) {
	if len(expr) == 0 {
		return nil, errors.New("arguments cannot be empty")
	}
	symbols := strings.Split(expr, ",")
	indices := make([]*idxRange, 0, len(symbols))
	for _, el := range symbols {
		el = strings.TrimSpace(el)
		switch el {
		case ":":
			indices = append(indices, Axis())
		case "":
			return nil, fmt.Errorf(
				"invalid expression: '%v'. Arguments should be numeric or ':' and separated by ','", expr,
			)
		default:
			if floatVal, err := strconv.ParseFloat(el, 64); err == nil {
				indices = append(indices, I(int(floatVal)))
				continue
			}
			return nil, fmt.Errorf(
				"found unknown symbol in expression: '%v'", el,
			)
		}
	}
	return indices, nil
}

// Advanced indexing allows to specify index ranges.
//
// Example: with given tensor:
//
// shape (2,3), strides (3,1)
//
//	[[1,2,3],
//	[4,5,6]]
//
// should return
// tensor.IndexAdv(":,0") ==> [1,4]
// tensor.IndexAdv(":,1") ==> [2,5]
//
// Getting a sub tensor by axis is similar to:
// a.TrC(2, 0, 1, 3, 4).Index(n) == a.IndexAdv(":,:,n,:,:")
func (tensor *Tensor[T]) IndexAdv(expr string) *Tensor[T] {
	indices, err := parse_indexes(expr)
	if err != nil {
		tensor.Err = err
		return tensor
	}
	return tensor.IndexAdv_(indices...)
}

func (tensor *Tensor[T]) IndexAdv_(indices ...*idxRange) *Tensor[T] {
	if tensor.Err != nil {
		return tensor
	}

	if len(indices) == 0 {
		tensor.Err = errors.New("at least one index is required")
		return tensor
	}
	if len(indices) > len(tensor.shape) {
		tensor.Err = errors.New("too many indices")
		return tensor
	}
	// remove trailing axis-wide idx
	// for example: [I(),Axis(),I(),Axis(),Axis()] => [I(),Axis(),I()]
	j := len(indices) - 1
	for i := len(indices) - 1; i >= 0; i-- {
		index_range := indices[i]
		if index_range.start == 0 && index_range.end == -1 {
			j -= 1
		} else {
			break
		}
	}
	indices = indices[:j+1]
	if len(indices) == 0 {
		return tensor.Copy()
	}

	are_constants_only := true
	for i := 0; i < len(indices); i++ {
		index_range := indices[i]
		if index_range.start == 0 && index_range.end == -1 {
			// axis wide
			are_constants_only = false
		} else if index_range.end != index_range.start {
			// sub axis
			are_constants_only = false
		}
	}

	if are_constants_only {
		idxs := make([]int, len(indices))
		for i, idx_range := range indices {
			idxs[i] = idx_range.start
		}
		return tensor.Index(idxs...)
	}

	// prepare axes to T()

	// Example: indices => [Axis(),I(i),I(j)]
	// init strides: [4,2,1]
	// init axes: [0,1,2]
	// It needs two (2 I()) transpositions with axes:
	// [1,0,2] then: [1,0]
	// T(1,0,2).Index(i).T(1,0).Index(j)
	shift := 0
	dims := len(tensor.shape)
	axes := make([]uint, dims)
	for j := range axes {
		axes[j] = uint(j)
	}
	for i, idx := range indices {
		if idx.start == idx.end {
			shifted_axes := make([]uint, dims-shift)
			copy(shifted_axes, axes[:dims-shift])
			prev := shifted_axes[0]
			shifted_axes[0] = shifted_axes[i-shift]

			for j := 1; j < dims-shift; j++ {
				shifted_axes[j], prev = prev, shifted_axes[j]
				if i-shift == j {
					break
				}
			}
			tensor = tensor.TrC(shifted_axes...).Index(idx.start)
			shift += 1
		}
	}
	return tensor
}

// DATA LAYOUT (move to other file?)

// Check if dimensions order is not shuffled and data layout is contiguous
func (tensor *Tensor[T]) IsContiguous() bool {
	dimOrder := tensor.dim_order
	switch len(dimOrder) {
	case 1:
		return true
	case 2:
		return dimOrder[0] == 0
	default:
		var min uint16 = 0
		var dim uint16
		for _, dim = range dimOrder {
			if dim > min {
				return false
			}
			min += 1
		}
		return true
	}
}

// reorders data layout to contiguous format.
// it is useful for optimizing indexing/iterating for transposed & other non-contiguous tensors
func (tensor *Tensor[T]) AsContiguous() *Tensor[T] {
	if tensor.Err != nil {
		return tensor
	}
	if tensor.IsContiguous() {
		return tensor
	}
	outTensor := CreateEmptyTensor[T](tensor.shape...)
	// for 2 dim tensor
	if len(tensor.shape) == 2 {
		// make matrix contiguous
		internal.TraverseAsContiguous2D(tensor.data(), outTensor.data(), tensor.shape)
		return outTensor
	}
	// for N Dim tensor
	internal.TraverseAsContiguousND[T](tensor.data(), outTensor.data(), tensor.strides, tensor.shape)
	return outTensor
}
