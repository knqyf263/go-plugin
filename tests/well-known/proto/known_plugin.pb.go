//go:build tinygo.wasm

// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v3.21.12
// source: tests/well-known/proto/known.proto

package proto

import (
	context "context"
	emptypb "github.com/knqyf263/go-plugin/types/known/emptypb"
	wasm "github.com/knqyf263/go-plugin/wasm"
)

const KnownTypesTestPluginAPIVersion = 1

//export known_types_test_api_version
func _known_types_test_api_version() uint64 {
	return KnownTypesTestPluginAPIVersion
}

var knownTypesTest KnownTypesTest

func RegisterKnownTypesTest(p KnownTypesTest) {
	knownTypesTest = p
}

//export known_types_test_test
func _known_types_test_test(ptr, size uint32) uint64 {
	b := wasm.PtrToByte(ptr, size)
	var req Request
	if err := req.UnmarshalVT(b); err != nil {
		return 0
	}
	response, err := knownTypesTest.Test(context.Background(), req)
	if err != nil {
		return 0
	}

	b, err = response.MarshalVT()
	if err != nil {
		return 0
	}
	ptr, size = wasm.ByteToPtr(b)
	return (uint64(ptr) << uint64(32)) | uint64(size)
}

const EmptyTestPluginAPIVersion = 1

//export empty_test_api_version
func _empty_test_api_version() uint64 {
	return EmptyTestPluginAPIVersion
}

var emptyTest EmptyTest

func RegisterEmptyTest(p EmptyTest) {
	emptyTest = p
}

//export empty_test_do_nothing
func _empty_test_do_nothing(ptr, size uint32) uint64 {
	b := wasm.PtrToByte(ptr, size)
	var req emptypb.Empty
	if err := req.UnmarshalVT(b); err != nil {
		return 0
	}
	response, err := emptyTest.DoNothing(context.Background(), req)
	if err != nil {
		return 0
	}

	b, err = response.MarshalVT()
	if err != nil {
		return 0
	}
	ptr, size = wasm.ByteToPtr(b)
	return (uint64(ptr) << uint64(32)) | uint64(size)
}
