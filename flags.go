package main

import (
	"github.com/urfave/cli"
)

var commonFlags = []cli.Flag{
	// Server
	cli.StringFlag{
		Name:   "bind-http",
		Usage:  "bind address for HTTP",
		EnvVar: "BIND_HTTP",
		Value:  "0.0.0.0:8080",
	},
	cli.StringFlag{
		Name:   "bind-grpc",
		Usage:  "bind address for gRPC",
		EnvVar: "BIND_GRPC",
		Value:  "0.0.0.0:2338",
	},

	// PostgresQL
	cli.StringFlag{
		Name:   "db-name",
		Usage:  "database name",
		EnvVar: "DB_NAME",
		Value:  "todo",
	},
	cli.StringFlag{
		Name:   "db-user",
		Usage:  "database username",
		EnvVar: "DB_USER",
		Value:  "postgres",
	},
	cli.StringFlag{
		Name:   "db-password",
		Usage:  "database password",
		EnvVar: "DB_PASSWORD",
	},
	cli.StringFlag{
		Name:   "db-host",
		Usage:  "postgres host",
		EnvVar: "DB_HOST",
		Value:  "127.0.0.1",
	},
	cli.IntFlag{
		Name:   "db-port",
		Usage:  "database port",
		EnvVar: "DB_PORT",
		Value:  5432,
	},
}
