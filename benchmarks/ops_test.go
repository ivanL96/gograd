package main

import (
	"gograd/tensor"
	"testing"
)

func BenchmarkAdd(b *testing.B) {
	a1 := tensor.CreateEmptyTensor[int32](100, 100).Fill(100)
	a2 := tensor.CreateEmptyTensor[int32](100, 100).Fill(99)
	for i := 0; i < b.N; i++ {
		a1.Add(a2, nil)
	}
}

// go test -benchmem -run=^$ -bench ^BenchmarkBigAdd$ gograd/benchmarks -benchmem -v -count=5
// goos: windows
// goarch: amd64
// pkg: gograd/benchmarks
// cpu: 11th Gen Intel(R) Core(TM) i5-11400H @ 2.70GHz
// scalar
// BenchmarkBigAdd-12           199           5888470 ns/op           60408 B/op          1 allocs/op
// vectorized
// BenchmarkBigAdd-12          1946            523187 ns/op            6175 B/op          0 allocs/op
// BenchmarkBigAdd-12          1957            529357 ns/op            6141 B/op          0 allocs/op
// vec + goroutines
// BenchmarkBigAdd-12          7246            158306 ns/op            3212 B/op         25 allocs/op
// BenchmarkBigAdd-12          6296            199704 ns/op            3460 B/op         25 allocs/op
// avx
// BenchmarkBigAdd-12          6844            171370 ns/op            3517 B/op         26 allocs/op
// BenchmarkBigAdd-12          5796            175902 ns/op            3833 B/op         26 allocs/op
// avx + inplace
// BenchmarkBigAdd-12          8274            141875 ns/op            2728 B/op         26 allocs/op
// BenchmarkBigAdd-12          7477            142808 ns/op            2831 B/op         26 allocs/op
// avx + goroutines
// BenchmarkBigAdd-12         10390            108.364 ns/op            2925 B/op         26 allocs/op
// BenchmarkBigAdd-12          9687            107.917 ns/op            3008 B/op         26 allocs/op
// BenchmarkBigAdd-12          9469            105.879 ns/op            3037 B/op         26 allocs/op

func BenchmarkBigAdd(b *testing.B) {
	a1 := tensor.Range[float32](1000000).Reshape(1000, 1000) //.Fill(1)
	a2 := tensor.Range[float32](1000000).Reshape(1000, 1000) //.Fill(1)
	out := tensor.CreateEmptyTensor[float32](1000, 1000)
	for i := 0; i < b.N; i++ {
		a1.Add(a2, out)
	}
	// fmt.Println(out.Get(999, 999))
}

// goos: windows
// goarch: amd64
// pkg: gograd/benchmarks
// cpu: 11th Gen Intel(R) Core(TM) i5-11400H @ 2.70GHz
// scalar
// BenchmarkBigMul-12           186           6.120.142 ns/op          107705 B/op          1 allocs/op
// BenchmarkBigMul-12           198           5.973.748 ns/op          101178 B/op          1 allocs/op
// BenchmarkBigMul-12           198           6.092.643 ns/op          101178 B/op          1 allocs/op
// BenchmarkBigMul-12           198           5.970.509 ns/op          101179 B/op          1 allocs/op
// BenchmarkBigMul-12           196           6.123.761 ns/op          102210 B/op          1 allocs/op
// vec + goroutines
// BenchmarkBigMul-12          1845            548.058 ns/op           10856 B/op          0 allocs/op
// BenchmarkBigMul-12          2286            514.924 ns/op            8762 B/op          0 allocs/op
// BenchmarkBigMul-12          2344            519.922 ns/op            8545 B/op          0 allocs/op
// AVX2 / AVX512 ~same speed
// BenchmarkBigMul-12          3577            343.849 ns/op            5599 B/op          0 allocs/op
// BenchmarkBigMul-12          3550            350.165 ns/op            5642 B/op          0 allocs/op
// BenchmarkBigMul-12          3417            343.890 ns/op            5861 B/op          0 allocs/op
// avx512 + goroutines combo
// BenchmarkBigMul-12         10958            106.339 ns/op            3596 B/op         26 allocs/op
// BenchmarkBigMul-12         11272            108.102 ns/op            3545 B/op         26 allocs/op
// BenchmarkBigMul-12         11263            106.109 ns/op            3547 B/op         26 allocs/op
// shape 1000x1000
func BenchmarkBigMul(b *testing.B) {
	rng := tensor.NewRNG(0)
	a1 := rng.RandomFloat32(1000, 1000)
	a2 := rng.RandomFloat32(1000, 1000)
	out := tensor.CreateEmptyTensor[float32](1000, 1000)
	for i := 0; i < b.N; i++ {
		a1.Mul(a2, out)
	}
	// fmt.Println(out.Get(999, 999))
}

// avx
// BenchmarkMulToConst-12              6510            171877 ns/op            1854 B/op          1 allocs/op
// BenchmarkMulToConst-12              6363            185830 ns/op            1897 B/op          1 allocs/op
// BenchmarkMulToConst-12              6712            179904 ns/op            1799 B/op          1 allocs/op
// BenchmarkMulToConst-12              6818            184437 ns/op            1771 B/op          1 allocs/op
// BenchmarkMulToConst-12              6506            180120 ns/op            1856 B/op          1 allocs/op
func BenchmarkMulToConst(b *testing.B) {
	rng := tensor.NewRNG(0)
	a1 := rng.RandomFloat32(1000, 1000)
	a2 := tensor.Scalar[float32](5)
	out := tensor.CreateEmptyTensor[float32](1000, 1000)
	for i := 0; i < b.N; i++ {
		a1.Mul(a2, out)
	}
	// fmt.Println(out.Get(999, 999))
}

func BenchmarkMulScalar(b *testing.B) {
	a1 := tensor.CreateEmptyTensor[int32](1).Fill(100)
	a2 := tensor.CreateEmptyTensor[int32](1).Fill(99)
	for i := 0; i < b.N; i++ {
		a1.Mul(a2, nil)
	}
}

// goos: windows
// goarch: amd64
// pkg: gograd/benchmarks
// cpu: 11th Gen Intel(R) Core(TM) i5-11400H @ 2.70GHz
// atomic impl
// BenchmarkSigmoid-12           16          62.976.994 ns/op         4507318 B/op          6 allocs/op
// BenchmarkSigmoid-12           19          61.038.279 ns/op         4427986 B/op          5 allocs/op
// BenchmarkSigmoid-12           19          60.590.647 ns/op         4427981 B/op          5 allocs/op
// BenchmarkSigmoid-12           19          60.433.121 ns/op         4427981 B/op          5 allocs/op
// BenchmarkSigmoid-12           19          62.144.426 ns/op         4427981 B/op          5 allocs/op
// vector + goroutines
// BenchmarkSigmoid-12          139           8.282.268 ns/op         4065871 B/op         32 allocs/op
// BenchmarkSigmoid-12          150           7.802.250 ns/op         4061361 B/op         31 allocs/op
// BenchmarkSigmoid-12          147           7.908.745 ns/op         4062321 B/op         31 allocs/op
// BenchmarkSigmoid-12          151           7.952.143 ns/op         4060873 B/op         31 allocs/op
// BenchmarkSigmoid-12          150           7.858.647 ns/op         4061227 B/op         31 allocs/op
// numpy sigmoid 							 10.001.420 ns/op
func BenchmarkSigmoid(b *testing.B) {
	rng := tensor.NewRNG(0)
	a1 := rng.RandomFloat32(1000, 1000)
	for i := 0; i < b.N; i++ {
		a1.Sigmoid(nil)
	}
}

// goos: windows
// goarch: amd64
// pkg: gograd/benchmarks
// cpu: 11th Gen Intel(R) Core(TM) i5-11400H @ 2.70GHz
// scalar
// BenchmarkNeg-12              517           2.252.472 ns/op         4021526 B/op          5 allocs/op
// BenchmarkNeg-12              512           2.272.280 ns/op         4021676 B/op          5 allocs/op
// BenchmarkNeg-12              519           2.247.524 ns/op         4021467 B/op          5 allocs/op
// vec
// BenchmarkNeg-12             2515             466.966 ns/op         4010974 B/op         31 allocs/op
// BenchmarkNeg-12             2566             465.331 ns/op         4010908 B/op         31 allocs/op
// BenchmarkNeg-12             2486             477.069 ns/op         4011007 B/op         31 allocs/op
func BenchmarkNeg(b *testing.B) {
	rng := tensor.NewRNG(0)
	a1 := rng.RandomFloat32(1000, 1000)
	for i := 0; i < b.N; i++ {
		a1.Neg(nil)
	}
}

// goos: windows
// goarch: amd64
// pkg: gograd/benchmarks
// cpu: 11th Gen Intel(R) Core(TM) i5-11400H @ 2.70GHz
// scalar
// BenchmarkPow-12               20          55.058.685 ns/op         4807766 B/op          6 allocs/op
// BenchmarkPow-12               21          56.090.957 ns/op         4769588 B/op          6 allocs/op
// BenchmarkPow-12               21          55.177.438 ns/op         4769593 B/op          6 allocs/op
// vec
// BenchmarkPow-12              145           8.629.586 ns/op         4118926 B/op         32 allocs/op
// BenchmarkPow-12              146           8.994.234 ns/op         4117723 B/op         31 allocs/op
// BenchmarkPow-12              133           8.472.005 ns/op         4128340 B/op         31 allocs/op
// BenchmarkPow-12              145           7.985.862 ns/op         4118395 B/op         31 allocs/op
// BenchmarkPow-12              144           8.125.274 ns/op         4119241 B/op         31 allocs/op
func BenchmarkPow(b *testing.B) {
	rng := tensor.NewRNG(0)
	a1 := rng.RandomFloat32(1000, 1000)
	b1 := rng.RandomFloat32(1000, 1000)
	for i := 0; i < b.N; i++ {
		a1.Pow(b1)
	}
}

func BenchmarkDiv(b *testing.B) {
	rng := tensor.NewRNG(0)
	a1 := rng.RandomFloat32(1000, 1000)
	b1 := rng.RandomFloat32(1000, 1000)
	for i := 0; i < b.N; i++ {
		a1.Div(b1)
	}
}

// goos: windows
// goarch: amd64
// pkg: gograd/benchmarks
// cpu: 11th Gen Intel(R) Core(TM) i5-11400H @ 2.70GHz
//
// BenchmarkSub-12             6736            166.824 ns/op            4735 B/op         26 allocs/op
// BenchmarkSub-12             7200            166.884 ns/op            4543 B/op         26 allocs/op
// BenchmarkSub-12             6364            166.159 ns/op            4909 B/op         26 allocs/op
// inplace
// BenchmarkSub-12             7800            139.553 ns/op            3818 B/op         26 allocs/op
// BenchmarkSub-12             8596            138.748 ns/op            3625 B/op         26 allocs/op
// BenchmarkSub-12             8678            138.503 ns/op            3608 B/op         26 allocs/op
func BenchmarkSub(b *testing.B) {
	rng := tensor.NewRNG(0)
	a1 := rng.RandomFloat32(1000, 1000)
	b1 := rng.RandomFloat32(1000, 1000)
	// out := tensor.CreateEmptyTensor[float32](1000, 1000)
	for i := 0; i < b.N; i++ {
		a1.Sub(b1, a1)
	}
}

// 1000x1000
// scalar
// BenchmarkRelu-12             538           2.207.686 ns/op         4020922 B/op          5 allocs/op
// BenchmarkRelu-12             534           2.204.835 ns/op         4021032 B/op          5 allocs/op
// BenchmarkRelu-12             531           2.198.954 ns/op         4021116 B/op          5 allocs/op
// vector
// BenchmarkRelu-12            2304             496.967 ns/op         4011268 B/op         31 allocs/op
// BenchmarkRelu-12            2036             501.141 ns/op         4011721 B/op         31 allocs/op
// BenchmarkRelu-12            2418             489.242 ns/op         4011099 B/op         31 allocs/op
func BenchmarkRelu(b *testing.B) {
	rng := tensor.NewRNG(0)
	a1 := rng.RandomFloat32(1000, 1000)
	for i := 0; i < b.N; i++ {
		a1.Relu()
	}
}
