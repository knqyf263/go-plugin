//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/knqyf263/go-plugin/examples/wasi/cat"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	cat.RegisterFileCat(CatPlugin{})
}

type CatPlugin struct{}

var _ cat.FileCat = (*CatPlugin)(nil)

func (CatPlugin) Cat(_ context.Context, request cat.FileCatRequest) (cat.FileCatReply, error) {
	// The message is shown in stdout as os.Stdout is attached.
	fmt.Println("File loading...")
	b, err := os.ReadFile(request.GetFilePath())
	if err != nil {
		return cat.FileCatReply{}, err
	}
	return cat.FileCatReply{
		Content: string(b),
	}, nil
}
