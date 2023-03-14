//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"

	"github.com/knqyf263/go-plugin/examples/known-types/known"
	"github.com/knqyf263/go-plugin/types/known/durationpb"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	known.RegisterWellKnown(WellKnownPlugin{})
}

type WellKnownPlugin struct{}

var _ known.WellKnown = (*WellKnownPlugin)(nil)

func (p WellKnownPlugin) Diff(_ context.Context, request *known.DiffRequest) (*known.DiffReply, error) {
	value := request.GetValue().AsInterface()
	if m, ok := value.(map[string]interface{}); ok {
		fmt.Printf("I love %s\n", m["A"])
		fmt.Printf("I love %s\n", m["B"])
	}
	return &known.DiffReply{
		Duration: durationpb.New(request.GetEnd().AsTime().Sub(request.GetStart().AsTime())),
	}, nil
}
