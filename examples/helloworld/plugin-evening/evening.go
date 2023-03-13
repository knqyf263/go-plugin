//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"

	"github.com/knqyf263/go-plugin/examples/helloworld/greeting"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	greeting.RegisterGreeter(GoodEvening{})
}

type GoodEvening struct{}

var _ greeting.Greeter = (*GoodEvening)(nil)

func (m GoodEvening) Greet(_ context.Context, request *greeting.GreetRequest) (*greeting.GreetReply, error) {
	return &greeting.GreetReply{
		Message: fmt.Sprintf("Good evening, %s", request.GetName()),
	}, nil
}
