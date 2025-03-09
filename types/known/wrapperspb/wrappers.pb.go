// Protocol Buffers - Google's data interchange format
// Copyright 2008 Google Inc.  All rights reserved.
// Copyright 2008 Teppei Fukuda.  All rights reserved.
// https://developers.google.com/protocol-buffers/

// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v5.29.3
// source: types/known/wrapperspb/wrappers.proto

package wrapperspb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Wrapper message for `double`.
//
// The JSON representation for `DoubleValue` is JSON number.
type DoubleValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The double value.
	Value float64 `protobuf:"fixed64,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *DoubleValue) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *DoubleValue) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

// Wrapper message for `float`.
//
// The JSON representation for `FloatValue` is JSON number.
type FloatValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The float value.
	Value float32 `protobuf:"fixed32,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *FloatValue) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *FloatValue) GetValue() float32 {
	if x != nil {
		return x.Value
	}
	return 0
}

// Wrapper message for `int64`.
//
// The JSON representation for `Int64Value` is JSON string.
type Int64Value struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The int64 value.
	Value int64 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Int64Value) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *Int64Value) GetValue() int64 {
	if x != nil {
		return x.Value
	}
	return 0
}

// Wrapper message for `uint64`.
//
// The JSON representation for `UInt64Value` is JSON string.
type UInt64Value struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The uint64 value.
	Value uint64 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *UInt64Value) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *UInt64Value) GetValue() uint64 {
	if x != nil {
		return x.Value
	}
	return 0
}

// Wrapper message for `int32`.
//
// The JSON representation for `Int32Value` is JSON number.
type Int32Value struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The int32 value.
	Value int32 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Int32Value) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *Int32Value) GetValue() int32 {
	if x != nil {
		return x.Value
	}
	return 0
}

// Wrapper message for `uint32`.
//
// The JSON representation for `UInt32Value` is JSON number.
type UInt32Value struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The uint32 value.
	Value uint32 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *UInt32Value) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *UInt32Value) GetValue() uint32 {
	if x != nil {
		return x.Value
	}
	return 0
}

// Wrapper message for `bool`.
//
// The JSON representation for `BoolValue` is JSON `true` and `false`.
type BoolValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The bool value.
	Value bool `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *BoolValue) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *BoolValue) GetValue() bool {
	if x != nil {
		return x.Value
	}
	return false
}

// Wrapper message for `string`.
//
// The JSON representation for `StringValue` is JSON string.
type StringValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The string value.
	Value string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *StringValue) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *StringValue) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

// Wrapper message for `bytes`.
//
// The JSON representation for `BytesValue` is JSON string.
type BytesValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The bytes value.
	Value []byte `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *BytesValue) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *BytesValue) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}
