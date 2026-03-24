package main

import (
	"context"
	"log"
	"os"

	_ "modernc.org/sqlite"

	appcli "github.com/axelrhd/referee-dashboard/cli"
	"github.com/axelrhd/referee-dashboard/config"
	"github.com/axelrhd/referee-dashboard/server"
)

var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	app := appcli.App(cfg, server.NewServer, version)
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
