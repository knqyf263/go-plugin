//go:build !wasip1

// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v5.29.3
// source: examples/helloworld/greeting/greet.proto

package greeting

import (
	context "context"
	errors "errors"
	fmt "fmt"
	wazero "github.com/tetratelabs/wazero"
	api "github.com/tetratelabs/wazero/api"
	sys "github.com/tetratelabs/wazero/sys"
	os "os"
)

const GreeterPluginAPIVersion = 1

type GreeterPlugin struct {
	newRuntime   func(context.Context) (wazero.Runtime, error)
	moduleConfig wazero.ModuleConfig
}

func NewGreeterPlugin(ctx context.Context, opts ...wazeroConfigOption) (*GreeterPlugin, error) {
	o := &WazeroConfig{
		newRuntime:   DefaultWazeroRuntime(),
		moduleConfig: wazero.NewModuleConfig().WithStartFunctions("_initialize"),
	}

	for _, opt := range opts {
		opt(o)
	}

	return &GreeterPlugin{
		newRuntime:   o.newRuntime,
		moduleConfig: o.moduleConfig,
	}, nil
}

type greeter interface {
	Close(ctx context.Context) error
	Greeter
}

func (p *GreeterPlugin) Load(ctx context.Context, pluginPath string) (greeter, error) {
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
	apiVersion := module.ExportedFunction("greeter_api_version")
	if apiVersion == nil {
		return nil, errors.New("greeter_api_version is not exported")
	}
	results, err := apiVersion.Call(ctx)
	if err != nil {
		return nil, err
	} else if len(results) != 1 {
		return nil, errors.New("invalid greeter_api_version signature")
	}
	if results[0] != GreeterPluginAPIVersion {
		return nil, fmt.Errorf("API version mismatch, host: %d, plugin: %d", GreeterPluginAPIVersion, results[0])
	}

	greet := module.ExportedFunction("greeter_greet")
	if greet == nil {
		return nil, errors.New("greeter_greet is not exported")
	}

	malloc := module.ExportedFunction("malloc")
	if malloc == nil {
		return nil, errors.New("malloc is not exported")
	}

	free := module.ExportedFunction("free")
	if free == nil {
		return nil, errors.New("free is not exported")
	}
	return &greeterPlugin{
		runtime: r,
		module:  module,
		malloc:  malloc,
		free:    free,
		greet:   greet,
	}, nil
}

func (p *greeterPlugin) Close(ctx context.Context) (err error) {
	if r := p.runtime; r != nil {
		r.Close(ctx)
	}
	return
}

type greeterPlugin struct {
	runtime wazero.Runtime
	module  api.Module
	malloc  api.Function
	free    api.Function
	greet   api.Function
}

func (p *greeterPlugin) Greet(ctx context.Context, request *GreetRequest) (*GreetReply, error) {
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

	ptrSize, err := p.greet.Call(ctx, dataPtr, dataSize)
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

	response := new(GreetReply)
	if err = response.UnmarshalVT(bytes); err != nil {
		return nil, err
	}

	return response, nil
}
