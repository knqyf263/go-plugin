package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/knqyf263/go-plugin/tests/host-functions/proto"
)

func TestHostFunctions(t *testing.T) {
	ctx := context.Background()
	p, err := proto.NewGreeterPlugin(ctx, proto.GreeterPluginOption{Stdout: os.Stdout})
	require.NoError(t, err)

	// Pass my host functions that are embedded into the plugin.
	plugin, err := p.Load(ctx, "plugin/plugin.wasm", myHostFunctions{})
	require.NoError(t, err)

	reply, err := plugin.Greet(ctx, proto.GreetRequest{
		Name: "Sato",
	})
	require.NoError(t, err)

	want := "Hello, Sato. This is Yamada (age 20)."
	assert.Equal(t, want, reply.GetMessage())
}

// myHostFunctions implements proto.HostFunctions
type myHostFunctions struct{}

// ParseJson is embedded into the plugin and can be called by the plugin.
func (myHostFunctions) ParseJson(ctx context.Context, request proto.ParseJsonRequest) (proto.ParseJsonResponse, error) {
	var person proto.Person
	if err := json.Unmarshal(request.GetContent(), &person); err != nil {
		return proto.ParseJsonResponse{}, err
	}

	return proto.ParseJsonResponse{Response: &person}, nil
}
