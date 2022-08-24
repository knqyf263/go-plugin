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

	if len(f.pluginServices) == 0 {
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
	// Define host functions
	g.P("var _hostFunctions ", f.hostService.GoName)
	g.P("func RegisterHostFunctions(h ", f.hostService.GoName, ") {")
	g.P("	_hostFunctions = h")
	g.P("}")
	g.P()

	g.P(fmt.Sprintf(`
			func readMemory(ctx %s, m %s, offset, size uint32) ([]byte, error) {
				buf, ok := m.Memory().Read(ctx, offset, size)
				if !ok {
					return nil, fmt.Errorf("Memory.Read(%%d, %%d) out of range", offset, size)
				}
				return buf, nil
			}
			
			func writeMemory(ctx %s, m %s, data []byte) (uint64, error) {
				malloc := m.ExportedFunction("malloc")
				if malloc == nil {
					return 0, %s("malloc is not exported")
				}
				results, err := malloc.Call(ctx, uint64(len(data)))
				if err != nil {
					return 0, err
				}
				dataPtr := results[0]
			
				// The pointer is a linear memory offset, which is where we write the name.
				if !m.Memory().Write(ctx, uint32(dataPtr), data) {
					return 0, %s("Memory.Write(%%d, %%d) out of range of memory size %%d",
						dataPtr, len(data), m.Memory().Size(ctx))
				}

				return dataPtr, nil
			}`,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		g.QualifiedGoIdent(wazeroAPIPackage.Ident("Module")),
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		g.QualifiedGoIdent(wazeroAPIPackage.Ident("Module")),
		g.QualifiedGoIdent(errorsPackage.Ident("New")),
		g.QualifiedGoIdent(fmtPackage.Ident("Errorf")),
	))
	g.P()

	// Define exporting functions
	g.P(`func hostFunctions() map[string]interface{} {
				if _hostFunctions == nil {
					return nil
				}
				return map[string]interface{}{`)
	for _, method := range f.hostService.Methods {
		g.P(fmt.Sprintf(`"%s": %s,`, toSnakeCase(method.GoName), method.GoName))
	}
	g.P("	}")
	g.P("}")

	errorHandling := `if err != nil {
			panic(err)
		}`
	for _, method := range f.hostService.Methods {
		g.P(method.Comments.Leading, fmt.Sprintf(`func %s(ctx %s, m %s, offset, size uint32) uint64 {`,
			method.GoName,
			g.QualifiedGoIdent(contextPackage.Ident("Context")),
			g.QualifiedGoIdent(wazeroAPIPackage.Ident("Module")),
		))
		g.P("buf, err := readMemory(ctx, m, offset, size)")
		g.P(errorHandling)

		g.P("var request ", g.QualifiedGoIdent(method.Input.GoIdent))
		g.P(`err = request.UnmarshalVT(buf)`)
		g.P(errorHandling)

		g.P("resp, err := _hostFunctions.", method.GoName, "(ctx, request)")
		g.P(errorHandling)

		g.P("buf, err = resp.MarshalVT()")
		g.P(errorHandling)

		g.P("ptr, err := writeMemory(ctx, m, buf)")
		g.P(errorHandling)

		g.P("return (ptr << uint64(32)) | uint64(len(buf))")
		g.P("}")
	}
}

func genHost(g *protogen.GeneratedFile, f *fileInfo, service *serviceInfo) {
	if service.Type != ServicePlugin {
		return
	}
	g.P(fmt.Sprintf(`type %sPlugin struct {
			runtime 		%s
			config  		%s
		}`, service.GoName,
		g.QualifiedGoIdent(wazeroPackage.Ident("Runtime")),
		g.QualifiedGoIdent(wazeroPackage.Ident("ModuleConfig")),
	))

	pluginName := service.GoName + "Plugin"
	g.P("func New", pluginName, "() (*", pluginName, ", error) {")
	g.P(fmt.Sprintf(`// Choose the context to use for function calls.
			ctx := %s()
		
			// Create a new WebAssembly Runtime.
			r := %s(ctx, %s().
				// WebAssembly 2.0 allows use of any version of TinyGo, including 0.24+.
				WithWasmCore2())
		
			// Combine the above into our baseline config, overriding defaults.
			config := %s().
				// By default, I/O streams are discarded and there's no file system.
				WithStdout(os.Stdout).WithStderr(os.Stderr)
			`, g.QualifiedGoIdent(contextPackage.Ident("Background")),
		g.QualifiedGoIdent(wazeroPackage.Ident("NewRuntimeWithConfig")),
		g.QualifiedGoIdent(wazeroPackage.Ident("NewRuntimeConfig")),
		g.QualifiedGoIdent(wazeroPackage.Ident("NewModuleConfig")),
	))

	g.P("return &", pluginName, `{
				runtime: r,
				config:  config,
			}, nil
		}`)

	// Plugin loading
	structName := strings.ToLower(service.GoName[:1]) + service.GoName[1:] + "Plugin"
	g.P(fmt.Sprintf("func (p *%s) Load(ctx %s, pluginPath string) (%s, error) {",
		pluginName,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		service.GoName,
	))
	g.P(fmt.Sprintf(`b, err := %s(pluginPath)
		if err != nil {
			return nil, err
		}

		// Create an empty namespace so that multiple modules will not conflict
		ns := p.runtime.NewNamespace(ctx)

		// Instantiate a Go-defined module named "env" that exports functions.
		_, err = p.runtime.NewModuleBuilder("env").
			ExportFunctions(hostFunctions()).
			Instantiate(ctx, ns)
		if err != nil {
			return nil, err
		}

		if _, err = %s(p.runtime).Instantiate(ctx, ns); err != nil {
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
			if exitErr, ok := err.(*%s); ok && exitErr.ExitCode() != 0 {
				return nil, %s("unexpected exit_code: %%d", exitErr.ExitCode())
			} else if !ok {
				return nil, err
			}
		}
`,
		g.QualifiedGoIdent(osPackage.Ident("ReadFile")),
		g.QualifiedGoIdent(wazeroWasiPackage.Ident("NewBuilder")),
		g.QualifiedGoIdent(wazeroSysPackage.Ident("ExitError")),
		g.QualifiedGoIdent(fmtPackage.Ident("Errorf")),
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
		", request ", method.Input.GoIdent.GoName, ")", "(response ", method.Output.GoIdent.GoName, ", err error) {")

	errorHandling := "if err != nil {return response , err}"
	g.P("data, err := request.MarshalVT()")
	g.P(errorHandling)
	g.P("dataSize := uint64(len(data))")

	g.P("results, err := p.malloc.Call(ctx, dataSize)")
	g.P(errorHandling)

	g.P(`
			dataPtr := results[0]
			// This pointer is managed by TinyGo, but TinyGo is unaware of external usage.
			// So, we have to free it when finished
			defer p.free.Call(ctx, dataPtr)

			// The pointer is a linear memory offset, which is where we write the name.
			if !p.module.Memory().Write(ctx, uint32(dataPtr), data) {`)

	errorMsg := `fmt.Errorf("Memory.Write(%d, %d) out of range of memory size %d", dataPtr, dataSize, p.module.Memory().Size(ctx))`
	g.P("return response, ", errorMsg, "}")
	g.P("ptrSize, err := p.", strings.ToLower(method.GoName[:1]+method.GoName[1:]),
		".Call(ctx, dataPtr, dataSize)")
	g.P(errorHandling)
	g.P(`
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
