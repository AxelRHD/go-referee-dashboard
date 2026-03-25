package cli

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"

	"github.com/axelrhd/referee-dashboard/config"
	"github.com/axelrhd/referee-dashboard/db"
)

type ServerFunc func(cfg config.Config) http.Handler

func App(cfg config.Config, newServer ServerFunc, version string) *cli.Command {
	return &cli.Command{
		Name:    "referee-dashboard",
		Usage:   "Referee Dashboard Server",
		Version: version,
		Commands: []*cli.Command{
			serveCmd(cfg, newServer),
			migrateCmd(cfg),
			healthCmd(cfg),
		},
	}
}

func serveCmd(cfg config.Config, newServer ServerFunc) *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Start the HTTP server",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			conn, err := openDB(cfg.DBPath)
			if err != nil {
				return err
			}
			if err := runMigrations(conn); err != nil {
				return fmt.Errorf("auto-migrate: %w", err)
			}
			conn.Close()

			handler := newServer(cfg)
			addr := fmt.Sprintf(":%d", cfg.Port)
			log.Printf("listening on %s", addr)
			return http.ListenAndServe(addr, handler)
		},
	}
}

func migrateCmd(cfg config.Config) *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Run database migrations",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			conn, err := openDB(cfg.DBPath)
			if err != nil {
				return err
			}
			defer conn.Close()
			return runMigrations(conn)
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

func openDB(path string) (*sql.DB, error) {
	return sql.Open("sqlite", path)
}

func runMigrations(conn *sql.DB) error {
	goose.SetBaseFS(db.Migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	return goose.Up(conn, "migrations")
}

