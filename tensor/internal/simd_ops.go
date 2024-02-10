package internal

import (
	"gograd/tensor/types"
	"math"
	"runtime"
	"sync"
)

func makeOutMat[T types.TensorType](out []T, size int) []T {
	if out == nil {
		return make([]T, size)
	}
	return out
}

func _range[T types.TensorType](slc []T, start, end int) []T {
	if len(slc) == 1 {
		return slc
	}
	if end >= len(slc) {
		return slc[start:]
	}
	return slc[start:end]
}

// here's the logic of elementwise addition between matrices.
// The "impl" argument can contain an implementation to accelerate inner loop using avx,etc
func RunSimdImpl[T types.TensorType](a, b, out []T, impl func([]T, []T, []T)) {
	chunk := func(start, end int, a, b, out []T, mu *sync.Mutex) {
		if start >= end {
			return
		}
		impl(_range(a, start, end), _range(b, start, end), _range(out, start, end))
	}
	parallel(chunk, a, b, makeOutMat(out, len(a)))
}

func RunSimdImplUnary[T types.TensorType](a, out []T, impl func([]T, []T)) {
	chunk := func(start, end int, a, dummy, out []T, mu *sync.Mutex) {
		if start >= end {
			return
		}
		impl(_range(a, start, end), _range(out, start, end))
	}
	parallel(chunk, a, []T{}, makeOutMat(out, len(a)))
}

func Dot[T types.TensorType](a, b []float32) float32 {
	var c float32
	for i := 0; i < len(a); i++ {
		c += a[i] * b[i]
	}
	return c
}

func PowMatx[T types.TensorType](a, b, out []T) {
	pow_chunk := func(start, end int, a, b, out []T, mu *sync.Mutex) {
		if len(a) == 1 {
			for i := start; i < end; i++ {
				out[i] = T(math.Pow(float64(a[0]), float64(b[i])))
			}
		} else if len(b) == 1 {
			for i := start; i < end; i++ {
				out[i] = T(math.Pow(float64(a[i]), float64(b[0])))
			}
		} else {
			for i := start; i < end; i++ {
				out[i] = T(math.Pow(float64(a[i]), float64(b[i])))
			}
		}
	}
	parallel(pow_chunk, a, b, makeOutMat(out, len(a)))
}

// unary
func SigmoidMatx[T types.TensorType](a, out []T) {
	sigm_chunk := func(start, end int, a, dummy, out []T, mu *sync.Mutex) {
		for i := start; i < end; i++ {
			out[i] = T(1. / (1. + math.Pow(math.E, float64(-a[i]))))
		}
	}
	parallel(sigm_chunk, a, nil, makeOutMat(out, len(a)))
}

func NegMatx[T types.TensorType](a, out []T) {
	neg_chunk := func(start, end int, a, dummy, out []T, mu *sync.Mutex) {
		for i := start; i < end; i++ {
			out[i] = -a[i]
		}
	}
	parallel(neg_chunk, a, nil, makeOutMat(out, len(a)))
}

func LnNegMatx[T types.TensorType](a, out []T) {
	lnneg_chunk := func(start, end int, a, dummy, out []T, mu *sync.Mutex) {
		for i := start; i < end; i++ {
			out[i] = -T(math.Log(float64(a[i])))
		}
	}
	parallel(lnneg_chunk, a, nil, makeOutMat(out, len(a)))
}

func ReluMatx[T types.TensorType](a, out []T) {
	relu_chunk := func(start, end int, a, dummy, out []T, mu *sync.Mutex) {
		for i := start; i < end; i++ {
			el := a[i]
			if el > 0 {
				out[i] = el
				continue
			}
			out[i] = 0
		}
	}
	parallel(relu_chunk, a, nil, makeOutMat(out, len(a)))
}

func ExpMatx[T types.TensorType](a, out []T) {
	exp_chunk := func(start, end int, a, dummy, out []T, mu *sync.Mutex) {
		for i := start; i < end; i++ {
			out[i] = T(math.Exp(float64(a[i])))
		}
	}
	parallel(exp_chunk, a, nil, makeOutMat(out, len(a)))
}

func ApplyFuncMatx[T types.TensorType](a []T, expr_fn func(T) T, out []T) {
	_chunk := func(start, end int, a, dummy, out []T, mu *sync.Mutex) {
		for i := start; i < end; i++ {
			out[i] = expr_fn(a[i])
		}
	}
	parallel(_chunk, a, nil, makeOutMat(out, len(a)))
}

// reduce
func SumMatx[T types.TensorType](a, out []T) {
	sum_chunk := func(start, end int, a, dummy, out []T, mu *sync.Mutex) {
		var chunk_sum T = 0
		for i := start; i < end; i++ {
			chunk_sum += a[i]
		}
		mu.Lock()
		defer mu.Unlock()
		out[0] += chunk_sum
	}
	parallel(sum_chunk, a, nil, makeOutMat(out, len(a)))
}

func SumAxisMatx[T types.TensorType](data, out []T, shape types.Shape, axis int) {
	inner_step := int(shape[axis])
	outer_step := len(data) / inner_step

	var stride_0 int = int(shape[1])
	var stride_1 int = 1
	if axis == 1 {
		stride_0 = 1
		stride_1 = int(shape[1])
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < outer_step; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			inner_sum := T(0)
			i_stride := i * stride_1
			for j := 0; j < inner_step; j++ {
				inner_sum += data[j*stride_0+i_stride]
			}
			mu.Lock()
			defer mu.Unlock()
			out[i] += inner_sum
		}(i)
	}
	wg.Wait()
}

func MaxMatx[T types.TensorType](a, out []T) {
	max_chunk := func(start, end int, a, dummy, out []T, mu *sync.Mutex) {
		var _max T = a[0]
		for i := start; i < end; i++ {
			v := a[i]
			if v > _max {
				_max = v
			}
		}
		mu.Lock()
		defer mu.Unlock()
		if _max > out[0] {
			out[0] = _max
		}
	}
	parallel(max_chunk, a, nil, makeOutMat(out, len(a)))
}

func MinMatx[T types.TensorType](a, out []T) {
	min_chunk := func(start, end int, a, dummy, out []T, mu *sync.Mutex) {
		var _min T = a[0]
		for i := start; i < end; i++ {
			v := a[i]
			if v < _min {
				_min = v
			}
		}
		mu.Lock()
		defer mu.Unlock()
		if _min < out[0] {
			out[0] = _min
		}
	}
	parallel(min_chunk, a, nil, makeOutMat(out, len(a)))
}

func SoftmaxMatx[T types.TensorType](a, out []T, strides []int) {
	var wg sync.WaitGroup

	c := strides[0]
	batchsize := len(a) / c
	for i := 0; i < batchsize; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			logits_start := c * i
			logits_end := (i + 1) * c
			logit_max := a[logits_start]
			for j := logits_start; j < logits_end; j++ {
				v := a[j]
				if v > logit_max {
					logit_max = v
				}
			}
			var logits_sum T = 0
			for j := logits_start; j < logits_end; j++ {
				exp := T(math.Exp(float64(a[j] - logit_max)))
				out[j] = exp
				logits_sum += exp
			}
			for j := logits_start; j < logits_end; j++ {
				out[j] /= logits_sum
			}
		}(i)
	}
	wg.Wait()
}

// matmul
func MatMulMatx(
	a_data, b_data, out_data []float32,
	a_shape, b_shape types.Shape,
	a_strides, b_strides, out_strides []int,
	dot_impl func([]float32, []float32) float32,
) {
	a_dim0 := int(a_shape[0])
	b_dim0 := int(b_shape[0])
	out_stride0 := out_strides[0]
	a_stride0 := a_strides[0]
	b_stride0 := b_strides[0]

	block_size := 64

	runtime.GOMAXPROCS(numCPU)

	var wg sync.WaitGroup
	// wg.Add(numCPU)
	// chunk_size := (a_dim0 + numCPU - 1) / numCPU

	// for ii := 0; ii < numCPU; ii++ {
	// 	start := ii * chunk_size
	// 	end := (ii + 1) * chunk_size
	// 	if end > a_dim0 {
	// 		end = a_dim0
	// 	}

	// 	go func(start, end int) {
	// 		defer wg.Done()
	for i := 0; i < a_dim0; i += block_size {
		// for i := start; i < end; i += block_size {
		for j := 0; j < b_dim0; j += block_size {
			wg.Add(1)

			// runtime.LockOSThread()
			// defer runtime.UnlockOSThread()

			go func(i, j int) {
				defer wg.Done()
				for bi := 0; bi < block_size; bi++ {
					for bj := 0; bj < block_size; bj++ {
						row := i + bi
						col := j + bj
						if row >= a_dim0 || col >= b_dim0 {
							continue
						}
						out_data[out_stride0*row+col] = dot_impl(
							a_data[a_stride0*row:a_stride0*(row+1)],
							b_data[b_stride0*col:b_stride0*(col+1)],
						)
					}
				}
			}(i, j)
		}
	} //(start, end)
	// }
	wg.Wait()
}