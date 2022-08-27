// Protocol Buffers - Google's data interchange format
// Copyright 2008 Google Inc.  All rights reserved.
// Copyright 2022 Teppei Fukuda.  All rights reserved.
// https://developers.google.com/protocol-buffers/

package wrapperspb

// Double stores v in a new DoubleValue and returns a pointer to it.
func Double(v float64) *DoubleValue {
	return &DoubleValue{Value: v}
}

// Float stores v in a new FloatValue and returns a pointer to it.
func Float(v float32) *FloatValue {
	return &FloatValue{Value: v}
}

// Int64 stores v in a new Int64Value and returns a pointer to it.
func Int64(v int64) *Int64Value {
	return &Int64Value{Value: v}
}

// UInt64 stores v in a new UInt64Value and returns a pointer to it.
func UInt64(v uint64) *UInt64Value {
	return &UInt64Value{Value: v}
}

// Int32 stores v in a new Int32Value and returns a pointer to it.
func Int32(v int32) *Int32Value {
	return &Int32Value{Value: v}
}

// UInt32 stores v in a new UInt32Value and returns a pointer to it.
func UInt32(v uint32) *UInt32Value {
	return &UInt32Value{Value: v}
}

// Bool stores v in a new BoolValue and returns a pointer to it.
func Bool(v bool) *BoolValue {
	return &BoolValue{Value: v}
}

// String stores v in a new StringValue and returns a pointer to it.
func String(v string) *StringValue {
	return &StringValue{Value: v}
}

// Bytes stores v in a new BytesValue and returns a pointer to it.
func Bytes(v []byte) *BytesValue {
	return &BytesValue{Value: v}
}
