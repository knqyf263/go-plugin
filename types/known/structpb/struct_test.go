// Copyright 2020 The Go Authors. All rights reserved.
// Copyright 2022 Teppei Fukuda. All rights reserved.

package structpb_test

import (
	"math"
	"testing"

	spb "github.com/knqyf263/go-plugin/types/known/structpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestToStruct(t *testing.T) {
	tests := []struct {
		name    string
		in      map[string]interface{}
		wantPB  *spb.Struct
		wantErr bool
	}{
		{
			name:   "nil",
			in:     nil,
			wantPB: &spb.Struct{Fields: make(map[string]*spb.Value, 0)},
		},
		{
			name:   "empty",
			in:     make(map[string]interface{}),
			wantPB: &spb.Struct{Fields: make(map[string]*spb.Value, 0)},
		},
		{
			name: "fields",
			in: map[string]interface{}{
				"nil":     nil,
				"bool":    bool(false),
				"int":     int(-123),
				"int32":   int32(math.MinInt32),
				"int64":   int64(math.MinInt64),
				"uint":    uint(123),
				"uint32":  uint32(math.MaxInt32),
				"uint64":  uint64(math.MaxInt64),
				"float32": float32(123.456),
				"float64": float64(123.456),
				"string":  string("hello, world!"),
				"bytes":   []byte("\xde\xad\xbe\xef"),
				"map":     map[string]interface{}{"k1": "v1", "k2": "v2"},
				"slice":   []interface{}{"one", "two", "three"},
			},
			wantPB: &spb.Struct{Fields: map[string]*spb.Value{
				"nil":     spb.NewNullValue(),
				"bool":    spb.NewBoolValue(false),
				"int":     spb.NewNumberValue(float64(-123)),
				"int32":   spb.NewNumberValue(float64(math.MinInt32)),
				"int64":   spb.NewNumberValue(float64(math.MinInt64)),
				"uint":    spb.NewNumberValue(float64(123)),
				"uint32":  spb.NewNumberValue(float64(math.MaxInt32)),
				"uint64":  spb.NewNumberValue(float64(math.MaxInt64)),
				"float32": spb.NewNumberValue(float64(float32(123.456))),
				"float64": spb.NewNumberValue(float64(float64(123.456))),
				"string":  spb.NewStringValue("hello, world!"),
				"bytes":   spb.NewStringValue("3q2+7w=="),
				"map": spb.NewStructValue(&spb.Struct{Fields: map[string]*spb.Value{
					"k1": spb.NewStringValue("v1"),
					"k2": spb.NewStringValue("v2"),
				}}),
				"slice": spb.NewListValue(&spb.ListValue{Values: []*spb.Value{
					spb.NewStringValue("one"),
					spb.NewStringValue("two"),
					spb.NewStringValue("three"),
				}}),
			}},
		},
		{
			name:    "invalid utf-8 key",
			in:      map[string]interface{}{"\xde\xad\xbe\xef": "<invalid UTF-8>"},
			wantErr: true,
		},
		{
			name:    "invalid utf-8 value",
			in:      map[string]interface{}{"<invalid UTF-8>": "\xde\xad\xbe\xef"},
			wantErr: true,
		},
		{
			name:    "invalid type",
			in:      map[string]interface{}{"key": protoreflect.Name("named string")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPB, gotErr := spb.NewStruct(tt.in)
			assert.Equal(t, tt.wantErr, gotErr != nil)
			assert.Equal(t, tt.wantPB, gotPB)
		})
	}
}

func TestToListValue(t *testing.T) {
	tests := []struct {
		name    string
		in      []interface{}
		wantPB  *spb.ListValue
		wantErr bool
	}{
		{
			name:   "nil",
			in:     nil,
			wantPB: &spb.ListValue{Values: make([]*spb.Value, 0)},
		},
		{
			name:   "empty",
			in:     make([]interface{}, 0),
			wantPB: &spb.ListValue{Values: make([]*spb.Value, 0)},
		},
		{
			name: "happy",
			in: []interface{}{
				nil,
				bool(false),
				int(-123),
				int32(math.MinInt32),
				int64(math.MinInt64),
				uint(123),
				uint32(math.MaxInt32),
				uint64(math.MaxInt64),
				float32(123.456),
				float64(123.456),
				string("hello, world!"),
				[]byte("\xde\xad\xbe\xef"),
				map[string]interface{}{"k1": "v1", "k2": "v2"},
				[]interface{}{"one", "two", "three"},
			},
			wantPB: &spb.ListValue{Values: []*spb.Value{
				spb.NewNullValue(),
				spb.NewBoolValue(false),
				spb.NewNumberValue(float64(-123)),
				spb.NewNumberValue(float64(math.MinInt32)),
				spb.NewNumberValue(float64(math.MinInt64)),
				spb.NewNumberValue(float64(123)),
				spb.NewNumberValue(float64(math.MaxInt32)),
				spb.NewNumberValue(float64(math.MaxInt64)),
				spb.NewNumberValue(float64(float32(123.456))),
				spb.NewNumberValue(float64(float64(123.456))),
				spb.NewStringValue("hello, world!"),
				spb.NewStringValue("3q2+7w=="),
				spb.NewStructValue(&spb.Struct{Fields: map[string]*spb.Value{
					"k1": spb.NewStringValue("v1"),
					"k2": spb.NewStringValue("v2"),
				}}),
				spb.NewListValue(&spb.ListValue{Values: []*spb.Value{
					spb.NewStringValue("one"),
					spb.NewStringValue("two"),
					spb.NewStringValue("three"),
				}}),
			}},
		},
		{
			name:    "invalid utf-8",
			in:      []interface{}{"\xde\xad\xbe\xef"},
			wantErr: true,
		},
		{
			name:    "invalid type",
			in:      []interface{}{protoreflect.Name("named string")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPB, gotErr := spb.NewList(tt.in)
			assert.Equal(t, tt.wantErr, gotErr != nil)
			assert.Equal(t, tt.wantPB, gotPB)
		})
	}
}

func TestToValue(t *testing.T) {
	tests := []struct {
		name    string
		in      interface{}
		wantPB  *spb.Value
		wantErr bool
	}{
		{
			name:   "nil",
			in:     nil,
			wantPB: spb.NewNullValue(),
		},
		{
			name:   "bool",
			in:     bool(false),
			wantPB: spb.NewBoolValue(false),
		},
		{
			name:   "int",
			in:     int(-123),
			wantPB: spb.NewNumberValue(float64(-123)),
		},
		{
			name:   "int32",
			in:     int32(math.MinInt32),
			wantPB: spb.NewNumberValue(float64(math.MinInt32)),
		},
		{
			name:   "int64",
			in:     int64(math.MinInt64),
			wantPB: spb.NewNumberValue(float64(math.MinInt64)),
		},
		{
			name:   "uint",
			in:     uint(123),
			wantPB: spb.NewNumberValue(float64(123)),
		},
		{
			name:   "uint32",
			in:     uint32(math.MaxInt32),
			wantPB: spb.NewNumberValue(float64(math.MaxInt32)),
		},
		{
			name:   "uint64",
			in:     uint64(math.MaxInt64),
			wantPB: spb.NewNumberValue(float64(math.MaxInt64)),
		},
		{
			name:   "float32",
			in:     float32(123.456),
			wantPB: spb.NewNumberValue(float64(float32(123.456))),
		},
		{
			name:   "float64",
			in:     float64(123.456),
			wantPB: spb.NewNumberValue(float64(float64(123.456))),
		},
		{
			name:   "string",
			in:     string("hello, world!"),
			wantPB: spb.NewStringValue("hello, world!"),
		},
		{
			name:   "bytes",
			in:     []byte("\xde\xad\xbe\xef"),
			wantPB: spb.NewStringValue("3q2+7w=="),
		},
		{
			name:   "map1",
			in:     map[string]interface{}(nil),
			wantPB: spb.NewStructValue(&spb.Struct{Fields: make(map[string]*spb.Value, 0)}),
		},
		{
			name:   "map2",
			in:     make(map[string]interface{}),
			wantPB: spb.NewStructValue(&spb.Struct{Fields: make(map[string]*spb.Value, 0)}),
		},
		{
			name: "map3",
			in:   map[string]interface{}{"k1": "v1", "k2": "v2"},
			wantPB: spb.NewStructValue(&spb.Struct{Fields: map[string]*spb.Value{
				"k1": spb.NewStringValue("v1"),
				"k2": spb.NewStringValue("v2"),
			}}),
		},
		{
			name:   "slice1",
			in:     []interface{}(nil),
			wantPB: spb.NewListValue(&spb.ListValue{Values: make([]*spb.Value, 0)}),
		},
		{
			name:   "slice2",
			in:     make([]interface{}, 0),
			wantPB: spb.NewListValue(&spb.ListValue{Values: make([]*spb.Value, 0)}),
		},
		{
			name: "slice3",
			in:   []interface{}{"one", "two", "three"},
			wantPB: spb.NewListValue(&spb.ListValue{Values: []*spb.Value{
				spb.NewStringValue("one"),
				spb.NewStringValue("two"),
				spb.NewStringValue("three"),
			}}),
		},
		{
			name:    "invalid utf-8",
			in:      "\xde\xad\xbe\xef",
			wantErr: true,
		}, {
			name:    "invalid type",
			in:      protoreflect.Name("named string"),
			wantErr: true,
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPB, gotErr := spb.NewValue(tt.in)
			assert.Equal(t, tt.wantErr, gotErr != nil)
			assert.Equal(t, tt.wantPB, gotPB)
		})
	}
}
