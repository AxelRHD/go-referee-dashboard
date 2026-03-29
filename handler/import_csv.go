package handler

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/axelrhd/referee-dashboard/store"
	"github.com/axelrhd/referee-dashboard/validation"
)

var gameColumns = map[string]string{
	"datum":              "game_date",
	"uhrzeit":            "game_time",
	"heimteam":           "home_team_id",
	"gastteam":           "away_team_id",
	"spielort":           "venue_id",
	"liga":               "league_id",
	"position":           "position",
	"honorar":            "referee_fee",
	"fahrtkosten":        "travel_costs",
	"kilometer":          "km_driven",
	"freundschaftsspiel": "exhibition",
	"bemerkungen":        "remarks",
}

var teamColumns = map[string]string{
	"name":        "name",
	"bundesland":  "state",
	"aktiv":       "is_active",
	"bemerkungen": "remarks",
}

var leagueColumns = map[string]string{
	"name":        "name",
	"kürzel":      "short_name",
	"kurzel":      "short_name",
	"sortierung":  "sorter",
	"bemerkungen": "remarks",
}

func (dh *DataHandler) importCSV(text, entity string) (int, []string) {
	columnMap := map[string]map[string]string{
		"games":   gameColumns,
		"teams":   teamColumns,
		"leagues": leagueColumns,
	}
	colMap, ok := columnMap[entity]
	if !ok {
		return 0, []string{fmt.Sprintf("Unbekannte Entity: %s", entity)}
	}

	// Detect delimiter
	delimiter := ';'
	firstLine := strings.SplitN(text, "\n", 2)[0]
	if !strings.Contains(firstLine, ";") {
		delimiter = ','
	}

	reader := csv.NewReader(strings.NewReader(text))
	reader.Comma = delimiter
	reader.LazyQuotes = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return 0, []string{"CSV-Header konnte nicht gelesen werden."}
	}

	// Map header indices to form field names
	fieldMap := make([]string, len(header))
	for i, col := range header {
		normalized := normalizeKey(col)
		for german, field := range colMap {
			if normalizeKey(german) == normalized {
				fieldMap[i] = field
				break
			}
		}
	}

	// Build name→ID lookup maps for FK resolution
	teams, _ := dh.s.ListTeams()
	leagues, _ := dh.s.ListLeagues()
	venues, _ := dh.s.ListVenues()

	teamByName := map[string]string{}
	for _, t := range teams {
		teamByName[t.Name] = t.ID
	}
	leagueByName := map[string]string{}
	for _, l := range leagues {
		leagueByName[l.Name] = l.ID
	}
	venueByCity := map[string]string{}
	for _, v := range venues {
		venueByCity[v.City] = v.ID
	}

	var errors []string
	count := 0

	for lineNum := 2; ; lineNum++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors = append(errors, fmt.Sprintf("Zeile %d: %v", lineNum, err))
			continue
		}

		// Build form values from CSV row
		form := map[string][]string{}
		for i, val := range record {
			if i >= len(fieldMap) || fieldMap[i] == "" {
				continue
			}
			field := fieldMap[i]
			val = transformCSVValue(field, val, teamByName, leagueByName, venueByCity)
			form[field] = []string{val}
		}

		// Validate and insert
		switch entity {
		case "games":
			data, errs := validation.ValidateGame(form)
			if len(errs) > 0 {
				msgs := fmtErrors(errs)
				errors = append(errors, fmt.Sprintf("Zeile %d: %s", lineNum, msgs))
				continue
			}
			game, err := buildGameFromCSV(dh.s, data)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Zeile %d: %v", lineNum, err))
				continue
			}
			if err := dh.s.PutGame(game); err != nil {
				errors = append(errors, fmt.Sprintf("Zeile %d: %v", lineNum, err))
				continue
			}
		case "teams":
			data, errs := validation.ValidateTeam(form)
			if len(errs) > 0 {
				errors = append(errors, fmt.Sprintf("Zeile %d: %s", lineNum, fmtErrors(errs)))
				continue
			}
			if err := dh.s.PutTeam(&store.Team{
				Name: data.Name, State: data.State,
				IsActive: data.IsActive, Remarks: data.Remarks,
			}); err != nil {
				errors = append(errors, fmt.Sprintf("Zeile %d: %v", lineNum, err))
				continue
			}
		case "leagues":
			data, errs := validation.ValidateLeague(form)
			if len(errs) > 0 {
				errors = append(errors, fmt.Sprintf("Zeile %d: %s", lineNum, fmtErrors(errs)))
				continue
			}
			if err := dh.s.PutLeague(&store.League{
				Name: data.Name, ShortName: data.ShortName,
				Sorter: int(data.Sorter), Remarks: data.Remarks,
			}); err != nil {
				errors = append(errors, fmt.Sprintf("Zeile %d: %v", lineNum, err))
				continue
			}
		}
		count++
	}

	return count, errors
}

func buildGameFromCSV(s *store.Store, data validation.GameData) (*store.Game, error) {
	homeTeam, err := s.GetTeam(data.HomeTeamID)
	if err != nil {
		return nil, fmt.Errorf("heimteam: %w", err)
	}
	awayTeam, err := s.GetTeam(data.AwayTeamID)
	if err != nil {
		return nil, fmt.Errorf("gastteam: %w", err)
	}
	league, err := s.GetLeague(data.LeagueID)
	if err != nil {
		return nil, fmt.Errorf("liga: %w", err)
	}

	var venueRef store.VenueRef
	if data.VenueID != "" {
		venue, err := s.GetVenue(data.VenueID)
		if err != nil {
			return nil, fmt.Errorf("spielort: %w", err)
		}
		venueRef = store.VenueRef{
			ID: venue.ID, City: venue.City, ShortName: venue.ShortName,
			Lat: venue.Lat, Lon: venue.Lon,
		}
	}

	return &store.Game{
		GameDate: data.GameDate, GameTime: data.GameTime,
		HomeTeam: store.TeamRef{ID: homeTeam.ID, Name: homeTeam.Name},
		AwayTeam: store.TeamRef{ID: awayTeam.ID, Name: awayTeam.Name},
		Venue:    venueRef,
		League:   store.LeagueRef{ID: league.ID, Name: league.Name, ShortName: league.ShortName},
		Position: data.Position, RefereeFee: data.RefereeFee,
		TravelCosts: data.TravelCosts, KmDriven: int(data.KmDriven),
		Exhibition: data.Exhibition, Remarks: data.Remarks,
	}, nil
}

func normalizeKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	key = strings.ReplaceAll(key, " ", "")
	key = strings.ReplaceAll(key, "_", "")
	return key
}

func transformCSVValue(field, val string, teamByName, leagueByName, venueByCity map[string]string) string {
	val = strings.TrimSpace(val)

	switch field {
	case "referee_fee", "travel_costs":
		val = strings.ReplaceAll(val, "€", "")
		val = strings.TrimSpace(val)
		val = strings.ReplaceAll(val, ".", "")  // thousands separator
		val = strings.ReplaceAll(val, ",", ".") // decimal comma → dot
	case "exhibition", "is_active":
		lower := strings.ToLower(val)
		if lower == "ja" || lower == "1" || lower == "true" || lower == "yes" || lower == "x" {
			val = "1"
		} else {
			val = ""
		}
	case "home_team_id", "away_team_id":
		if id, ok := teamByName[val]; ok {
			val = id
		}
	case "league_id":
		if id, ok := leagueByName[val]; ok {
			val = id
		}
	case "venue_id":
		if id, ok := venueByCity[val]; ok {
			val = id
		}
	}

	return val
}

func fmtErrors(errs map[string]string) string {
	var parts []string
	for k, v := range errs {
		parts = append(parts, k+": "+v)
	}
	return strings.Join(parts, "; ")
}
