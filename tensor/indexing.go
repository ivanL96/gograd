package tensor

import "fmt"

func (tensor *Tensor[T]) getFlatIndex(indices ...int) int {
	if len(indices) != len(tensor.shape) {
		panic(fmt.Sprintf(
			"Incorrect number of indices. Must be %v got %v", len(tensor.shape), len(indices)))
	}
	flatIndex := 0
	for i, ind := range indices {
		// bound check
		if ind >= int(tensor.shape[i]) {
			panic(fmt.Sprintf("Index %v is out of bounds for dim %v", ind, tensor.shape[i]))
		}
		flatIndex += tensor.strides[i] * ind
	}
	return flatIndex
}

func (tensor *Tensor[T]) Get(indices ...int) T {
	flatIndex := tensor.getFlatIndex(indices...)
	return tensor.data[flatIndex]
}

// returns sub data for given indices.
func (tensor *Tensor[T]) Index(indices ...int) *Tensor[T] {
	// TODO advanced indexing
	n_indices := len(indices)
	n_dims := len(tensor.shape)
	if n_indices == 0 {
		panic("At leat one index is required in View")
	}
	if n_indices > n_dims {
		panic("Too many indices")
	}

	flatIndex := 0 // index of the first elem in the sub tensor
	for i, ind := range indices {
		// resolve negative indexes
		if ind < 0 {
			ind = int(tensor.shape[i]) + ind
		}
		if ind < 0 {
			panic(fmt.Sprintf("Index %v is out of bounds", ind))
		}
		flatIndex += tensor.strides[i] * ind
	}
	if n_indices == n_dims {
		return InitTensor([]T{tensor.data[flatIndex]}, Shape{1})
	}
	innerShape := tensor.shape[n_indices:]
	var innerShapeProd Dim = 1
	for _, dim := range innerShape {
		innerShapeProd *= dim
	}

	// tensor with shape (1,1,...1)
	if innerShapeProd == 1 {
		return InitTensor([]T{tensor.data[0]}, innerShape)
	}

	// continuous data
	// if data layout is continuous we can just take a slice start:end from data
	if isDimOrderInit(tensor.dim_order) {
		endFlatIndex := flatIndex + tensor.strides[n_indices-1]
		subData := tensor.data[flatIndex:endFlatIndex]
		return InitTensor(subData, innerShape)
	}

	// not continuous data. i.e. transposed tensor

	innerStrides := tensor.strides[n_indices:]
	subShape := innerShape
	// expand innerShape
	// TODO this is extra step, better to do something with the loop
	if len(innerShape) == 1 {
		innerShape = Shape{1, innerShape[0]}
		innerStrides = []int{innerStrides[0], innerStrides[0]}
	}

	// prealloc output
	subData := make([]T, innerShapeProd)
	innermostStride := tensor.strides[len(tensor.strides)-1]
	row := int(innerShape[len(innerShape)-1]) // innermost axis
	//number of dims around the 'row'. Cannot be zero
	numDims := len(innerStrides) - 2
	for i := numDims; i >= 0; i-- {
		stride := innerStrides[i]
		subDataIdx := 0
		for s := 0; s < int(innerShape[i]); s++ {
			for j := 0; j < row; j++ {
				// from innermost to outermost
				deepIndex := flatIndex + innermostStride*j + stride*s
				subData[subDataIdx] = tensor.data[deepIndex]
				subDataIdx++
			}
		}
	}
	return InitTensor(subData, subShape)
}

type TensorIterator[T TensorType] struct {
	tensor         *Tensor[T]
	currentIndexes []int
	index          int
}

func (tensor *Tensor[T]) CreateIterator() *TensorIterator[T] {
	ti := TensorIterator[T]{
		tensor:         tensor,
		currentIndexes: make([]int, len(tensor.shape)),
		index:          0,
	}
	return &ti
}

func (ti *TensorIterator[T]) Index() int {
	return ti.index
}

func (ti *TensorIterator[T]) Iterate() bool {
	return ti.index != len(ti.tensor.Data())
}

func (ti *TensorIterator[T]) Next() []int {
	if ti.index == 0 {
		ti.index++
		return ti.currentIndexes
	}

	indices := ti.currentIndexes
	for j := len(indices) - 1; j >= 0; j-- {
		indices[j]++
		if indices[j] < int(ti.tensor.shape[j]) {
			break
		}
		indices[j] = 0
	}
	ti.currentIndexes = indices
	ti.index++
	return ti.currentIndexes
}

func (tensor *Tensor[T]) AsContinuous() *Tensor[T] {
	if isDimOrderInit(tensor.dim_order) {
		return tensor
	}
	outTensor := InitEmptyTensor[T](tensor.shape...)

	iter := tensor.CreateIterator()
	for iter.Iterate() {
		dataIndex := iter.Index()
		valueIndexes := iter.Next()
		val := tensor.Get(valueIndexes...)
		outTensor.data[dataIndex] = val
	}
	return outTensor
}
