# Hello World Example
This example shows how to define plugin interfaces, generate Go code and use it.

## Generate Go code
A proto file is under `./greeting`.

```shell
$ protoc --go-plugin_out=. --go-plugin_opt=paths=source_relative greeting/greet.proto 
```

## Compile a plugin
Use TinyGo to compile the plugin to Wasm.
This example contains two plugins, `plugin-morning` and `plugin-evening`.

```shell
$ tinygo build -o plugin-morning/morning.wasm -scheduler=none -target=wasi --no-debug plugin-morning/morning.go
$ tinygo build -o plugin-evening/evening.wasm -scheduler=none -target=wasi --no-debug plugin-evening/evening.go
```

## Run
`main.go` loads the above two plugins.

```shell
$ go run main.go
Good morning, go-plugin
Good evening, go-plugin
```