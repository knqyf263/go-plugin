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
	p, err := greeting.NewGreeterPlugin(ctx)
	if err != nil {
		return err
	}

	morningPlugin, err := p.Load(ctx, "plugin-morning/morning.wasm")
	if err != nil {
		return err
	}
	defer morningPlugin.Close(ctx)

	eveningPlugin, err := p.Load(ctx, "plugin-evening/evening.wasm")
	if err != nil {
		return err
	}
	defer eveningPlugin.Close(ctx)

	reply, err := morningPlugin.Greet(ctx, &greeting.GreetRequest{
		Name: "go-plugin",
	})
	if err != nil {
		return err
	}

	fmt.Println(reply.GetMessage())

	reply, err = eveningPlugin.Greet(ctx, &greeting.GreetRequest{
		Name: "go-plugin",
	})
	if err != nil {
		return err
	}

	fmt.Println(reply.GetMessage())

	return nil
}
