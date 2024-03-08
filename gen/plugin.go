package gen

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func (gg *Generator) generatePluginFile(f *fileInfo) {
	// This file will be imported by plugins written in TinyGo
	filename := f.GeneratedFilenamePrefix + "_plugin.pb.go"
	g := gg.plugin.NewGeneratedFile(filename, f.GoImportPath)

	if len(f.pluginServices) == 0 && f.hostService == nil {
		g.Skip()
	}

	// Build constraints
	g.P("//go:build tinygo.wasm")

	// Generate header
	gg.generateHeader(g, f)

	// Generate exported functions that wrap interfaces
	for _, service := range f.pluginServices {
		genPlugin(g, f, service)
	}

	genHostFunctions(g, f)
}

func genPlugin(g *protogen.GeneratedFile, f *fileInfo, service *serviceInfo) {
	serviceVar := strings.ToLower(service.GoName[:1]) + service.GoName[1:]

	// API version
	g.P("const ", service.GoName, "PluginAPIVersion = ", service.Version)
	g.P(fmt.Sprintf(`
		//export %s_api_version
		func _%s_api_version() uint64 {
			return %sPluginAPIVersion
		}`,
		toSnakeCase(service.GoName), toSnakeCase(service.GoName), service.GoName,
	))

	// Variable definition
	g.P("var ", serviceVar, " ", service.GoName)

	// Register function
	g.P("func Register", service.GoName, "(p ", service.GoName, ") {")
	g.P(serviceVar, "= p")
	g.P("}")

	// Exported functions
	for _, method := range service.Methods {
		exportedName := toSnakeCase(service.GoName + method.GoName)
		g.P("//export ", exportedName)
		g.P("func _", exportedName, "(ptr, size uint32) uint64 {")
		g.P("b := ", g.QualifiedGoIdent(pluginWasmPackage.Ident("PtrToByte")), "(ptr, size)")

		g.P("req := new(", g.QualifiedGoIdent(method.Input.GoIdent), ")")
		g.P(`if err := req.UnmarshalVT(b); err != nil {
						return 0
					  }`)
		g.P(fmt.Sprintf(`response, err := %s.%s(%s(), req)`,
			serviceVar, method.GoName, g.QualifiedGoIdent(contextPackage.Ident("Background"))))
		g.P(fmt.Sprintf(`if err != nil {
					ptr, size = %s([]byte(err.Error()))
					return (uint64(ptr) << uint64(32)) | uint64(size) |
						// Indicate that this is the error string by setting the 32-th bit, assuming that
						// no data exceeds 31-bit size (2 GiB).
						%s
				}

				b, err = response.MarshalVT()
				if err != nil {
					return 0
				}
				ptr, size = %s(b)
				return (uint64(ptr) << uint64(32)) | uint64(size)`,
			g.QualifiedGoIdent(pluginWasmPackage.Ident("ByteToPtr")),
			ErrorMaskBit,
			g.QualifiedGoIdent(pluginWasmPackage.Ident("ByteToPtr"))))
		g.P("}")
	}
}

func genHostFunctions(g *protogen.GeneratedFile, f *fileInfo) {
	if f.hostService == nil {
		return
	}

	g.Import(unsafePackage)

	// Host functions
	structName := strings.ToLower(f.hostService.GoName[:1]) + f.hostService.GoName[1:]
	g.P("type ", structName, " struct{}")
	g.P()
	g.P("func New", f.hostService.GoName, "()", f.hostService.GoName, "{")
	g.P("	return ", structName, "{}")
	g.P("}")

	for _, method := range f.hostService.Methods {
		importedName := toSnakeCase(method.GoName)
		g.P(fmt.Sprintf(`
		//go:wasmimport %s %s
		func _%s(ptr uint32, size uint32) uint64

		func (h %s) %s(ctx %s, request *%s) (*%s, error) {
			buf, err := request.MarshalVT()
			if err != nil {
				return nil, err
			}
			ptr, size := %s(buf)
			ptrSize := _%s(ptr, size)
			%s(ptr)

			ptr = uint32(ptrSize >> 32)
			size = uint32(ptrSize)
			buf = %s(ptr, size)

			response := new(%s)
			if err = response.UnmarshalVT(buf); err != nil {
				return nil, err
			}
			return response, nil
		}`,
			f.hostService.Module, importedName, importedName, structName, method.GoName,
			g.QualifiedGoIdent(contextPackage.Ident("Context")),
			g.QualifiedGoIdent(method.Input.GoIdent),
			g.QualifiedGoIdent(method.Output.GoIdent),
			g.QualifiedGoIdent(pluginWasmPackage.Ident("ByteToPtr")),
			importedName,
			g.QualifiedGoIdent(pluginWasmPackage.Ident("FreePtr")),
			g.QualifiedGoIdent(pluginWasmPackage.Ident("PtrToByte")),
			g.QualifiedGoIdent(method.Output.GoIdent),
		))
	}
}
