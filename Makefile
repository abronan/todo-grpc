ROOTDIR=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))
PROJECT_ROOT=github.com/abronan/todo-grpc
PB_FILES=$(shell find . -path '*.pb.go' | grep -v "vendor")
PROTO_FILES=$(shell find . -path '*.proto' | grep -v "vendor")
GOOGLE_APIS=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis
PROTOC_FLAGS=-I/usr/local/include -I. -I$(GOPATH)/src -I$(GOPATH)/src/$(GOOGLE_APIS)
GRPC_GATEWAY=github.com/grpc-ecosystem/grpc-gateway
PACKAGES=$(shell go list ./... | grep -v /vendor/)
BINARIES=$(addprefix bin/,$(COMMANDS))
COMMANDS=protoc-gen-gogotodo
DESTDIR=/usr/local

.PHONY: clean setup generate protos build lint test binaries install uninstall help
.DEFAULT: default

setup: ## install dependencies
	@go get -u github.com/stevvooe/protobuild
	@go get -u github.com/favadi/protoc-go-inject-tag
	@go get -d $(GRPC_GATEWAY)/...
	@cd $(GOPATH)/src/$(GRPC_GATEWAY)/protoc-gen-grpc-gateway && go install
	@cd $(GOPATH)/src/$(GRPC_GATEWAY)/protoc-gen-swagger && go install
	@go get -u gopkg.in/alecthomas/gometalinter.v2
	@gometalinter.v2 --install

generate: protos
	@echo "$@"
	@PATH=$(ROOTDIR)/bin:$(PATH) go generate -x $(PACKAGES)

protos: bin/protoc-gen-gogotodo ## generate protobuf
	@PATH=$(ROOTDIR)/bin:$(PATH) protobuild $(PACKAGES)
	@$(foreach file,$(PB_FILES),protoc-go-inject-tag -input=$(file);)
	@$(foreach file,$(PROTO_FILES),protoc $(PROTOC_FLAGS) --grpc-gateway_out=logtostderr=true:. $(file);)
	@$(foreach file,$(PROTO_FILES),protoc $(PROTOC_FLAGS) --swagger_out=logtostderr=true:. $(file);)

build: ## build the go packages
	@go build $(PACKAGES)

lint: ## run go lint
	@gometalinter.v2

test: ## run test suite (requires docker or a local postgresql instance)
	@docker run -d --name todo-test -p 5432:5432 -e POSTGRES_DB="todo" postgres
	@go test $(PACKAGES)
	@docker rm -f todo-test

FORCE:

# Build a binary from a cmd.
bin/%: cmd/% FORCE
	@test $$(go list) = "$(PROJECT_ROOT)" || \
		(echo "Please correctly set up your Go build environment. This project must be located at <GOPATH>/src/$(PROJECT_ROOT)" && false)
	@echo "$@"
	@go build -i -tags "$(DOCKER_BUILDTAGS)" -o $@ $(GO_LDFLAGS) $(GO_GCFLAGS) ./$<

binaries: $(BINARIES) ## build binaries
	@echo "$@"

install: $(BINARIES) ## install binaries
	@echo "$@"
	@mkdir -p $(DESTDIR)/bin
	@install $(BINARIES) $(DESTDIR)/bin

uninstall: ## uninstall binaries
	@echo "$@"
	@rm -f $(addprefix $(DESTDIR)/bin/,$(notdir $(BINARIES)))

clean: ## clean up binaries
	@echo "$@"
	@rm -f $(BINARIES)

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort