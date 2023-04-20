package main

import (
	"context"
	"testing"

	"github.com/knqyf263/go-plugin/tests/fields/proto"
)

func BenchmarkFields(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()
	p, err := proto.NewFieldTestPlugin(ctx)
	if err != nil {
		b.Fatal(err)
	}

	plugin, err := p.Load(ctx, "plugin/plugin.wasm")
	if err != nil {
		b.Fatal(err)
	}
	defer plugin.Close(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = plugin.Test(ctx, &proto.Request{
			A: 1.2,
			B: 3.4,
			C: 5,
			D: -6,
			E: 7,
			F: 8,
			G: 9,
			H: -10,
			I: 11,
			J: 12,
			K: 13,
			L: -14,
			M: false,
			N: "foo",
			O: []byte("hoge"),
			P: []string{"a", "b"},
			Q: map[string]*proto.IntValue{
				"key": {A: 15},
			},
			R: &proto.Request_Nested{
				A: "samurai",
			},
			S: proto.Enum_A,
		}); err != nil {
			b.Fatal(err)
		}
	}
}
