# todo-grpc

Todo app with gRPC/REST with the goal of using the least amount of source code with an extended feature set.

## Development Environment

- **go** >= v1.9
- **git**
- **grpc** (libprotoc >= 3.5.0)
- **gnu-make**
- **docker**

## Setup

### Makefile

```text
binaries                       build binaries
build                          build the go packages
clean                          clean up binaries
help                           this help
install                        install binaries
lint                           run go lint
protos                         generate protobuf
setup                          install dependencies
test                           run test suite (requires docker or a local postgresql instance)
uninstall                      uninstall binaries
```

## Usage

```text
NAME:
   todo-grpc - Todo app

USAGE:
   todo-grpc [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --bind-http value             bind address for HTTP (default: ":8080") [$BIND_HTTP]
   --bind-grpc value             bind address for gRPC (default: ":2338") [$BIND_GRPC]
   --bind-prometheus-http value  bind prometheus address for HTTP (default: ":8081") [$BIND_PROMETHEUS_HTTP]
   --db-name value               database name (default: "todo") [$DB_NAME]
   --db-user value               database username (default: "postgres") [$DB_USER]
   --db-password value           database password [$DB_PASSWORD]
   --db-host value               postgres host (default: "127.0.0.1") [$DB_HOST]
   --db-port value               database port (default: 5432) [$DB_PORT]
   --jaeger-host value           Jaeger hostname (default: "127.0.0.1") [$JAEGER_HOST]
   --jaeger-port value           Jaeger port (default: 5775) [$JAEGER_PORT]
   --jaeger-sampler value        Jaeger sampler (default: 0.05) [$JAEGER_SAMPLER]
   --jaeger-tags value           Jaeger tags (default: "todo") [$JAEGER_TAGS]
   --help, -h                    show help
   --version, -v                 print the version
```

### Rest API

- Create a new Todo:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"title":"Test","description":"Test"}' "http://localhost:8080/v1/todo"
{"id":"34d63bd4-56b3-4795-80d4-86e5db6fa0b5"}
```

- Get an existing Todo:

```bash
curl -X GET "http://localhost:8080/v1/todo/34d63bd4-56b3-4795-80d4-86e5db6fa0b5"
{"item":{"id":"34d63bd4-56b3-4795-80d4-86e5db6fa0b5","title":"Test","description":"Test","created_at":"2018-03-30T20:13:25.291887Z"}}
```

- List Todos (example with limit and non completed items in query parameters):

```bash
curl -X GET "http://localhost:8080/v1/todo?limit=10&not_completed=true"
```

- Update a Todo:

```bash
curl -X PUT -H "Content-Type: application/json" -d '{"id": "34d63bd4-56b3-4795-80d4-86e5db6fa0b5", "title":"TestBis", "description":"TestBis", "completed": true}' "http://localhost:8080/v1/todo"
{}
```

- Delete a Todo:

```bash
curl -X DELETE "http://localhost:8080/v1/todo/34d63bd4-56b3-4795-80d4-86e5db6fa0b5"
{}
```

- Bulk Insert Todos:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"items": [{"title":"Todo_1","description":"Todo_1"},{"title":"Todo_2","description":"Todo_2"}]}' "http://localhost:8080/v1/todo/bulk"
{"ids":["e8924469-8847-4840-ae16-21be734173f4","0db11e34-4707-4a5d-92fe-f4952213d940"]}
```

- Bulk Update Todos:

```bash
curl -X PUT -H "Content-Type: application/json" -d '{"items": [{"id":"e94a6d0b-953b-4dad-aecb-318f183db4c7","title":"Todo_1","description":"Todo_1","completed":true},{"id":"d53daa2c-e6af-45ba-b192-3e1dc443b165","title":"Todo_2","description":"Todo_2","completed":true}]}' "http://localhost:8080/v1/todo/bulk"
{}
```

## Language/Libraries

- golang
- grpc-gateway
- go-grpc-middleware
- postgresql (access through go-pg orm)
- gRPC: Protocol buffers v3 ([documentation](https://developers.google.com/protocol-buffers/))

## Author

Alexandre Beslic

- [abronan.com](https://abronan.com)
- [@abronan](https://twitter.com/abronan)

## License

This work is released under the MIT license. A copy of the license is provided in the LICENSE file.
