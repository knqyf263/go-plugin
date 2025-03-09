GOPATH := $(shell go env GOPATH)
GOBIN := $(if $(GOPATH),$(GOPATH)/bin,/usr/local/bin)

go_sources := $(shell find cmd encoding gen genid version wasm -name "*.go")
ifdef VERSION
	LDFLAGS := -ldflags="-X github.com/knqyf263/go-plugin/version.Version=${VERSION}"
endif

.PHONY: build
build: $(GOBIN)/protoc-gen-go-plugin

$(GOBIN)/protoc-gen-go-plugin: $(go_sources)
	go build ${LDFLAGS} -o $(GOPATH)/bin/protoc-gen-go-plugin cmd/protoc-gen-go-plugin/main.go

go_examples := $(shell find examples -path "*/plugin*/*.go")
.PHONY: build.examples
build.examples: $(go_examples:.go=.wasm)

go_tests := $(shell find tests -path "*/plugin*/*.go")
.PHONY: build.tests
build.tests: $(go_tests:.go=.wasm)

%.wasm: %.go $(GOBIN)/protoc-gen-go-plugin
	GOOS=wasip1 GOARCH=wasm go build -o $@ -buildmode=c-shared $<

proto_files := $(shell find . -name "*.proto")
.PHONY: protoc
protoc: $(proto_files:.proto=.pb.go) $(proto_files:.proto=_vtproto.pb.go)

%.pb.go: %.proto $(GOBIN)/protoc-gen-go-plugin
	protoc --go-plugin_out=. --go-plugin_opt=paths=source_relative $<;

.PHONY: fmt
fmt: $(proto_files)
	@for f in $^; do \
		clang-format -i $$f; \
	done

.PHONY: test
test: build.tests build.examples
	go test -v -short ./...
