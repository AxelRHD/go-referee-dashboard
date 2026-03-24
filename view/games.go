package view

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"github.com/axelrhd/referee-dashboard/model"
)

const gamesPerPage = 25

var monthNames = map[string]string{
	"01": "Januar", "02": "Februar", "03": "März",
	"04": "April", "05": "Mai", "06": "Juni",
	"07": "Juli", "08": "August", "09": "September",
	"10": "Oktober", "11": "November", "12": "Dezember",
}

func eurFormat(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	parts := strings.Split(s, ".")
	integer := parts[0]
	decimal := parts[1]

	// German thousands separator
	var result []byte
	for i, c := range integer {
		if i > 0 && (len(integer)-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	return string(result) + "," + decimal + " €"
}

type GameStats struct {
	Count       int
	TotalFee    float64
	TotalTravel float64
	TotalKm     int64
	Positions   map[string]int
}

type FilterOptions struct {
	Seasons   []string
	Months    []struct{ Value, Label string }
	Leagues   []model.League
	Positions []model.Position
}

type GameFilters struct {
	Season   string
	Month    string
	LeagueID string
	Position string
	Query    string
	Page     int
}

func GameList(w http.ResponseWriter, r *http.Request, games []model.ListGamesRow, stats GameStats, filters GameFilters, opts FilterOptions) g.Node {
	return BasePage("Spiele",
		FlashAlert(w, r),
		h.H1(g.Text("Spiele")),
		h.Div(h.Class("d-flex justify-content-between mb-3"),
			h.Div(
				h.A(h.Class("btn btn-success"), g.Attr("href", "/games/new"), g.Text("Neues Spiel")),
			),
			h.Div(
				h.A(h.Class("btn btn-sm btn-outline-secondary me-1"), g.Attr("href", "/games/export/csv"), g.Text("CSV")),
				h.A(h.Class("btn btn-sm btn-outline-secondary"), g.Attr("href", "/games/export/sql"), g.Text("SQL")),
			),
		),
		filterBar(filters, opts),
		GameTable(games, stats, filters),
	)
}

func filterBar(f GameFilters, opts FilterOptions) g.Node {
	seasonOpts := []g.Node{h.Option(g.Attr("value", ""), g.Text("Alle"))}
	for _, s := range opts.Seasons {
		attrs := []g.Node{g.Attr("value", s)}
		if s == f.Season {
			attrs = append(attrs, g.Attr("selected", ""))
		}
		seasonOpts = append(seasonOpts, h.Option(append(attrs, g.Text(s))...))
	}

	monthOpts := []g.Node{h.Option(g.Attr("value", ""), g.Text("Alle"))}
	for _, m := range opts.Months {
		attrs := []g.Node{g.Attr("value", m.Value)}
		if m.Value == f.Month {
			attrs = append(attrs, g.Attr("selected", ""))
		}
		monthOpts = append(monthOpts, h.Option(append(attrs, g.Text(m.Label))...))
	}

	leagueOpts := []g.Node{h.Option(g.Attr("value", ""), g.Text("Alle"))}
	for _, l := range opts.Leagues {
		attrs := []g.Node{g.Attr("value", fmt.Sprintf("%d", l.ID))}
		if fmt.Sprintf("%d", l.ID) == f.LeagueID {
			attrs = append(attrs, g.Attr("selected", ""))
		}
		leagueOpts = append(leagueOpts, h.Option(append(attrs, g.Text(l.Name))...))
	}

	posOpts := []g.Node{h.Option(g.Attr("value", ""), g.Text("Alle"))}
	for _, p := range opts.Positions {
		attrs := []g.Node{g.Attr("value", p.Position)}
		if p.Position == f.Position {
			attrs = append(attrs, g.Attr("selected", ""))
		}
		posOpts = append(posOpts, h.Option(append(attrs, g.Text(p.Position))...))
	}

	filterSelect := func(name string, opts []g.Node) g.Node {
		return h.Select(h.Class("form-select form-select-sm"), g.Attr("name", name), g.Group(opts))
	}

	filterCol := func(label, name string, opts []g.Node) g.Node {
		return h.Div(h.Class("col-auto"),
			h.Small(h.Class("text-muted"), g.Text(label)),
			filterSelect(name, opts),
		)
	}

	return h.FormEl(h.Class("mb-3"), g.Attr("method", "get"), g.Attr("action", "/games"),
		g.Attr("hx-get", "/games"),
		g.Attr("hx-target", "#games-content"),
		g.Attr("hx-trigger", "change"),
		g.Attr("hx-push-url", "true"),
		h.Div(h.Class("row g-2 align-items-end"),
			filterCol("Saison", "season", seasonOpts),
			filterCol("Monat", "month", monthOpts),
			filterCol("Liga", "league_id", leagueOpts),
			filterCol("Position", "position", posOpts),
			h.Div(h.Class("col"),
				h.Small(h.Class("text-muted"), g.Text("Suche")),
				h.Input(
					h.Class("form-control form-control-sm"),
					g.Attr("type", "text"),
					g.Attr("name", "q"),
					g.Attr("placeholder", "Team, Spielort, Bemerkung..."),
					g.Attr("value", f.Query),
					g.Attr("hx-get", "/games"),
					g.Attr("hx-target", "#games-content"),
					g.Attr("hx-push-url", "true"),
					g.Attr("hx-trigger", "input changed delay:400ms, search"),
				),
			),
			h.Div(h.Class("col-auto"),
				h.A(h.Class("btn btn-sm btn-outline-secondary"), g.Attr("href", "/games"), g.Text("Zurücksetzen")),
			),
		),
	)
}

func GameTable(games []model.ListGamesRow, stats GameStats, f GameFilters) g.Node {
	total := len(games)
	totalPages := int(math.Max(1, math.Ceil(float64(total)/float64(gamesPerPage))))
	page := f.Page
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * gamesPerPage
	end := start + gamesPerPage
	if end > total {
		end = total
	}
	pageGames := games[start:end]

	var rows []g.Node
	for _, gm := range pageGames {
		venueDisplay := ""
		if gm.VenueStadium != "" && gm.VenueCity != "" {
			venueDisplay = gm.VenueCity + ", " + gm.VenueStadium
		} else if gm.VenueCity != "" {
			venueDisplay = gm.VenueCity
		}
		totalFee := gm.RefereeFee + gm.TravelCosts
		rows = append(rows, h.Tr(
			h.Td(g.Text(gm.GameDate)),
			h.Td(g.Text(gm.GameTime)),
			h.Td(g.Text(gm.Position)),
			h.Td(g.Text(gm.LeagueName)),
			h.Td(g.Text(gm.HomeTeamName)),
			h.Td(g.Text(gm.AwayTeamName)),
			h.Td(g.Text(venueDisplay)),
			h.Td(h.Class("text-end"), g.Text(eurFormat(gm.RefereeFee))),
			h.Td(h.Class("text-end"), g.Text(eurFormat(gm.TravelCosts))),
			h.Td(h.Class("text-end"), g.Text(eurFormat(totalFee))),
			ActionLinks(
				fmt.Sprintf("/games/%d/edit", gm.ID),
				fmt.Sprintf("/games/%d/delete", gm.ID),
			),
		))
	}

	return h.Div(g.Attr("id", "games-content"),
		statsRow(stats),
		DataTable(
			[]string{"Datum", "Zeit", "Pos", "Liga", "Heim", "Gast", "Spielort", "Vergütung", "Fahrtkosten", "Gesamt", "Aktionen"},
			rows,
		),
		pagination(page, totalPages, f),
	)
}

func statsRow(stats GameStats) g.Node {
	totalAll := stats.TotalFee + stats.TotalTravel

	posOrder := []string{"R", "CJ", "U", "LJ", "LM", "BJ", "FJ", "SJ"}
	var posCards []g.Node
	for _, p := range posOrder {
		if count, ok := stats.Positions[p]; ok {
			posCards = append(posCards, gameStatCard(p, fmt.Sprintf("%d", count)))
		}
	}

	return g.Group([]g.Node{
		h.Div(h.Class("row g-2 mb-2"),
			gameStatCard("Spiele", fmt.Sprintf("%d", stats.Count)),
			gameStatCard("Vergütung", eurFormat(stats.TotalFee)),
			gameStatCard("Fahrtkosten", eurFormat(stats.TotalTravel)),
			gameStatCard("Gesamt", eurFormat(totalAll)),
			gameStatCard("Kilometer", fmt.Sprintf("%d", stats.TotalKm)),
		),
		h.Div(h.Class("row g-2 mb-3"),
			g.Group(posCards),
		),
	})
}

func gameStatCard(label, value string) g.Node {
	return h.Div(h.Class("col"),
		h.Div(h.Class("card"),
			h.Div(h.Class("card-body p-2 text-center"),
				h.Small(h.Class("text-muted"), g.Text(label)),
				h.Div(h.Class("fw-bold"), g.Text(value)),
			),
		),
	)
}

func pagination(page, totalPages int, f GameFilters) g.Node {
	if totalPages <= 1 {
		return nil
	}

	pageURL := func(p int) string {
		parts := []string{fmt.Sprintf("page=%d", p)}
		if f.Season != "" {
			parts = append(parts, "season="+f.Season)
		}
		if f.Month != "" {
			parts = append(parts, "month="+f.Month)
		}
		if f.LeagueID != "" {
			parts = append(parts, "league_id="+f.LeagueID)
		}
		if f.Position != "" {
			parts = append(parts, "position="+f.Position)
		}
		if f.Query != "" {
			parts = append(parts, "q="+f.Query)
		}
		return "/games?" + strings.Join(parts, "&")
	}

	var items []g.Node

	// Previous
	if page == 1 {
		items = append(items, h.Li(h.Class("page-item disabled"),
			h.Span(h.Class("page-link"), g.Text("«")),
		))
	} else {
		items = append(items, h.Li(h.Class("page-item"),
			h.A(h.Class("page-link"), g.Attr("href", pageURL(page-1)), g.Text("«")),
		))
	}

	// Page numbers
	for i := 1; i <= totalPages; i++ {
		cls := "page-item"
		if i == page {
			cls += " active"
		}
		items = append(items, h.Li(h.Class(cls),
			h.A(h.Class("page-link"), g.Attr("href", pageURL(i)), g.Text(fmt.Sprintf("%d", i))),
		))
	}

	// Next
	if page == totalPages {
		items = append(items, h.Li(h.Class("page-item disabled"),
			h.Span(h.Class("page-link"), g.Text("»")),
		))
	} else {
		items = append(items, h.Li(h.Class("page-item"),
			h.A(h.Class("page-link"), g.Attr("href", pageURL(page+1)), g.Text("»")),
		))
	}

	return h.Nav(
		h.Ul(h.Class("pagination pagination-sm"), g.Group(items)),
	)
}

func GameForm(game *model.Game, errors map[string]string, data map[string]string,
	teams []model.Team, leagues []model.League, positions []model.Position, venues []model.Venue) g.Node {

	title := "Neues Spiel"
	action := "/games/new"
	if game != nil {
		title = "Spiel bearbeiten"
		action = fmt.Sprintf("/games/%d/edit", game.ID)
	}

	val := func(field string) string {
		if data != nil {
			if v, ok := data[field]; ok {
				return v
			}
		}
		if game == nil {
			if field == "referee_fee" || field == "travel_costs" || field == "km_driven" {
				return "0"
			}
			return ""
		}
		switch field {
		case "game_date":
			return game.GameDate
		case "game_time":
			return game.GameTime
		case "home_team_id":
			return fmt.Sprintf("%d", game.HomeTeamID)
		case "away_team_id":
			return fmt.Sprintf("%d", game.AwayTeamID)
		case "venue_id":
			if game.VenueID != 0 {
				return fmt.Sprintf("%d", game.VenueID)
			}
			return ""
		case "league_id":
			return fmt.Sprintf("%d", game.LeagueID)
		case "position":
			return game.Position
		case "referee_fee":
			return fmt.Sprintf("%.2f", game.RefereeFee)
		case "travel_costs":
			return fmt.Sprintf("%.2f", game.TravelCosts)
		case "km_driven":
			return fmt.Sprintf("%d", game.KmDriven)
		case "exhibition":
			return fmt.Sprintf("%d", game.Exhibition)
		case "remarks":
			return game.Remarks
		}
		return ""
	}

	errFor := func(field string) string {
		if errors != nil {
			return errors[field]
		}
		return ""
	}

	// Select helpers
	teamSelect := func(name string) g.Node {
		opts := []g.Node{h.Option(g.Attr("value", ""), g.Text("— Team wählen —"))}
		for _, t := range teams {
			attrs := []g.Node{g.Attr("value", fmt.Sprintf("%d", t.ID))}
			if fmt.Sprintf("%d", t.ID) == val(name) {
				attrs = append(attrs, g.Attr("selected", ""))
			}
			opts = append(opts, h.Option(append(attrs, g.Text(t.Name))...))
		}
		cls := "form-select"
		if errFor(name) != "" {
			cls += " is-invalid"
		}
		return h.Select(h.Class(cls), g.Attr("name", name), g.Attr("id", name), g.Group(opts))
	}

	venueOpts := []g.Node{h.Option(g.Attr("value", ""), g.Text("— Spielort wählen —"))}
	for _, v := range venues {
		display := v.City
		if v.Stadium != "" {
			display = v.City + ", " + v.Stadium
		}
		attrs := []g.Node{g.Attr("value", fmt.Sprintf("%d", v.ID))}
		if fmt.Sprintf("%d", v.ID) == val("venue_id") {
			attrs = append(attrs, g.Attr("selected", ""))
		}
		venueOpts = append(venueOpts, h.Option(append(attrs, g.Text(display))...))
	}

	leagueOpts := []g.Node{h.Option(g.Attr("value", ""), g.Text("— Liga wählen —"))}
	for _, l := range leagues {
		attrs := []g.Node{g.Attr("value", fmt.Sprintf("%d", l.ID))}
		if fmt.Sprintf("%d", l.ID) == val("league_id") {
			attrs = append(attrs, g.Attr("selected", ""))
		}
		leagueOpts = append(leagueOpts, h.Option(append(attrs, g.Text(l.Name))...))
	}

	posOpts := []g.Node{h.Option(g.Attr("value", ""), g.Text("— Position wählen —"))}
	for _, p := range positions {
		attrs := []g.Node{g.Attr("value", p.Position)}
		if p.Position == val("position") {
			attrs = append(attrs, g.Attr("selected", ""))
		}
		label := fmt.Sprintf("%s — %s", p.Position, p.Long)
		posOpts = append(posOpts, h.Option(append(attrs, g.Text(label))...))
	}

	leagueCls := "form-select"
	if errFor("league_id") != "" {
		leagueCls += " is-invalid"
	}
	posCls := "form-select"
	if errFor("position") != "" {
		posCls += " is-invalid"
	}

	dateInput := h.Input(h.Class("form-control"), g.Attr("type", "date"), g.Attr("name", "game_date"), g.Attr("id", "game_date"), g.Attr("value", val("game_date")), g.Attr("required", ""))
	timeInput := h.Input(h.Class("form-control"), g.Attr("type", "time"), g.Attr("name", "game_time"), g.Attr("id", "game_time"), g.Attr("value", val("game_time")))

	feeInput := h.Input(h.Class("form-control"), g.Attr("type", "number"), g.Attr("step", "0.01"), g.Attr("name", "referee_fee"), g.Attr("id", "referee_fee"), g.Attr("value", val("referee_fee")))
	travelInput := h.Input(h.Class("form-control"), g.Attr("type", "number"), g.Attr("step", "0.01"), g.Attr("name", "travel_costs"), g.Attr("id", "travel_costs"), g.Attr("value", val("travel_costs")))
	kmInput := h.Input(h.Class("form-control"), g.Attr("type", "number"), g.Attr("step", "1"), g.Attr("name", "km_driven"), g.Attr("id", "km_driven"), g.Attr("value", val("km_driven")))

	isExhibition := val("exhibition") == "1"
	checkboxAttrs := []g.Node{
		h.Class("form-check-input"), g.Attr("type", "checkbox"),
		g.Attr("name", "exhibition"), g.Attr("id", "exhibition"), g.Attr("value", "1"),
	}
	if isExhibition {
		checkboxAttrs = append(checkboxAttrs, g.Attr("checked", ""))
	}

	return BasePage(title,
		h.H1(g.Text(title)),
		h.Form(g.Attr("method", "post"), g.Attr("action", action),
			FormRow(
				FormField("Datum", dateInput, errFor("game_date")),
				FormField("Uhrzeit", timeInput, ""),
			),
			FormRow(
				FormField("Heimteam", teamSelect("home_team_id"), errFor("home_team_id")),
				FormField("Gastteam", teamSelect("away_team_id"), errFor("away_team_id")),
			),
			FormRow(
				FormField("Spielort", h.Select(h.Class("form-select"), g.Attr("name", "venue_id"), g.Attr("id", "venue_id"), g.Group(venueOpts)), ""),
				FormField("Liga", h.Select(h.Class(leagueCls), g.Attr("name", "league_id"), g.Attr("id", "league_id"), g.Group(leagueOpts)), errFor("league_id")),
				FormField("Position", h.Select(h.Class(posCls), g.Attr("name", "position"), g.Attr("id", "position"), g.Group(posOpts)), errFor("position")),
			),
			FormRow(
				FormField("Honorar", feeInput, errFor("referee_fee")),
				FormField("Fahrtkosten", travelInput, errFor("travel_costs")),
				FormField("Kilometer", kmInput, errFor("km_driven")),
			),
			h.Div(h.Class("mb-3 form-check"),
				h.Input(checkboxAttrs...),
				h.Label(h.Class("form-check-label"), g.Attr("for", "exhibition"), g.Text("Freundschaftsspiel")),
			),
			FormField("Bemerkungen", TextareaInput("remarks", val("remarks"), 2, ""), ""),
			SubmitButton("Speichern", "/games"),
		),
	)
}
