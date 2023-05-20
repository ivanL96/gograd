package tensor

import (
	"fmt"
	"reflect"
)

func Log(value ...interface{}) {
	str := fmt.Sprintf("%x", value)
	fmt.Println("Debug: " + str)
}

func squeeze_shape(shape Shape) Shape {
	result := Shape{1}
	for _, v := range shape {
		if v > 1 {
			result = append(result, v)
		}
	}
	return result
}

func isScalarLike(shape Shape) bool {
	if len(shape) <= 1 && shape[0] <= 1 {
		return true
	}
	return false
}

func getStrides(shape Shape) []int {
	strides := make([]int, len(shape))
	stride := 1
	for i := len(shape) - 1; i >= 0; i-- {
		strides[i] = stride
		stride *= int(shape[i])
	}
	return strides
}

func initDimOrder(shape Shape) []int {
	dimOrder := make([]int, len(shape))
	for i := range dimOrder {
		dimOrder[i] = i
	}
	return dimOrder
}

func isDimOrderInit(dimOrder []int) bool {
	min := 0
	for _, dim := range dimOrder {
		if dim > min {
			return false
		}
		min = dim
	}
	return true
}

func isIntKind(tensorDType reflect.Type) bool {
	switch tensorDType.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}
