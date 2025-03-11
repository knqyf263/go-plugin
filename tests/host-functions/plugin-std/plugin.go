//go:build wasip1

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/knqyf263/go-plugin/tests/host-functions/proto"
)

// main is required for Go to compile to Wasm.
func main() {}

func init() {
	proto.RegisterGreeter(TestPlugin{})
}

type TestPlugin struct{}

var _ proto.Greeter = (*TestPlugin)(nil)

func (p TestPlugin) Greet(ctx context.Context, request *proto.GreetRequest) (*proto.GreetReply, error) {
	// Use encoding/json to parse JSON
	var person proto.Person
	if err := json.Unmarshal([]byte(`{"name": "Suzuki", "age": 30}`), &person); err != nil {
		return nil, err
	}

	return &proto.GreetReply{
		Message: fmt.Sprintf("Hello, %s. This is %s (age %d).", request.GetName(), person.GetName(), person.GetAge()),
	}, nil
}
