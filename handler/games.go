package handler

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/model"
	"github.com/axelrhd/referee-dashboard/validation"
	"github.com/axelrhd/referee-dashboard/view"
)

type GameHandler struct {
	q *model.Queries
}

func NewGameHandler(q *model.Queries) *GameHandler {
	return &GameHandler{q: q}
}

func (h *GameHandler) Routes(r chi.Router) {
	r.Get("/games", h.List)
	r.Get("/games/new", h.NewForm)
	r.Post("/games/new", h.Create)
	r.Get("/games/{id}/edit", h.EditForm)
	r.Post("/games/{id}/edit", h.Update)
	r.Post("/games/{id}/delete", h.Delete)
	r.Get("/games/export/csv", h.ExportCSV)
	r.Get("/games/export/sql", h.ExportSQL)
}

func (h *GameHandler) APIRoutes(r chi.Router) {
	r.Get("/games", h.APIList)
	r.Get("/games/{id}", h.APIGet)
}

// HTML Handlers

func (gh *GameHandler) List(w http.ResponseWriter, r *http.Request) {
	games, err := gh.q.ListGames(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filters := parseGameFilters(r)
	games = filterGames(games, filters)
	stats := calcStats(games)
	opts := gh.filterOptions(r)

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
	teams, leagues, positions, venues := gh.formData(r)
	page := view.GameForm(nil, nil, nil, teams, leagues, positions, venues)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (gh *GameHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	data, errors := validation.ValidateGame(r.Form)

	if len(errors) > 0 {
		teams, leagues, positions, venues := gh.formData(r)
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.GameForm(nil, errors, formValues(r), teams, leagues, positions, venues)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	_, err := gh.q.CreateGame(r.Context(), model.CreateGameParams{
		GameDate:    data.GameDate,
		GameTime:    data.GameTime,
		HomeTeamID:  data.HomeTeamID,
		AwayTeamID:  data.AwayTeamID,
		VenueID:     data.VenueID,
		LeagueID:    data.LeagueID,
		Position:    data.Position,
		RefereeFee:  data.RefereeFee,
		TravelCosts: data.TravelCosts,
		KmDriven:    data.KmDriven,
		Exhibition:  data.Exhibition,
		Remarks:     data.Remarks,
	})
	if err != nil {
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
	teams, leagues, positions, venues := gh.formData(r)
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
	data, errors := validation.ValidateGame(r.Form)

	if len(errors) > 0 {
		teams, leagues, positions, venues := gh.formData(r)
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.GameForm(&game, errors, formValues(r), teams, leagues, positions, venues)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	err = gh.q.UpdateGame(r.Context(), model.UpdateGameParams{
		ID:          game.ID,
		GameDate:    data.GameDate,
		GameTime:    data.GameTime,
		HomeTeamID:  data.HomeTeamID,
		AwayTeamID:  data.AwayTeamID,
		VenueID:     data.VenueID,
		LeagueID:    data.LeagueID,
		Position:    data.Position,
		RefereeFee:  data.RefereeFee,
		TravelCosts: data.TravelCosts,
		KmDriven:    data.KmDriven,
		Exhibition:  data.Exhibition,
		Remarks:     data.Remarks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Spiel wurde aktualisiert.")
	http.Redirect(w, r, "/games", http.StatusSeeOther)
}

func (gh *GameHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	if err := gh.q.DeleteGame(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Spiel wurde gelöscht.")
	http.Redirect(w, r, "/games", http.StatusSeeOther)
}

// JSON API

func (gh *GameHandler) APIList(w http.ResponseWriter, r *http.Request) {
	games, err := gh.q.ListGames(r.Context())
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
	games, err := gh.q.ListGames(r.Context())
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
		venueDisplay := ""
		if gm.VenueStadium != "" && gm.VenueCity != "" {
			venueDisplay = gm.VenueCity + ", " + gm.VenueStadium
		} else if gm.VenueCity != "" {
			venueDisplay = gm.VenueCity
		}
		fee := strings.Replace(fmt.Sprintf("%.2f", gm.RefereeFee), ".", ",", 1)
		travel := strings.Replace(fmt.Sprintf("%.2f", gm.TravelCosts), ".", ",", 1)
		exhibition := "Nein"
		if gm.Exhibition == 1 {
			exhibition = "Ja"
		}
		cw.Write([]string{
			gm.GameDate, gm.GameTime, gm.HomeTeamName, gm.AwayTeamName,
			venueDisplay, gm.LeagueName, gm.Position,
			fee, travel, fmt.Sprintf("%d", gm.KmDriven), exhibition, gm.Remarks,
		})
	}
	cw.Flush()
}

func (gh *GameHandler) ExportSQL(w http.ResponseWriter, r *http.Request) {
	games, err := gh.q.ListGames(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", exportFilename("spiele", "sql"))

	for _, gm := range games {
		fmt.Fprintf(w, "INSERT INTO games (game_date, game_time, home_team_id, away_team_id, venue_id, league_id, position, referee_fee, travel_costs, km_driven, exhibition, remarks) VALUES (%s, %s, %d, %d, %d, %d, %s, %f, %f, %d, %d, %s);\n",
			sqlEscape(gm.GameDate), sqlEscape(gm.GameTime),
			gm.HomeTeamID, gm.AwayTeamID, gm.VenueID, gm.LeagueID,
			sqlEscape(gm.Position), gm.RefereeFee, gm.TravelCosts,
			gm.KmDriven, gm.Exhibition, sqlEscape(gm.Remarks))
	}
}

// Helpers

func (gh *GameHandler) getGame(w http.ResponseWriter, r *http.Request) (model.Game, error) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return model.Game{}, err
	}
	game, err := gh.q.GetGame(r.Context(), id)
	if err == sql.ErrNoRows {
		http.Error(w, "Spiel nicht gefunden", http.StatusNotFound)
		return model.Game{}, err
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return model.Game{}, err
	}
	return game, nil
}

func (gh *GameHandler) formData(r *http.Request) ([]model.Team, []model.League, []model.Position, []model.Venue) {
	teams, _ := gh.q.ListTeams(r.Context())
	leagues, _ := gh.q.ListLeagues(r.Context())
	positions, _ := gh.q.GetAllPositions(r.Context())
	venues, _ := gh.q.ListVenues(r.Context())
	return teams, leagues, positions, venues
}

func (gh *GameHandler) filterOptions(r *http.Request) view.FilterOptions {
	seasons, _ := gh.q.ListSeasons(r.Context())
	leagues, _ := gh.q.ListLeagues(r.Context())
	positions, _ := gh.q.GetAllPositions(r.Context())

	// Build month options from distinct months in games
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

func filterGames(games []model.ListGamesRow, f view.GameFilters) []model.ListGamesRow {
	var result []model.ListGamesRow
	for _, gm := range games {
		if f.Season != "" && !strings.HasPrefix(gm.GameDate, f.Season+"-") {
			continue
		}
		if f.Month != "" && len(gm.GameDate) >= 7 && gm.GameDate[5:7] != f.Month {
			continue
		}
		if f.LeagueID != "" && fmt.Sprintf("%d", gm.LeagueID) != f.LeagueID {
			continue
		}
		if f.Position != "" && gm.Position != f.Position {
			continue
		}
		if f.Query != "" {
			q := strings.ToLower(f.Query)
			if !strings.Contains(strings.ToLower(gm.HomeTeamName), q) &&
				!strings.Contains(strings.ToLower(gm.AwayTeamName), q) &&
				!strings.Contains(strings.ToLower(gm.VenueCity), q) &&
				!strings.Contains(strings.ToLower(gm.VenueStadium), q) &&
				!strings.Contains(strings.ToLower(gm.Remarks), q) {
				continue
			}
		}
		result = append(result, gm)
	}
	return result
}

func calcStats(games []model.ListGamesRow) view.GameStats {
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
