package amd64

import "unsafe"

// implemented in /intrinsics/amd64/dot_avx256.s
//go:noescape
func _mm256_dot(a, b, n, ret unsafe.Pointer)

//go:noescape
func _mm256_mul_to(a, b, c, n unsafe.Pointer)
