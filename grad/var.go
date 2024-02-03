package grad

import (
	"fmt"
	"gograd/tensor"
	"gograd/tensor/types"
	"reflect"
)

var intKinds map[reflect.Kind]bool = map[reflect.Kind]bool{
	reflect.Uint:   true,
	reflect.Uint8:  true,
	reflect.Uint16: true,
	reflect.Uint32: true,
	reflect.Uint64: true,
	reflect.Int:    true,
	reflect.Int8:   true,
	reflect.Int16:  true,
	reflect.Int32:  true,
	reflect.Int64:  true,
}

type variable[T types.TensorType] struct {
	Value         *tensor.Tensor[T]
	Grad          *tensor.Tensor[T]
	backward_fn   func() *tensor.Tensor[T]
	Alias         string
	Children      []*variable[T]
	Requires_grad bool
}

func Variable[T types.TensorType](
	tensor_val *tensor.Tensor[T],
	children ...*variable[T],
) *variable[T] {
	tensor_val.MustAssert()
	v := &variable[T]{
		Value:         tensor_val,
		Grad:          tensor.Zeros[T](tensor_val.Shape()...),
		Children:      children,
		Requires_grad: true,
	}
	if v.Requires_grad && intKinds[tensor_val.DType().Kind()] {
		panic("Cannot create variable of Int type that requires gradient.")
	}
	return v
}

func VarFrom[T types.TensorType](data []T, shape types.Shape) *variable[T] {
	return Variable(tensor.CreateTensor[T](data, shape))
}

func Constant[T types.TensorType](
	tensor_val *tensor.Tensor[T],
	children ...*variable[T],
) *variable[T] {
	v := Variable[T](tensor_val, children...)
	v.Requires_grad = false
	return v
}

func (v *variable[T]) MustAssert() *variable[T] {
	v.Value.MustAssert()
	return v
}

func (v *variable[T]) ZeroGrad() {
	v.Grad = tensor.Zeros[T](v.Value.Shape()...)
}

func (v *variable[T]) ToString() string {
	name := "Const"
	if v.Requires_grad {
		name = "variable"
	}
	return fmt.Sprintf("%v(%v)", name, v.Value.ToString())
}

// VARIABLE OPS

// reduce gradient shape
func unbroadcast[T types.TensorType](
	grad,
	other *tensor.Tensor[T],
) *tensor.Tensor[T] {
	if !grad.Shape().Equals(other.Shape()) {
		return grad.SumAlongAxis(0, true).Reshape(other.Shape()...)
	}
	return grad
}

func (this *variable[T]) Add(other *variable[T]) *variable[T] {
	out := Variable(this.Value.Add(other.Value), this, other)
	out.Alias = "Add"
	if this.Requires_grad {
		this.backward_fn = func() *tensor.Tensor[T] {
			return unbroadcast(out.Grad, this.Value)
		}
	}
	if other.Requires_grad {
		other.backward_fn = func() *tensor.Tensor[T] {
			return unbroadcast(out.Grad, other.Value)
		}
	}
	return out
}

func (this *variable[T]) Sub(other *variable[T]) *variable[T] {
	out := Variable(this.Value.Sub(other.Value), this, other)
	if this.Requires_grad {
		this.backward_fn = func() *tensor.Tensor[T] {
			// out.g
			return unbroadcast(out.Grad, this.Value)
		}
	}
	if other.Requires_grad {
		other.backward_fn = func() *tensor.Tensor[T] {
			// -out.g
			return unbroadcast(out.Grad.Neg(), other.Value)
		}
	}
	return out
}

func (this *variable[T]) Mul(other *variable[T]) *variable[T] {
	out := Variable(this.Value.Mul(other.Value), this, other)
	if this.Requires_grad {
		this.backward_fn = func() *tensor.Tensor[T] {
			grad := other.Value.Mul(out.Grad) // other * out.g
			return unbroadcast(grad, this.Value)
		}
	}
	if other.Requires_grad {
		other.backward_fn = func() *tensor.Tensor[T] {
			grad := this.Value.Mul(out.Grad) // this * out.g
			return unbroadcast(grad, other.Value)
		}
	}
	return out
}

func (this *variable[T]) Pow(other *variable[T]) *variable[T] {
	out := Variable(this.Value.Pow(other.Value), this, other)
	if this.Requires_grad {
		this.backward_fn = func() *tensor.Tensor[T] {
			// out.g * other * this**(other-1)
			// or out.g * other * out / this ),
			grad := out.Grad.Mul(other.Value.Mul(out.Value.Div(this.Value)))
			return unbroadcast(grad, this.Value)
		}
	}
	if other.Requires_grad {
		other.backward_fn = func() *tensor.Tensor[T] {
			// out.g * out * this.ln()
			grad := out.Grad.Mul(out.Value.Mul(this.Value.Ln()))
			return unbroadcast(grad, other.Value)
		}
	}
	return out
}

// this/other
// => d(this): 1/other
// => d(other): (-this) / (other**2)
func (this *variable[T]) Div(other *variable[T]) *variable[T] {
	out := Variable(this.Value.Div(other.Value), this, other)
	if this.Requires_grad {
		// one := tensor.Scalar[T](1)
		this.backward_fn = func() *tensor.Tensor[T] {
			grad := out.Grad.Div(other.Value) // this.g += out.g / other.val
			return unbroadcast(grad, this.Value)

		}
	}
	if other.Requires_grad {
		other.backward_fn = func() *tensor.Tensor[T] {
			grad := this.Value.Neg().Div(other.Value.Mul(other.Value)) // other.g += -this.val * other.
			return unbroadcast(grad, other.Value)
		}
	}
	return out
}

func (this *variable[T]) MatMul(other *variable[T]) *variable[T] {
	out := Variable(this.Value.MatMul(other.Value), this, other)
	out.Alias = "MatMul"
	if this.Requires_grad {
		this.backward_fn = func() *tensor.Tensor[T] {
			// out.g @ other.T
			return out.Grad.MatMul(other.Value.T())
		}
	}
	if other.Requires_grad {
		other.backward_fn = func() *tensor.Tensor[T] {
			// this.T @ out.g
			return this.Value.TrC().MatMul(out.Grad)
		}
	}
	return out
}

// activations
func (this *variable[T]) Sigmoid() *variable[T] {
	out := Variable(this.Value.Sigmoid(), this)
	if this.Requires_grad {
		this.backward_fn = func() *tensor.Tensor[T] {
			one := tensor.Ones[T](out.Value.Shape()...)
			return out.Value.Mul(one.Sub(out.Value))
		}
	}
	return out
}

func (this *variable[T]) Relu() *variable[T] {
	out := Variable(this.Value.Relu(), this)
	if this.Requires_grad {
		this.backward_fn = func() *tensor.Tensor[T] {
			expr := func(a T) T {
				if a > 0 {
					return 1
				}
				return 0
			}
			return out.Grad.Mul(out.Value.ApplyFunc(expr))
		}
	}
	return out
}

func (this *variable[T]) Softmax() *variable[T] {
	n := types.Dim(len(this.Value.Shape()))

	e := this.Value.Exp()
	se := e.Sum(false)
	_softmax := Variable(e.Div(se), this)
	if this.Requires_grad {
		this.backward_fn = func() *tensor.Tensor[T] {
			ds := _softmax.Value.Mul(tensor.Eye[T](n, n).Sub(_softmax.Value.Copy().Unsqueeze(1)))
			return _softmax.Grad.Mul(ds)
		}
	}
	return _softmax
}

// reduce
func (this *variable[T]) Mean() *variable[T] {
	out := Variable(this.Value.Mean(false), this)
	if this.Requires_grad {
		this.backward_fn = func() *tensor.Tensor[T] {
			filler := tensor.Scalar(T(1. / float32(this.Value.Size())))
			return out.Grad.Mul(filler)
		}
	}
	return out
}
