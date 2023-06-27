//go:build !tinygo.wasm

// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v3.21.1
// source: examples/host-function-library/library/json-parser/export/library.proto

package export

import (
	context "context"
	wasm "github.com/knqyf263/go-plugin/wasm"
	wazero "github.com/tetratelabs/wazero"
	api "github.com/tetratelabs/wazero/api"
)

const (
	i32 = api.ValueTypeI32
	i64 = api.ValueTypeI64
)

type _parserLibrary struct {
	ParserLibrary
}

// Instantiate a Go-defined module named "json-parser" that exports host functions.
func Instantiate(ctx context.Context, r wazero.Runtime, hostFunctions ParserLibrary) error {
	envBuilder := r.NewHostModuleBuilder("json-parser")
	h := _parserLibrary{hostFunctions}

	envBuilder.NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(h._ParseJson), []api.ValueType{i32, i32}, []api.ValueType{i64}).
		WithParameterNames("offset", "size").
		Export("parse_json")

	_, err := envBuilder.Instantiate(ctx)
	return err
}

func (h _parserLibrary) _ParseJson(ctx context.Context, m api.Module, stack []uint64) {
	offset, size := uint32(stack[0]), uint32(stack[1])
	buf, err := wasm.ReadMemory(m.Memory(), offset, size)
	if err != nil {
		panic(err)
	}
	request := new(ParseJsonRequest)
	err = request.UnmarshalVT(buf)
	if err != nil {
		panic(err)
	}
	resp, err := h.ParseJson(ctx, request)
	if err != nil {
		panic(err)
	}
	buf, err = resp.MarshalVT()
	if err != nil {
		panic(err)
	}
	ptr, err := wasm.WriteMemory(ctx, m, buf)
	if err != nil {
		panic(err)
	}
	ptrLen := (ptr << uint64(32)) | uint64(len(buf))
	stack[0] = ptrLen
}
