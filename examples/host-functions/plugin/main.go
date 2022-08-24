//go:build tinygo.wasm

package main

import (
	"context"

	greeting "github.com/knqyf263/go-plugin/examples/host-functions/greeting"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	greeting.RegisterGreeter(GreetingPlugin{})

}

type GreetingPlugin struct{}

func (m GreetingPlugin) Greet(ctx context.Context, request greeting.GreetRequest) (greeting.GreetReply, error) {
	hostFunctions := greeting.NewHostFunctions()
	resp, err := hostFunctions.HttpGet(ctx, greeting.HttpGetRequest{Url: "http://ifconfig.me"})
	if err != nil {
		return greeting.GreetReply{}, err
	}
	return greeting.GreetReply{
		Message: "Hello, " + request.GetName() + " from " + string(resp.Response),
	}, nil
}
