//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"

	"github.com/knqyf263/go-plugin/tests/import-host-functions/host/foo/export"
	"github.com/knqyf263/go-plugin/tests/import-host-functions/proto"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	proto.RegisterGreeter(TestPlugin{})
}

type TestPlugin struct{}

func (p TestPlugin) Greet(ctx context.Context, request proto.GreetRequest) (proto.GreetReply, error) {
	fooHostsFunctions := export.NewForeignHostFunctions()
	localHostFunctions := proto.NewHostFunctions()
	sanr, err := localHostFunctions.San(ctx, proto.SanRequest{Message: "Yamada"})
	if err != nil {
		return proto.GreetReply{}, err
	}

	// Call the host function to parse JSON
	resp, err := fooHostsFunctions.ParseJson(ctx, export.ParseJsonRequest{
		Content: []byte(fmt.Sprintf(`{"name": "%s", "age": 20}`, sanr.Message)),
	})
	if err != nil {
		return proto.GreetReply{}, err
	}

	return proto.GreetReply{
		Message: fmt.Sprintf("Hello, %s. This is %s (age %d).",
			request.GetName(), resp.GetResponse().GetName(), resp.GetResponse().GetAge()),
	}, nil
}
