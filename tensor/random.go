package tensor

import (
	"math/rand"
	"time"
)

func createRand(seed int64) *rand.Rand {
	// use -1 for non deterministic rand
	if seed == -1 {
		seed = time.Now().UnixNano()
	}
	randSource := rand.NewSource(seed)
	return rand.New(randSource)
}

func CreateRandFloat32Slice(length int, seed int64) []float32 {
	_rand := createRand(seed)
	slice := make([]float32, length)
	for i := range slice {
		slice[i] = _rand.Float32()
	}
	return slice
}

func CreateRandFloat64Slice(length int, seed int64) []float64 {
	_rand := createRand(seed)
	slice := make([]float64, length)
	for i := range slice {
		slice[i] = _rand.Float64()
	}
	return slice
}

func RandomFloat64Tensor(shape Shape, seed int64) *Tensor[float64] {
	randTensor := InitEmptyTensor[float64](shape...)
	value := CreateRandFloat64Slice(len(randTensor.data), seed)
	randTensor.data = value
	return randTensor
}

func RandomFloat32Tensor(shape Shape, seed int64) *Tensor[float32] {
	randTensor := InitEmptyTensor[float32](shape...)
	value := CreateRandFloat32Slice(len(randTensor.data), seed)
	randTensor.data = value
	return randTensor
}
