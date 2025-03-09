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

	// Call the host function with nil request
	resp, err := hostFunctions.ParseJson(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &proto.GreetReply{
		Message: fmt.Sprintf("Hello, empty request '%s' and empty '%s' host function request", request.GetName(), resp.GetResponse().GetName()),
	}, nil
}
