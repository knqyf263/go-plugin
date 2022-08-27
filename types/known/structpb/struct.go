// Protocol Buffers - Google's data interchange format
// Copyright 2008 Google Inc.  All rights reserved.
// Copyright 2022 Teppei Fukuda.  All rights reserved.
// https://developers.google.com/protocol-buffers/

package structpb

import (
	"encoding/base64"
	"math"
	"unicode/utf8"

	"google.golang.org/protobuf/runtime/protoimpl"
)

// NewValue constructs a Value from a general-purpose Go interface.
//
//	╔════════════════════════╤════════════════════════════════════════════╗
//	║ Go type                │ Conversion                                 ║
//	╠════════════════════════╪════════════════════════════════════════════╣
//	║ nil                    │ stored as NullValue                        ║
//	║ bool                   │ stored as BoolValue                        ║
//	║ int, int32, int64      │ stored as NumberValue                      ║
//	║ uint, uint32, uint64   │ stored as NumberValue                      ║
//	║ float32, float64       │ stored as NumberValue                      ║
//	║ string                 │ stored as StringValue; must be valid UTF-8 ║
//	║ []byte                 │ stored as StringValue; base64-encoded      ║
//	║ map[string]interface{} │ stored as StructValue                      ║
//	║ []interface{}          │ stored as ListValue                        ║
//	╚════════════════════════╧════════════════════════════════════════════╝
//
// When converting an int64 or uint64 to a NumberValue, numeric precision loss
// is possible since they are stored as a float64.
func NewValue(v interface{}) (*Value, error) {
	switch v := v.(type) {
	case nil:
		return NewNullValue(), nil
	case bool:
		return NewBoolValue(v), nil
	case int:
		return NewNumberValue(float64(v)), nil
	case int32:
		return NewNumberValue(float64(v)), nil
	case int64:
		return NewNumberValue(float64(v)), nil
	case uint:
		return NewNumberValue(float64(v)), nil
	case uint32:
		return NewNumberValue(float64(v)), nil
	case uint64:
		return NewNumberValue(float64(v)), nil
	case float32:
		return NewNumberValue(float64(v)), nil
	case float64:
		return NewNumberValue(float64(v)), nil
	case string:
		if !utf8.ValidString(v) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in string: %q", v)
		}
		return NewStringValue(v), nil
	case []byte:
		s := base64.StdEncoding.EncodeToString(v)
		return NewStringValue(s), nil
	case map[string]interface{}:
		v2, err := NewStruct(v)
		if err != nil {
			return nil, err
		}
		return NewStructValue(v2), nil
	case []interface{}:
		v2, err := NewList(v)
		if err != nil {
			return nil, err
		}
		return NewListValue(v2), nil
	default:
		return nil, protoimpl.X.NewError("invalid type: %T", v)
	}
}

// NewNullValue constructs a new null Value.
func NewNullValue() *Value {
	return &Value{Kind: &Value_NullValue{NullValue: NullValue_NULL_VALUE}}
}

// NewBoolValue constructs a new boolean Value.
func NewBoolValue(v bool) *Value {
	return &Value{Kind: &Value_BoolValue{BoolValue: v}}
}

// NewNumberValue constructs a new number Value.
func NewNumberValue(v float64) *Value {
	return &Value{Kind: &Value_NumberValue{NumberValue: v}}
}

// NewStringValue constructs a new string Value.
func NewStringValue(v string) *Value {
	return &Value{Kind: &Value_StringValue{StringValue: v}}
}

// NewStructValue constructs a new struct Value.
func NewStructValue(v *Struct) *Value {
	return &Value{Kind: &Value_StructValue{StructValue: v}}
}

// NewListValue constructs a new list Value.
func NewListValue(v *ListValue) *Value {
	return &Value{Kind: &Value_ListValue{ListValue: v}}
}

// AsInterface converts x to a general-purpose Go interface.
//
// Calling Value.MarshalJSON and "encoding/json".Marshal on this output produce
// semantically equivalent JSON (assuming no errors occur).
//
// Floating-point values (i.e., "NaN", "Infinity", and "-Infinity") are
// converted as strings to remain compatible with MarshalJSON.
func (x *Value) AsInterface() interface{} {
	switch v := x.GetKind().(type) {
	case *Value_NumberValue:
		if v != nil {
			switch {
			case math.IsNaN(v.NumberValue):
				return "NaN"
			case math.IsInf(v.NumberValue, +1):
				return "Infinity"
			case math.IsInf(v.NumberValue, -1):
				return "-Infinity"
			default:
				return v.NumberValue
			}
		}
	case *Value_StringValue:
		if v != nil {
			return v.StringValue
		}
	case *Value_BoolValue:
		if v != nil {
			return v.BoolValue
		}
	case *Value_StructValue:
		if v != nil {
			return v.StructValue.AsMap()
		}
	case *Value_ListValue:
		if v != nil {
			return v.ListValue.AsSlice()
		}
	}
	return nil
}

// NewStruct constructs a Struct from a general-purpose Go map.
// The map keys must be valid UTF-8.
// The map values are converted using NewValue.
func NewStruct(v map[string]interface{}) (*Struct, error) {
	x := &Struct{Fields: make(map[string]*Value, len(v))}
	for k, v := range v {
		if !utf8.ValidString(k) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in string: %q", k)
		}
		var err error
		x.Fields[k], err = NewValue(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// AsMap converts x to a general-purpose Go map.
// The map values are converted by calling Value.AsInterface.
func (x *Struct) AsMap() map[string]interface{} {
	vs := make(map[string]interface{})
	for k, v := range x.GetFields() {
		vs[k] = v.AsInterface()
	}
	return vs
}

// NewList constructs a ListValue from a general-purpose Go slice.
// The slice elements are converted using NewValue.
func NewList(v []interface{}) (*ListValue, error) {
	x := &ListValue{Values: make([]*Value, len(v))}
	for i, v := range v {
		var err error
		x.Values[i], err = NewValue(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// AsSlice converts x to a general-purpose Go slice.
// The slice elements are converted by calling Value.AsInterface.
func (x *ListValue) AsSlice() []interface{} {
	vs := make([]interface{}, len(x.GetValues()))
	for i, v := range x.GetValues() {
		vs[i] = v.AsInterface()
	}
	return vs
}
