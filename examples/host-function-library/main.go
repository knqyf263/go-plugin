//go:build !tinygo.wasm

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tetratelabs/wazero"

	"github.com/knqyf263/go-plugin/examples/host-function-library/library/json-parser/export"
	"github.com/knqyf263/go-plugin/examples/host-function-library/library/json-parser/impl"
	"github.com/knqyf263/go-plugin/examples/host-function-library/proto"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	p, err := proto.NewGreeterPlugin(ctx, proto.WazeroRuntime(func(ctx context.Context) (wazero.Runtime, error) {
		r, err := proto.DefaultWazeroRuntime()(ctx)
		if err != nil {
			return nil, err
		}
		return r, export.Instantiate(ctx, r, impl.ParserLibraryImpl{})
	}))

	// Pass my host functions that are embedded into the plugin.
	plugin, err := p.Load(ctx, "plugin/plugin.wasm", myHostFunctions{})
	if err != nil {
		return err
	}

	defer plugin.Close(ctx)

	reply, err := plugin.Greet(ctx, &proto.GreetRequest{
		Name: "Sato",
	})

	fmt.Println(reply.GetMessage())

	return nil
}

type myHostFunctions struct{}

var _ proto.HostFunctions = (*myHostFunctions)(nil)

func (m myHostFunctions) San(_ context.Context, request *proto.SanRequest) (*proto.SanResponse, error) {
	return &proto.SanResponse{Message: fmt.Sprintf("%s-san", request.GetMessage())}, nil
}
