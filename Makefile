PB_FILES=$(shell find . -path '*.pb.go' | grep -v "vendor")
PROTO_FILES=$(shell find . -path '*.proto' | grep -v "vendor")
GOOGLE_APIS=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis
PROTOC_FLAGS=-I/usr/local/include -I. -I$(GOPATH)/src -I$(GOPATH)/src/$(GOOGLE_APIS)
GRPC_GATEWAY=github.com/grpc-ecosystem/grpc-gateway
PACKAGES=$(shell go list ./... | grep -v /vendor/)

.PHONY: build setup generate generate help
.DEFAULT: default

setup:
	@go get -u github.com/stevvooe/protobuild
	@go get -u github.com/favadi/protoc-go-inject-tag
	@go get -d $(GRPC_GATEWAY)/...
	@cd $(GOPATH)/src/$(GRPC_GATEWAY)/protoc-gen-grpc-gateway && go install
	@cd $(GOPATH)/src/$(GRPC_GATEWAY)/protoc-gen-swagger && go install
	@go get -u gopkg.in/alecthomas/gometalinter.v2
	@gometalinter.v2 --install

generate:
	@protobuild $(PACKAGES)
	@$(foreach file,$(PB_FILES),protoc-go-inject-tag -input=$(file);)
	@$(foreach file,$(PROTO_FILES),protoc $(PROTOC_FLAGS) --grpc-gateway_out=logtostderr=true:. $(file);)

build:
	@go build

lint:
	@gometalinter.v2

test:
	@docker run -d --name todo-test -p 5432:5432 -e POSTGRES_DB="todo" postgres
	@sleep 3 # Wait for postgres startup (╯°□°）╯︵ ┻━┻
	@go test $(PACKAGES)
	@docker rm -f todo-test
