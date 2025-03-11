//go:build wasip1

// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v5.29.3
// source: examples/host-functions/greeting/greet.proto

package greeting

import (
	context "context"
	emptypb "github.com/knqyf263/go-plugin/types/known/emptypb"
	wasm "github.com/knqyf263/go-plugin/wasm"
	_ "unsafe"
)

const GreeterPluginAPIVersion = 1

//go:wasmexport greeter_api_version
func _greeter_api_version() uint64 {
	return GreeterPluginAPIVersion
}

var greeter Greeter

func RegisterGreeter(p Greeter) {
	greeter = p
}

//go:wasmexport greeter_greet
func _greeter_greet(ptr, size uint32) uint64 {
	b := wasm.PtrToByte(ptr, size)
	req := new(GreetRequest)
	if err := req.UnmarshalVT(b); err != nil {
		return 0
	}
	response, err := greeter.Greet(context.Background(), req)
	if err != nil {
		ptr, size = wasm.ByteToPtr([]byte(err.Error()))
		return (uint64(ptr) << uint64(32)) | uint64(size) |
			// Indicate that this is the error string by setting the 32-th bit, assuming that
			// no data exceeds 31-bit size (2 GiB).
			(1 << 31)
	}

	b, err = response.MarshalVT()
	if err != nil {
		return 0
	}
	ptr, size = wasm.ByteToPtr(b)
	return (uint64(ptr) << uint64(32)) | uint64(size)
}

type hostFunctions struct{}

func NewHostFunctions() HostFunctions {
	return hostFunctions{}
}

//go:wasmimport env http_get
func _http_get(ptr uint32, size uint32) uint64

func (h hostFunctions) HttpGet(ctx context.Context, request *HttpGetRequest) (*HttpGetResponse, error) {
	buf, err := request.MarshalVT()
	if err != nil {
		return nil, err
	}
	ptr, size := wasm.ByteToPtr(buf)
	ptrSize := _http_get(ptr, size)
	wasm.Free(ptr)

	ptr = uint32(ptrSize >> 32)
	size = uint32(ptrSize)
	buf = wasm.PtrToByte(ptr, size)

	response := new(HttpGetResponse)
	if err = response.UnmarshalVT(buf); err != nil {
		return nil, err
	}
	return response, nil
}

//go:wasmimport env log
func _log(ptr uint32, size uint32) uint64

func (h hostFunctions) Log(ctx context.Context, request *LogRequest) (*emptypb.Empty, error) {
	buf, err := request.MarshalVT()
	if err != nil {
		return nil, err
	}
	ptr, size := wasm.ByteToPtr(buf)
	ptrSize := _log(ptr, size)
	wasm.Free(ptr)

	ptr = uint32(ptrSize >> 32)
	size = uint32(ptrSize)
	buf = wasm.PtrToByte(ptr, size)

	response := new(emptypb.Empty)
	if err = response.UnmarshalVT(buf); err != nil {
		return nil, err
	}
	return response, nil
}
