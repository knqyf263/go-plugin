package options

import (
	"context"

	"github.com/tetratelabs/wazero"
)

type WazeroConfigOption func(plugin *WazeroConfig)

type WazeroNewRuntime func(context.Context) (wazero.Runtime, error)

type WazeroConfig struct {
	NewRuntime   func(context.Context) (wazero.Runtime, error)
	ModuleConfig wazero.ModuleConfig
}

func NewWazeroConfig() *WazeroConfig {
	return &WazeroConfig{
		NewRuntime: func(ctx context.Context) (wazero.Runtime, error) {
			return wazero.NewRuntime(ctx), nil
		},
		ModuleConfig: wazero.NewModuleConfig(),
	}
}

func WazeroRuntime(newRuntime WazeroNewRuntime) WazeroConfigOption {
	return func(h *WazeroConfig) {
		h.NewRuntime = newRuntime
	}
}

func WazeroModuleConfig(moduleConfig wazero.ModuleConfig) WazeroConfigOption {
	return func(h *WazeroConfig) {
		h.ModuleConfig = moduleConfig
	}
}
