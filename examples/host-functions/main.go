package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/knqyf263/go-plugin/examples/host-functions/greeting"
	"github.com/knqyf263/go-plugin/types/known/emptypb"
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

	// Pass my host functions that are embedded into the plugin.
	greetingPlugin, err := p.Load(ctx, "plugin/plugin.wasm", myHostFunctions{})
	if err != nil {
		return err
	}
	defer greetingPlugin.Close(ctx)

	reply, err := greetingPlugin.Greet(ctx, greeting.GreetRequest{
		Name: "go-plugin",
	})
	if err != nil {
		return err
	}

	fmt.Println(reply.GetMessage())

	return nil
}

// myHostFunctions implements greeting.HostFunctions
type myHostFunctions struct{}

var _ greeting.HostFunctions = (*myHostFunctions)(nil)

// HttpGet is embedded into the plugin and can be called by the plugin.
func (myHostFunctions) HttpGet(ctx context.Context, request greeting.HttpGetRequest) (greeting.HttpGetResponse, error) {
	resp, err := http.Get(request.Url)
	if err != nil {
		return greeting.HttpGetResponse{}, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return greeting.HttpGetResponse{}, err
	}

	return greeting.HttpGetResponse{Response: buf}, nil
}

// Log is embedded into the plugin and can be called by the plugin.
func (myHostFunctions) Log(ctx context.Context, request greeting.LogRequest) (emptypb.Empty, error) {
	// Use the host logger
	log.Println(request.GetMessage())
	return emptypb.Empty{}, nil
}
