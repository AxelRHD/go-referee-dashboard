package view

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

// --- Helper components (match Python exactly) ---

func statCard(label, xText string) g.Node {
	return h.Div(h.Class("card"),
		h.Div(h.Class("card-body py-2 px-3"),
			h.Small(h.Class("text-muted d-block"), g.Text(label)),
			h.Span(h.Class("fw-bold"), g.Attr("x-text", xText)),
		),
	)
}

func statRow(leftLabel, leftText, rightLabel, rightText string) g.Node {
	return h.Div(h.Class("card"),
		h.Div(h.Class("card-body py-2 px-3 d-flex justify-content-between"),
			h.Div(
				h.Small(h.Class("text-muted d-block"), g.Text(leftLabel)),
				h.Span(h.Class("fw-bold"), g.Attr("x-text", leftText)),
			),
			h.Div(h.Class("text-end"),
				h.Small(h.Class("text-muted d-block"), g.Text(rightLabel)),
				h.Span(h.Class("fw-bold"), g.Attr("x-text", rightText)),
			),
		),
	)
}

func sidebarHeading(text string) g.Node {
	return h.Small(h.Class("text-muted d-block mb-1 mt-2 fw-semibold"), g.Text(text))
}

func collapsibleSection(sectionID, iconClass, title string, content ...g.Node) g.Node {
	return h.Div(h.Class("card mt-3"),
		h.Div(h.Class("card-header d-flex justify-content-between align-items-center py-2"),
			g.Attr("role", "button"),
			g.Attr("data-bs-toggle", "collapse"),
			g.Attr("data-bs-target", "#"+sectionID),
			g.Attr("aria-expanded", "true"),
			h.H6(h.Class("mb-0 text-muted"),
				h.I(h.Class(iconClass+" me-1")),
				g.Text(title),
			),
			h.I(h.Class("bi bi-chevron-down text-muted")),
		),
		h.Div(h.Class("collapse show"), g.Attr("id", sectionID),
			h.Div(h.Class("card-body"), g.Group(content)),
		),
	)
}

func widgetCard(title, chartID string) g.Node {
	return h.Div(h.Class("card h-100"),
		h.Div(h.Class("card-body"),
			h.H6(h.Class("card-title text-muted"), g.Text(title)),
			h.Div(g.Attr("id", chartID)),
		),
	)
}

func alpineSelect(label, xModel string, optionsTemplate g.Node) g.Node {
	return h.Div(h.Class("mb-3"),
		h.Small(h.Class("text-muted d-block mb-1"), g.Text(label)),
		h.Select(h.Class("form-select form-select-sm"), g.Attr("x-model", xModel),
			h.Option(g.Attr("value", ""), g.Text("Alle")),
			optionsTemplate,
		),
	)
}

func gameTableCard(title, icon, xShow, xFor string, countExpr string) g.Node {
	headerContent := []g.Node{
		h.I(h.Class("bi " + icon + " me-1")),
		g.Text(title),
	}
	if countExpr != "" {
		headerContent = append(headerContent,
			h.Small(h.Class("ms-2 text-muted"), g.Attr("x-text", "'(' + "+countExpr+" + ')'")),
		)
	}

	return h.Div(h.Class("card mt-3"), g.Attr("x-show", xShow),
		h.Div(h.Class("card-header py-2"),
			h.H6(h.Class("mb-0 text-muted"), g.Group(headerContent)),
		),
		h.Div(h.Class("card-body p-0"),
			h.Div(h.Class("table-responsive"),
				h.Table(h.Class("table table-sm table-striped table-hover mb-0"),
					h.THead(h.Tr(
						h.Th(h.Class("text-nowrap ps-3"), g.Text("Datum")),
						h.Th(h.Class("text-nowrap"), g.Text("Zeit")),
						h.Th(h.Class("text-nowrap"), g.Text("Pos")),
						h.Th(h.Class("text-nowrap"), g.Text("Liga")),
						h.Th(h.Class("text-nowrap"), g.Text("Heim")),
						h.Th(h.Class("text-nowrap"), g.Text("Gast")),
						h.Th(h.Class("text-nowrap pe-3"), g.Text("Spielort")),
					)),
					h.TBody(
						h.Template(g.Attr("x-for", "(g, __idx) in "+xFor[5:]), g.Attr(":key", "__idx"),
							h.Tr(
								h.Td(h.Class("ps-3"), g.Attr("x-text", "fmtDate(g.game_date)")),
								h.Td(h.Class("text-muted"), g.Attr("x-text", "g.game_time")),
								h.Td(h.Span(h.Class("badge bg-secondary"), g.Attr("x-text", "g.position"))),
								h.Td(g.Attr("x-text", "g.league.name")),
								h.Td(h.Class("fw-semibold"), g.Attr("x-text", "g.home")),
								h.Td(h.Class("fw-semibold"), g.Attr("x-text", "g.away")),
								h.Td(h.Class("text-muted pe-3"), g.Attr("x-text", "g.venue_label")),
							),
						),
					),
				),
			),
		),
	)
}

// --- Main Dashboard Page ---

func DashboardPage(seasons []string, defaultSeason string) g.Node {
	// Season options for the select
	var seasonOpts []g.Node
	for _, s := range seasons {
		attrs := []g.Node{g.Attr("value", s)}
		if s == defaultSeason {
			attrs = append(attrs, g.Attr("selected", ""))
		}
		seasonOpts = append(seasonOpts, h.Option(append(attrs, g.Text(s))...))
	}

	// --- Sidebar ---
	sidebar := h.Div(h.Class("col-md-2"),
		h.Div(g.Attr("style", "position:sticky;top:4rem"),
			h.H1(h.Class("h5 mb-3"), g.Text("Dashboard")),
			// View toggle
			h.Div(h.Class("btn-group btn-group-sm w-100 mb-3"), g.Attr("role", "group"),
				h.Button(h.Class("btn"), g.Attr("type", "button"),
					g.Attr(":class", "view === 'year' ? 'btn-primary' : 'btn-outline-primary'"),
					g.Attr("@click", "view = 'year'"),
					g.Text("Jahr"),
				),
				h.Button(h.Class("btn"), g.Attr("type", "button"),
					g.Attr(":class", "view === 'overview' ? 'btn-primary' : 'btn-outline-primary'"),
					g.Attr("@click", "view = 'overview'"),
					g.Text("Übersicht"),
				),
			),
			// Year view filters
			h.Div(g.Attr("x-show", "view === 'year'"),
				// Season select
				h.Div(h.Class("mb-3"),
					h.Small(h.Class("text-muted d-block mb-1"), g.Text("Saison")),
					h.Select(h.Class("form-select form-select-sm"), g.Attr("x-model", "season"),
						g.Group(seasonOpts),
					),
				),
				// Position toggle buttons
				h.Div(h.Class("mb-3"),
					h.Small(h.Class("text-muted d-block mb-1"), g.Text("Position")),
					h.Div(h.Class("d-flex flex-wrap gap-1"),
						h.Template(g.Attr("x-for", "p in availablePositions"), g.Attr(":key", "p"),
							h.Button(h.Class("btn btn-sm"), g.Attr("type", "button"),
								g.Attr(":class", "filterPositions.includes(p) ? 'btn-primary' : 'btn-outline-secondary'"),
								g.Attr("@click", "togglePosition(filterPositions, p)"),
								g.Attr("x-text", "p"),
							),
						),
					),
				),
				// League filter
				alpineSelect("Liga", "filterLeague",
					h.Template(g.Attr("x-for", "lg in availableLeagues"), g.Attr(":key", "lg.id"),
						h.Option(g.Attr(":value", "'' + lg.id"), g.Attr("x-text", "lg.name")),
					),
				),
				// Games section
				sidebarHeading("Spiele"),
				h.Div(h.Class("d-flex align-items-center flex-wrap gap-2 mb-3"),
					h.Div(h.Class("form-check mb-0"),
						h.Input(h.Class("form-check-input"), g.Attr("type", "checkbox"),
							g.Attr("id", "only-completed"), g.Attr("x-model", "onlyCompleted")),
						h.Label(h.Class("form-check-label small"), g.Attr("for", "only-completed"),
							g.Text("Ohne offene")),
					),
					h.Div(h.Class("form-check mb-0"), g.Attr("x-show", "view === 'year'"),
						h.Input(h.Class("form-check-input"), g.Attr("type", "checkbox"),
							g.Attr("id", "show-recent"), g.Attr("x-model", "showRecentGames")),
						h.Label(h.Class("form-check-label small"), g.Attr("for", "show-recent"),
							g.Text("Absolviert")),
					),
					h.Select(h.Class("form-select form-select-sm"),
						g.Attr("style", "width:auto"),
						g.Attr("x-model", "recentGamesLimit"),
						g.Attr("x-show", "view === 'year' && showRecentGames"),
						h.Option(g.Attr("value", "0"), g.Text("Alle")),
						h.Option(g.Attr("value", "5"), g.Text("5")),
						h.Option(g.Attr("value", "10"), g.Text("10")),
						h.Option(g.Attr("value", "15"), g.Text("15")),
						h.Option(g.Attr("value", "20"), g.Text("20")),
					),
				),
				// Reset button
				h.Button(h.Class("btn btn-outline-secondary btn-sm w-100 mb-3"),
					g.Attr("type", "button"),
					g.Attr("@click", "filterLeague = ''; filterPositions = []; onlyCompleted = true; showRecentGames = false; recentGamesLimit = '10'"),
					g.Attr("x-show", "filterLeague || filterPositions.length || !onlyCompleted || showRecentGames"),
					h.I(h.Class("bi bi-x-circle me-1")),
					g.Text("Filter zurücksetzen"),
				),
				// Stats
				h.Div(h.Class("d-flex flex-column gap-2"),
					statRow("Spiele", "stats.count", "Gesamt", "eur(stats.total)"),
					statRow("Vergütung", "eur(stats.fee)", "Fahrtkosten", "eur(stats.travel)"),
					statRow("Kilometer", "stats.km.toLocaleString('de-DE')", "ct/km",
						"stats.km > 0 ? (stats.travel / stats.km * 100).toFixed(1).replace('.', ',') + ' ct' : '–'"),
				),
			),
			// Overview sidebar
			h.Div(g.Attr("x-show", "view === 'overview'"),
				// Position toggle
				h.Div(h.Class("mb-3"),
					h.Small(h.Class("text-muted d-block mb-1"), g.Text("Position")),
					h.Div(h.Class("d-flex flex-wrap gap-1"),
						h.Template(g.Attr("x-for", "p in overviewPositions"), g.Attr(":key", "p"),
							h.Button(h.Class("btn btn-sm"), g.Attr("type", "button"),
								g.Attr(":class", "ovFilterPositions.includes(p) ? 'btn-primary' : 'btn-outline-secondary'"),
								g.Attr("@click", "togglePosition(ovFilterPositions, p)"),
								g.Attr("x-text", "p"),
							),
						),
					),
				),
				// League filter
				alpineSelect("Liga", "ovFilterLeague",
					h.Template(g.Attr("x-for", "lg in overviewLeagues"), g.Attr(":key", "lg.id"),
						h.Option(g.Attr(":value", "'' + lg.id"), g.Attr("x-text", "lg.name")),
					),
				),
				// Year range
				h.Div(h.Class("mb-3"),
					h.Small(h.Class("text-muted d-block mb-1"), g.Text("Zeitraum")),
					h.Div(h.Class("d-flex gap-1 align-items-center"),
						h.Select(h.Class("form-select form-select-sm"), g.Attr("x-model", "ovYearFrom"),
							h.Template(g.Attr("x-for", "y in allOverviewYears"), g.Attr(":key", "'from'+y"),
								h.Option(g.Attr(":value", "y"), g.Attr("x-text", "y")),
							),
						),
						h.Small(h.Class("text-muted"), g.Text("–")),
						h.Select(h.Class("form-select form-select-sm"), g.Attr("x-model", "ovYearTo"),
							h.Template(g.Attr("x-for", "y in allOverviewYears"), g.Attr(":key", "'to'+y"),
								h.Option(g.Attr(":value", "y"), g.Attr("x-text", "y")),
							),
						),
					),
				),
				// Games section
				sidebarHeading("Spiele"),
				h.Div(h.Class("form-check mb-3"),
					h.Input(h.Class("form-check-input"), g.Attr("type", "checkbox"),
						g.Attr("id", "only-completed-overview"), g.Attr("x-model", "onlyCompleted")),
					h.Label(h.Class("form-check-label small"), g.Attr("for", "only-completed-overview"),
						g.Text("Ohne offene")),
				),
				// Reset button
				h.Button(h.Class("btn btn-outline-secondary btn-sm w-100 mb-3"),
					g.Attr("type", "button"),
					g.Attr("@click", "ovFilterLeague = ''; ovFilterPositions = []; ovYearTo = allOverviewYears[0] || ''; ovYearFrom = ovYearTo ? '' + (parseInt(ovYearTo) - 9) : ''; onlyCompleted = true"),
					g.Attr("x-show", "ovFilterLeague || ovFilterPositions.length || ovYearFrom || ovYearTo || !onlyCompleted"),
					h.I(h.Class("bi bi-x-circle me-1")),
					g.Text("Filter zurücksetzen"),
				),
				// Stats
				h.Div(h.Class("d-flex flex-column gap-2"),
					statRow("Spiele", "overviewStats.count", "Gesamt", "eur(overviewStats.total)"),
					statRow("Vergütung", "eur(overviewStats.fee)", "Fahrtkosten", "eur(overviewStats.travel)"),
					statRow("Kilometer", "overviewStats.km.toLocaleString('de-DE')", "ct/km",
						"overviewStats.km > 0 ? (overviewStats.travel / overviewStats.km * 100).toFixed(1).replace('.', ',') + ' ct' : '–'"),
				),
			),
		),
	)

	// --- Year view charts ---
	yearCharts := h.Div(h.Class("col-md-10"), g.Attr("x-show", "view === 'year'"),
		gameTableCard("Anstehende Spiele", "bi-calendar-event",
			"view === 'year' && upcomingGames.length > 0", "g in upcomingGames", ""),
		gameTableCard("Absolvierte Spiele", "bi-calendar-check",
			"view === 'year' && showRecentGames && recentGames.length > 0", "g in recentGames", "recentGames.length"),
		h.Div(g.Attr("x-show", "filtered.length > 0"),
			collapsibleSection("charts-body", "bi bi-bar-chart-fill", "Einsätze",
				h.Div(h.Class("row g-3"),
					h.Div(h.Class("col-md-3"), widgetCard("nach Position", "chart-positions")),
					h.Div(h.Class("col-md-5"), widgetCard("nach Monat", "chart-monthly")),
					h.Div(h.Class("col-md-4"), widgetCard("nach Liga", "chart-leagues")),
				),
				h.Div(h.Class("row g-3 mt-2"),
					h.Div(h.Class("col-12"), widgetCard("Saisonkalender", "chart-calendar")),
				),
			),
			collapsibleSection("fees-body", "bi bi-currency-euro", "Vergütung",
				h.Div(h.Class("row g-3"),
					h.Div(h.Class("col-md-4"), widgetCard("Gesamt", "chart-fee-total")),
					h.Div(h.Class("col-md-4"), widgetCard("Pauschale", "chart-fee-base")),
					h.Div(h.Class("col-md-4"), widgetCard("Fahrtkosten", "chart-fee-travel")),
				),
			),
		),
		h.Div(g.Attr("x-show", "filtered.some(g => g.venue && g.venue.lat && g.venue.lon && (g.venue.lat !== 0 || g.venue.lon !== 0))"),
			collapsibleSection("map-body", "bi bi-geo-alt-fill", "Spielorte",
				h.Div(h.Class("row g-3"),
					h.Div(h.Class("col-md-4"), widgetCard("Top 10 Spielorte", "chart-venues-top")),
					h.Div(h.Class("col-md-8"), widgetCard("Einsatzorte", "chart-map")),
				),
			),
		),
	)

	// --- Overview charts ---
	overviewCharts := h.Div(h.Class("col-md-10"), g.Attr("x-show", "view === 'overview'"),
		collapsibleSection("overview-games-body", "bi bi-bar-chart-fill", "Einsätze",
			h.Div(h.Class("row g-3 mb-3"),
				h.Div(h.Class("col-md-6"), widgetCard("Spiele pro Jahr", "chart-overview-games")),
				h.Div(h.Class("col-md-6"), widgetCard("Positionstrend", "chart-overview-trend")),
			),
			h.Div(h.Class("row g-3"),
				h.Div(h.Class("col-md-4"), widgetCard("Positionsverteilung", "chart-overview-pie")),
				h.Div(h.Class("col-md-8"), widgetCard("Spiele pro Liga", "chart-overview-leagues")),
			),
		),
		collapsibleSection("overview-fees-body", "bi bi-currency-euro", "Vergütung",
			h.Div(h.Class("row g-3"),
				h.Div(h.Class("col-md-4"), widgetCard("Vergütung pro Jahr", "chart-overview-fee")),
				h.Div(h.Class("col-md-4"), widgetCard("Durchschnitt pro Spiel", "chart-overview-avg")),
				h.Div(h.Class("col-md-4"), widgetCard("Kilometer pro Jahr", "chart-overview-km")),
			),
		),
		collapsibleSection("overview-flow-body", "bi bi-diagram-3", "Verteilung",
			h.Div(h.Class("row g-3 mb-3"),
				h.Div(h.Class("col-12"), widgetCard("Positionsfluss", "chart-overview-river")),
			),
			h.Div(h.Class("row g-3"),
				h.Div(h.Class("col-12"), widgetCard("Positionen pro Jahr", "chart-overview-sankey")),
			),
		),
		h.Div(g.Attr("x-show", "overviewFiltered.some(g => g.venue && g.venue.lat && g.venue.lon && (g.venue.lat !== 0 || g.venue.lon !== 0))"),
			collapsibleSection("overview-map-body", "bi bi-geo-alt-fill", "Spielorte",
				h.Div(h.Class("row g-3"),
					h.Div(h.Class("col-md-4"), widgetCard("Top 10 Spielorte", "chart-overview-venues-top")),
					h.Div(h.Class("col-md-8"), widgetCard("Einsatzorte", "chart-overview-map")),
				),
			),
		),
	)

	return BasePageRaw("Dashboard",
		h.Div(g.Attr("id", "dashboard-app"), h.Class("container-dashboard"),
			g.Attr("x-data", "dashboard"),
			g.Attr("data-default-season", defaultSeason),
			// Loading spinner
			h.Div(h.Class("text-center py-5"), g.Attr("x-show", "loading"),
				h.Div(h.Class("spinner-border text-secondary"), g.Attr("role", "status"),
					h.Span(h.Class("visually-hidden"), g.Text("Laden...")),
				),
			),
			// Content
			h.Div(g.Attr("x-show", "!loading"), g.Attr("x-cloak", ""),
				h.Div(h.Class("row g-4"),
					sidebar,
					yearCharts,
					overviewCharts,
				),
			),
		),
	)
}

