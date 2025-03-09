//go:build wasip1

package main

import (
	"context"
	"fmt"

	"github.com/knqyf263/go-plugin/tests/host-functions/proto"
)

// main is required for Go to compile to Wasm.
func main() {}

func init() {
	proto.RegisterGreeter(TestPlugin{})
}

type TestPlugin struct{}

var _ proto.Greeter = (*TestPlugin)(nil)

func (p TestPlugin) Greet(ctx context.Context, request *proto.GreetRequest) (*proto.GreetReply, error) {
	hostFunctions := proto.NewHostFunctions()

	// Call the host function to parse JSON
	resp, err := hostFunctions.ParseJson(ctx, &proto.ParseJsonRequest{
		Content: []byte(`{"name": "Yamada", "age": 20}`),
	})
	if err != nil {
		return nil, err
	}

	return &proto.GreetReply{
		Message: fmt.Sprintf("Hello, %s. This is %s (age %d).",
			request.GetName(), resp.GetResponse().GetName(), resp.GetResponse().GetAge()),
	}, nil
}
