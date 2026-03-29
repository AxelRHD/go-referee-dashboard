package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/store"
	"github.com/axelrhd/referee-dashboard/validation"
	"github.com/axelrhd/referee-dashboard/view"
)

type GameHandler struct {
	s *store.Store
}

func NewGameHandler(s *store.Store) *GameHandler {
	return &GameHandler{s: s}
}

func (h *GameHandler) Routes(r chi.Router) {
	r.Get("/games", h.List)
	r.Get("/games/new", h.NewForm)
	r.Post("/games/new", h.Create)
	r.Get("/games/{id}/edit", h.EditForm)
	r.Post("/games/{id}/edit", h.Update)
	r.Post("/games/{id}/delete", h.Delete)
	r.Get("/games/export/csv", h.ExportCSV)
}

func (h *GameHandler) APIRoutes(r chi.Router) {
	r.Get("/games", h.APIList)
	r.Get("/games/{id}", h.APIGet)
}

// HTML Handlers

func (gh *GameHandler) List(w http.ResponseWriter, r *http.Request) {
	games, err := gh.s.ListGames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filters := parseGameFilters(r)
	games = filterGames(games, filters)
	stats := calcStats(games)
	opts := gh.filterOptions()

	// HTMX partial
	if r.Header.Get("HX-Request") != "" {
		partial := view.GameTable(games, stats, filters)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		partial.Render(w)
		return
	}

	page := view.GameList(w, r, games, stats, filters, opts)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (gh *GameHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	teams, leagues, positions, venues := gh.formData()
	page := view.GameForm(nil, nil, nil, teams, leagues, positions, venues)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (gh *GameHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	data, errs := validation.ValidateGame(r.Form)

	if len(errs) > 0 {
		teams, leagues, positions, venues := gh.formData()
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.GameForm(nil, errs, formValues(r), teams, leagues, positions, venues)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	game, err := gh.buildGame("", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := gh.s.PutGame(game); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Spiel wurde erstellt.")
	http.Redirect(w, r, "/games", http.StatusSeeOther)
}

func (gh *GameHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	game, err := gh.getGame(w, r)
	if err != nil {
		return
	}
	teams, leagues, positions, venues := gh.formData()
	page := view.GameForm(&game, nil, nil, teams, leagues, positions, venues)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (gh *GameHandler) Update(w http.ResponseWriter, r *http.Request) {
	game, err := gh.getGame(w, r)
	if err != nil {
		return
	}

	r.ParseForm()
	data, errs := validation.ValidateGame(r.Form)

	if len(errs) > 0 {
		teams, leagues, positions, venues := gh.formData()
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.GameForm(&game, errs, formValues(r), teams, leagues, positions, venues)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	updated, err := gh.buildGame(game.ID, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := gh.s.PutGame(updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Spiel wurde aktualisiert.")
	http.Redirect(w, r, "/games", http.StatusSeeOther)
}

func (gh *GameHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	if err := gh.s.DeleteGame(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Spiel wurde gelöscht.")
	http.Redirect(w, r, "/games", http.StatusSeeOther)
}

// JSON API

func (gh *GameHandler) APIList(w http.ResponseWriter, r *http.Request) {
	games, err := gh.s.ListGames()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, games)
}

func (gh *GameHandler) APIGet(w http.ResponseWriter, r *http.Request) {
	game, err := gh.getGame(w, r)
	if err != nil {
		return
	}
	jsonResponse(w, game)
}

// Export

func (gh *GameHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	games, err := gh.s.ListGames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", exportFilename("spiele", "csv"))
	w.Write([]byte("\xEF\xBB\xBF"))

	cw := csv.NewWriter(w)
	cw.Comma = ';'
	cw.Write([]string{"Datum", "Uhrzeit", "Heimteam", "Gastteam", "Spielort", "Liga", "Position",
		"Honorar", "Fahrtkosten", "Kilometer", "Freundschaftsspiel", "Bemerkungen"})

	for _, gm := range games {
		fee := strings.Replace(fmt.Sprintf("%.2f", gm.RefereeFee), ".", ",", 1)
		travel := strings.Replace(fmt.Sprintf("%.2f", gm.TravelCosts), ".", ",", 1)
		exhibition := "Nein"
		if gm.Exhibition {
			exhibition = "Ja"
		}
		cw.Write([]string{
			gm.GameDate, gm.GameTime, gm.HomeTeam.Name, gm.AwayTeam.Name,
			gm.Venue.City, gm.League.Name, gm.Position,
			fee, travel, fmt.Sprintf("%d", gm.KmDriven), exhibition, gm.Remarks,
		})
	}
	cw.Flush()
}

// Helpers

func (gh *GameHandler) buildGame(id string, data validation.GameData) (*store.Game, error) {
	homeTeam, err := gh.s.GetTeam(data.HomeTeamID)
	if err != nil {
		return nil, fmt.Errorf("heimteam: %w", err)
	}
	awayTeam, err := gh.s.GetTeam(data.AwayTeamID)
	if err != nil {
		return nil, fmt.Errorf("gastteam: %w", err)
	}
	league, err := gh.s.GetLeague(data.LeagueID)
	if err != nil {
		return nil, fmt.Errorf("liga: %w", err)
	}

	var venueRef store.VenueRef
	if data.VenueID != "" {
		venue, err := gh.s.GetVenue(data.VenueID)
		if err != nil {
			return nil, fmt.Errorf("spielort: %w", err)
		}
		venueRef = store.VenueRef{
			ID:        venue.ID,
			City:      venue.City,
			ShortName: venue.ShortName,
			Lat:       venue.Lat,
			Lon:       venue.Lon,
		}
	}

	return &store.Game{
		ID:       id,
		GameDate: data.GameDate,
		GameTime: data.GameTime,
		HomeTeam: store.TeamRef{
			ID:   homeTeam.ID,
			Name: homeTeam.Name,
		},
		AwayTeam: store.TeamRef{
			ID:   awayTeam.ID,
			Name: awayTeam.Name,
		},
		Venue: venueRef,
		League: store.LeagueRef{
			ID:        league.ID,
			Name:      league.Name,
			ShortName: league.ShortName,
		},
		Position:    data.Position,
		RefereeFee:  data.RefereeFee,
		TravelCosts: data.TravelCosts,
		KmDriven:    int(data.KmDriven),
		Exhibition:  data.Exhibition,
		Remarks:     data.Remarks,
	}, nil
}

func (gh *GameHandler) getGame(w http.ResponseWriter, r *http.Request) (store.Game, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return store.Game{}, fmt.Errorf("empty id")
	}
	game, err := gh.s.GetGame(id)
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "Spiel nicht gefunden", http.StatusNotFound)
		return store.Game{}, err
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return store.Game{}, err
	}
	return game, nil
}

func (gh *GameHandler) formData() ([]store.Team, []store.League, []store.Position, []store.Venue) {
	teams, _ := gh.s.ListTeams()
	leagues, _ := gh.s.ListLeagues()
	positions, _ := gh.s.ListPositions()
	venues, _ := gh.s.ListVenues()
	return teams, leagues, positions, venues
}

func (gh *GameHandler) filterOptions() view.FilterOptions {
	seasons, _ := gh.s.ListSeasons()
	leagues, _ := gh.s.ListLeagues()
	positions, _ := gh.s.ListPositions()

	months := []struct{ Value, Label string }{
		{"01", "Januar"}, {"02", "Februar"}, {"03", "März"},
		{"04", "April"}, {"05", "Mai"}, {"06", "Juni"},
		{"07", "Juli"}, {"08", "August"}, {"09", "September"},
		{"10", "Oktober"}, {"11", "November"}, {"12", "Dezember"},
	}

	return view.FilterOptions{
		Seasons:   seasons,
		Months:    months,
		Leagues:   leagues,
		Positions: positions,
	}
}

func parseGameFilters(r *http.Request) view.GameFilters {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	return view.GameFilters{
		Season:   r.URL.Query().Get("season"),
		Month:    r.URL.Query().Get("month"),
		LeagueID: r.URL.Query().Get("league_id"),
		Position: r.URL.Query().Get("position"),
		Query:    strings.TrimSpace(r.URL.Query().Get("q")),
		Page:     page,
	}
}

func filterGames(games []store.Game, f view.GameFilters) []store.Game {
	var result []store.Game
	for _, gm := range games {
		if f.Season != "" && !strings.HasPrefix(gm.GameDate, f.Season+"-") {
			continue
		}
		if f.Month != "" && len(gm.GameDate) >= 7 && gm.GameDate[5:7] != f.Month {
			continue
		}
		if f.LeagueID != "" && gm.League.ID != f.LeagueID {
			continue
		}
		if f.Position != "" && gm.Position != f.Position {
			continue
		}
		if f.Query != "" {
			q := strings.ToLower(f.Query)
			if !strings.Contains(strings.ToLower(gm.HomeTeam.Name), q) &&
				!strings.Contains(strings.ToLower(gm.AwayTeam.Name), q) &&
				!strings.Contains(strings.ToLower(gm.Venue.City), q) &&
				!strings.Contains(strings.ToLower(gm.Venue.ShortName), q) &&
				!strings.Contains(strings.ToLower(gm.Remarks), q) {
				continue
			}
		}
		result = append(result, gm)
	}
	return result
}

func calcStats(games []store.Game) view.GameStats {
	stats := view.GameStats{
		Count:     len(games),
		Positions: make(map[string]int),
	}
	for _, gm := range games {
		stats.TotalFee += gm.RefereeFee
		stats.TotalTravel += gm.TravelCosts
		stats.TotalKm += gm.KmDriven
		stats.Positions[gm.Position]++
	}
	return stats
}
