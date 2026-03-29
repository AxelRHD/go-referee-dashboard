package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/store"
	"github.com/axelrhd/referee-dashboard/view"
)

type DashboardHandler struct {
	s *store.Store
}

func NewDashboardHandler(s *store.Store) *DashboardHandler {
	return &DashboardHandler{s: s}
}

func (h *DashboardHandler) Routes(r chi.Router) {
	r.Get("/dashboard", h.Page)
	r.Get("/dashboard/", h.Page)
}

func (h *DashboardHandler) APIRoutes(r chi.Router) {
	r.Get("/dashboard/overview", h.APIOverview)
	r.Get("/dashboard/season/{season}", h.APISeason)
}

func (dh *DashboardHandler) Page(w http.ResponseWriter, r *http.Request) {
	seasons, _ := dh.s.ListSeasons()
	if len(seasons) == 0 {
		http.Redirect(w, r, "/games", http.StatusFound)
		return
	}
	defaultSeason := seasons[0]
	page := view.DashboardPage(seasons, defaultSeason)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (dh *DashboardHandler) APIOverview(w http.ResponseWriter, r *http.Request) {
	games, _ := dh.s.ListGames()
	positions, _ := dh.s.ListPositions()
	leagues, _ := dh.s.ListLeagues()

	gamePosSet := map[string]bool{}
	leagueIDSet := map[string]bool{}
	for _, g := range games {
		gamePosSet[g.Position] = true
		leagueIDSet[g.League.ID] = true
	}

	jsonResponse(w, map[string]any{
		"games":             games,
		"positions":         filterPositions(positions, gamePosSet),
		"available_leagues": filterLeagues(leagues, leagueIDSet),
	})
}

func (dh *DashboardHandler) APISeason(w http.ResponseWriter, r *http.Request) {
	season := chi.URLParam(r, "season")
	games, _ := dh.s.ListGamesBySeason(season)
	positions, _ := dh.s.ListPositions()
	leagues, _ := dh.s.ListLeagues()

	gamePosSet := map[string]bool{}
	leagueIDSet := map[string]bool{}
	for _, g := range games {
		gamePosSet[g.Position] = true
		leagueIDSet[g.League.ID] = true
	}

	allPos := make([]string, 0, len(positions))
	for _, p := range positions {
		allPos = append(allPos, p.Position)
	}

	jsonResponse(w, map[string]any{
		"season":              season,
		"games":               games,
		"positions":           allPos,
		"available_positions": filterPositions(positions, gamePosSet),
		"available_leagues":   filterLeagues(leagues, leagueIDSet),
	})
}

func filterPositions(positions []store.Position, present map[string]bool) []string {
	var result []string
	for _, p := range positions {
		if present[p.Position] {
			result = append(result, p.Position)
		}
	}
	return result
}

type leagueJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func filterLeagues(leagues []store.League, present map[string]bool) []leagueJSON {
	var result []leagueJSON
	for _, lg := range leagues {
		if present[lg.ID] {
			result = append(result, leagueJSON{ID: lg.ID, Name: lg.Name})
		}
	}
	return result
}
