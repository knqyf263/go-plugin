//go:build wasip1

package main

import (
	"context"

	"github.com/knqyf263/go-plugin/tests/import/proto/bar"
	"github.com/knqyf263/go-plugin/tests/import/proto/foo"
)

// main is required for Go to compile to Wasm.
func main() {}

func init() {
	foo.RegisterFoo(TestPlugin{})
}

type TestPlugin struct{}

var _ foo.Foo = (*TestPlugin)(nil)

func (p TestPlugin) Hello(_ context.Context, request *foo.Request) (*bar.Reply, error) {
	return &bar.Reply{
		A: request.GetA() + ", bar",
	}, nil
}
