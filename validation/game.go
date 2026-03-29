package validation

import (
	"net/url"
	"regexp"
	"strings"
)

var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

type GameData struct {
	GameDate    string
	GameTime    string
	HomeTeamID  string
	AwayTeamID  string
	VenueID     string
	LeagueID    string
	Position    string
	RefereeFee  float64
	TravelCosts float64
	KmDriven    int64
	Exhibition  bool
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

	// home_team_id: required string
	homeTeamID := requireField(form, "home_team_id", "Heimteam", errors)

	// away_team_id: required string
	awayTeamID := requireField(form, "away_team_id", "Gastteam", errors)

	// Cross-field: teams must differ
	if homeTeamID != "" && awayTeamID != "" && homeTeamID == awayTeamID {
		errors["away_team_id"] = "Heim- und Gastteam dürfen nicht identisch sein."
	}

	// league_id: required string
	leagueID := requireField(form, "league_id", "Liga", errors)

	// position: required
	position := requireField(form, "position", "Position", errors)

	// optional numeric fields
	refereeFee := parseFloatField(form, "referee_fee", "Honorar", errors, 0)
	travelCosts := parseFloatField(form, "travel_costs", "Fahrtkosten", errors, 0)
	kmDriven := parseIntField(form, "km_driven", "Kilometer", errors, 0)

	// venue_id: optional
	venueID := strings.TrimSpace(form.Get("venue_id"))

	// exhibition: checkbox
	exhibition := form.Get("exhibition") != ""

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
