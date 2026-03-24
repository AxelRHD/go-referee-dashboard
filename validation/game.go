package validation

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

type GameData struct {
	GameDate    string
	GameTime    string
	HomeTeamID  int64
	AwayTeamID  int64
	VenueID     int64
	LeagueID    int64
	Position    string
	RefereeFee  float64
	TravelCosts float64
	KmDriven    int64
	Exhibition  int64
	Remarks     string
}

func ValidateGame(form url.Values) (GameData, map[string]string) {
	errors := make(map[string]string)

	// game_date: required, ISO format
	gameDate := requireField(form, "game_date", "Datum", errors)
	if gameDate != "" && !dateRegex.MatchString(gameDate) {
		errors["game_date"] = "Datum muss im Format JJJJ-MM-TT sein."
	}

	gameTime := optionalField(form, "game_time")

	// home_team_id: required integer
	homeTeamID := parseRequiredID(form, "home_team_id", "Heimteam", errors)

	// away_team_id: required integer
	awayTeamID := parseRequiredID(form, "away_team_id", "Gastteam", errors)

	// Cross-field: teams must differ
	if homeTeamID != 0 && awayTeamID != 0 && homeTeamID == awayTeamID {
		errors["away_team_id"] = "Heim- und Gastteam dürfen nicht identisch sein."
	}

	// league_id: required integer
	leagueID := parseRequiredID(form, "league_id", "Liga", errors)

	// position: required
	position := requireField(form, "position", "Position", errors)

	// optional numeric fields
	refereeFee := parseFloatField(form, "referee_fee", "Honorar", errors, 0)
	travelCosts := parseFloatField(form, "travel_costs", "Fahrtkosten", errors, 0)
	kmDriven := parseIntField(form, "km_driven", "Kilometer", errors, 0)

	// venue_id: optional
	var venueID int64
	if raw := strings.TrimSpace(form.Get("venue_id")); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
			venueID = v
		}
	}

	// exhibition: checkbox
	var exhibition int64
	if form.Get("exhibition") != "" {
		exhibition = 1
	}

	remarks := optionalField(form, "remarks")

	data := GameData{
		GameDate:    gameDate,
		GameTime:    gameTime,
		HomeTeamID:  homeTeamID,
		AwayTeamID:  awayTeamID,
		VenueID:     venueID,
		LeagueID:    leagueID,
		Position:    position,
		RefereeFee:  refereeFee,
		TravelCosts: travelCosts,
		KmDriven:    kmDriven,
		Exhibition:  exhibition,
		Remarks:     remarks,
	}

	return data, errors
}

func parseRequiredID(form url.Values, field, label string, errors map[string]string) int64 {
	raw := strings.TrimSpace(form.Get(field))
	if raw == "" {
		errors[field] = label + " ist ein Pflichtfeld."
		return 0
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		errors[field] = label + " ist ungültig."
		return 0
	}
	return v
}
