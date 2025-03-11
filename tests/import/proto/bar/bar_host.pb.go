//go:build !wasip1

// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v5.29.3
// source: tests/import/proto/bar/bar.proto

package bar

import (
	context "context"
	errors "errors"
	fmt "fmt"
	wazero "github.com/tetratelabs/wazero"
	api "github.com/tetratelabs/wazero/api"
	sys "github.com/tetratelabs/wazero/sys"
	os "os"
)

const BarPluginAPIVersion = 1

type BarPlugin struct {
	newRuntime   func(context.Context) (wazero.Runtime, error)
	moduleConfig wazero.ModuleConfig
}

func NewBarPlugin(ctx context.Context, opts ...wazeroConfigOption) (*BarPlugin, error) {
	o := &WazeroConfig{
		newRuntime:   DefaultWazeroRuntime(),
		moduleConfig: wazero.NewModuleConfig().WithStartFunctions("_initialize"),
	}

	for _, opt := range opts {
		opt(o)
	}

	return &BarPlugin{
		newRuntime:   o.newRuntime,
		moduleConfig: o.moduleConfig,
	}, nil
}

type bar interface {
	Close(ctx context.Context) error
	Bar
}

func (p *BarPlugin) Load(ctx context.Context, pluginPath string) (bar, error) {
	b, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, err
	}

	// Create a new runtime so that multiple modules will not conflict
	r, err := p.newRuntime(ctx)
	if err != nil {
		return nil, err
	}

	// Compile the WebAssembly module using the default configuration.
	code, err := r.CompileModule(ctx, b)
	if err != nil {
		return nil, err
	}

	// InstantiateModule runs the "_start" function, WASI's "main".
	module, err := r.InstantiateModule(ctx, code, p.moduleConfig)
	if err != nil {
		// Note: Most compilers do not exit the module after running "_start",
		// unless there was an Error. This allows you to call exported functions.
		if exitErr, ok := err.(*sys.ExitError); ok && exitErr.ExitCode() != 0 {
			return nil, fmt.Errorf("unexpected exit_code: %d", exitErr.ExitCode())
		} else if !ok {
			return nil, err
		}
	}

	// Compare API versions with the loading plugin
	apiVersion := module.ExportedFunction("bar_api_version")
	if apiVersion == nil {
		return nil, errors.New("bar_api_version is not exported")
	}
	results, err := apiVersion.Call(ctx)
	if err != nil {
		return nil, err
	} else if len(results) != 1 {
		return nil, errors.New("invalid bar_api_version signature")
	}
	if results[0] != BarPluginAPIVersion {
		return nil, fmt.Errorf("API version mismatch, host: %d, plugin: %d", BarPluginAPIVersion, results[0])
	}

	hello := module.ExportedFunction("bar_hello")
	if hello == nil {
		return nil, errors.New("bar_hello is not exported")
	}

	malloc := module.ExportedFunction("malloc")
	if malloc == nil {
		return nil, errors.New("malloc is not exported")
	}

	free := module.ExportedFunction("free")
	if free == nil {
		return nil, errors.New("free is not exported")
	}
	return &barPlugin{
		runtime: r,
		module:  module,
		malloc:  malloc,
		free:    free,
		hello:   hello,
	}, nil
}

func (p *barPlugin) Close(ctx context.Context) (err error) {
	if r := p.runtime; r != nil {
		r.Close(ctx)
	}
	return
}

type barPlugin struct {
	runtime wazero.Runtime
	module  api.Module
	malloc  api.Function
	free    api.Function
	hello   api.Function
}

func (p *barPlugin) Hello(ctx context.Context, request *Request) (*Reply, error) {
	data, err := request.MarshalVT()
	if err != nil {
		return nil, err
	}
	dataSize := uint64(len(data))

	var dataPtr uint64
	// If the input data is not empty, we must allocate the in-Wasm memory to store it, and pass to the plugin.
	if dataSize != 0 {
		results, err := p.malloc.Call(ctx, dataSize)
		if err != nil {
			return nil, err
		}
		dataPtr = results[0]
		// This pointer is managed by the Wasm module, which is unaware of external usage.
		// So, we have to free it when finished
		defer p.free.Call(ctx, dataPtr)

		// The pointer is a linear memory offset, which is where we write the name.
		if !p.module.Memory().Write(uint32(dataPtr), data) {
			return nil, fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d", dataPtr, dataSize, p.module.Memory().Size())
		}
	}

	ptrSize, err := p.hello.Call(ctx, dataPtr, dataSize)
	if err != nil {
		return nil, err
	}

	resPtr := uint32(ptrSize[0] >> 32)
	resSize := uint32(ptrSize[0])
	var isErrResponse bool
	if (resSize & (1 << 31)) > 0 {
		isErrResponse = true
		resSize &^= (1 << 31)
	}

	// We don't need the memory after deserialization: make sure it is freed.
	if resPtr != 0 {
		defer p.free.Call(ctx, uint64(resPtr))
	}

	// The pointer is a linear memory offset, which is where we write the name.
	bytes, ok := p.module.Memory().Read(resPtr, resSize)
	if !ok {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d",
			resPtr, resSize, p.module.Memory().Size())
	}

	if isErrResponse {
		return nil, errors.New(string(bytes))
	}

	response := new(Reply)
	if err = response.UnmarshalVT(bytes); err != nil {
		return nil, err
	}

	return response, nil
}
