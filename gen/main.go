// Copyright 2018 The Go Authors. All rights reserved.
// Copyright 2022 Teppei Fukuda. All rights reserved.

package gen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	_ "github.com/planetscale/vtprotobuf/features/marshal"
	_ "github.com/planetscale/vtprotobuf/features/size"
	_ "github.com/planetscale/vtprotobuf/features/unmarshal"
	vtgenerator "github.com/planetscale/vtprotobuf/generator"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/knqyf263/go-plugin/encoding/tag"
	"github.com/knqyf263/go-plugin/genid"
	"github.com/knqyf263/go-plugin/version"
)

// SupportedFeatures reports the set of supported protobuf language features.
var SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

// GenerateVersionMarkers specifies whether to generate version markers.
var GenerateVersionMarkers = true

const (
	// Standard library dependencies.
	contextPackage = protogen.GoImportPath("context")
	errorsPackage  = protogen.GoImportPath("errors")
	fmtPackage     = protogen.GoImportPath("fmt")
	unsafePackage  = protogen.GoImportPath("unsafe")
	osPackage      = protogen.GoImportPath("os")
	ioPackage      = protogen.GoImportPath("io")
	fsPackage      = protogen.GoImportPath("io/fs")
	base64Package  = protogen.GoImportPath("encoding/base64")
	mathPackage    = protogen.GoImportPath("math")
	reflectPackage = protogen.GoImportPath("reflect")
	sortPackage    = protogen.GoImportPath("sort")
	stringsPackage = protogen.GoImportPath("strings")
	syncPackage    = protogen.GoImportPath("sync")
	timePackage    = protogen.GoImportPath("time")

	knownTypesPrefix = "google.golang.org/protobuf/types/known/"

	// ErrorMaskBit bit values to indicate if there is an error in the returned data.
	ErrorMaskBit = "(1 << 31)"
)

// Protobuf library dependencies.
//
// These are declared as an interface type so that they can be more easily
// patched to support unique build environments that impose restrictions
// on the dependencies of generated source code.
var (
	protoPackage         goImportPath = protogen.GoImportPath("google.golang.org/protobuf/proto")
	protoifacePackage    goImportPath = protogen.GoImportPath("google.golang.org/protobuf/runtime/protoiface")
	protoimplPackage     goImportPath = protogen.GoImportPath("google.golang.org/protobuf/runtime/protoimpl")
	protojsonPackage     goImportPath = protogen.GoImportPath("google.golang.org/protobuf/encoding/protojson")
	protoreflectPackage  goImportPath = protogen.GoImportPath("google.golang.org/protobuf/reflect/protoreflect")
	protoregistryPackage goImportPath = protogen.GoImportPath("google.golang.org/protobuf/reflect/protoregistry")

	wazeroPackage     goImportPath = protogen.GoImportPath("github.com/tetratelabs/wazero")
	wazeroAPIPackage  goImportPath = protogen.GoImportPath("github.com/tetratelabs/wazero/api")
	wazeroSysPackage  goImportPath = protogen.GoImportPath("github.com/tetratelabs/wazero/sys")
	wazeroWasiPackage goImportPath = protogen.GoImportPath("github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1")

	pluginWasmPackage goImportPath = protogen.GoImportPath("github.com/knqyf263/go-plugin/wasm")
)

type goImportPath interface {
	String() string
	Ident(string) protogen.GoIdent
}

type Generator struct {
	plugin *protogen.Plugin
	vtgen  *vtgenerator.Generator
}

func NewGenerator(plugin *protogen.Plugin) (*Generator, error) {
	ext := &vtgenerator.Extensions{}
	featureNames := []string{"marshal", "unmarshal", "size"}

	vtgen, err := vtgenerator.NewGenerator(plugin.Files, featureNames, ext)
	if err != nil {
		return nil, err
	}

	for _, f := range plugin.Files {
		if !f.Generate {
			continue
		}

		// Replace google known types with custom types so that TinyGo can build them.
		walkMessages(f.Messages, func(message *protogen.Message) {
			for _, field := range message.Fields {
				replaceImport(field.Message)
			}
			replaceImport(message)
		})
		for _, svc := range f.Services {
			for _, method := range svc.Methods {
				walkMessages([]*protogen.Message{method.Input, method.Output}, func(message *protogen.Message) {
					replaceImport(message)
				})
			}
		}
	}

	return &Generator{
		plugin: plugin,
		vtgen:  vtgen,
	}, nil
}

func replaceImport(m *protogen.Message) {
	if m == nil {
		return
	}
	if strings.HasPrefix(string(m.GoIdent.GoImportPath), knownTypesPrefix) {
		m.GoIdent.GoImportPath = protogen.GoImportPath(
			strings.ReplaceAll(string(m.GoIdent.GoImportPath),
				knownTypesPrefix, "github.com/knqyf263/go-plugin/types/known/"),
		)
	}
}

// GenerateFiles generates the contents of a .pb.go file.
func (gg *Generator) GenerateFiles(file *protogen.File) *protogen.GeneratedFile {
	f := gg.newFileInfo(file)
	gg.generatePBFile(f)
	gg.generateHostFile(f)
	gg.generatePluginFile(f)
	gg.generateVTFile(f)
	return nil
}

func (gg *Generator) generatePBFile(f *fileInfo) {
	filename := f.GeneratedFilenamePrefix + ".pb.go"
	g := gg.plugin.NewGeneratedFile(filename, f.GoImportPath)

	gg.generateHeader(g, f)

	// Emit a static check that enforces a minimum version of the proto package.
	if GenerateVersionMarkers {
		g.P("const (")
		g.P("// Verify that this generated code is sufficiently up-to-date.")
		g.P("_ = ", protoimplPackage.Ident("EnforceVersion"), "(", protoimpl.GenVersion, " - ", protoimplPackage.Ident("MinVersion"), ")")
		g.P("// Verify that runtime/protoimpl is sufficiently up-to-date.")
		g.P("_ = ", protoimplPackage.Ident("EnforceVersion"), "(", protoimplPackage.Ident("MaxVersion"), " - ", protoimpl.GenVersion, ")")
		g.P(")")
		g.P()
	}

	for i, imps := 0, f.Desc.Imports(); i < imps.Len(); i++ {
		gg.genImport(g, f, imps.Get(i))
	}
	for _, enum := range f.allEnums {
		genEnum(g, f, enum)
	}
	for _, message := range f.allMessages {
		genMessage(g, f, message)
	}
	for _, service := range f.allServices {
		genServiceInterface(g, f, service)
	}
}

func (gg *Generator) generateHeader(g *protogen.GeneratedFile, f *fileInfo) {
	genStandaloneComments(g, f, int32(genid.FileDescriptorProto_Syntax_field_number))
	genGeneratedHeader(gg.plugin, g, f)
	genStandaloneComments(g, f, int32(genid.FileDescriptorProto_Package_field_number))

	// TODO
	//packageDoc := genPackageKnownComment(f)
	g.P("package ", f.GoPackageName)
	g.P()
}

// genStandaloneComments prints all leading comments for a FileDescriptorProto
// location identified by the field number n.
func genStandaloneComments(g *protogen.GeneratedFile, f *fileInfo, n int32) {
	loc := f.Desc.SourceLocations().ByPath(protoreflect.SourcePath{n})
	for _, s := range loc.LeadingDetachedComments {
		g.P(protogen.Comments(s))
		g.P()
	}
	if s := loc.LeadingComments; s != "" {
		g.P(protogen.Comments(s))
		g.P()
	}
}

func genGeneratedHeader(gen *protogen.Plugin, g *protogen.GeneratedFile, f *fileInfo) {
	g.P("// Code generated by protoc-gen-go-plugin. DO NOT EDIT.")

	if GenerateVersionMarkers {
		g.P("// versions:")
		protocGenGoVersion := version.Version
		protocVersion := "(unknown)"
		if v := gen.Request.GetCompilerVersion(); v != nil {
			protocVersion = fmt.Sprintf("v%v.%v.%v", v.GetMajor(), v.GetMinor(), v.GetPatch())
		}
		g.P("// \tprotoc-gen-go-plugin ", protocGenGoVersion)
		g.P("// \tprotoc               ", protocVersion)
	}

	if f.Proto.GetOptions().GetDeprecated() {
		g.P("// ", f.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", f.Desc.Path())
	}
	g.P()
}

func (gg *Generator) genImport(g *protogen.GeneratedFile, f *fileInfo, imp protoreflect.FileImport) {
	impFile, ok := gg.plugin.FilesByPath[imp.Path()]
	if !ok {
		return
	}
	if impFile.GoImportPath == f.GoImportPath {
		// Don't generate imports or aliases for types in the same Go package.
		return
	}

	if strings.HasPrefix(string(impFile.GoImportPath), knownTypesPrefix) {
		// Don't generate imports for well-known types as it cannot be compiled by TinyGo.
		return
	}
	// Generate imports for all non-weak dependencies, even if they are not
	// referenced, because other code and tools depend on having the
	// full transitive closure of protocol buffer types in the binary.
	if !imp.IsWeak {
		g.Import(impFile.GoImportPath)
	}
	if !imp.IsPublic {
		return
	}

	// Generate public imports by generating the imported file, parsing it,
	// and extracting every symbol that should receive a forwarding declaration.
	impGen := gg.GenerateFiles(impFile)
	impGen.Skip()
	b, err := impGen.Content()
	if err != nil {
		gg.plugin.Error(err)
		return
	}
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "", b, parser.ParseComments)
	if err != nil {
		gg.plugin.Error(err)
		return
	}
	genForward := func(tok token.Token, name string, expr ast.Expr) {
		// Don't import unexported symbols.
		r, _ := utf8.DecodeRuneInString(name)
		if !unicode.IsUpper(r) {
			return
		}
		// Don't import the FileDescriptor.
		if name == impFile.GoDescriptorIdent.GoName {
			return
		}
		// Don't import decls referencing a symbol defined in another package.
		// i.e., don't import decls which are themselves public imports:
		//
		//	type T = somepackage.T
		if _, ok := expr.(*ast.SelectorExpr); ok {
			return
		}
		g.P(tok, " ", name, " = ", impFile.GoImportPath.Ident(name))
	}
	g.P("// Symbols defined in public import of ", imp.Path(), ".")
	g.P()
	for _, decl := range astFile.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					genForward(decl.Tok, spec.Name.Name, spec.Type)
				case *ast.ValueSpec:
					for i, name := range spec.Names {
						var expr ast.Expr
						if i < len(spec.Values) {
							expr = spec.Values[i]
						}
						genForward(decl.Tok, name.Name, expr)
					}
				case *ast.ImportSpec:
				default:
					panic(fmt.Sprintf("can't generate forward for spec type %T", spec))
				}
			}
		}
	}
	g.P()
}

func genEnum(g *protogen.GeneratedFile, f *fileInfo, e *enumInfo) {
	// Enum type declaration.
	g.Annotate(e.GoIdent.GoName, e.Location)
	leadingComments := appendDeprecationSuffix(e.Comments.Leading,
		e.Desc.Options().(*descriptorpb.EnumOptions).GetDeprecated())
	g.P(leadingComments,
		"type ", e.GoIdent, " int32")

	// Enum value constants.
	g.P("const (")
	for _, value := range e.Values {
		g.Annotate(value.GoIdent.GoName, value.Location)
		leadingComments := appendDeprecationSuffix(value.Comments.Leading,
			value.Desc.Options().(*descriptorpb.EnumValueOptions).GetDeprecated())
		g.P(leadingComments,
			value.GoIdent, " ", e.GoIdent, " = ", value.Desc.Number(),
			trailingComment(value.Comments.Trailing))
	}
	g.P(")")
	g.P()

	// Enum value maps.
	g.P("// Enum value maps for ", e.GoIdent, ".")
	g.P("var (")
	g.P(e.GoIdent.GoName+"_name", " = map[int32]string{")
	for _, value := range e.Values {
		duplicate := ""
		if value.Desc != e.Desc.Values().ByNumber(value.Desc.Number()) {
			duplicate = "// Duplicate value: "
		}
		g.P(duplicate, value.Desc.Number(), ": ", strconv.Quote(string(value.Desc.Name())), ",")
	}
	g.P("}")
	g.P(e.GoIdent.GoName+"_value", " = map[string]int32{")
	for _, value := range e.Values {
		g.P(strconv.Quote(string(value.Desc.Name())), ": ", value.Desc.Number(), ",")
	}
	g.P("}")
	g.P(")")
	g.P()

	// Enum method.
	//
	// NOTE: A pointer value is needed to represent presence in proto2.
	// Since a proto2 message can reference a proto3 enum, it is useful to
	// always generate this method (even on proto3 enums) to support that case.
	g.P("func (x ", e.GoIdent, ") Enum() *", e.GoIdent, " {")
	g.P("p := new(", e.GoIdent, ")")
	g.P("*p = x")
	g.P("return p")
	g.P("}")
	g.P()
}

func genMessage(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	if m.Desc.IsMapEntry() {
		return
	}

	// Message type declaration.
	g.Annotate(m.GoIdent.GoName, m.Location)
	leadingComments := appendDeprecationSuffix(m.Comments.Leading,
		m.Desc.Options().(*descriptorpb.MessageOptions).GetDeprecated())
	g.P(leadingComments,
		"type ", m.GoIdent, " struct {")
	genMessageFields(g, f, m)
	g.P("}")
	g.P()

	genMessageDefaultDecls(g, f, m)
	genMessageMethods(g, f, m)
	genMessageOneofWrapperTypes(g, f, m)
}

func genMessageFields(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	sf := f.allMessageFieldsByPtr[m]
	genMessageInternalFields(g, f, m, sf)
	for _, field := range m.Fields {
		genMessageField(g, f, m, field, sf)
	}
}

func genMessageInternalFields(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo, sf *structFields) {
	g.P(genid.State_goname, " ", protoimplPackage.Ident("MessageState"))
	sf.append(genid.State_goname)
	g.P(genid.SizeCache_goname, " ", protoimplPackage.Ident("SizeCache"))
	sf.append(genid.SizeCache_goname)
	if m.hasWeak {
		g.P(genid.WeakFields_goname, " ", protoimplPackage.Ident("WeakFields"))
		sf.append(genid.WeakFields_goname)
	}
	g.P(genid.UnknownFields_goname, " ", protoimplPackage.Ident("UnknownFields"))
	sf.append(genid.UnknownFields_goname)
	if sf.count > 0 {
		g.P()
	}
}

func genMessageField(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo, field *protogen.Field, sf *structFields) {
	if oneof := field.Oneof; oneof != nil && !oneof.Desc.IsSynthetic() {
		// It would be a bit simpler to iterate over the oneofs below,
		// but generating the field here keeps the contents of the Go
		// struct in the same order as the contents of the source
		// .proto file.
		if oneof.Fields[0] != field {
			return // only generate for first appearance
		}

		tags := structTags{
			{"protobuf_oneof", string(oneof.Desc.Name())},
		}
		if m.isTracked {
			tags = append(tags, gotrackTags...)
		}

		g.Annotate(m.GoIdent.GoName+"."+oneof.GoName, oneof.Location)
		leadingComments := oneof.Comments.Leading
		if leadingComments != "" {
			leadingComments += "\n"
		}
		ss := []string{fmt.Sprintf(" Types that are assignable to %s:\n", oneof.GoName)}
		for _, field := range oneof.Fields {
			ss = append(ss, "\t*"+field.GoIdent.GoName+"\n")
		}
		leadingComments += protogen.Comments(strings.Join(ss, ""))
		g.P(leadingComments,
			oneof.GoName, " ", oneofInterfaceName(oneof), tags)
		sf.append(oneof.GoName)
		return
	}
	goType, pointer := fieldGoType(g, f, field)
	if pointer {
		goType = "*" + goType
	}
	tags := structTags{
		{"protobuf", fieldProtobufTagValue(field)},
		{"json", fieldJSONTagValue(field)},
	}
	if field.Desc.IsMap() {
		key := field.Message.Fields[0]
		val := field.Message.Fields[1]
		tags = append(tags, structTags{
			{"protobuf_key", fieldProtobufTagValue(key)},
			{"protobuf_val", fieldProtobufTagValue(val)},
		}...)
	}
	if m.isTracked {
		tags = append(tags, gotrackTags...)
	}

	name := field.GoName
	if field.Desc.IsWeak() {
		name = genid.WeakFieldPrefix_goname + name
	}
	g.Annotate(m.GoIdent.GoName+"."+name, field.Location)
	leadingComments := appendDeprecationSuffix(field.Comments.Leading,
		field.Desc.Options().(*descriptorpb.FieldOptions).GetDeprecated())
	g.P(leadingComments,
		name, " ", goType, tags,
		trailingComment(field.Comments.Trailing))
	sf.append(field.GoName)
}

// genMessageDefaultDecls generates consts and vars holding the default
// values of fields.
func genMessageDefaultDecls(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	var consts, vars []string
	for _, field := range m.Fields {
		if !field.Desc.HasDefault() {
			continue
		}
		name := "Default_" + m.GoIdent.GoName + "_" + field.GoName
		goType, _ := fieldGoType(g, f, field)
		defVal := field.Desc.Default()
		switch field.Desc.Kind() {
		case protoreflect.StringKind:
			consts = append(consts, fmt.Sprintf("%s = %s(%q)", name, goType, defVal.String()))
		case protoreflect.BytesKind:
			vars = append(vars, fmt.Sprintf("%s = %s(%q)", name, goType, defVal.Bytes()))
		case protoreflect.EnumKind:
			idx := field.Desc.DefaultEnumValue().Index()
			val := field.Enum.Values[idx]
			consts = append(consts, fmt.Sprintf("%s = %s", name, g.QualifiedGoIdent(val.GoIdent)))
		case protoreflect.FloatKind, protoreflect.DoubleKind:
			if f := defVal.Float(); math.IsNaN(f) || math.IsInf(f, 0) {
				var fn, arg string
				switch f := defVal.Float(); {
				case math.IsInf(f, -1):
					fn, arg = g.QualifiedGoIdent(mathPackage.Ident("Inf")), "-1"
				case math.IsInf(f, +1):
					fn, arg = g.QualifiedGoIdent(mathPackage.Ident("Inf")), "+1"
				case math.IsNaN(f):
					fn, arg = g.QualifiedGoIdent(mathPackage.Ident("NaN")), ""
				}
				vars = append(vars, fmt.Sprintf("%s = %s(%s(%s))", name, goType, fn, arg))
			} else {
				consts = append(consts, fmt.Sprintf("%s = %s(%v)", name, goType, f))
			}
		default:
			consts = append(consts, fmt.Sprintf("%s = %s(%v)", name, goType, defVal.Interface()))
		}
	}
	if len(consts) > 0 {
		g.P("// Default values for ", m.GoIdent, " fields.")
		g.P("const (")
		for _, s := range consts {
			g.P(s)
		}
		g.P(")")
	}
	if len(vars) > 0 {
		g.P("// Default values for ", m.GoIdent, " fields.")
		g.P("var (")
		for _, s := range vars {
			g.P(s)
		}
		g.P(")")
	}
	g.P()
}

func genMessageMethods(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	genMessageReflectMethods(g, f, m)
	genMessageGetterMethods(g, f, m)
	genMessageSetterMethods(g, f, m)
}

func genMessageGetterMethods(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	for _, field := range m.Fields {
		genNoInterfacePragma(g, m.isTracked)

		// Getter for parent oneof.
		if oneof := field.Oneof; oneof != nil && oneof.Fields[0] == field && !oneof.Desc.IsSynthetic() {
			g.Annotate(m.GoIdent.GoName+".Get"+oneof.GoName, oneof.Location)
			g.P("func (m *", m.GoIdent.GoName, ") Get", oneof.GoName, "() ", oneofInterfaceName(oneof), " {")
			g.P("if m != nil {")
			g.P("return m.", oneof.GoName)
			g.P("}")
			g.P("return nil")
			g.P("}")
			g.P()
		}

		// Getter for message field.
		goType, pointer := fieldGoType(g, f, field)
		defaultValue := fieldDefaultValue(g, m, field)
		g.Annotate(m.GoIdent.GoName+".Get"+field.GoName, field.Location)
		leadingComments := appendDeprecationSuffix("",
			field.Desc.Options().(*descriptorpb.FieldOptions).GetDeprecated())
		switch {
		case field.Desc.IsWeak():
			g.P(leadingComments, "func (x *", m.GoIdent, ") Get", field.GoName, "() ", protoPackage.Ident("Message"), "{")
			g.P("var w ", protoimplPackage.Ident("WeakFields"))
			g.P("if x != nil {")
			g.P("w = x.", genid.WeakFields_goname)
			if m.isTracked {
				g.P("_ = x.", genid.WeakFieldPrefix_goname+field.GoName)
			}
			g.P("}")
			g.P("return ", protoimplPackage.Ident("X"), ".GetWeak(w, ", field.Desc.Number(), ", ", strconv.Quote(string(field.Message.Desc.FullName())), ")")
			g.P("}")
		case field.Oneof != nil && !field.Oneof.Desc.IsSynthetic():
			g.P(leadingComments, "func (x *", m.GoIdent, ") Get", field.GoName, "() ", goType, " {")
			g.P("if x, ok := x.Get", field.Oneof.GoName, "().(*", field.GoIdent, "); ok {")
			g.P("return x.", field.GoName)
			g.P("}")
			g.P("return ", defaultValue)
			g.P("}")
		default:
			g.P(leadingComments, "func (x *", m.GoIdent, ") Get", field.GoName, "() ", goType, " {")
			if !field.Desc.HasPresence() || defaultValue == "nil" {
				g.P("if x != nil {")
			} else {
				g.P("if x != nil && x.", field.GoName, " != nil {")
			}
			star := ""
			if pointer {
				star = "*"
			}
			g.P("return ", star, " x.", field.GoName)
			g.P("}")
			g.P("return ", defaultValue)
			g.P("}")
		}
		g.P()
	}
}

func genMessageSetterMethods(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	for _, field := range m.Fields {
		if !field.Desc.IsWeak() {
			continue
		}

		genNoInterfacePragma(g, m.isTracked)

		g.Annotate(m.GoIdent.GoName+".Set"+field.GoName, field.Location)
		leadingComments := appendDeprecationSuffix("",
			field.Desc.Options().(*descriptorpb.FieldOptions).GetDeprecated())
		g.P(leadingComments, "func (x *", m.GoIdent, ") Set", field.GoName, "(v ", protoPackage.Ident("Message"), ") {")
		g.P("var w *", protoimplPackage.Ident("WeakFields"))
		g.P("if x != nil {")
		g.P("w = &x.", genid.WeakFields_goname)
		if m.isTracked {
			g.P("_ = x.", genid.WeakFieldPrefix_goname+field.GoName)
		}
		g.P("}")
		g.P(protoimplPackage.Ident("X"), ".SetWeak(w, ", field.Desc.Number(), ", ", strconv.Quote(string(field.Message.Desc.FullName())), ", v)")
		g.P("}")
		g.P()
	}
}

// fieldGoType returns the Go type used for a field.
//
// If it returns pointer=true, the struct field is a pointer to the type.
func fieldGoType(g *protogen.GeneratedFile, f *fileInfo, field *protogen.Field) (goType string, pointer bool) {
	if field.Desc.IsWeak() {
		return "struct{}", false
	}

	pointer = field.Desc.HasPresence()
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		goType = "bool"
	case protoreflect.EnumKind:
		goType = g.QualifiedGoIdent(field.Enum.GoIdent)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		goType = "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		goType = "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		goType = "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		goType = "uint64"
	case protoreflect.FloatKind:
		goType = "float32"
	case protoreflect.DoubleKind:
		goType = "float64"
	case protoreflect.StringKind:
		goType = "string"
	case protoreflect.BytesKind:
		goType = "[]byte"
		pointer = false // rely on nullability of slices for presence
	case protoreflect.MessageKind, protoreflect.GroupKind:
		goType = "*" + g.QualifiedGoIdent(field.Message.GoIdent)
		pointer = false // pointer captured as part of the type
	}
	switch {
	case field.Desc.IsList():
		return "[]" + goType, false
	case field.Desc.IsMap():
		keyType, _ := fieldGoType(g, f, field.Message.Fields[0])
		valType, _ := fieldGoType(g, f, field.Message.Fields[1])
		return fmt.Sprintf("map[%v]%v", keyType, valType), false
	}
	return goType, pointer
}

func fieldProtobufTagValue(field *protogen.Field) string {
	var enumName string
	if field.Desc.Kind() == protoreflect.EnumKind {
		enumName = protoimpl.X.LegacyEnumName(field.Enum.Desc)
	}
	return tag.Marshal(field.Desc, enumName)
}

func fieldDefaultValue(g *protogen.GeneratedFile, m *messageInfo, field *protogen.Field) string {
	if field.Desc.IsList() {
		return "nil"
	}
	if field.Desc.HasDefault() {
		defVarName := "Default_" + m.GoIdent.GoName + "_" + field.GoName
		if field.Desc.Kind() == protoreflect.BytesKind {
			return "append([]byte(nil), " + defVarName + "...)"
		}
		return defVarName
	}
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "false"
	case protoreflect.StringKind:
		return `""`
	case protoreflect.MessageKind, protoreflect.GroupKind, protoreflect.BytesKind:
		return "nil"
	case protoreflect.EnumKind:
		return g.QualifiedGoIdent(field.Enum.Values[0].GoIdent)
	default:
		return "0"
	}
}

func fieldJSONTagValue(field *protogen.Field) string {
	return string(field.Desc.Name()) + ",omitempty"
}

func genServiceInterface(g *protogen.GeneratedFile, f *fileInfo, s *serviceInfo) {
	if s.Type == ServiceUnknown || s.Type == ServiceNone {
		return
	}

	// Service interface declaration.
	g.Annotate(s.GoName, s.Location)
	g.P(s.Comments.Leading,
		"type ", s.GoName, " interface{")
	genServiceInterfaceMethod(g, f, s)
	g.P("}")

}

func genServiceInterfaceMethod(g *protogen.GeneratedFile, f *fileInfo, s *serviceInfo) {
	for _, method := range s.Methods {
		g.P(method.Comments.Leading, method.GoName,
			"(", g.QualifiedGoIdent(contextPackage.Ident("Context")), ", ", g.QualifiedGoIdent(method.Input.GoIdent), ") ",
			"(", g.QualifiedGoIdent(method.Output.GoIdent), ", error)")
	}
}

// genMessageOneofWrapperTypes generates the oneof wrapper types and
// associates the types with the parent message type.
func genMessageOneofWrapperTypes(g *protogen.GeneratedFile, f *fileInfo, m *messageInfo) {
	for _, oneof := range m.Oneofs {
		if oneof.Desc.IsSynthetic() {
			continue
		}
		ifName := oneofInterfaceName(oneof)
		g.P("type ", ifName, " interface {")
		g.P(ifName, "()")
		g.P("}")
		g.P()
		for _, field := range oneof.Fields {
			g.Annotate(field.GoIdent.GoName, field.Location)
			g.Annotate(field.GoIdent.GoName+"."+field.GoName, field.Location)
			g.P("type ", field.GoIdent, " struct {")
			goType, _ := fieldGoType(g, f, field)
			tags := structTags{
				{"protobuf", fieldProtobufTagValue(field)},
			}
			if m.isTracked {
				tags = append(tags, gotrackTags...)
			}
			leadingComments := appendDeprecationSuffix(field.Comments.Leading,
				field.Desc.Options().(*descriptorpb.FieldOptions).GetDeprecated())
			g.P(leadingComments,
				field.GoName, " ", goType, tags,
				trailingComment(field.Comments.Trailing))
			g.P("}")
			g.P()
		}
		for _, field := range oneof.Fields {
			g.P("func (*", field.GoIdent, ") ", ifName, "() {}")
			g.P()
		}
	}
}

// oneofInterfaceName returns the name of the interface type implemented by
// the oneof field value types.
func oneofInterfaceName(oneof *protogen.Oneof) string {
	return "is" + oneof.GoIdent.GoName
}

// genNoInterfacePragma generates a standalone "nointerface" pragma to
// decorate methods with field-tracking support.
func genNoInterfacePragma(g *protogen.GeneratedFile, tracked bool) {
	if tracked {
		g.P("//go:nointerface")
		g.P()
	}
}

var gotrackTags = structTags{{"go", "track"}}

// structTags is a data structure for build idiomatic Go struct tags.
// Each [2]string is a key-value pair, where value is the unescaped string.
//
// Example: structTags{{"key", "value"}}.String() -> `key:"value"`
type structTags [][2]string

func (tags structTags) String() string {
	if len(tags) == 0 {
		return ""
	}
	var ss []string
	for _, tag := range tags {
		// NOTE: When quoting the value, we need to make sure the backtick
		// character does not appear. Convert all cases to the escaped hex form.
		key := tag[0]
		val := strings.Replace(strconv.Quote(tag[1]), "`", `\x60`, -1)
		ss = append(ss, fmt.Sprintf("%s:%s", key, val))
	}
	return "`" + strings.Join(ss, " ") + "`"
}

// appendDeprecationSuffix optionally appends a deprecation notice as a suffix.
func appendDeprecationSuffix(prefix protogen.Comments, deprecated bool) protogen.Comments {
	if !deprecated {
		return prefix
	}
	if prefix != "" {
		prefix += "\n"
	}
	return prefix + " Deprecated: Do not use.\n"
}

// trailingComment is like protogen.Comments, but lacks a trailing newline.
type trailingComment protogen.Comments

func (c trailingComment) String() string {
	s := strings.TrimSuffix(protogen.Comments(c).String(), "\n")
	if strings.Contains(s, "\n") {
		// We don't support multi-lined trailing comments as it is unclear
		// how to best render them in the generated code.
		return ""
	}
	return s
}

type Parameter struct {
	APIVersion int
	Type       ServiceType
}

type ServiceType string

const (
	ServiceHost    ServiceType = "host"
	ServicePlugin  ServiceType = "plugin"
	ServiceUnknown ServiceType = "unknown"
	ServiceNone    ServiceType = "none"
)

// parseParam parses a comment and extract parameters for go-plugin
// e.g. // go:plugin type=plugin version=2
func parseParam(comment string) (Parameter, error) {
	param := Parameter{
		APIVersion: 1,
		Type:       ServiceNone,
	}
	for _, line := range strings.Split(comment, "\n") {
		line = strings.TrimPrefix(line, "//")
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "go:plugin") {
			continue
		}
		line = strings.TrimPrefix(line, "go:plugin")

		for _, field := range strings.Fields(line) {
			var key, value string
			if i := strings.Index(field, "="); i >= 0 {
				value = field[i+1:]
				key = field[0:i]
			}
			if key == "" || value == "" {
				continue
			}
			switch key {
			case "type":
				switch value {
				case "host":
					param.Type = ServiceHost
				case "plugin":
					param.Type = ServicePlugin
				default:
					param.Type = ServiceUnknown
				}
			case "version":
				ver, err := strconv.Atoi(value)
				if err != nil {
					return Parameter{}, fmt.Errorf("invalid version: %w", err)
				}
				param.APIVersion = ver
			}
		}
	}
	return param, nil
}
