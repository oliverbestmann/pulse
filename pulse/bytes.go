package pulse

import "unsafe"

func AsByteSlice[T any](value *T) []byte {
	var zeroT T

	n := unsafe.Sizeof(zeroT)
	ptr := (*byte)(unsafe.Pointer(value))

	return unsafe.Slice(ptr, n)
}
