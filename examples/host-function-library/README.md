# Register Host Function Library Example
This example shows how to embed functions defined in a host into plugins to expand the plugin functionalities.
Host functions can be defined either in the current plugin distribution in the proto file or imported from another module defined as a source for these host functions.
This allows host functions to be distributed as a library or SDK instead of having to copy them every time you create plugins.

## Generate Go code for distributed host functions
A proto file is under `library/json-parser/export`.

```protobuf
// Distributing host functions without plugin code
// go:plugin type=host module=json-parser
service ParserLibrary {
  rpc ParseJson(ParseJsonRequest) returns (ParseJsonResponse) {}
}
```

> **_NOTE:_** You must specify `type=host` in the comment `module=json-parser` and module name to be unique for the host function redistribution and registration.
It represents the service is for host functions.

Then, generate source code for `json-parser` module hosts functions.

```shell
$ protoc --go-plugin_out=. --go-plugin_opt=paths=source_relative library/json-parser/export/library.proto 
```

## Implement `json-parser` module host functions
The following interface is generated.

```go
// Distributing host functions without plugin code
// go:plugin type=host module=json-parser
type ParserLibrary interface {
	ParseJson(context.Context, *ParseJsonRequest) (*ParseJsonResponse, error)
}
```

Implement that interface in separate package which will be exported to the main plugin implementation.

```go
// ParserLibraryImpl implements export.ParserLibrary functions
type ParserLibraryImpl struct{}

// ParseJson is embedded into the plugin and can be called by the plugin.
func (ParserLibraryImpl) ParseJson(_ context.Context, request *export.ParseJsonRequest) (*export.ParseJsonResponse, error) {
    var person export.Person
    if err := json.Unmarshal(request.GetContent(), &person); err != nil {
        return nil, err
    }

    return &export.ParseJsonResponse{Response: &person}, nil
}
```

Then, generate source code stubs for the base plugin.

```shell
$ protoc --go-plugin_out=. --go-plugin_opt=paths=source_relative proto/greer.proto 
```

Register exported hosts functions module int while providing new WazeroRuntime in the `proto.NewGreeterPlugin()`.

```go
ctx := context.Background()
p, err := proto.NewGreeterPlugin(ctx, proto.WazeroRuntime(func(ctx context.Context) (wazero.Runtime, error) {
    r, err := proto.DefaultWazeroRuntime()(ctx)
    if err != nil {
        return nil, err
    }
    return r, export.Instantiate(ctx, r, impl.ParserLibraryImpl{})
}))
```

## Call host functions in a plugin
The exported as a library or SDK host functions can be called in plugins.

```go
parserLibrary := export.NewParserLibrary()

// Call the host function to parse JSON
resp, err := parserLibrary.ParseJson(ctx, &export.ParseJsonRequest{
    Content: []byte(fmt.Sprintf(`{"name": "%s", "age": 20}`, sanrequest.Message)),
})
if err != nil {
    return nil, err
}
```

## Compile a plugin
Use TinyGo to compile the plugin to Wasm.

```shell
$ go generate main.go
```

## Run
`main.go` loads the above plugin.

```shell
$ go run main.go
```
```shll
2022/08/28 10:13:57 Sending a HTTP request...
Hello, Sato. This is Yamada-san (age 20).
```


