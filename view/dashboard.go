package view

import (
	"fmt"

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
						h.Template(g.Attr("x-for", xFor), g.Attr(":key", "g.date+g.home+g.away"),
							h.Tr(
								h.Td(h.Class("ps-3"), g.Attr("x-text", "fmtDate(g.date)")),
								h.Td(h.Class("text-muted"), g.Attr("x-text", "g.time")),
								h.Td(h.Span(h.Class("badge bg-secondary"), g.Attr("x-text", "g.position"))),
								h.Td(g.Attr("x-text", "g.league_long || g.league")),
								h.Td(h.Class("fw-semibold"), g.Attr("x-text", "g.home")),
								h.Td(h.Class("fw-semibold"), g.Attr("x-text", "g.away")),
								h.Td(h.Class("text-muted pe-3"), g.Attr("x-text", "g.venue")),
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
		collapsibleSection("charts-body", "bi bi-bar-chart-fill", "Einsätze",
			h.Div(h.Class("row g-3"),
				h.Div(h.Class("col-md-3"), widgetCard("nach Position", "chart-positions")),
				h.Div(h.Class("col-md-5"), widgetCard("nach Monat", "chart-monthly")),
				h.Div(h.Class("col-md-4"), widgetCard("nach Liga", "chart-leagues")),
			),
		),
		collapsibleSection("fees-body", "bi bi-currency-euro", "Vergütung",
			h.Div(h.Class("row g-3"),
				h.Div(h.Class("col-md-4"), widgetCard("Gesamt", "chart-fee-total")),
				h.Div(h.Class("col-md-4"), widgetCard("Pauschale", "chart-fee-base")),
				h.Div(h.Class("col-md-4"), widgetCard("Fahrtkosten", "chart-fee-travel")),
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
			h.Div(h.Class("row g-3"),
				h.Div(h.Class("col-12"), widgetCard("Positionen pro Jahr", "chart-overview-sankey")),
			),
		),
	)

	// --- Alpine.js script ---
	alpineScript := h.Script(g.Raw(fmt.Sprintf(dashboardJS, defaultSeason)))

	return BasePageRaw("Dashboard",
		h.Div(g.Attr("id", "dashboard-app"), h.Class("container-dashboard"),
			g.Attr("x-data", "dashboard"),
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
		alpineScript,
	)
}

// dashboardJS is the Alpine.js data component — 1:1 from Python.
// %s is replaced with the default season.
const dashboardJS = `
document.addEventListener('alpine:init', () => {
    Alpine.data('dashboard', () => ({
        season: localStorage.getItem('db_season') || '%s',
        games: [],
        positions: [],
        availableLeagues: [],
        availablePositions: [],
        filterLeague: localStorage.getItem('db_filterLeague') || '',
        filterPositions: JSON.parse(localStorage.getItem('db_filterPositions') || '[]'),
        loading: true,
        view: localStorage.getItem('db_view') || 'year',
        onlyCompleted: localStorage.getItem('db_onlyCompleted') === 'true',
        showRecentGames: localStorage.getItem('db_showRecentGames') === 'true',
        recentGamesLimit: localStorage.getItem('db_recentGamesLimit') || '10',
        today: new Date().toISOString().slice(0, 10),

        togglePosition(arr, p) {
            var i = arr.indexOf(p);
            if (i >= 0) arr.splice(i, 1); else arr.push(p);
        },

        get filtered() {
            return this.games.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.filterLeague && g.league_id != this.filterLeague) return false;
                if (this.filterPositions.length
                    && !this.filterPositions.includes(g.position)) return false;
                return true;
            });
        },

        get stats() {
            var f = this.filtered;
            var fee = f.reduce((s, g) => s + g.fee, 0);
            var travel = f.reduce((s, g) => s + g.travel, 0);
            return {
                count: f.length,
                fee: fee,
                travel: travel,
                total: fee + travel,
                km: f.reduce((s, g) => s + g.km, 0),
            };
        },

        get upcomingGames() {
            return this.games
                .filter(g => g.date >= this.today)
                .sort((a, b) => a.date.localeCompare(b.date)
                    || (a.time || '').localeCompare(b.time || ''));
        },

        get recentGames() {
            var all = this.games
                .filter(g => {
                    if (g.date >= this.today) return false;
                    if (this.filterLeague && g.league_id != this.filterLeague) return false;
                    if (this.filterPositions.length
                        && !this.filterPositions.includes(g.position)) return false;
                    return true;
                })
                .sort((a, b) => b.date.localeCompare(a.date)
                    || (b.time || '').localeCompare(a.time || ''));
            var limit = parseInt(this.recentGamesLimit);
            return limit > 0 ? all.slice(0, limit) : all;
        },

        overviewGames: [],
        overviewPositions: [],
        overviewLeagues: [],
        overviewLoaded: false,
        ovFilterLeague: '',
        ovFilterPositions: [],
        ovYearFrom: '',
        ovYearTo: '',

        get allOverviewYears() {
            var seen = new Set();
            this.overviewGames.forEach(g => seen.add(g.year));
            return [...seen].sort().reverse();
        },

        get overviewFiltered() {
            return this.overviewGames.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.ovFilterLeague && g.league_id != this.ovFilterLeague) return false;
                if (this.ovFilterPositions.length
                    && !this.ovFilterPositions.includes(g.position)) return false;
                if (this.ovYearFrom && g.year < this.ovYearFrom) return false;
                if (this.ovYearTo && g.year > this.ovYearTo) return false;
                return true;
            });
        },

        get overviewForPositions() {
            return this.overviewGames.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.ovFilterLeague && g.league_id != this.ovFilterLeague) return false;
                if (this.ovFilterPositions.length > 1
                    && !this.ovFilterPositions.includes(g.position)) return false;
                if (this.ovYearFrom && g.year < this.ovYearFrom) return false;
                if (this.ovYearTo && g.year > this.ovYearTo) return false;
                return true;
            });
        },

        get overviewYears() {
            var seen = new Set();
            this.overviewFiltered.forEach(g => seen.add(g.year));
            return [...seen].sort();
        },

        overviewAggByYear(games) {
            var byYear = {};
            games.forEach(g => {
                if (!byYear[g.year]) byYear[g.year] = {
                    count: 0, fee: 0, travel: 0, km: 0,
                    by_position: {},
                };
                var y = byYear[g.year];
                y.count++;
                y.fee += g.fee;
                y.travel += g.travel;
                y.km += g.km;
                y.by_position[g.position] =
                    (y.by_position[g.position] || 0) + 1;
            });
            return byYear;
        },

        get overviewByYear() {
            return this.overviewAggByYear(this.overviewFiltered);
        },

        get overviewByYearForPositions() {
            return this.overviewAggByYear(this.overviewForPositions);
        },

        get overviewStats() {
            var f = this.overviewFiltered;
            var fee = f.reduce((s, g) => s + g.fee, 0);
            var travel = f.reduce((s, g) => s + g.travel, 0);
            return {
                count: f.length,
                fee: fee,
                travel: travel,
                total: fee + travel,
                km: f.reduce((s, g) => s + g.km, 0),
            };
        },

        init() {
            this.loadSeason();
            if (this.view === 'overview') {
                this.loadOverview();
            }
            this.$watch('season', (v) => {
                localStorage.setItem('db_season', v);
                this.loadSeason();
            });
            this.$watch('filtered', () => this.renderCharts());
            this.$watch('filterLeague', (v) => {
                localStorage.setItem('db_filterLeague', v);
                this.renderPositionChart();
            });
            this.$watch('filterPositions', (v) => {
                localStorage.setItem('db_filterPositions', JSON.stringify(v));
            });
            this.$watch('onlyCompleted', (v) => {
                localStorage.setItem('db_onlyCompleted', v);
                this.renderCharts();
                if (this.view === 'overview') {
                    this.$nextTick(() => this.renderOverviewCharts());
                }
            });
            this.$watch('showRecentGames', (v) => {
                localStorage.setItem('db_showRecentGames', v);
            });
            this.$watch('recentGamesLimit', (v) => {
                localStorage.setItem('db_recentGamesLimit', v);
            });
            this.$watch('ovFilterLeague', (v) => {
                localStorage.setItem('db_ovFilterLeague', v);
                if (this.view === 'overview') {
                    this.$nextTick(() => this.renderOverviewCharts());
                }
            });
            this.$watch('ovFilterPositions', (v) => {
                localStorage.setItem('db_ovFilterPositions', JSON.stringify(v));
                if (this.view === 'overview') {
                    this.$nextTick(() => this.renderOverviewCharts());
                }
            });
            this.$watch('ovYearFrom', (v) => {
                localStorage.setItem('db_ovYearFrom', v);
                if (this.view === 'overview') {
                    this.$nextTick(() => this.renderOverviewCharts());
                }
            });
            this.$watch('ovYearTo', (v) => {
                localStorage.setItem('db_ovYearTo', v);
                if (this.view === 'overview') {
                    this.$nextTick(() => this.renderOverviewCharts());
                }
            });
            this.$watch('view', (v) => {
                localStorage.setItem('db_view', v);
                if (v === 'overview') {
                    this.loadOverview();
                } else {
                    this.$nextTick(() => this.renderCharts());
                }
            });
            window.addEventListener('theme-changed', () => {
                this.$nextTick(() => {
                    this.renderCharts();
                    if (this.view === 'overview') this.renderOverviewCharts();
                });
            });
        },

        _seasonLoaded: false,
        async loadSeason() {
            this.loading = true;
            if (this._seasonLoaded) {
                this.filterLeague = '';
                this.filterPositions = [];
            }
            this._seasonLoaded = true;
            var res = await fetch('/api/dashboard/season/' + this.season);
            var data = await res.json();
            this.games = data.games;
            this.positions = data.positions;
            this.availableLeagues = data.available_leagues;
            this.availablePositions = data.available_positions;
            this.loading = false;
            this.$nextTick(() => this.renderCharts());
        },

        renderCharts() {
            this.renderPositionChart();
            this.renderMonthlyChart();
            this.renderLeagueChart();
            this.renderFeeBar('chart-fee-total', g => g.fee + g.travel);
            this.renderFeeBar('chart-fee-base', g => g.fee);
            this.renderFeeBar('chart-fee-travel', g => g.travel);
        },

        fmtDate(d) {
            if (!d || d.length < 10) return d;
            return d.slice(8,10) + '.' + d.slice(5,7) + '.' + d.slice(0,4);
        },

        eur(v) {
            return v.toFixed(2).replace('.', ',').replace(/\B(?=(\d{3})+(?!\d))/g, '.') + ' €';
        },

        baseLayout(height, extra) {
            var dark = document.documentElement.getAttribute('data-bs-theme') === 'dark';
            var fg = getComputedStyle(document.body).color;
            var gc = dark ? '#3B4252' : '#D8DEE9';
            var layout = {
                height: height,
                margin: {t: 10, b: 30, l: 40, r: 10},
                paper_bgcolor: 'rgba(0,0,0,0)',
                plot_bgcolor: 'rgba(0,0,0,0)',
                font: {color: fg, size: 12},
                xaxis: {gridcolor: gc},
                yaxis: {gridcolor: gc},
            };
            return Object.assign(layout, extra || {});
        },

        stackAnnotations(categories, totals, horizontal, fmt, angle) {
            var fg = getComputedStyle(document.body).color;
            return categories.map((cat, i) => {
                var val = totals[i] || 0;
                if (val === 0) return null;
                var label = fmt ? fmt(val) : String(val);
                var a = horizontal ? {
                    x: val, y: cat, text: label,
                    xanchor: 'left', yanchor: 'middle',
                    xshift: 5, showarrow: false,
                    font: {size: 11, color: fg},
                } : {
                    x: cat, y: val, text: label,
                    xanchor: 'center', yanchor: 'bottom',
                    yshift: 3, showarrow: false,
                    font: {size: 11, color: fg},
                };
                if (angle) a.textangle = angle;
                return a;
            }).filter(a => a);
        },

        renderPositionChart() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var games = this.games.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.filterLeague && g.league_id != this.filterLeague) return false;
                if (this.filterPositions.length > 1
                    && !this.filterPositions.includes(g.position)) return false;
                return true;
            });
            var counts = {};
            games.forEach(g => { counts[g.position] = (counts[g.position] || 0) + 1; });
            var entries = Object.entries(counts).sort((a, b) => a[1] - b[1]);
            var labels = entries.map(e => e[0]);
            var values = entries.map(e => e[1]);
            Plotly.newPlot('chart-positions', [{
                x: values, y: labels, type: 'bar', orientation: 'h',
                marker: {color: labels.map((_, i) => colors[i %% colors.length])},
                text: values, textposition: 'auto',
            }], this.baseLayout(350, {
                yaxis: {
                    gridcolor: this.baseLayout(0).yaxis.gridcolor,
                    categoryorder: 'array', categoryarray: labels,
                    ticksuffix: '  ',
                },
            }), {responsive: true, displayModeBar: false});
        },

        renderMonthlyChart() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var allMonths = ['Jan','Feb','Mär','Apr','Mai','Jun','Jul','Aug','Sep','Okt','Nov','Dez'];
            var f = this.filtered;
            var raw = {};
            var activeMonths = new Set();
            f.forEach(g => {
                var m = parseInt(g.month) - 1;
                raw[g.position] = raw[g.position] || {};
                raw[g.position][m] = (raw[g.position][m] || 0) + 1;
                activeMonths.add(m);
            });
            var months = [...activeMonths].sort((a, b) => a - b);
            var monthLabels = months.map(m => allMonths[m]);
            var traces = this.positions
                .filter(p => raw[p])
                .map((p, idx) => ({
                    x: monthLabels,
                    y: months.map(m => raw[p][m] || 0),
                    type: 'bar', name: p,
                    marker: {color: colors[idx %% colors.length]},
                }));
            var totals = months.map(m => {
                var t = 0;
                this.positions.forEach(p => {
                    t += (raw[p] && raw[p][m]) || 0;
                });
                return t;
            });
            Plotly.newPlot('chart-monthly', traces, this.baseLayout(350, {
                barmode: 'stack',
                showlegend: true,
                legend: {font: {size: 10}, traceorder: 'normal'},
                margin: {t: 20, b: 30, l: 30, r: 10},
                annotations: this.stackAnnotations(monthLabels, totals, false),
            }), {responsive: true, displayModeBar: false});
        },

        renderLeagueChart() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var games = this.games.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.filterPositions.length
                    && !this.filterPositions.includes(g.position)) return false;
                return true;
            });
            var leagueData = {};
            var posPresent = new Set();
            games.forEach(g => {
                leagueData[g.league] = leagueData[g.league] || {};
                leagueData[g.league][g.position] =
                    (leagueData[g.league][g.position] || 0) + 1;
                posPresent.add(g.position);
            });
            var leagues = Object.entries(leagueData)
                .map(([lg, pos]) => [lg, Object.values(pos).reduce((a, b) => a + b, 0)])
                .sort((a, b) => a[1] - b[1])
                .map(e => e[0]);
            var posOrder = this.positions.filter(p => posPresent.has(p));
            var traces = posOrder.map((p, i) => ({
                x: leagues.map(lg =>
                    (leagueData[lg] && leagueData[lg][p]) || 0
                ),
                y: leagues,
                type: 'bar', orientation: 'h', name: p,
                marker: {color: colors[i %% colors.length]},
            }));
            var leagueTotals = leagues.map(lg => {
                return Object.values(leagueData[lg])
                    .reduce((a, b) => a + b, 0);
            });
            Plotly.newPlot('chart-leagues', traces, this.baseLayout(350, {
                barmode: 'stack',
                showlegend: true,
                legend: {font: {size: 10}, traceorder: 'normal'},
                yaxis: {
                    gridcolor: this.baseLayout(0).yaxis.gridcolor,
                    categoryorder: 'array', categoryarray: leagues,
                    ticksuffix: '  ',
                },
                margin: {t: 10, b: 30, l: 100, r: 30},
                annotations: this.stackAnnotations(leagues, leagueTotals, true),
            }), {responsive: true, displayModeBar: false});
        },

        async loadOverview() {
            if (this.overviewLoaded) {
                this.$nextTick(() => this.renderOverviewCharts());
                return;
            }
            var res = await fetch('/api/dashboard/overview');
            var data = await res.json();
            this.overviewGames = data.games;
            this.overviewPositions = data.positions;
            this.overviewLeagues = data.available_leagues;
            this.ovFilterLeague = localStorage.getItem('db_ovFilterLeague') || '';
            this.ovFilterPositions = JSON.parse(
                localStorage.getItem('db_ovFilterPositions') || '[]'
            );
            var allYrs = this.allOverviewYears;
            var maxY = allYrs[0] || '';
            var minDefault = maxY ? '' + (parseInt(maxY) - 9) : '';
            this.ovYearFrom = localStorage.getItem('db_ovYearFrom') || minDefault;
            this.ovYearTo = localStorage.getItem('db_ovYearTo') || maxY;
            this.overviewLoaded = true;
            this.$nextTick(() => this.renderOverviewCharts());
        },

        yearLabels(years) {
            return years.map(y => '\u200b' + y);
        },

        yearTickVals(xl, years) {
            if (years.length <= 10) return xl;
            return xl.filter((_, i) => parseInt(years[i]) %% 5 === 0);
        },

        renderOverviewCharts() {
            if (!this.overviewLoaded) return;
            this.renderOverviewGamesPerYear();
            this.renderOverviewPositionTrend();
            this.renderOverviewPositionPie();
            this.renderOverviewFeePerYear();
            this.renderOverviewAvgPerGame();
            this.renderOverviewKmPerYear();
            this.renderOverviewSankey();
            this.renderOverviewLeagues();
        },

        renderOverviewGamesPerYear() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var byYear = this.overviewByYear;
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var traces = this.overviewPositions
                .filter(p => years.some(y =>
                    (byYear[y].by_position[p] || 0) > 0
                ))
                .map((p, i) => ({
                    x: xl,
                    y: years.map(y =>
                        byYear[y].by_position[p] || 0
                    ),
                    type: 'bar', name: p,
                    marker: {color: colors[i %% colors.length]},
                }));
            var totals = years.map(y => byYear[y].count);
            Plotly.newPlot('chart-overview-games', traces,
                this.baseLayout(350, {
                    barmode: 'stack',
                    showlegend: true,
                    legend: {orientation: 'h', font: {size: 10},
                        y: -0.15, x: 0.5, xanchor: 'center',
                        traceorder: 'normal'},
                    margin: {t: 20, b: 60, l: 30, r: 10},
                    xaxis: {tickvals: this.yearTickVals(xl, years)},
                    annotations: this.stackAnnotations(
                        xl, totals, false
                    ),
                }),
                {responsive: true, displayModeBar: false});
        },

        renderOverviewPositionTrend() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var byYear = this.overviewByYear;
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var traces = this.overviewPositions
                .filter(p => years.some(y =>
                    (byYear[y].by_position[p] || 0) > 0
                ))
                .map((p, i) => ({
                    x: xl,
                    y: years.map(y =>
                        byYear[y].by_position[p] || 0
                    ),
                    type: 'scatter', mode: 'lines+markers',
                    name: p,
                    line: {
                        color: colors[i %% colors.length], width: 2,
                    },
                    marker: {size: 5},
                }));
            Plotly.newPlot('chart-overview-trend', traces,
                this.baseLayout(350, {
                    showlegend: true,
                    legend: {orientation: 'h', font: {size: 10},
                        y: -0.15, x: 0.5, xanchor: 'center',
                        traceorder: 'normal'},
                    margin: {t: 20, b: 60, l: 30, r: 10},
                    xaxis: {tickvals: this.yearTickVals(xl, years)},
                }),
                {responsive: true, displayModeBar: false});
        },

        renderOverviewPositionPie() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var byYear = this.overviewByYearForPositions;
            var totals = {};
            Object.values(byYear).forEach(y => {
                Object.entries(y.by_position).forEach(([p, c]) => {
                    totals[p] = (totals[p] || 0) + c;
                });
            });
            var sorted = Object.entries(totals)
                .sort((a, b) => b[1] - a[1]);
            Plotly.newPlot('chart-overview-pie', [{
                labels: sorted.map(e => e[0]),
                values: sorted.map(e => e[1]),
                type: 'pie', hole: 0.4,
                marker: {colors: colors},
                textinfo: 'label+percent',
                textposition: 'outside',
                textfont: {size: 11},
                domain: {x: [0.1, 0.9], y: [0.1, 0.9]},
            }], this.baseLayout(350, {
                showlegend: false,
                margin: {t: 30, b: 30, l: 30, r: 30},
            }), {responsive: true, displayModeBar: false});
        },

        renderOverviewFeePerYear() {
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var byYear = this.overviewByYear;
            var feeTrace = {
                x: xl,
                y: years.map(y =>
                    byYear[y] ? byYear[y].fee : 0
                ),
                type: 'bar', name: 'Pauschale',
                marker: {color: '#5E81AC'},
            };
            var travelTrace = {
                x: xl,
                y: years.map(y =>
                    byYear[y] ? byYear[y].travel : 0
                ),
                type: 'bar', name: 'Fahrtkosten',
                marker: {color: '#88C0D0'},
            };
            var totals = years.map(y => {
                var d = byYear[y];
                return d ? d.fee + d.travel : 0;
            });
            var fmtEur = v => Math.round(v) + ' €';
            var traces = [feeTrace, travelTrace];
            Plotly.newPlot('chart-overview-fee', traces,
                this.baseLayout(350, {
                    barmode: 'stack',
                    showlegend: true,
                    legend: {orientation: 'h', font: {size: 10},
                        y: -0.15, x: 0.5, xanchor: 'center',
                        traceorder: 'normal'},
                    margin: {t: 20, b: 60, l: 50, r: 10},
                    xaxis: {tickvals: this.yearTickVals(xl, years)},
                    annotations: this.stackAnnotations(
                        xl, totals, false, fmtEur, -45
                    ),
                }),
                {responsive: true, displayModeBar: false});
        },

        renderOverviewAvgPerGame() {
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var byYear = this.overviewByYear;
            var avgs = years.map(y => {
                var d = byYear[y];
                if (!d || d.count === 0) return 0;
                return (d.fee + d.travel) / d.count;
            });
            var fmtEur = v => v.toFixed(0) + ' €';
            Plotly.newPlot('chart-overview-avg', [{
                x: xl, y: avgs, type: 'bar',
                marker: {color: '#A3BE8C'},
                text: avgs.map(fmtEur), textposition: 'auto',
            }], this.baseLayout(350, {
                margin: {t: 20, b: 30, l: 50, r: 10},
                xaxis: {tickvals: this.yearTickVals(xl, years)},
            }), {responsive: true, displayModeBar: false});
        },

        renderOverviewKmPerYear() {
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var byYear = this.overviewByYear;
            var kms = years.map(y =>
                byYear[y] ? byYear[y].km : 0
            );
            Plotly.newPlot('chart-overview-km', [{
                x: xl, y: kms, type: 'bar',
                marker: {color: '#81A1C1'},
                text: kms.map(v => v.toLocaleString('de-DE')),
                textposition: 'auto',
            }], this.baseLayout(350, {
                margin: {t: 20, b: 30, l: 50, r: 10},
                xaxis: {tickvals: this.yearTickVals(xl, years)},
            }), {responsive: true, displayModeBar: false});
        },

        renderOverviewSankey() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var games = this.overviewFiltered;
            if (!games.length) return;
            var years = this.overviewYears;
            var xl = this.yearLabels(years);
            var positions = this.overviewPositions.filter(
                p => games.some(g => g.position === p)
            );
            var counts = {};
            games.forEach(g => {
                var key = g.position + '|' + g.year;
                counts[key] = (counts[key] || 0) + 1;
            });
            var z = positions.map(p =>
                years.map(y => counts[p + '|' + y] || 0)
            );
            var annotations = [];
            positions.forEach((p, pi) => {
                years.forEach((y, yi) => {
                    var val = counts[p + '|' + y] || 0;
                    if (val > 0) {
                        annotations.push({
                            x: xl[yi], y: p,
                            text: '' + val,
                            showarrow: false,
                            font: {color: val > 8 ? '#2E3440' : '#ECEFF4', size: 11},
                        });
                    }
                });
            });
            Plotly.newPlot('chart-overview-sankey', [{
                x: xl, y: positions, z: z,
                type: 'heatmap',
                colorscale: [
                    [0, '#3B4252'],
                    [0.3, '#5E81AC'],
                    [0.6, '#88C0D0'],
                    [0.8, '#D08770'],
                    [1, '#BF616A'],
                ],
                showscale: false,
                hoverongaps: false,
                xgap: 2,
                ygap: 2,
            }], this.baseLayout(Math.max(300, positions.length * 45), {
                margin: {t: 10, b: 30, l: 40, r: 10},
                xaxis: {tickvals: xl, side: 'bottom'},
                yaxis: {autorange: 'reversed'},
                annotations: annotations,
            }), {responsive: true, displayModeBar: false});
        },

        renderOverviewLeagues() {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var games = this.overviewGames.filter(g => {
                if (this.onlyCompleted && g.date >= this.today) return false;
                if (this.ovFilterPositions.length
                    && !this.ovFilterPositions.includes(g.position)) return false;
                if (this.ovYearFrom && g.year < this.ovYearFrom) return false;
                if (this.ovYearTo && g.year > this.ovYearTo) return false;
                return true;
            });
            if (!games.length) return;
            var seenYears = new Set();
            games.forEach(g => seenYears.add(g.year));
            var years = [...seenYears].sort();
            var xl = this.yearLabels(years);
            var leagueTotals = {};
            games.forEach(g => {
                leagueTotals[g.league] = (leagueTotals[g.league] || 0) + 1;
            });
            var leagues = Object.keys(leagueTotals)
                .sort((a, b) => leagueTotals[b] - leagueTotals[a]);
            var counts = {};
            games.forEach(g => {
                var key = g.league + '|' + g.year;
                counts[key] = (counts[key] || 0) + 1;
            });
            var traces = years.map((y, i) => ({
                x: leagues,
                y: leagues.map(lg => counts[lg + '|' + y] || 0),
                type: 'bar',
                name: y,
                marker: {color: colors[i %% colors.length]},
            }));
            Plotly.newPlot('chart-overview-leagues', traces,
                this.baseLayout(350, {
                barmode: 'stack',
                showlegend: true,
                legend: {font: {size: 10}, traceorder: 'normal'},
                margin: {t: 10, b: 80, l: 40, r: 10},
                xaxis: {tickangle: -45},
            }), {responsive: true, displayModeBar: false});
        },

        renderFeeBar(chartId, valueFn) {
            var colors = ['#5E81AC','#81A1C1','#88C0D0','#8FBCBB','#A3BE8C','#EBCB8B','#D08770','#BF616A','#B48EAD'];
            var months = ['Jan','Feb','Mär','Apr','Mai','Jun','Jul','Aug','Sep','Okt','Nov','Dez'];
            var f = this.filtered;
            var data = {};
            var monthsPresent = new Set();
            f.forEach(g => {
                var m = parseInt(g.month) - 1;
                data[g.position] = data[g.position] || {};
                data[g.position][m] = (data[g.position][m] || 0) + valueFn(g);
                monthsPresent.add(m);
            });
            var sortedMonths = [...monthsPresent].sort((a, b) => a - b);
            var monthLabels = sortedMonths.map(m => months[m] + ' ' + this.season);
            var traces = this.positions
                .filter(p => data[p])
                .map((p, i) => ({
                    y: monthLabels,
                    x: sortedMonths.map(m => data[p][m] || 0),
                    type: 'bar', orientation: 'h', name: p,
                    marker: {color: colors[i %% colors.length]},
                }));
            var totalArr = sortedMonths.map(m => {
                var t = 0;
                this.positions.forEach(p => {
                    t += (data[p] && data[p][m]) || 0;
                });
                return t;
            });
            var fmtEur = v => Math.round(v) + ' €';
            Plotly.newPlot(chartId, traces, this.baseLayout(350, {
                barmode: 'stack',
                showlegend: true,
                legend: {font: {size: 10}, traceorder: 'normal'},
                yaxis: {
                    gridcolor: this.baseLayout(0).yaxis.gridcolor,
                    categoryorder: 'array',
                    categoryarray: monthLabels,
                    ticksuffix: '  ',
                },
                margin: {t: 10, b: 30, l: 90, r: 60},
                annotations: this.stackAnnotations(
                    monthLabels, totalArr, true, fmtEur
                ),
            }), {responsive: true, displayModeBar: false});
        },
    }));
});
`
