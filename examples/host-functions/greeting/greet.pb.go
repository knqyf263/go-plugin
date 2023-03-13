// Code generated by protoc-gen-go-plugin. DO NOT EDIT.
// versions:
// 	protoc-gen-go-plugin v0.1.0
// 	protoc               v3.21.12
// source: examples/host-functions/greeting/greet.proto

package greeting

import (
	context "context"
	emptypb "github.com/knqyf263/go-plugin/types/known/emptypb"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// The request message containing the user's name.
type GreetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *GreetRequest) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *GreetRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// The response message containing the greetings
type GreetReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *GreetReply) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *GreetReply) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type HttpGetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *HttpGetRequest) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *HttpGetRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

type HttpGetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Response []byte `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *HttpGetResponse) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *HttpGetResponse) GetResponse() []byte {
	if x != nil {
		return x.Response
	}
	return nil
}

type LogRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *LogRequest) ProtoReflect() protoreflect.Message {
	panic(`not implemented`)
}

func (x *LogRequest) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

// The greeting service definition.
// go:plugin type=plugin version=1
type Greeter interface {
	// Sends a greeting
	Greet(context.Context, *GreetRequest) (*GreetReply, error)
}

// The host functions embedded into the plugin
// go:plugin type=host
type HostFunctions interface {
	// Sends a HTTP GET request
	HttpGet(context.Context, *HttpGetRequest) (*HttpGetResponse, error)
	// Shows a log message
	Log(context.Context, *LogRequest) (*emptypb.Empty, error)
}
