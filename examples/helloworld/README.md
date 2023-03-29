# Hello World Example
This example shows how to define plugin interfaces, generate Go code and use it.

## Generate Go code
A proto file is under `./greeting`.

```shell
$ protoc --go-plugin_out=. --go-plugin_opt=paths=source_relative greeting/greet.proto 
```

## Compile a plugin
Use TinyGo to compile the plugin to Wasm.

```shell
$ go generate main.go
```

## Run
`main.go` loads the above two plugins.

```shell
$ go run main.go
Good morning, go-plugin
Good evening, go-plugin
```