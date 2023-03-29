package gen

import (
	"fmt"
)

func (gg *Generator) generateOptionsFile(f *fileInfo) {
	filename := f.GeneratedFilenamePrefix + "_options.pb.go"
	g := gg.plugin.NewGeneratedFile(filename, f.GoImportPath)

	if len(f.pluginServices) == 0 {
		g.Skip()
	}

	// Build constraints
	g.P("//go:build !tinygo.wasm")
	gg.generateHeader(g, f)

	g.P(fmt.Sprintf(`type wazeroConfigOption func(plugin *WazeroConfig)

	type WazeroNewRuntime func(%s) (%s, error)

	type WazeroConfig struct {
		newRuntime   func(%s) (%s, error)
		moduleConfig %s
	}

	func WazeroRuntime(newRuntime WazeroNewRuntime) wazeroConfigOption {
		return func(h *WazeroConfig) {
			h.newRuntime = newRuntime
		}
	}

	func DefaultWazeroRuntime() WazeroNewRuntime {
		return func(ctx %s) (%s, error) {
			r := %s(ctx)
			if _, err := %s(ctx, r); err != nil {
				return nil, err
			}

			return r, nil
		}
	}

	func WazeroModuleConfig(moduleConfig %s) wazeroConfigOption {
		return func(h *WazeroConfig) {
			h.moduleConfig = moduleConfig
		}
	}
`,
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		g.QualifiedGoIdent(wazeroPackage.Ident("Runtime")),
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		g.QualifiedGoIdent(wazeroPackage.Ident("Runtime")),
		g.QualifiedGoIdent(wazeroPackage.Ident("ModuleConfig")),
		g.QualifiedGoIdent(contextPackage.Ident("Context")),
		g.QualifiedGoIdent(wazeroPackage.Ident("Runtime")),
		g.QualifiedGoIdent(wazeroPackage.Ident("NewRuntime")),
		g.QualifiedGoIdent(wazeroWasiPackage.Ident("Instantiate")),
		g.QualifiedGoIdent(wazeroPackage.Ident("ModuleConfig")),
	))
}
