//go:build !tinygo.wasm

// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v3.21.12
// source: examples/helloworld/greeting/greet.proto

package greeting

import (
	context "context"
	wazero "github.com/tetratelabs/wazero"
)

type wazeroConfigOption func(plugin *WazeroConfig)

type WazeroNewRuntime func(context.Context) (wazero.Runtime, error)

type WazeroConfig struct {
	newRuntime   func(context.Context) (wazero.Runtime, error)
	moduleConfig wazero.ModuleConfig
}

func WazeroRuntime(newRuntime WazeroNewRuntime) wazeroConfigOption {
	return func(h *WazeroConfig) {
		h.newRuntime = newRuntime
	}
}

func WazeroModuleConfig(moduleConfig wazero.ModuleConfig) wazeroConfigOption {
	return func(h *WazeroConfig) {
		h.moduleConfig = moduleConfig
	}
}
