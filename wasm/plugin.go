//go:build tinygo.wasm

// This file is designed to be imported by plugins.

package wasm

// #include <stdlib.h>
import "C"

import (
	"reflect"
	"unsafe"
)

func PtrToByte(ptr, size uint32) []byte {
	var b []byte
	s := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	s.Len = uintptr(size)
	s.Cap = uintptr(size)
	s.Data = uintptr(ptr)
	return b
}

func ByteToPtr(buf []byte) (uint32, uint32) {
	if len(buf) == 0 {
		return 0, 0
	}

	size := C.ulong(len(buf))
	ptr := unsafe.Pointer(C.malloc(size))

	copy(unsafe.Slice((*byte)(ptr), size), buf)

	return uint32(uintptr(ptr)), uint32(len(buf))
}

func FreePtr(ptr uint32) {
	C.free(unsafe.Pointer(uintptr(ptr)))
}
