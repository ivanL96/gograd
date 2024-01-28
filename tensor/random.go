package tensor

import (
	types "gograd/tensor/types"
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

func createRandFloat32Slice(length int, seed int64) []float32 {
	_rand := createRand(seed)
	slice := make([]float32, length)
	for i := range slice {
		slice[i] = _rand.Float32()
	}
	return slice
}

func createRandFloat64Slice(length int, seed int64) []float64 {
	_rand := createRand(seed)
	slice := make([]float64, length)
	for i := range slice {
		slice[i] = _rand.Float64()
	}
	return slice
}

type RNG struct {
	Seed int64
}

func (rng *RNG) RandomFloat64(shape ...types.Dim) *Tensor[float64] {
	randTensor := CreateEmptyTensor[float64](shape...)
	value := createRandFloat64Slice(len(randTensor.data()), rng.Seed)
	randTensor.SetData(value)
	return randTensor
}

func (rng *RNG) RandomFloat32(shape ...types.Dim) *Tensor[float32] {
	randTensor := CreateEmptyTensor[float32](shape...)
	value := createRandFloat32Slice(len(randTensor.data()), rng.Seed)
	randTensor.SetData(value)
	return randTensor
}
