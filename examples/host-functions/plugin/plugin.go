//go:build tinygo.wasm

package main

import (
	"context"

	"github.com/knqyf263/go-plugin/examples/host-functions/greeting"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	greeting.RegisterGreeter(GreetingPlugin{})
}

type GreetingPlugin struct{}

var _ greeting.Greeter = (*GreetingPlugin)(nil)

func (m GreetingPlugin) Greet(ctx context.Context, request greeting.GreetRequest) (greeting.GreetReply, error) {
	hostFunctions := greeting.NewHostFunctions()

	// Logging via the host function
	hostFunctions.Log(ctx, greeting.LogRequest{
		Message: "Sending a HTTP request...",
	})

	// HTTP GET via the host function
	resp, err := hostFunctions.HttpGet(ctx, greeting.HttpGetRequest{Url: "http://ifconfig.me"})
	if err != nil {
		return greeting.GreetReply{}, err
	}

	return greeting.GreetReply{
		Message: "Hello, " + request.GetName() + " from " + string(resp.Response),
	}, nil
}
