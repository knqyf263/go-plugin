//go:build tinygo.wasm

package main

import (
	"context"

	"github.com/knqyf263/go-plugin/tests/fields/proto"
)

// main is required for TinyGo to compile to Wasm.
func main() {
	proto.RegisterFieldTest(TestPlugin{})

}

type TestPlugin struct{}

func (p TestPlugin) Test(_ context.Context, request proto.Request) (proto.Response, error) {
	return proto.Response{
		A: request.GetA() * 2,
		B: request.GetB() * 2,
		C: request.GetC() * 2,
		D: request.GetD() * 2,
		E: request.GetE() * 2,
		F: request.GetF() * 2,
		G: request.GetG() * 2,
		H: request.GetH() * 2,
		I: request.GetI() * 2,
		J: request.GetJ() * 2,
		K: request.GetK() * 2,
		L: request.GetL() * 2,
		M: !request.GetM(),
		N: request.GetN() + "bar",
		O: append(request.GetO(), []byte("fuga")...),
		P: request.GetP()[1:],
		Q: func() map[string]*proto.IntValue {
			q := request.GetQ()
			q["key"].A++
			return q
		}(),
		R: func() *proto.Response_Nested {
			r := request.GetR()
			if r.A == "samurai" {
				return &proto.Response_Nested{
					A: "ninja",
				}
			}
			return nil
		}(),
		S: request.GetS() + 1,
	}, nil
}
