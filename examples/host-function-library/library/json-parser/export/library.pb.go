// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v5.29.3
// source: examples/host-function-library/library/json-parser/export/library.proto

package export

import (
	context "context"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ParseJsonRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content []byte `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *ParseJsonRequest) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *ParseJsonRequest) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

type ParseJsonResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Response *Person `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *ParseJsonResponse) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *ParseJsonResponse) GetResponse() *Person {
	if x != nil {
		return x.Response
	}
	return nil
}

type Person struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Age  int64  `protobuf:"varint,2,opt,name=age,proto3" json:"age,omitempty"`
}

func (x *Person) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *Person) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Person) GetAge() int64 {
	if x != nil {
		return x.Age
	}
	return 0
}

// Distributing host functions without plugin code
// go:plugin type=host module=json-parser
type ParserLibrary interface {
	ParseJson(context.Context, *ParseJsonRequest) (*ParseJsonResponse, error)
}
