# Host Functions Example
This example shows how to embed functions defined in a host into plugins to expand the plugin functionalities.

## Generate Go code
A proto file is under `./greeting`.

```protobuf
// The host functions embedded into the plugin
// go:plugin type=host
service HostFunctions {
  // Sends a HTTP GET request
  rpc HttpGet(HttpGetRequest) returns (HttpGetResponse) {}
  // Shows a log message
  rpc Log(LogRequest) returns (google.protobuf.Empty) {}
}
```

NOTE: You must specify `type=host` in the comment. It represents the service is for host functions.

Then, generate source code.

```shell
$ protoc --go-plugin_out=. --go-plugin_opt=paths=source_relative greeting/greet.proto 
```

## Implement host functions
The following interface is generated.

```go
type HostFunctions interface {
	HttpGet(context.Context, HttpGetRequest) (HttpGetResponse, error)
	Log(context.Context, LogRequest) (emptypb.Empty, error)
}
```

Implement that interface.

```go
// myHostFunctions implements greeting.HostFunctions
type myHostFunctions struct{}

// HttpGet is embedded into the plugin and can be called by the plugin.
func (myHostFunctions) HttpGet(ctx context.Context, request greeting.HttpGetRequest) (greeting.HttpGetResponse, error) {
	resp, err := http.Get(request.Url)
	if err != nil {
		return greeting.HttpGetResponse{}, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return greeting.HttpGetResponse{}, err
	}

	return greeting.HttpGetResponse{Response: buf}, nil
}

// Log is embedded into the plugin and can be called by the plugin.
func (myHostFunctions) Log(ctx context.Context, request greeting.LogRequest) (emptypb.Empty, error) {
	// Use the host logger
	log.Println(request.GetMessage())
	return emptypb.Empty{}, nil
}
```

Pass it to a plugin in `Load()`.

```go
// Pass my host functions that are embedded into the plugin.
greetingPlugin, err := p.Load(ctx, "plugin/plugin.wasm", myHostFunctions{})
```

## Call host functions in a plugin
The embedded host functions can be called in plugins.

```go
hostFunctions := greeting.NewHostFunctions()

// Logging via the host function
hostFunctions.Log(ctx, greeting.LogRequest{
	Message: "Sending a HTTP request...",
})

// HTTP GET via the host function
resp, err := hostFunctions.HttpGet(ctx, greeting.HttpGetRequest{Url: "http://ifconfig.me"})
```

## Compile a plugin
Use TinyGo to compile the plugin to Wasm.

```shell
$ tinygo build -o plugin/plugin.wasm -scheduler=none -target=wasi --no-debug plugin/plugin.go
```

## Run
`main.go` loads the above plugin.

```shell
$ go run main.go
2022/08/28 10:13:57 Sending a HTTP request...
Hello, go-plugin from x.x.x.x
```