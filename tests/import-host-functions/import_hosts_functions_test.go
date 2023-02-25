//go:build !tinygo.wasm

package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"

	"github.com/knqyf263/go-plugin/tests/import-host-functions/host/foo/export"
	"github.com/knqyf263/go-plugin/tests/import-host-functions/host/foo/impl"
	"github.com/knqyf263/go-plugin/tests/import-host-functions/proto"
)

func TestHostFunctions(t *testing.T) {
	ctx := context.Background()
	p, err := proto.NewGreeterPlugin(ctx, proto.GreeterPluginOption{Stdout: os.Stdout})
	require.NoError(t, err)
	defer p.Close(ctx)

	// Pass my host functions that are embedded into the plugin.
	plugin, err := p.Load(ctx, "plugin/plugin.wasm", myHostFunctions{}, func(ctx context.Context, runtime wazero.Runtime) error {
		return export.Instantiate(ctx, runtime, impl.FooHostFunctions{})
	})
	require.NoError(t, err)

	reply, err := plugin.Greet(ctx, proto.GreetRequest{
		Name: "Sato",
	})
	require.NoError(t, err)

	want := "Hello, Sato. This is Yamada-san (age 20)."
	assert.Equal(t, want, reply.GetMessage())
}

type myHostFunctions struct{}

var _ proto.HostFunctions = myHostFunctions{}

func (m myHostFunctions) San(ctx context.Context, request proto.SanRequest) (proto.SanResponse, error) {
	return proto.SanResponse{Message: fmt.Sprintf("%s-san", request.GetMessage())}, nil
}
