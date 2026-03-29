package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/axelrhd/referee-dashboard/config"
	"github.com/axelrhd/referee-dashboard/handler"
	"github.com/axelrhd/referee-dashboard/store"
)

func NewServer(cfg config.Config) (http.Handler, *store.Store) {
	s, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	// Setup handler + redirect middleware
	setup := handler.NewSetupHandler(s)

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

	leagues := handler.NewLeagueHandler(s)
	leagues.Routes(r)

	teams := handler.NewTeamHandler(s)
	teams.Routes(r)

	venues := handler.NewVenueHandler(s)
	venues.Routes(r)

	games := handler.NewGameHandler(s)
	games.Routes(r)

	dashboard := handler.NewDashboardHandler(s)
	dashboard.Routes(r)

	data := handler.NewDataHandler(s)
	data.Routes(r)

	r.Route("/api", func(r chi.Router) {
		leagues.APIRoutes(r)
		teams.APIRoutes(r)
		venues.APIRoutes(r)
		games.APIRoutes(r)
		dashboard.APIRoutes(r)
	})

	return r, s
}
