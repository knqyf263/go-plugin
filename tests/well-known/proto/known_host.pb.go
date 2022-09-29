//go:build !tinygo.wasm

// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v3.21.5
// source: tests/well-known/proto/known.proto

package proto

import (
	context "context"
	errors "errors"
	fmt "fmt"
	emptypb "github.com/knqyf263/go-plugin/types/known/emptypb"
	wazero "github.com/tetratelabs/wazero"
	api "github.com/tetratelabs/wazero/api"
	wasi_snapshot_preview1 "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	sys "github.com/tetratelabs/wazero/sys"
	io "io"
	fs "io/fs"
	os "os"
)

const KnownTypesTestPluginAPIVersion = 1

type KnownTypesTestPluginOption struct {
	Stdout io.Writer
	Stderr io.Writer
	FS     fs.FS
}

type KnownTypesTestPlugin struct {
	runtime wazero.Runtime
	config  wazero.ModuleConfig
}

func NewKnownTypesTestPlugin(ctx context.Context, opt KnownTypesTestPluginOption) (*KnownTypesTestPlugin, error) {

	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntime(ctx)

	// Combine the above into our baseline config, overriding defaults.
	config := wazero.NewModuleConfig().
		// By default, I/O streams are discarded and there's no file system.
		WithStdout(opt.Stdout).WithStderr(opt.Stderr).WithFS(opt.FS)

	return &KnownTypesTestPlugin{
		runtime: r,
		config:  config,
	}, nil
}
func (p *KnownTypesTestPlugin) Load(ctx context.Context, pluginPath string) (KnownTypesTest, error) {
	b, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, err
	}

	// Create an empty namespace so that multiple modules will not conflict
	ns := p.runtime.NewNamespace(ctx)

	if _, err = wasi_snapshot_preview1.NewBuilder(p.runtime).Instantiate(ctx, ns); err != nil {
		return nil, err
	}

	// Compile the WebAssembly module using the default configuration.
	code, err := p.runtime.CompileModule(ctx, b)
	if err != nil {
		return nil, err
	}

	// InstantiateModule runs the "_start" function, WASI's "main".
	module, err := ns.InstantiateModule(ctx, code, p.config)
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
	apiVersion := module.ExportedFunction("known_types_test_api_version")
	if apiVersion == nil {
		return nil, errors.New("known_types_test_api_version is not exported")
	}
	results, err := apiVersion.Call(ctx)
	if err != nil {
		return nil, err
	} else if len(results) != 1 {
		return nil, errors.New("invalid known_types_test_api_version signature")
	}
	if results[0] != KnownTypesTestPluginAPIVersion {
		return nil, fmt.Errorf("API version mismatch, host: %d, plugin: %d", KnownTypesTestPluginAPIVersion, results[0])
	}

	test := module.ExportedFunction("known_types_test_test")
	if test == nil {
		return nil, errors.New("known_types_test_test is not exported")
	}

	malloc := module.ExportedFunction("malloc")
	if malloc == nil {
		return nil, errors.New("malloc is not exported")
	}

	free := module.ExportedFunction("free")
	if free == nil {
		return nil, errors.New("free is not exported")
	}
	return &knownTypesTestPlugin{module: module,
		malloc: malloc,
		free:   free,
		test:   test,
	}, nil
}

type knownTypesTestPlugin struct {
	module api.Module
	malloc api.Function
	free   api.Function
	test   api.Function
}

func (p *knownTypesTestPlugin) Test(ctx context.Context, request Request) (response Response, err error) {
	data, err := request.MarshalVT()
	if err != nil {
		return response, err
	}
	dataSize := uint64(len(data))
	results, err := p.malloc.Call(ctx, dataSize)
	if err != nil {
		return response, err
	}

	dataPtr := results[0]
	// This pointer is managed by TinyGo, but TinyGo is unaware of external usage.
	// So, we have to free it when finished
	defer p.free.Call(ctx, dataPtr)

	// The pointer is a linear memory offset, which is where we write the name.
	if !p.module.Memory().Write(ctx, uint32(dataPtr), data) {
		return response, fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d", dataPtr, dataSize, p.module.Memory().Size(ctx))
	}
	ptrSize, err := p.test.Call(ctx, dataPtr, dataSize)
	if err != nil {
		return response, err
	}

	// Note: This pointer is still owned by TinyGo, so don't try to free it!
	resPtr := uint32(ptrSize[0] >> 32)
	resSize := uint32(ptrSize[0])

	// The pointer is a linear memory offset, which is where we write the name.
	bytes, ok := p.module.Memory().Read(ctx, resPtr, resSize)
	if !ok {
		return response, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d",
			resPtr, resSize, p.module.Memory().Size(ctx))
	}

	if err = response.UnmarshalVT(bytes); err != nil {
		return response, err
	}

	return response, nil
}

const EmptyTestPluginAPIVersion = 1

type EmptyTestPluginOption struct {
	Stdout io.Writer
	Stderr io.Writer
	FS     fs.FS
}

type EmptyTestPlugin struct {
	runtime wazero.Runtime
	config  wazero.ModuleConfig
}

func NewEmptyTestPlugin(ctx context.Context, opt EmptyTestPluginOption) (*EmptyTestPlugin, error) {

	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntime(ctx)

	// Combine the above into our baseline config, overriding defaults.
	config := wazero.NewModuleConfig().
		// By default, I/O streams are discarded and there's no file system.
		WithStdout(opt.Stdout).WithStderr(opt.Stderr).WithFS(opt.FS)

	return &EmptyTestPlugin{
		runtime: r,
		config:  config,
	}, nil
}
func (p *EmptyTestPlugin) Load(ctx context.Context, pluginPath string) (EmptyTest, error) {
	b, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, err
	}

	// Create an empty namespace so that multiple modules will not conflict
	ns := p.runtime.NewNamespace(ctx)

	if _, err = wasi_snapshot_preview1.NewBuilder(p.runtime).Instantiate(ctx, ns); err != nil {
		return nil, err
	}

	// Compile the WebAssembly module using the default configuration.
	code, err := p.runtime.CompileModule(ctx, b)
	if err != nil {
		return nil, err
	}

	// InstantiateModule runs the "_start" function, WASI's "main".
	module, err := ns.InstantiateModule(ctx, code, p.config)
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
	apiVersion := module.ExportedFunction("empty_test_api_version")
	if apiVersion == nil {
		return nil, errors.New("empty_test_api_version is not exported")
	}
	results, err := apiVersion.Call(ctx)
	if err != nil {
		return nil, err
	} else if len(results) != 1 {
		return nil, errors.New("invalid empty_test_api_version signature")
	}
	if results[0] != EmptyTestPluginAPIVersion {
		return nil, fmt.Errorf("API version mismatch, host: %d, plugin: %d", EmptyTestPluginAPIVersion, results[0])
	}

	donothing := module.ExportedFunction("empty_test_do_nothing")
	if donothing == nil {
		return nil, errors.New("empty_test_do_nothing is not exported")
	}

	malloc := module.ExportedFunction("malloc")
	if malloc == nil {
		return nil, errors.New("malloc is not exported")
	}

	free := module.ExportedFunction("free")
	if free == nil {
		return nil, errors.New("free is not exported")
	}
	return &emptyTestPlugin{module: module,
		malloc:    malloc,
		free:      free,
		donothing: donothing,
	}, nil
}

type emptyTestPlugin struct {
	module    api.Module
	malloc    api.Function
	free      api.Function
	donothing api.Function
}

func (p *emptyTestPlugin) DoNothing(ctx context.Context, request emptypb.Empty) (response emptypb.Empty, err error) {
	data, err := request.MarshalVT()
	if err != nil {
		return response, err
	}
	dataSize := uint64(len(data))
	results, err := p.malloc.Call(ctx, dataSize)
	if err != nil {
		return response, err
	}

	dataPtr := results[0]
	// This pointer is managed by TinyGo, but TinyGo is unaware of external usage.
	// So, we have to free it when finished
	defer p.free.Call(ctx, dataPtr)

	// The pointer is a linear memory offset, which is where we write the name.
	if !p.module.Memory().Write(ctx, uint32(dataPtr), data) {
		return response, fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d", dataPtr, dataSize, p.module.Memory().Size(ctx))
	}
	ptrSize, err := p.donothing.Call(ctx, dataPtr, dataSize)
	if err != nil {
		return response, err
	}

	// Note: This pointer is still owned by TinyGo, so don't try to free it!
	resPtr := uint32(ptrSize[0] >> 32)
	resSize := uint32(ptrSize[0])

	// The pointer is a linear memory offset, which is where we write the name.
	bytes, ok := p.module.Memory().Read(ctx, resPtr, resSize)
	if !ok {
		return response, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d",
			resPtr, resSize, p.module.Memory().Size(ctx))
	}

	if err = response.UnmarshalVT(bytes); err != nil {
		return response, err
	}

	return response, nil
}
