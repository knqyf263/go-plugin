# Well-Known Types Example
This example shows how to use [well-known types][well-known-types] in Protocol Buffers.

## Generate Go code
A proto file is under `./known`.
You can use well-known types by importing `google/protobuf/xxx.proto`.

```protobuf
import "google/protobuf/timestamp.proto";
```

Then, you can use those types as usual.

```protobuf
message DiffRequest {
  google.protobuf.Value     value = 1;
  google.protobuf.Timestamp start = 2;
  google.protobuf.Timestamp end   = 3;
}
```

Generate code.

```shell
$ protoc --go-plugin_out=. --go-plugin_opt=paths=source_relative known/known.proto
```

NOTE: The generated code imports packages under https://github.com/knqyf263/go-plugin/tree/main/types/known.
TinyGo cannot compile packages under https://github.com/protocolbuffers/protobuf-go/tree/master/types/known.

## Compile a plugin
Use TinyGo to compile the plugin to Wasm.

```shell
$ tinygo build -o plugin/plugin.wasm -scheduler=none -target=wasi --no-debug plugin/plugin.go
```

## Run
`main.go` loads the above plugin.

```shell
$ go run main.go
I love Sushi
I love Tempura
Duration: 1h0m0s
```

[well-known-types]: https://developers.google.com/protocol-buffers/docs/reference/google.protobuf