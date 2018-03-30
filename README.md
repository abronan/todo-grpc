# todo-grpc

Todo app with gRPC/REST with the goal of using the least amount of source code with an extended feature set.

## Development Environment

- go >= v1.9
- git
- gnu-make
- grpc (libprotoc >= 3.5.0)

## Setup

### Makefile

```text
setup                          install dependencies
generate                       generate protobuf files
lint                           run gometalinter
build                          build the go packages
```

## Language/Libraries

- golang
- grpc-gateway
- go-grpc-middleware
- postgresql (access through go-pg orm)
- gRPC: Protocol buffers v3 ([documentation](https://developers.google.com/protocol-buffers/))
