package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/knqyf263/go-plugin/tests/host-functions/proto"
)

func TestHostFunctions(t *testing.T) {
	ctx := context.Background()
	mc := wazero.NewModuleConfig().WithStdout(os.Stdout)
	p, err := proto.NewGreeterPlugin(ctx, proto.WazeroRuntime(func(ctx context.Context) (wazero.Runtime, error) {
		r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCompilationCache(wazero.NewCompilationCache()))
		if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
			return nil, err
		}

		return r, nil
	}), proto.WazeroModuleConfig(mc))
	require.NoError(t, err)

	// Pass my host functions that are embedded into the plugin.
	plugin, err := p.Load(ctx, "plugin/plugin.wasm", myHostFunctions{})
	require.NoError(t, err)
	defer plugin.Close(ctx)

	reply, err := plugin.Greet(ctx, proto.GreetRequest{
		Name: "Sato",
	})
	require.NoError(t, err)

	want := "Hello, Sato. This is Yamada (age 20)."
	assert.Equal(t, want, reply.GetMessage())
}

// myHostFunctions implements proto.HostFunctions
type myHostFunctions struct{}

var _ proto.HostFunctions = (*myHostFunctions)(nil)

// ParseJson is embedded into the plugin and can be called by the plugin.
func (myHostFunctions) ParseJson(_ context.Context, request proto.ParseJsonRequest) (proto.ParseJsonResponse, error) {
	var person proto.Person
	if err := json.Unmarshal(request.GetContent(), &person); err != nil {
		return proto.ParseJsonResponse{}, err
	}

	return proto.ParseJsonResponse{Response: &person}, nil
}
