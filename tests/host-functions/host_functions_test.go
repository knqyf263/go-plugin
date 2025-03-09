package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/knqyf263/go-plugin/tests/host-functions/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
)

func TestHostFunctions(t *testing.T) {
	ctx := context.Background()
	mc := wazero.NewModuleConfig().WithStdout(os.Stdout).WithStartFunctions("_initialize")
	p, err := proto.NewGreeterPlugin(ctx, proto.WazeroRuntime(func(ctx context.Context) (wazero.Runtime, error) {
		return proto.DefaultWazeroRuntime()(ctx)
	}), proto.WazeroModuleConfig(mc))
	require.NoError(t, err)

	// Pass my host functions that are embedded into the plugin.
	plugin, err := p.Load(ctx, "plugin/plugin.wasm", myHostFunctions{})
	require.NoError(t, err)
	defer plugin.Close(ctx)

	reply, err := plugin.Greet(ctx, &proto.GreetRequest{
		Name: "Sato",
	})
	require.NoError(t, err)

	want := "Hello, Sato. This is Yamada (age 20)."
	assert.Equal(t, want, reply.GetMessage())
}

func TestEmptyRequest(t *testing.T) {
	ctx := context.Background()
	p, err := proto.NewGreeterPlugin(ctx)
	require.NoError(t, err)

	plugin, err := p.Load(ctx, "plugin-empty/plugin.wasm", myEmptyHostFunctions{})
	require.NoError(t, err)
	defer plugin.Close(ctx)

	reply, err := plugin.Greet(ctx, nil)
	require.NoError(t, err)
	want := "Hello, empty request '' and empty '' host function request"
	assert.Equal(t, want, reply.GetMessage())
}

// myHostFunctions implements proto.HostFunctions
type myHostFunctions struct{}

var _ proto.HostFunctions = (*myHostFunctions)(nil)

// ParseJson is embedded into the plugin and can be called by the plugin.
func (myHostFunctions) ParseJson(_ context.Context, request *proto.ParseJsonRequest) (*proto.ParseJsonResponse, error) {
	var person proto.Person
	if err := json.Unmarshal(request.GetContent(), &person); err != nil {
		return nil, err
	}

	return &proto.ParseJsonResponse{Response: &person}, nil
}

type myEmptyHostFunctions struct{}

var _ proto.HostFunctions = (*myEmptyHostFunctions)(nil)

// ParseJson is embedded into the plugin and can be called by the plugin.
func (myEmptyHostFunctions) ParseJson(_ context.Context, _ *proto.ParseJsonRequest) (*proto.ParseJsonResponse, error) {
	return &proto.ParseJsonResponse{Response: &proto.Person{}}, nil
}
