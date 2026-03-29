package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/axelrhd/referee-dashboard/config"
	"github.com/axelrhd/referee-dashboard/store"
)

type ServerFunc func(cfg config.Config) (http.Handler, *store.Store)

func App(cfg config.Config, newServer ServerFunc, version string) *cli.Command {
	serve := serveCmd(cfg, newServer)
	return &cli.Command{
		Name:           "referee-dashboard",
		Usage:          "Referee Dashboard Server",
		Version:        version,
		DefaultCommand: serve.Name,
		Commands: []*cli.Command{
			serve,
			healthCmd(cfg),
		},
	}
}

func serveCmd(cfg config.Config, newServer ServerFunc) *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Start the HTTP server",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			handler, s := newServer(cfg)
			defer s.Close()

			addr := fmt.Sprintf(":%d", cfg.Port)
			log.Printf("listening on %s", addr)
			return http.ListenAndServe(addr, handler)
		},
	}
}

func healthCmd(cfg config.Config) *cli.Command {
	return &cli.Command{
		Name:  "health",
		Usage: "Check if the server is healthy",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			url := fmt.Sprintf("http://localhost:%d/health", cfg.Port)
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("health check failed: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return fmt.Errorf("health check returned %d", resp.StatusCode)
			}
			fmt.Println("ok")
			return nil
		},
	}
}
