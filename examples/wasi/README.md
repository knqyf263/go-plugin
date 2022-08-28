# WASI Example
This example shows how to access the local files.

## Generate Go code
A proto file is under `./cat`.

```shell
$ protoc --go-plugin_out=. --go-plugin_opt=paths=source_relative cat/cat.proto
```

## Pass fs.FS in a host
A host can control which files/dirs plugins can access.

```go
//go:embed testdata/hello.txt
var f embed.FS

func run() error {
        ctx := context.Background()
        p, err := cat.NewFileCatPlugin(ctx, cat.FileCatPluginOption{
                Stdout: os.Stdout, // Attach stdout so that the plugin can write outputs to stdout
                Stderr: os.Stderr, // Attach stderr so that the plugin can write errors to stderr
                FS:     f,         // Loaded plugins can access only files that the host allows.
        })
```

In this example, the host just passes `testdata/hello.txt` via `FileCatPluginOption`, but you can pass whatever you want.
Please refer to [io/fs][io/fs].

## Open a file in a plugin
A plugin can open a file as usual.

```go
b, err := os.ReadFile(request.GetFilePath())
if err != nil {
    return cat.FileCatReply{}, err
}
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
File loading...
Hello WASI!
```

[io/fs]: https://pkg.go.dev/io/fs