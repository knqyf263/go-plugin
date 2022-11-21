package main

import (
	"context"
	"fmt"
	"log"

	"github.com/knqyf263/go-plugin/examples/helloworld/greeting"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	p, err := greeting.NewGreeterPlugin(ctx, greeting.GreeterPluginOption{})
	if err != nil {
		return err
	}
	defer p.Close(ctx)

	morningPlugin, err := p.Load(ctx, "plugin-morning/morning.wasm")
	if err != nil {
		return err
	}

	eveningPlugin, err := p.Load(ctx, "plugin-evening/evening.wasm")
	if err != nil {
		return err
	}

	reply, err := morningPlugin.Greet(ctx, greeting.GreetRequest{
		Name: "go-plugin",
	})
	if err != nil {
		return err
	}

	fmt.Println(reply.GetMessage())

	reply, err = eveningPlugin.Greet(ctx, greeting.GreetRequest{
		Name: "go-plugin",
	})
	if err != nil {
		return err
	}

	fmt.Println(reply.GetMessage())

	return nil
}
