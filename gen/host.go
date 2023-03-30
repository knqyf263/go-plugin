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
	// If it is only distributable host-functions, i.e. there is no plugin service definition
	if len(f.pluginServices) == 0 {
		g.P(fmt.Sprintf(`
		// Instantiate a Go-defined module named "%s" that exports host functions.
		func Instantiate(ctx %s, r %s, hostFunctions %s) error {
			envBuilder := r.NewHostModuleBuilder("%s")
			h := %s{hostFunctions}`,
			f.hostService.Module,
			g.QualifiedGoIdent(contextPackage.Ident("Context")),
			g.QualifiedGoIdent(wazeroPackage.Ident("Runtime")),
			f.hostService.GoName,
			f.hostService.Module,
			structName))
	} else {
		g.P(fmt.Sprintf(`
		// Instantiate a Go-defined module named "%s" that exports host functions.
		func (h %s) Instantiate(ctx %s, r %s) error {
			envBuilder := r.NewHostModuleBuilder("%s")`,
			f.hostService.Module,
			structName,
			g.QualifiedGoIdent(contextPackage.Ident("Context")),
			g.QualifiedGoIdent(wazeroPackage.Ident("Runtime")),
			f.hostService.Module))
	}

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

		g.P("request := new(", g.QualifiedGoIdent(method.Input.GoIdent), ")")
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
		type %s struct {
			newRuntime   func(%s) (%s, error)
			moduleConfig %s
		}`,
		pluginName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		g.QualifiedGoIdent(wazeroPackage.Ident("Runtime")),
		g.QualifiedGoIdent(wazeroPackage.Ident("ModuleConfig")),
	))

	g.P(fmt.Sprintf(
		"func New%s(ctx %s, opts ...wazeroConfigOption) (*%s, error) {",
		pluginName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		pluginName,
	))
	g.P(fmt.Sprintf(`o := &WazeroConfig{
				newRuntime: DefaultWazeroRuntime(),
				moduleConfig: %s(),
			}

			for _, opt := range opts {
				opt(o)
			}

			return &%s{
				newRuntime:   o.newRuntime,
				moduleConfig: o.moduleConfig,
			}, nil
		}
	`,
		g.QualifiedGoIdent(wazeroPackage.Ident("NewModuleConfig")),
		pluginName,
	))

	// Plugin loading
	structName := strings.ToLower(service.GoName[:1]) + service.GoName[1:]
	var hostFunctionsArg, exportHostFunctions string
	if f.hostService != nil {
		hostFunctionsArg = ", hostFunctions " + f.hostService.GoName
		exportHostFunctions = `
		h := _` + strings.ToLower(f.hostService.GoName[:1]) + f.hostService.GoName[1:] + `{hostFunctions}

		if err := h.Instantiate(ctx, r); err != nil {
			return nil, err
		}`
	}

	g.P(fmt.Sprintf(`
		type %s interface {
			Close(ctx %s) error
			%s
		}
	`,
		structName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		service.GoName,
	))

	g.P(fmt.Sprintf("func (p *%s) Load(ctx %s, pluginPath string %s) (%s, error) {",
		pluginName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		hostFunctionsArg,
		structName,
	))

	g.P(fmt.Sprintf(`b, err := %s(pluginPath)
		if err != nil {
			return nil, err
		}

		// Create a new runtime so that multiple modules will not conflict
		r, err := p.newRuntime(ctx)
		if err != nil {
			return nil, err
		}
		%s

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
		exportHostFunctions,
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

	g.P("return &", structName, "Plugin {",
		`
         runtime: r,
         module: module,
		 malloc: malloc,
		 free: free,`)

	for _, method := range service.Methods {
		varName := strings.ToLower(method.GoName[:1] + method.GoName[1:])
		g.P(varName, ": ", varName, ",")
	}
	g.P("}, nil")
	g.P("}")
	g.P()

	// Close plugin instance
	g.P(fmt.Sprintf(`func (p *%sPlugin) Close(ctx %s) (err error) {
		if r := p.runtime; r != nil {
			r.Close(ctx)
		}
		return
	}
	`,
		structName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
	))

	// Struct definition
	moduleType := g.QualifiedGoIdent(wazeroAPIPackage.Ident("Module"))
	funcType := g.QualifiedGoIdent(wazeroAPIPackage.Ident("Function"))
	g.P("type ", structName, "Plugin struct{")
	g.P(fmt.Sprintf(`
		runtime  %s
		module   %s
		malloc   %s 
		free     %s`,
		g.QualifiedGoIdent(wazeroPackage.Ident("Runtime")), moduleType, funcType, funcType))

	for _, method := range service.Methods {
		varName := strings.ToLower(method.GoName[:1] + method.GoName[1:])
		g.P(varName, " ", funcType)
	}
	g.P("}")

	for _, method := range service.Methods {
		genPluginMethod(g, f, method, structName+"Plugin")
	}
}

func genPluginMethod(g *protogen.GeneratedFile, f *fileInfo, method *protogen.Method, structName string) {
	g.P("func (p *", structName, ")", method.GoName, "(ctx ", g.QualifiedGoIdent(contextPackage.Ident("Context")),
		", request *", g.QualifiedGoIdent(method.Input.GoIdent), ")",
		"(*", g.QualifiedGoIdent(method.Output.GoIdent), ", error) {")

	errorHandling := "if err != nil {return nil , err}"
	g.P("data, err := request.MarshalVT()")
	g.P(errorHandling)
	g.P("dataSize := uint64(len(data))")
	g.P(fmt.Sprintf(`
			var dataPtr uint64
			// If the input data is not empty, we must allocate the in-Wasm memory to store it, and pass to the plugin.
			if dataSize != 0 {
				results, err := p.malloc.Call(ctx, dataSize)
				%s
				dataPtr = results[0]
				// This pointer is managed by TinyGo, but TinyGo is unaware of external usage.
				// So, we have to free it when finished
				defer p.free.Call(ctx, dataPtr)

				// The pointer is a linear memory offset, which is where we write the name.
				if !p.module.Memory().Write(uint32(dataPtr), data) {
					return nil, fmt.Errorf("Memory.Write(%%d, %%d) out of range of memory size %%d", dataPtr, dataSize, p.module.Memory().Size())
				}
			}
`, errorHandling))

	g.P("ptrSize, err := p.", strings.ToLower(method.GoName[:1]+method.GoName[1:]),
		".Call(ctx, dataPtr, dataSize)")
	g.P(errorHandling)
	g.P(fmt.Sprintf(`
			resPtr := uint32(ptrSize[0] >> 32)
			resSize := uint32(ptrSize[0])
			var isErrResponse bool
			if (resSize & %s) > 0 {
				isErrResponse = true
				resSize &^= %s
			}

			// We don't need the memory after deserialization: make sure it is freed.
			if resPtr != 0 {
				defer p.free.Call(ctx, uint64(resPtr))     
			}

			// The pointer is a linear memory offset, which is where we write the name.
			bytes, ok := p.module.Memory().Read(resPtr, resSize)
			if !ok {
				return nil, fmt.Errorf("Memory.Read(%%d, %%d) out of range of memory size %%d",
					resPtr, resSize, p.module.Memory().Size())
			}

			if isErrResponse {
				return nil, %s(string(bytes))
			}

			response := new(%s)
			if err = response.UnmarshalVT(bytes); err != nil {
				return nil, err
			}

			return response, nil`,
		ErrorMaskBit,
		ErrorMaskBit,
		g.QualifiedGoIdent(errorsPackage.Ident("New")),
		g.QualifiedGoIdent(method.Output.GoIdent)),
	)
	g.P("}")
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
