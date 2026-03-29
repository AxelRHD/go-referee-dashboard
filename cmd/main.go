package main

import (
	"context"
	"log"
	"os"

	appcli "github.com/axelrhd/referee-dashboard/cli"
	"github.com/axelrhd/referee-dashboard/config"
	"github.com/axelrhd/referee-dashboard/server"
	"github.com/axelrhd/referee-dashboard/view"
)

var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	view.Version = version
	app := appcli.App(cfg, server.NewServer, version)
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
