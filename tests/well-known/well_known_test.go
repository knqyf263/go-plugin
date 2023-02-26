package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/knqyf263/go-plugin/tests/well-known/proto"
	"github.com/knqyf263/go-plugin/types/known/durationpb"
	"github.com/knqyf263/go-plugin/types/known/emptypb"
	"github.com/knqyf263/go-plugin/types/known/structpb"
	"github.com/knqyf263/go-plugin/types/known/timestamppb"
	"github.com/knqyf263/go-plugin/types/known/wrapperspb"
)

func TestWellKnownTypes(t *testing.T) {
	ctx := context.Background()
	p, err := proto.NewKnownTypesTestPlugin(ctx)
	require.NoError(t, err)
	defer p.Close(ctx)

	plugin, err := p.Load(ctx, "plugin/plugin.wasm")
	require.NoError(t, err)

	b := time.Date(2022, 1, 2, 3, 4, 5, 6, time.UTC)
	c, err := structpb.NewValue(map[string]interface{}{
		"CA": "AAA",
		"CB": true,
		"CC": 100,
		"CD": []interface{}{
			map[string]interface{}{
				"CE": "CF",
			},
			map[string]interface{}{
				"CG": "CH",
				"CI": "CJ",
			},
		},
		"CK": nil,
	})
	require.NoError(t, err)

	got, err := plugin.Test(ctx, proto.Request{
		// duration
		A: durationpb.New(1 * time.Hour),

		// timestamp
		B: timestamppb.New(b),

		// struct
		C: c,

		// wrappers
		D: wrapperspb.Bool(true),
		E: wrapperspb.Bytes([]byte(`Bytes`)),
		F: wrapperspb.Double(1.2),
		G: wrapperspb.Float(3.4),
		H: wrapperspb.Int32(1),
		I: wrapperspb.Int64(2),
		J: wrapperspb.String("String"),
		K: wrapperspb.UInt32(3),
		L: wrapperspb.UInt64(4),
	})

	c, err = structpb.NewValue(map[string]interface{}{
		"CA": "AAABBB",
		"CB": false,
		"CC": 200,
		"CD": []interface{}{
			map[string]interface{}{
				"CE": "CF",
			},
			map[string]interface{}{
				"CG": "CH",
				"CI": "CJ",
			},
			"FOO",
		},
		"CK": nil,
	})
	require.NoError(t, err)

	want := proto.Response{
		A: durationpb.New(2 * time.Minute),
		B: timestamppb.New(b.Add(1 * time.Hour)),
		C: c,
		D: wrapperspb.Bool(false),
		E: wrapperspb.Bytes([]byte(`BytesValue`)),
		F: wrapperspb.Double(2.4),
		G: wrapperspb.Float(6.8),
		H: wrapperspb.Int32(2),
		I: wrapperspb.Int64(4),
		J: wrapperspb.String("StringValue"),
		K: wrapperspb.UInt32(6),
		L: wrapperspb.UInt64(8),
	}
	assert.Equal(t, want, got)
}

func TestEmpty(t *testing.T) {
	ctx := context.Background()
	p, err := proto.NewEmptyTestPlugin(ctx)
	require.NoError(t, err)

	plugin, err := p.Load(ctx, "plugin/plugin.wasm")
	require.NoError(t, err)

	got, err := plugin.DoNothing(ctx, emptypb.Empty{})
	require.NoError(t, err)
	assert.Equal(t, emptypb.Empty{}, got)
}
