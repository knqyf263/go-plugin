//go:build tinygo.wasm

package main

import (
	"context"
	"log"

	"github.com/knqyf263/go-plugin/examples/known-types/known"
	"github.com/knqyf263/go-plugin/types/known/durationpb"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	known.RegisterWellKnown(WellKnownPlugin{})
}

type WellKnownPlugin struct{}

func (p WellKnownPlugin) Diff(_ context.Context, request known.DiffRequest) (known.DiffReply, error) {
	value := request.GetValue().AsInterface()
	if m, ok := value.(map[string]interface{}); ok {
		for _, v := range m {
			log.Printf("I love %s\n", v)
		}
	}
	return known.DiffReply{
		Duration: durationpb.New(request.GetEnd().AsTime().Sub(request.GetStart().AsTime())),
	}, nil
}