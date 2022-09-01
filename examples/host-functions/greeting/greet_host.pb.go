//go:build !tinygo.wasm

// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v3.21.5
// source: examples/host-functions/greeting/greet.proto

package greeting

import (
	context "context"
	errors "errors"
	fmt "fmt"
	wasm "github.com/knqyf263/go-plugin/wasm"
	wazero "github.com/tetratelabs/wazero"
	api "github.com/tetratelabs/wazero/api"
	wasi_snapshot_preview1 "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	sys "github.com/tetratelabs/wazero/sys"
	io "io"
	fs "io/fs"
	os "os"
)

type _hostFunctions struct {
	HostFunctions
}

func (h _hostFunctions) Export() map[string]interface{} {
	return map[string]interface{}{
		"http_get": h._HttpGet(),
		"log":      h._Log(),
	}
}

// Sends a HTTP GET request

func (h _hostFunctions) _HttpGet() func(ctx context.Context, m api.Module, offset, size uint32) uint64 {
	return func(ctx context.Context, m api.Module, offset, size uint32) uint64 {
		buf, err := wasm.ReadMemory(ctx, m, offset, size)
		if err != nil {
			panic(err)
		}
		var request HttpGetRequest
		err = request.UnmarshalVT(buf)
		if err != nil {
			panic(err)
		}
		resp, err := h.HttpGet(ctx, request)
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
		return (ptr << uint64(32)) | uint64(len(buf))
	}
}

// Shows a log message

func (h _hostFunctions) _Log() func(ctx context.Context, m api.Module, offset, size uint32) uint64 {
	return func(ctx context.Context, m api.Module, offset, size uint32) uint64 {
		buf, err := wasm.ReadMemory(ctx, m, offset, size)
		if err != nil {
			panic(err)
		}
		var request LogRequest
		err = request.UnmarshalVT(buf)
		if err != nil {
			panic(err)
		}
		resp, err := h.Log(ctx, request)
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
		return (ptr << uint64(32)) | uint64(len(buf))
	}
}

const GreeterPluginAPIVersion = 1

type GreeterPluginOption struct {
	Stdout io.Writer
	Stderr io.Writer
	FS     fs.FS
}

type GreeterPlugin struct {
	runtime wazero.Runtime
	config  wazero.ModuleConfig
}

func NewGreeterPlugin(ctx context.Context, opt GreeterPluginOption) (*GreeterPlugin, error) {

	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		// WebAssembly 2.0 allows use of any version of TinyGo, including 0.24+.
		WithWasmCore2())

	// Combine the above into our baseline config, overriding defaults.
	config := wazero.NewModuleConfig().
		// By default, I/O streams are discarded and there's no file system.
		WithStdout(opt.Stdout).WithStderr(opt.Stderr).WithFS(opt.FS)

	return &GreeterPlugin{
		runtime: r,
		config:  config,
	}, nil
}
func (p *GreeterPlugin) Load(ctx context.Context, pluginPath string, hostFunctions HostFunctions) (Greeter, error) {
	b, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, err
	}

	// Create an empty namespace so that multiple modules will not conflict
	ns := p.runtime.NewNamespace(ctx)

	h := _hostFunctions{hostFunctions}

	// Instantiate a Go-defined module named "env" that exports functions.
	_, err = p.runtime.NewModuleBuilder("env").ExportFunctions(h.Export()).Instantiate(ctx, ns)
	if err != nil {
		return nil, err
	}

	if _, err = wasi_snapshot_preview1.NewBuilder(p.runtime).Instantiate(ctx, ns); err != nil {
		return nil, err
	}

	// Compile the WebAssembly module using the default configuration.
	code, err := p.runtime.CompileModule(ctx, b, wazero.NewCompileConfig())
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
	return &greeterPlugin{module: module,
		malloc: malloc,
		free:   free,
		greet:  greet,
	}, nil
}

type greeterPlugin struct {
	module api.Module
	malloc api.Function
	free   api.Function
	greet  api.Function
}

func (p *greeterPlugin) Greet(ctx context.Context, request GreetRequest) (response GreetReply, err error) {
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
	ptrSize, err := p.greet.Call(ctx, dataPtr, dataSize)
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
