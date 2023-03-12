//go:build tinygo.wasm

package main

import (
	"context"
	"time"

	"github.com/knqyf263/go-plugin/tests/well-known/proto"
	"github.com/knqyf263/go-plugin/types/known/durationpb"
	"github.com/knqyf263/go-plugin/types/known/emptypb"
	"github.com/knqyf263/go-plugin/types/known/structpb"
	"github.com/knqyf263/go-plugin/types/known/timestamppb"
	"github.com/knqyf263/go-plugin/types/known/wrapperspb"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	proto.RegisterKnownTypesTest(TestPlugin{})
	proto.RegisterEmptyTest(TestPlugin{})
}

var _ proto.EmptyTest = (*TestPlugin)(nil)

type TestPlugin struct{}

func (p TestPlugin) Test(_ context.Context, request proto.Request) (proto.Response, error) {
	c, err := p.GetC(request.GetC())
	if err != nil {
		return proto.Response{}, err
	}
	return proto.Response{
		A: durationpb.New(2 * time.Minute),
		B: timestamppb.New(request.GetB().AsTime().Add(request.GetA().AsDuration())),
		C: c,
		D: wrapperspb.Bool(!request.GetD().Value),
		E: wrapperspb.Bytes(append(request.GetE().Value, []byte(`Value`)...)),
		F: wrapperspb.Double(request.GetF().Value * 2),
		G: wrapperspb.Float(request.GetG().Value * 2),
		H: wrapperspb.Int32(request.GetH().Value * 2),
		I: wrapperspb.Int64(request.GetI().Value * 2),
		J: wrapperspb.String(request.GetJ().Value + "Value"),
		K: wrapperspb.UInt32(request.GetK().Value * 2),
		L: wrapperspb.UInt64(request.GetL().Value * 2),
	}, nil
}

func (p TestPlugin) GetC(v *structpb.Value) (*structpb.Value, error) {
	c := v.AsInterface().(map[string]interface{})
	c["CA"] = c["CA"].(string) + "BBB"
	c["CB"] = !c["CB"].(bool)
	c["CC"] = c["CC"].(float64) * 2
	c["CD"] = append(c["CD"].([]interface{}), "FOO")
	return structpb.NewValue(c)
}

func (p TestPlugin) DoNothing(_ context.Context, _ emptypb.Empty) (emptypb.Empty, error) {
	return emptypb.Empty{}, nil
}
