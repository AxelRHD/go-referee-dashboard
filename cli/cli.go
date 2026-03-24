package cli

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"

	"github.com/axelrhd/referee-dashboard/config"
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
			seedCmd(cfg),
		},
	}
}

func serveCmd(cfg config.Config, newServer ServerFunc) *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Start the HTTP server",
		Action: func(ctx context.Context, cmd *cli.Command) error {
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
			db, err := openDB(cfg.DBPath)
			if err != nil {
				return err
			}
			defer db.Close()

			goose.SetBaseFS(os.DirFS("db/migrations"))
			if err := goose.SetDialect("sqlite3"); err != nil {
				return err
			}
			return goose.Up(db, ".")
		},
	}
}

func seedCmd(cfg config.Config) *cli.Command {
	return &cli.Command{
		Name:  "seed",
		Usage: "Seed initial data (positions)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			db, err := openDB(cfg.DBPath)
			if err != nil {
				return err
			}
			defer db.Close()

			return seedPositions(db)
		},
	}
}

func openDB(path string) (*sql.DB, error) {
	return sql.Open("sqlite", path)
}

func seedPositions(db *sql.DB) error {
	positions := []struct {
		Position string
		Long     string
		Sorter   int
	}{
		{"R", "Referee", 10},
		{"CJ", "Center Judge", 20},
		{"U", "Umpire", 30},
		{"LJ", "Line Judge", 40},
		{"LM", "Linesman", 50},
		{"BJ", "Back Judge", 60},
		{"FJ", "Field Judge", 70},
		{"SJ", "Side Judge", 80},
	}

	for _, p := range positions {
		_, err := db.Exec(
			`INSERT OR IGNORE INTO positions (position, long, sorter) VALUES (?, ?, ?)`,
			p.Position, p.Long, p.Sorter,
		)
		if err != nil {
			return fmt.Errorf("seed position %s: %w", p.Position, err)
		}
	}

	log.Printf("seeded %d positions", len(positions))
	return nil
}
