package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/knqyf263/go-plugin/tests/fields/proto"
	"github.com/knqyf263/go-plugin/types/known/emptypb"
)

func TestFields(t *testing.T) {
	ctx := context.Background()
	plugin := loadPlugin(ctx, t)

	res, err := plugin.TestEmptyInput(ctx, emptypb.Empty{})
	require.NoError(t, err)
	require.True(t, res.GetOk())

	got, err := plugin.Test(ctx, proto.Request{
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
	})

	want := proto.Response{
		A: 2.4,
		B: 6.8,
		C: 10,
		D: -12,
		E: 14,
		F: 16,
		G: 18,
		H: -20,
		I: 22,
		J: 24,
		K: 26,
		L: -28,
		M: true,
		N: "foobar",
		O: []byte("hogefuga"),
		P: []string{"b"},
		Q: map[string]*proto.IntValue{
			"key": {A: 16},
		},
		R: &proto.Response_Nested{
			A: "ninja",
		},
		S: proto.Enum_B,
	}
	assert.Equal(t, want, got)
}

func TestErrorResponse(t *testing.T) {
	ctx := context.Background()
	plugin := loadPlugin(ctx, t)

	_, err := plugin.TestError(ctx, emptypb.Empty{})
	require.Error(t, err)
	require.Equal(t, err.Error(), "error from plugin")
}

func loadPlugin(ctx context.Context, t *testing.T) proto.FieldTest {
	p, err := proto.NewFieldTestPlugin(ctx, proto.FieldTestPluginOption{})
	require.NoError(t, err)
	defer p.Close(ctx)

	plugin, err := p.Load(ctx, "plugin/plugin.wasm")
	require.NoError(t, err)

	return plugin
}
