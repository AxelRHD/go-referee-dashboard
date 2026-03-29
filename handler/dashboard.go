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

	// Collect league IDs present in games
	leagueIDSet := map[string]bool{}
	for _, g := range games {
		leagueIDSet[g.League.ID] = true
	}

	// Build response
	type gameJSON struct {
		Year     string  `json:"year"`
		Date     string  `json:"date"`
		Position string  `json:"position"`
		LeagueID string  `json:"league_id"`
		League   string  `json:"league"`
		Fee      float64 `json:"fee"`
		Travel   float64 `json:"travel"`
		Km       int     `json:"km"`
		Venue    string  `json:"venue"`
		VenueLat float64 `json:"venue_lat"`
		VenueLon float64 `json:"venue_lon"`
	}

	data := make([]gameJSON, 0, len(games))
	for _, g := range games {
		year := ""
		if len(g.GameDate) >= 4 {
			year = g.GameDate[:4]
		}
		leagueShort := g.League.ShortName
		if leagueShort == "" {
			leagueShort = g.League.Name
		}
		data = append(data, gameJSON{
			Year:     year,
			Date:     g.GameDate,
			Position: g.Position,
			LeagueID: g.League.ID,
			League:   leagueShort,
			Fee:      g.RefereeFee,
			Travel:   g.TravelCosts,
			Km:       g.KmDriven,
			Venue:    venueLabel(g.Venue),
			VenueLat: g.Venue.Lat,
			VenueLon: g.Venue.Lon,
		})
	}

	gamePosSet := map[string]bool{}
	for _, g := range games {
		gamePosSet[g.Position] = true
	}
	posNames := make([]string, 0, len(positions))
	for _, p := range positions {
		if gamePosSet[p.Position] {
			posNames = append(posNames, p.Position)
		}
	}

	type leagueJSON struct {
		ID   string `json:"id"`
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
	games, _ := dh.s.ListGamesBySeason(season)
	positions, _ := dh.s.ListPositions()
	leagues, _ := dh.s.ListLeagues()

	leagueShortMap := map[string]string{}
	leagueLongMap := map[string]string{}
	for _, lg := range leagues {
		if lg.ShortName != "" {
			leagueShortMap[lg.ID] = lg.ShortName
		} else {
			leagueShortMap[lg.ID] = lg.Name
		}
		leagueLongMap[lg.ID] = lg.Name
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
		LeagueID   string  `json:"league_id"`
		Position   string  `json:"position"`
		Fee        float64 `json:"fee"`
		Travel     float64 `json:"travel"`
		Km         int     `json:"km"`
		Exhibition bool    `json:"exhibition"`
	}

	data := make([]gameJSON, 0, len(games))
	leagueIDSet := map[string]bool{}
	gamePosSet := map[string]bool{}
	for _, g := range games {
		leagueIDSet[g.League.ID] = true
		gamePosSet[g.Position] = true
		month := ""
		if len(g.GameDate) >= 7 {
			month = g.GameDate[5:7]
		}
		data = append(data, gameJSON{
			Date:       g.GameDate,
			Time:       g.GameTime,
			Month:      month,
			Home:       g.HomeTeam.Name,
			Away:       g.AwayTeam.Name,
			Venue:      venueLabel(g.Venue),
			VenueLat:   g.Venue.Lat,
			VenueLon:   g.Venue.Lon,
			League:     leagueShortMap[g.League.ID],
			LeagueLong: leagueLongMap[g.League.ID],
			LeagueID:   g.League.ID,
			Position:   g.Position,
			Fee:        g.RefereeFee,
			Travel:     g.TravelCosts,
			Km:         g.KmDriven,
			Exhibition: g.Exhibition,
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
		ID   string `json:"id"`
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

func venueLabel(v store.VenueRef) string {
	if v.ShortName != "" {
		return v.ShortName
	}
	return v.City
}
