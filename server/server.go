package server

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/axelrhd/referee-dashboard/config"
	"github.com/axelrhd/referee-dashboard/handler"
	"github.com/axelrhd/referee-dashboard/model"
)

func NewServer(cfg config.Config) http.Handler {
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")

	queries := model.New(db)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	// Setup handler + redirect middleware
	setup := handler.NewSetupHandler(db)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			// Skip redirect for setup, static, health, and API routes
			if strings.HasPrefix(path, "/setup") ||
				strings.HasPrefix(path, "/static") ||
				strings.HasPrefix(path, "/api") ||
				path == "/health" {
				next.ServeHTTP(w, r)
				return
			}
			if setup.NeedsSetup() {
				http.Redirect(w, r, "/setup", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	setup.Routes(r)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard/", http.StatusFound)
	})

	leagues := handler.NewLeagueHandler(queries)
	leagues.Routes(r)

	teams := handler.NewTeamHandler(queries)
	teams.Routes(r)

	venues := handler.NewVenueHandler(queries)
	venues.Routes(r)

	games := handler.NewGameHandler(queries)
	games.Routes(r)

	dashboard := handler.NewDashboardHandler(queries)
	dashboard.Routes(r)

	data := handler.NewDataHandler(queries, db)
	data.Routes(r)

	r.Route("/api", func(r chi.Router) {
		leagues.APIRoutes(r)
		teams.APIRoutes(r)
		venues.APIRoutes(r)
		games.APIRoutes(r)
		dashboard.APIRoutes(r)
	})

	return r
}
