go_sources   := $(shell find cmd encoding gen genid version wasm -name "*.go")

.PHONY: build
build: $(GOPATH)/bin/protoc-gen-go-plugin

$(GOPATH)/bin/protoc-gen-go-plugin: $(go_sources)
	go build -o $(GOPATH)/bin/protoc-gen-go-plugin cmd/protoc-gen-go-plugin/main.go

tinygo_examples := $(shell find examples -path "*/plugin*/*.go")
.PHONY: build.examples
build.examples: $(tinygo_examples:.go=.wasm)

%.wasm: %.go
	tinygo build -o $@ -scheduler=none --no-debug --target=wasi $<

proto_files := $(shell find . -name "*.proto")
.PHONY: protoc
protoc: $(proto_files:.proto=.pb.go) $(proto_files:.proto=_vtproto.pb.go)

%.pb.go: %.proto $(GOPATH)/bin/protoc-gen-go-plugin
	protoc --go-plugin_out=. --go-plugin_opt=paths=source_relative $<;

.PHONY: fmt
fmt: $(proto_files)
	@for f in $^; do \
		clang-format -i $$f; \
	done