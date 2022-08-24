package main

import (
	"context"
	"fmt"
	"log"

	"github.com/knqyf263/go-plugin/examples/basic/greeting"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	p, err := greeting.NewGreeterPlugin()
	if err != nil {
		return err
	}

	ctx := context.Background()

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
