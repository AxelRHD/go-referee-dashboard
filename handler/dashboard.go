package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/model"
	"github.com/axelrhd/referee-dashboard/view"
)

type DashboardHandler struct {
	q *model.Queries
}

func NewDashboardHandler(q *model.Queries) *DashboardHandler {
	return &DashboardHandler{q: q}
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
	seasons, _ := dh.q.ListSeasons(r.Context())
	defaultSeason := ""
	if len(seasons) > 0 {
		defaultSeason = seasons[0]
	}
	page := view.DashboardPage(seasons, defaultSeason)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (dh *DashboardHandler) APIOverview(w http.ResponseWriter, r *http.Request) {
	games, _ := dh.q.ListGamesFull(r.Context())
	positions, _ := dh.q.GetAllPositions(r.Context())
	leagues, _ := dh.q.ListLeagues(r.Context())

	// Collect league IDs present in games
	leagueIDSet := map[int64]bool{}
	for _, g := range games {
		leagueIDSet[g.LeagueID] = true
	}

	// Build response
	type gameJSON struct {
		Year     string  `json:"year"`
		Date     string  `json:"date"`
		Position string  `json:"position"`
		LeagueID int64   `json:"league_id"`
		League   string  `json:"league"`
		Fee      float64 `json:"fee"`
		Travel   float64 `json:"travel"`
		Km       int64   `json:"km"`
		Venue    string  `json:"venue"`
		VenueLat float64 `json:"venue_lat"`
		VenueLon float64 `json:"venue_lon"`
	}

	data := make([]gameJSON, 0, len(games))
	for _, g := range games {
		data = append(data, gameJSON{
			Year:     g.Year,
			Date:     g.GameDate,
			Position: g.Position,
			LeagueID: g.LeagueID,
			League:   g.LeagueShort,
			Fee:      g.RefereeFee,
			Travel:   g.TravelCosts,
			Km:       g.KmDriven,
			Venue:    g.VenueDisplay,
			VenueLat: g.VenueLat,
			VenueLon: g.VenueLon,
		})
	}

	posNames := make([]string, 0, len(positions))
	for _, p := range positions {
		posNames = append(posNames, p.Position)
	}

	type leagueJSON struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	availLeagues := make([]leagueJSON, 0)
	for _, lg := range leagues {
		if leagueIDSet[lg.ID] {
			availLeagues = append(availLeagues, leagueJSON{ID: lg.ID, Name: lg.Name})
		}
	}

	jsonResponse(w, map[string]any{
		"games":             data,
		"positions":         posNames,
		"available_leagues": availLeagues,
	})
}

func (dh *DashboardHandler) APISeason(w http.ResponseWriter, r *http.Request) {
	season := chi.URLParam(r, "season")
	games, _ := dh.q.ListGamesBySeason(r.Context(), season)
	positions, _ := dh.q.GetAllPositions(r.Context())
	leagues, _ := dh.q.ListLeagues(r.Context())

	leagueShort := map[int64]string{}
	leagueLong := map[int64]string{}
	for _, lg := range leagues {
		if lg.ShortName != "" {
			leagueShort[lg.ID] = lg.ShortName
		} else {
			leagueShort[lg.ID] = lg.Name
		}
		leagueLong[lg.ID] = lg.Name
	}

	type gameJSON struct {
		Date       string  `json:"date"`
		Time       string  `json:"time"`
		Month      string  `json:"month"`
		Home       string  `json:"home"`
		Away       string  `json:"away"`
		Venue      string  `json:"venue"`
		VenueLat   float64 `json:"venue_lat"`
		VenueLon   float64 `json:"venue_lon"`
		League     string  `json:"league"`
		LeagueLong string  `json:"league_long"`
		LeagueID   int64   `json:"league_id"`
		Position   string  `json:"position"`
		Fee        float64 `json:"fee"`
		Travel     float64 `json:"travel"`
		Km         int64   `json:"km"`
		Exhibition bool    `json:"exhibition"`
	}

	data := make([]gameJSON, 0, len(games))
	leagueIDSet := map[int64]bool{}
	gamePosSet := map[string]bool{}
	for _, g := range games {
		leagueIDSet[g.LeagueID] = true
		gamePosSet[g.Position] = true
		data = append(data, gameJSON{
			Date:       g.GameDate,
			Time:       g.GameTime,
			Month:      g.Month,
			Home:       g.HomeTeamName,
			Away:       g.AwayTeamName,
			Venue:      g.VenueDisplay,
			VenueLat:   g.VenueLat,
			VenueLon:   g.VenueLon,
			League:     leagueShort[g.LeagueID],
			LeagueLong: leagueLong[g.LeagueID],
			LeagueID:   g.LeagueID,
			Position:   g.Position,
			Fee:        g.RefereeFee,
			Travel:     g.TravelCosts,
			Km:         g.KmDriven,
			Exhibition: g.Exhibition == 1,
		})
	}

	posNames := make([]string, 0, len(positions))
	availPos := make([]string, 0)
	for _, p := range positions {
		posNames = append(posNames, p.Position)
		if gamePosSet[p.Position] {
			availPos = append(availPos, p.Position)
		}
	}

	type leagueJSON struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	availLeagues := make([]leagueJSON, 0)
	for _, lg := range leagues {
		if leagueIDSet[lg.ID] {
			availLeagues = append(availLeagues, leagueJSON{ID: lg.ID, Name: lg.Name})
		}
	}

	jsonResponse(w, map[string]any{
		"season":              season,
		"games":               data,
		"positions":           posNames,
		"available_positions": availPos,
		"available_leagues":   availLeagues,
	})
}
