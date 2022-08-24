package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/knqyf263/go-plugin/examples/host-functions/greeting"
)

func init() {
	greeting.RegisterHostFunctions(hostFunctions{})
}

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

	greetingPlugin, err := p.Load(ctx, "plugin/greeting.wasm")
	if err != nil {
		return err
	}

	reply, err := greetingPlugin.Greet(ctx, greeting.GreetRequest{
		Name: "go-plugin",
	})
	if err != nil {
		return err
	}

	fmt.Println(reply.GetMessage())

	return nil
}

type hostFunctions struct{}

func (hostFunctions) HttpGet(ctx context.Context, request greeting.HttpGetRequest) (greeting.HttpGetResponse, error) {
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
