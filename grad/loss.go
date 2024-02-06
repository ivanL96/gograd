package grad

import (
	"gograd/tensor"
)

// MSE impl. mean((y_true-y_pred)**2)
//
// 'y_true' is a cosnt by definition
func (y_pred *Var[T]) MSE(y_true *Var[T]) *Var[T] {
	squared := tensor.Scalar[T](2)
	mean := y_true.Value.Sub(y_pred.Value).Pow(squared).Mean(false)
	out := Variable(mean, y_pred)
	out.Alias = "MSE"
	if y_pred.Requires_grad {
		y_pred.backward_fn = func() *tensor.Tensor[T] {
			n := tensor.Scalar[T](T(len(y_true.Value.Data())))
			_const := tensor.Scalar[T](2).Div(n).Neg()
			return out.Grad.Mul(_const.Mul(y_true.Value.Sub(y_pred.Value)))
		}
	}
	return out
}

func (logits *Var[T]) SoftmaxCrossEntropy(y_true *Var[T]) *Var[T] {

	y_pred := logits.Value.Softmax(nil)

	var epsilon float32 = 1e-15
	cross_entropy := y_pred.IndexMask(y_true.Value, true).Clip(epsilon, 1-epsilon).LnNeg()
	out := Variable(cross_entropy, logits).SetAlias("SoftmaxCrossEntropy")

	if logits.Requires_grad {
		n_classes := uint(logits.Value.Shape()[1])
		y_onehot := y_true.ToOneHot(n_classes)
		logits.backward_fn = func() *tensor.Tensor[T] {
			return out.Grad.Mul(y_pred.Sub(y_onehot.Value))
		}
	}
	return out
}
