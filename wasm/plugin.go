//go:build tinygo.wasm

// This file is designed to be imported by plugins.

package wasm

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
	ptr := &buf[0]
	unsafePtr := uintptr(unsafe.Pointer(ptr))
	return uint32(unsafePtr), uint32(len(buf))
}
