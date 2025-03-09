//go:build wasip1

package main

import (
	"context"
	"fmt"

	"github.com/knqyf263/go-plugin/examples/host-function-library/library/json-parser/export"
	"github.com/knqyf263/go-plugin/examples/host-function-library/proto"
)

// main is required for Go to compile to Wasm.
func main() {}

func init() {
	proto.RegisterGreeter(TestPlugin{})
}

type TestPlugin struct{}

var _ proto.Greeter = (*TestPlugin)(nil)

func (p TestPlugin) Greet(ctx context.Context, request *proto.GreetRequest) (*proto.GreetReply, error) {
	parserLibrary := export.NewParserLibrary()
	localHostFunctions := proto.NewHostFunctions()

	sanrequest, err := localHostFunctions.San(ctx, &proto.SanRequest{Message: "Yamada"})
	if err != nil {
		return nil, err
	}

	// Call the host function to parse JSON
	resp, err := parserLibrary.ParseJson(ctx, &export.ParseJsonRequest{
		Content: []byte(fmt.Sprintf(`{"name": "%s", "age": 20}`, sanrequest.Message)),
	})
	if err != nil {
		return nil, err
	}

	return &proto.GreetReply{
		Message: fmt.Sprintf("Hello, %s. This is %s (age %d).",
			request.GetName(), resp.GetResponse().GetName(), resp.GetResponse().GetAge()),
	}, nil
}
