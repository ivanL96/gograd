package tensor

import (
	"fmt"
	ops "gograd/tensor/ops"
	types "gograd/tensor/types"
	"math"
)

type BinaryScalarOp[T types.TensorType] func(T, T) T
type UnaryScalarOp[T types.TensorType] func(T) T

func BaseBinElementwiseOp[T types.TensorType](
	tensor_a,
	tensor_b *Tensor[T],
	binOp BinaryScalarOp[T],
	out *Tensor[T],
) *Tensor[T] {
	var outTensor *Tensor[T]
	if IsScalarLike(tensor_a.shape) && IsScalarLike(tensor_b.shape) {
		outTensor = PrepareOutTensor(out, tensor_a.shape)
		// most trivial case (1,) & (1,)
		outTensor.data[0] = binOp(tensor_a.data[0], tensor_b.data[0])
		return outTensor
	}

	// tensors should have equal shapes or at least one of them should be scalar-like
	if !AreBroadcastable(tensor_a.Shape(), tensor_b.Shape()) {
		panic(
			fmt.Sprintf("Shapes: %x, %x are not broadcastable", tensor_a.Shape(), tensor_b.Shape()),
		)
	}
	if len(tensor_a.data) == len(tensor_b.data) {
		// same broadcastable shapes (N,M) & (N,M)
		outTensor = PrepareOutTensor(out, tensor_a.shape)

		// if two tensors are filled with same values. For example [2,2,2] and [3,3,3]
		if tensor_a.hasFlag(SameValuesFlag) && tensor_b.hasFlag(SameValuesFlag) {
			outTensor.Fill(binOp(tensor_a.data[0], tensor_b.data[0]))
			return outTensor
		}

		iter := tensor_a.CreateIterator()
		for iter.Iterate() {
			dataIndex := iter.Index()
			idx := iter.Next()
			outTensor.data[dataIndex] = binOp(tensor_a.Get_fast(idx...), tensor_b.Get_fast(idx...))
		}
	} else if len(tensor_b.data) == 1 {
		// tensor_b is scalar
		// (N, M, ...) & (1,)
		outTensor = PrepareOutTensor(out, tensor_a.shape)
		value := tensor_b.data[0]
		for i, val := range tensor_a.data {
			outTensor.data[i] = binOp(val, value)
		}
	} else if len(tensor_a.data) == 1 {
		// tensor_a is scalar
		// (1,) & (N, M, ...)
		outTensor = PrepareOutTensor(out, tensor_b.shape)
		value := tensor_a.data[0]
		for i, val := range tensor_b.data {
			outTensor.data[i] = binOp(val, value)
		}
	} else {
		// (A, B ...) & (N, M, ...)
		// apply operation for non scalar broadcastable tensors
		broadcasted_shape, _ := BroadcastShapes(tensor_a.Shape(), tensor_b.Shape())

		// determine which tensor (at least 1) should be broadcasted
		var broadcasted_tensor_a *Tensor[T] = nil
		var broadcasted_tensor_b *Tensor[T] = nil
		for i, brc_dim := range broadcasted_shape {
			if i < len(tensor_a.Shape()) {
				dim_a := tensor_a.Shape()[i]
				if dim_a < brc_dim && broadcasted_tensor_a == nil {
					// need to broadcast tensor_a
					// TODO minor. how to avoid additional Broadcast()
					broadcasted_tensor_a = tensor_a.Broadcast(broadcasted_shape...)
				}
			}
			if i < len(tensor_b.Shape()) {
				dim_b := tensor_b.Shape()[i]
				if dim_b < brc_dim && broadcasted_tensor_b == nil {
					// need to broadcast tensor_b
					broadcasted_tensor_b = tensor_b.Broadcast(broadcasted_shape...)
				}
			}
		}
		outTensor = PrepareOutTensor(out, broadcasted_shape)
		if broadcasted_tensor_a == nil {
			broadcasted_tensor_a = tensor_a
		}
		if broadcasted_tensor_b == nil {
			broadcasted_tensor_b = tensor_b
		}
		iter := broadcasted_tensor_a.CreateIterator()
		for iter.Iterate() {
			dataIndex := iter.Index()
			idx := iter.Next()
			outTensor.data[dataIndex] = binOp(
				broadcasted_tensor_a.Get(idx...), broadcasted_tensor_b.Get(idx...))
		}
	}
	outTensor.ResetFlags()
	return outTensor
}

func unaryElementwiseRoutine[T types.TensorType](
	tensor *Tensor[T],
	unaryOp UnaryScalarOp[T],
	out *Tensor[T],
) *Tensor[T] {
	outTensor := PrepareOutTensor(out, tensor.Shape())
	if IsScalarLike(tensor.Shape()) {
		outTensor.data[0] = unaryOp(tensor.data[0])
		return outTensor
	}
	for i, val := range tensor.data {
		outTensor.data[i] = unaryOp(val)
	}
	return outTensor
}

// binary ops
func _add[T types.TensorType](a, b T) T {
	return a + b
}

func (tensor *Tensor[T]) Add(other_tensor, out *Tensor[T]) *Tensor[T] {
	return BaseBinElementwiseOp(tensor, other_tensor, _add[T], out)
}

func _sub[T types.TensorType](a, b T) T {
	return a - b
}
func (tensor *Tensor[T]) Sub(other_tensor, out *Tensor[T]) *Tensor[T] {
	return BaseBinElementwiseOp(tensor, other_tensor, _sub[T], out)
}

func _mul[T types.TensorType](a, b T) T {
	return a * b
}
func (tensor *Tensor[T]) Mul(other_tensor, out *Tensor[T]) *Tensor[T] {
	return BaseBinElementwiseOp(tensor, other_tensor, _mul[T], out)
}

func _div[T types.TensorType](a, b T) T {
	return a / b
}
func (tensor *Tensor[T]) Div(other_tensor, out *Tensor[T]) *Tensor[T] {
	return BaseBinElementwiseOp(tensor, other_tensor, _div[T], out)
}

func _sigmoid[T types.TensorType](a T) T {
	return T(1. / (1. + math.Pow(math.E, float64(-a))))
}
func (tensor *Tensor[T]) Sigmoid(out *Tensor[T]) *Tensor[T] {
	return unaryElementwiseRoutine(tensor, _sigmoid[T], out)
}

// input A and B, both n by n matrices
// initialize C to be an n by n matrix of all zeros
// for i from 1 to n:
//     for j from 1 to n:
//         for k from 1 to n:
//             C[i][j] = C[i][j] + A[i][k]*B[k][j]
// output C (as A*B)

func (tensor *Tensor[T]) MatMul(other_tensor *Tensor[T]) *Tensor[T] {
	if len(tensor.shape) != 2 || len(other_tensor.shape) != 2 {
		panic("Tensors must be two-dim.")
	}
	if tensor.shape[1] != other_tensor.shape[0] {
		panic(fmt.Sprintf(
			"Tensors inner shapes are different. %v != %v", tensor.shape[1], other_tensor.shape[0],
		))
	}
	var a types.ITensor[T] = tensor.AsContinuous(nil)
	var b types.ITensor[T] = other_tensor.AsContinuous(nil)
	var outTensor types.ITensor[T] = InitEmptyTensor[T](a.Shape()[0], b.Shape()[1])
	ops.MatMulImplSimple(a, b, outTensor)
	return outTensor.(*Tensor[T])
}

// matmul ops
func SplitTensor[T types.TensorType](
	tensor, outA, outB, outC, outD *Tensor[T],
) (a, b, c, d *Tensor[T]) {
	if len(tensor.shape) != 2 {
		panic("Tensor must have (N,N) shape")
	}
	rows := int(tensor.shape[0])
	row2 := rows / 2
	sub_tensor_shape := types.Shape{types.Dim(row2), types.Dim(row2)}
	a = PrepareOutTensor(outA, sub_tensor_shape)
	b = PrepareOutTensor(outB, sub_tensor_shape)
	c = PrepareOutTensor(outC, sub_tensor_shape)
	d = PrepareOutTensor(outD, sub_tensor_shape)
	ops.SplitTensorImpl(tensor.data, rows, a.data, b.data, c.data, d.data)
	return
}

// func matMulStrassen[T types.TensorType](tensor_a, tensor_b, outTensor *Tensor[T]) {

// }
