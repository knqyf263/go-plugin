package gen

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func (gg *Generator) generateHostFile(f *fileInfo) {
	filename := f.GeneratedFilenamePrefix + "_host.pb.go"
	g := gg.plugin.NewGeneratedFile(filename, f.GoImportPath)

	if len(f.pluginServices) == 0 && f.hostService == nil {
		g.Skip()
	}

	// Build constraints
	g.P("//go:build !tinygo.wasm")

	gg.generateHeader(g, f)
	gg.genHostFunctions(g, f)

	for _, service := range f.pluginServices {
		genHost(g, f, service)
	}
}

func (gg *Generator) genHostFunctions(g *protogen.GeneratedFile, f *fileInfo) {
	if f.hostService == nil {
		return
	}

	// Define host functions
	structName := "_" + strings.ToLower(f.hostService.GoName[:1]) + f.hostService.GoName[1:]
	g.P(fmt.Sprintf(`
		const (
			i32 = api.ValueTypeI32
			i64 = api.ValueTypeI64
		)

		type %s struct {
			%s
		}
		`, structName, f.hostService.GoName))

	// Define exporting functions
	g.P(fmt.Sprintf(`
		// Instantiate a Go-defined module named "env" that exports host functions.
		func (h %s) Instantiate(ctx %s, r %s) error {
			envBuilder := r.NewHostModuleBuilder("env")`,
		structName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		g.QualifiedGoIdent(wazeroPackage.Ident("Runtime"))))
	for _, method := range f.hostService.Methods {
		g.P(fmt.Sprintf(`
			envBuilder.NewFunctionBuilder().
			WithGoModuleFunction(api.GoModuleFunc(h._%s), []api.ValueType{i32, i32}, []api.ValueType{i64}).
			WithParameterNames("offset", "size").
			Export("%s")`,
			method.GoName, toSnakeCase(method.GoName)))
	}
	g.P(`
			_, err := envBuilder.Instantiate(ctx)
			return err
		}
`)

	errorHandling := `if err != nil {
			panic(err)
		}`
	for _, method := range f.hostService.Methods {
		g.P(method.Comments.Leading, fmt.Sprintf(`
			func (h %s) _%s(ctx %s, m %s, stack []uint64) {`,
			structName,
			method.GoName,
			g.QualifiedGoIdent(contextPackage.Ident("Context")),
			g.QualifiedGoIdent(wazeroAPIPackage.Ident("Module")),
		))
		g.P("offset, size := uint32(stack[0]), uint32(stack[1])")
		g.P("buf, err := ", g.QualifiedGoIdent(pluginWasmPackage.Ident("ReadMemory")), "(m.Memory(), offset, size)")
		g.P(errorHandling)

		g.P("var request ", g.QualifiedGoIdent(method.Input.GoIdent))
		g.P(`err = request.UnmarshalVT(buf)`)
		g.P(errorHandling)

		g.P("resp, err := h.", method.GoName, "(ctx, request)")
		g.P(errorHandling)

		g.P("buf, err = resp.MarshalVT()")
		g.P(errorHandling)

		g.P("ptr, err := ", g.QualifiedGoIdent(pluginWasmPackage.Ident("WriteMemory")), "(ctx, m, buf)")
		g.P(errorHandling)

		g.P("ptrLen := (ptr << uint64(32)) | uint64(len(buf))")
		g.P("stack[0] = ptrLen")
		g.P("}")
	}
}

func genHost(g *protogen.GeneratedFile, f *fileInfo, service *serviceInfo) {
	pluginName := service.GoName + "Plugin"

	g.P("const ", pluginName, "APIVersion = ", service.Version)
	g.P(fmt.Sprintf(`
		type %sOption struct {
			Stdout %s
			Stderr %s
			FS     %s
		}

		type %s struct {
			cache   %s
			config 	%s
		}`,
		pluginName,
		g.QualifiedGoIdent(ioPackage.Ident("Writer")),
		g.QualifiedGoIdent(ioPackage.Ident("Writer")),
		g.QualifiedGoIdent(fsPackage.Ident("FS")),
		pluginName,
		g.QualifiedGoIdent(wazeroPackage.Ident("CompilationCache")),
		g.QualifiedGoIdent(wazeroPackage.Ident("ModuleConfig")),
	))

	g.P(fmt.Sprintf(
		"func New%s(ctx %s, opt %sOption) (*%s, error) {",
		pluginName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		pluginName,
		pluginName,
	))
	g.P(fmt.Sprintf(`
			// Create a new WebAssembly CompilationCache.
			cache := %s()
		
			// Combine the above into our baseline config, overriding defaults.
			config := %s().
				// By default, I/O streams are discarded and there's no file system.
				WithStdout(opt.Stdout).WithStderr(opt.Stderr).WithFS(opt.FS)
			`,
		g.QualifiedGoIdent(wazeroPackage.Ident("NewCompilationCache")),
		g.QualifiedGoIdent(wazeroPackage.Ident("NewModuleConfig")),
	))

	g.P("return &", pluginName, `{
				cache: cache,
				config:  config,
			}, nil
		}
	`)

	// Close plugin
	g.P(fmt.Sprintf(`func (p *%s) Close(ctx %s) (err error) {
	if c := p.cache; c != nil {
		err = c.Close(ctx)
	}
	return
}
`,
		pluginName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
	))

	// Plugin loading
	structName := strings.ToLower(service.GoName[:1]) + service.GoName[1:] + "Plugin"
	var hostFunctionsArg, exportHostFunctions string
	if f.hostService != nil {
		hostFunctionsArg = ", hostFunctions " + f.hostService.GoName
		exportHostFunctions = `
		h := _` + strings.ToLower(f.hostService.GoName[:1]) + f.hostService.GoName[1:] + `{hostFunctions}

		if err := h.Instantiate(ctx, r); err != nil {
			return nil, err
		}`
	}

	g.P(fmt.Sprintf("type %sHandlerFn func(ctx %s, runtime %s) error",
		service.GoName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		g.QualifiedGoIdent(wazeroPackage.Ident("Runtime")),
	))

	g.P(fmt.Sprintf("func (p *%s) Load(ctx %s, pluginPath string %s, handlers ...%sHandlerFn) (%s, error) {",
		pluginName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		hostFunctionsArg,
		service.GoName,
		service.GoName,
	))

	g.P(fmt.Sprintf(`b, err := %s(pluginPath)
		if err != nil {
			return nil, err
		}

		// Create an empty namespace so that multiple modules will not conflict
		r := %s(ctx, %s().WithCompilationCache(p.cache))
		%s

		for _, hf := range handlers {
			if err := hf(ctx, r); err != nil {
				return nil, err
			}
		}

		if _, err = %s(r).Instantiate(ctx); err != nil {
			return nil, err
		}

		// Compile the WebAssembly module using the default configuration.
		code, err := r.CompileModule(ctx, b)
		if err != nil {
			return nil, err
		}
	
		// InstantiateModule runs the "_start" function, WASI's "main".
		module, err := r.InstantiateModule(ctx, code, p.config)
		if err != nil {
			// Note: Most compilers do not exit the module after running "_start",
			// unless there was an Error. This allows you to call exported functions.
			if exitErr, ok := err.(*%s); ok && exitErr.ExitCode() != 0 {
				return nil, %s("unexpected exit_code: %%d", exitErr.ExitCode())
			} else if !ok {
				return nil, err
			}
		}

		// Compare API versions with the loading plugin
		apiVersion := module.ExportedFunction("%s_api_version")
		if apiVersion == nil {
			return nil, %s("%s_api_version is not exported")
		}
		results, err := apiVersion.Call(ctx)
		if err != nil {
			return nil, err
		} else if len(results) != 1 {
			return nil, %s("invalid %s_api_version signature")
		}
		if results[0] != %sAPIVersion {
			return nil, fmt.Errorf("API version mismatch, host: %%d, plugin: %%d", %sAPIVersion, results[0])
		}
`,
		g.QualifiedGoIdent(osPackage.Ident("ReadFile")),
		g.QualifiedGoIdent(wazeroPackage.Ident("NewRuntimeWithConfig")),
		g.QualifiedGoIdent(wazeroPackage.Ident("NewRuntimeConfig")),
		exportHostFunctions,
		g.QualifiedGoIdent(wazeroWasiPackage.Ident("NewBuilder")),
		g.QualifiedGoIdent(wazeroSysPackage.Ident("ExitError")),
		g.QualifiedGoIdent(fmtPackage.Ident("Errorf")),
		toSnakeCase(service.GoName),
		g.QualifiedGoIdent(errorsPackage.Ident("New")),
		toSnakeCase(service.GoName),
		g.QualifiedGoIdent(errorsPackage.Ident("New")),
		toSnakeCase(service.GoName),
		pluginName, pluginName,
	))

	errorsNew := g.QualifiedGoIdent(errorsPackage.Ident("New"))
	for _, method := range service.Methods {
		varName := strings.ToLower(method.GoName[:1] + method.GoName[1:])
		funcName := toSnakeCase(service.GoName + method.GoName)
		g.P(varName, `:= module.ExportedFunction("`, funcName, `")`)
		g.P("if ", varName, `== nil { return nil, `, errorsNew, `("`, funcName, ` is not exported")}`)
	}

	g.P(fmt.Sprintf(`
		malloc := module.ExportedFunction("malloc")
		if malloc == nil {
			return nil, %s("malloc is not exported")
		}

		free := module.ExportedFunction("free")
		if free == nil {
			return nil, %s("free is not exported")
		}`,
		errorsNew, errorsNew))

	g.P("return &", structName, "{",
		`module: module,
		 malloc: malloc,
		 free: free,`)

	for _, method := range service.Methods {
		varName := strings.ToLower(method.GoName[:1] + method.GoName[1:])
		g.P(varName, ": ", varName, ",")
	}
	g.P("}, nil")
	g.P("}")
	g.P()

	// Struct definition
	moduleType := g.QualifiedGoIdent(wazeroAPIPackage.Ident("Module"))
	funcType := g.QualifiedGoIdent(wazeroAPIPackage.Ident("Function"))
	g.P("type ", structName, " struct{")
	g.P(fmt.Sprintf(`
		module   %s
		malloc   %s 
		free     %s`,
		moduleType, funcType, funcType))

	for _, method := range service.Methods {
		varName := strings.ToLower(method.GoName[:1] + method.GoName[1:])
		g.P(varName, " ", funcType)
	}
	g.P("}")

	for _, method := range service.Methods {
		genPluginMethod(g, f, method, structName)
	}
}

func genPluginMethod(g *protogen.GeneratedFile, f *fileInfo, method *protogen.Method, structName string) {
	g.P("func (p *", structName, ")", method.GoName, "(ctx ", g.QualifiedGoIdent(contextPackage.Ident("Context")),
		", request ", g.QualifiedGoIdent(method.Input.GoIdent), ")",
		"(response ", g.QualifiedGoIdent(method.Output.GoIdent), ", err error) {")

	errorHandling := "if err != nil {return response , err}"
	g.P("data, err := request.MarshalVT()")
	g.P(errorHandling)
	g.P("dataSize := uint64(len(data))")
	g.P(`
			var dataPtr uint64
			// If the input data is not empty, we must allocate the in-Wasm memory to store it, and pass to the plugin.
			if dataSize != 0 {
				results, err := p.malloc.Call(ctx, dataSize)
				if err != nil {return response , err}
				dataPtr = results[0]
				// This pointer is managed by TinyGo, but TinyGo is unaware of external usage.
				// So, we have to free it when finished
				defer p.free.Call(ctx, dataPtr)

				// The pointer is a linear memory offset, which is where we write the name.
				if !p.module.Memory().Write(uint32(dataPtr), data) {
					return response, fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d", dataPtr, dataSize, p.module.Memory().Size())
				}
			}
`)
	g.P("ptrSize, err := p.", strings.ToLower(method.GoName[:1]+method.GoName[1:]),
		".Call(ctx, dataPtr, dataSize)")
	g.P(errorHandling)
	g.P(`
			// Note: This pointer is still owned by TinyGo, so don't try to free it!
			resPtr := uint32(ptrSize[0] >> 32)
			resSize := uint32(ptrSize[0])

			// The pointer is a linear memory offset, which is where we write the name.
			bytes, ok := p.module.Memory().Read(resPtr, resSize)
			if !ok {
				return response, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d",
					resPtr, resSize, p.module.Memory().Size())
			}

			if err = response.UnmarshalVT(bytes); err != nil {
				return response, err
			}

			return response, nil`)
	g.P("}")
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
