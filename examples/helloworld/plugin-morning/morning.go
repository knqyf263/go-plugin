//go:build tinygo.wasm

package main

import (
	"context"

	"github.com/knqyf263/go-plugin/examples/helloworld/greeting"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	greeting.RegisterGreeter(GoodMorning{})
}

type GoodMorning struct{}

var _ greeting.Greeter = (*GoodMorning)(nil)

func (m GoodMorning) Greet(_ context.Context, request *greeting.GreetRequest) (*greeting.GreetReply, error) {
	return &greeting.GreetReply{
		Message: "Good morning, " + request.GetName(),
	}, nil
}
