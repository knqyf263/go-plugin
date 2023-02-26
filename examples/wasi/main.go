package main

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/tetratelabs/wazero"

	"github.com/knqyf263/go-plugin/examples/wasi/cat"
)

//go:embed testdata/hello.txt
var f embed.FS

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	mc := wazero.NewModuleConfig().
		WithStdout(os.Stdout). // Attach stdout so that the plugin can write outputs to stdout
		WithStderr(os.Stderr). // Attach stderr so that the plugin can write errors to stderr
		WithFS(f)              // Loaded plugins can access only files that the host allows.
	p, err := cat.NewFileCatPlugin(ctx, cat.WazeroModuleConfig(mc))
	if err != nil {
		return err
	}

	wasiPlugin, err := p.Load(ctx, "plugin/plugin.wasm")
	defer wasiPlugin.Close(ctx)
	if err != nil {
		return err
	}

	reply, err := wasiPlugin.Cat(ctx, cat.FileCatRequest{
		FilePath: "testdata/hello.txt",
	})
	if err != nil {
		return err
	}

	fmt.Println(reply.GetContent())

	return nil
}
