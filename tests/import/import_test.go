package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/knqyf263/go-plugin/tests/import/proto/bar"
	"github.com/knqyf263/go-plugin/tests/import/proto/foo"
)

func TestImport(t *testing.T) {
	ctx := context.Background()
	p, err := foo.NewFooPlugin(ctx)
	require.NoError(t, err)

	plugin, err := p.Load(ctx, "plugin/plugin.wasm")
	require.NoError(t, err)
	defer plugin.Close(ctx)

	got, err := plugin.Hello(ctx, &foo.Request{
		A: "Hi",
	})

	want := &bar.Reply{
		A: "Hi, bar",
	}
	assert.Equal(t, want, got)
}

func TestImportReader(t *testing.T) {
	ctx := context.Background()
	p, err := foo.NewFooPlugin(ctx)
	require.NoError(t, err)

	f, err := os.Open("plugin/plugin.wasm")
	require.NoError(t, err)
	defer f.Close()

	plugin, err := p.LoadReader(ctx, f)
	require.NoError(t, err)
	defer plugin.Close(ctx)

	got, err := plugin.Hello(ctx, &foo.Request{
		A: "Hi",
	})

	want := &bar.Reply{
		A: "Hi, bar",
	}
	assert.Equal(t, want, got)
}
