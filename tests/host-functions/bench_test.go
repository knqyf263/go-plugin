package main

import (
	"context"
	"os"
	"testing"

	"github.com/tetratelabs/wazero"

	"github.com/knqyf263/go-plugin/tests/host-functions/proto"
)

func BenchmarkHostFunctions(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	mc := wazero.NewModuleConfig().WithStdout(os.Stdout)
	p, err := proto.NewGreeterPlugin(ctx, proto.WazeroRuntime(func(ctx context.Context) (wazero.Runtime, error) {
		return proto.DefaultWazeroRuntime()(ctx)
	}), proto.WazeroModuleConfig(mc))
	if err != nil {
		b.Fatal(err)
	}

	// Pass my host functions that are embedded into the plugin.
	plugin, err := p.Load(ctx, "plugin/plugin.wasm", myHostFunctions{})
	if err != nil {
		b.Fatal(err)
	}
	defer plugin.Close(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := plugin.Greet(ctx, &proto.GreetRequest{
			Name: "Sato",
		}); err != nil {
			b.Fatal(err)
		}
	}
}
