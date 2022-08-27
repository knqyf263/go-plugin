package main

import (
	"context"
	"testing"

	"github.com/knqyf263/go-plugin/test/import/proto/bar"

	"github.com/knqyf263/go-plugin/test/import/proto/foo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImport(t *testing.T) {
	ctx := context.Background()
	p, err := foo.NewFooPlugin(ctx, foo.FooPluginOption{})
	require.NoError(t, err)

	plugin, err := p.Load(ctx, "plugin/plugin.wasm")
	require.NoError(t, err)

	got, err := plugin.Hello(ctx, foo.Request{
		A: "Hi",
	})

	want := bar.Reply{
		A: "Hi, bar",
	}
	assert.Equal(t, want, got)
}
