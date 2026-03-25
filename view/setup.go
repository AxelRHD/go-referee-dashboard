package view

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func SetupPage(messages []string, setupDone bool) g.Node {
	var alerts []g.Node
	for _, msg := range messages {
		alerts = append(alerts, h.Div(h.Class("alert alert-success alert-dismissible fade show"),
			g.Text(msg),
			h.Button(h.Class("btn-close"), g.Attr("type", "button"), g.Attr("data-bs-dismiss", "alert")),
		))
	}

	if setupDone {
		return BasePage("Setup abgeschlossen",
			g.Group(alerts),
			h.Div(h.Class("row justify-content-center"),
				h.Div(h.Class("col-md-8 text-center py-5"),
					h.I(h.Class("bi bi-check-circle-fill text-success"), g.Attr("style", "font-size:4rem")),
					h.H2(h.Class("mt-3 mb-3"), g.Text("Setup abgeschlossen")),
					h.P(h.Class("text-muted mb-4"), g.Text("Die Datenbank wurde erfolgreich eingerichtet.")),
					h.A(h.Class("btn btn-primary btn-lg"), g.Attr("href", "/dashboard/"),
						h.I(h.Class("bi bi-arrow-right me-2")),
						g.Text("Weiter zum Dashboard"),
					),
				),
			),
		)
	}

	return BasePage("Setup",
		g.Group(alerts),
		h.Div(h.Class("row justify-content-center"),
			h.Div(h.Class("col-md-8"),
				h.H1(h.Class("mb-4"), g.Text("Datenbank einrichten")),
				h.P(h.Class("text-muted mb-4"),
					g.Text("Die Datenbank ist leer. Wählen Sie eine der folgenden Optionen:"),
				),

				// Option 1: Seeding
				h.Div(h.Class("card mb-4"),
					h.Div(h.Class("card-body"),
						h.H5(h.Class("card-title"),
							h.I(h.Class("bi bi-database-add me-2")),
							g.Text("Neues Setup"),
						),
						h.P(h.Class("text-muted"),
							g.Text("Stammdaten für einen Neustart importieren. Positionen, Ligen, Teams und Spielorte werden angelegt."),
						),
						h.FormEl(g.Attr("method", "post"), g.Attr("action", "/setup/seed"),
							h.Div(h.Class("d-flex flex-column gap-2 mb-3"),
								seedCheckbox("positions", "Positionen (R, CJ, U, LJ, DJ, LM, BJ, FJ, SJ)", true),
								seedCheckbox("leagues", "Ligen", false),
								seedCheckbox("teams", "Teams", false),
								seedCheckbox("venues", "Spielorte", false),
							),
							h.Button(h.Class("btn btn-primary"),
								g.Attr("type", "submit"),
								h.I(h.Class("bi bi-download me-1")),
								g.Text("Ausgewählte Daten importieren"),
							),
						),
					),
				),

				// Option 2: Restore
				h.Div(h.Class("card mb-4"),
					h.Div(h.Class("card-body"),
						h.H5(h.Class("card-title"),
							h.I(h.Class("bi bi-arrow-counterclockwise me-2")),
							g.Text("Wiederherstellung"),
						),
						h.P(h.Class("text-muted"),
							g.Text("Kompletten SQL-Dump einer bestehenden Datenbank importieren."),
						),
						// Warning
						h.Div(h.Class("alert alert-danger"),
							h.I(h.Class("bi bi-exclamation-triangle-fill me-2")),
							h.Strong(g.Text("Wichtige Hinweise:")),
							h.Ul(h.Class("mb-0 mt-2"),
								h.Li(g.Text("Importieren Sie nur Dumps aus Ihrer eigenen Datenbank. Fremde Dumps können inkonsistente Referenzen enthalten.")),
								h.Li(g.Text("Übernehmen Sie keine Spieldaten aus einer anderen Datenbank in eine neu erstellte DB — die internen IDs (Teams, Ligen, Spielorte) stimmen nicht überein.")),
								h.Li(g.Text("Ein vollständiger Dump enthält bereits alle Stammdaten. Kein zusätzliches Seeding nötig.")),
							),
						),
						h.FormEl(g.Attr("method", "post"), g.Attr("action", "/setup/restore"),
							g.Attr("enctype", "multipart/form-data"),
							h.Div(h.Class("row g-3 align-items-end mb-3"),
								h.Div(h.Class("col"),
									h.Label(h.Class("form-label"), g.Text("SQL-Datei")),
									h.Input(h.Class("form-control"),
										g.Attr("type", "file"), g.Attr("name", "file"),
										g.Attr("accept", ".sql,.txt")),
								),
							),
							h.Div(h.Class("mb-3"),
								h.Label(h.Class("form-label"), g.Text("Oder direkt einfügen")),
								h.Textarea(h.Class("form-control font-monospace"),
									g.Attr("name", "content"), g.Attr("rows", "8"),
									g.Attr("placeholder", "SQL-Dump hier einfügen...")),
							),
							h.Button(h.Class("btn btn-warning"),
								g.Attr("type", "submit"),
								h.I(h.Class("bi bi-upload me-1")),
								g.Text("Wiederherstellen"),
							),
						),
					),
				),

				// Option 3: Skip
				h.Div(h.Class("card mb-4"),
					h.Div(h.Class("card-body"),
						h.H5(h.Class("card-title"),
							h.I(h.Class("bi bi-skip-forward me-2")),
							g.Text("Leere Datenbank"),
						),
						h.P(h.Class("text-muted"),
							g.Text("Nur die Tabellenstruktur — keine Stammdaten. Daten können später manuell angelegt oder über die Datenverwaltung importiert werden."),
						),
						h.FormEl(g.Attr("method", "post"), g.Attr("action", "/setup/skip"),
							h.Button(h.Class("btn btn-outline-secondary"),
								g.Attr("type", "submit"),
								g.Text("Überspringen"),
							),
						),
					),
				),
			),
		),
	)
}

func seedCheckbox(name, label string, checked bool) g.Node {
	attrs := []g.Node{
		h.Class("form-check-input"),
		g.Attr("type", "checkbox"),
		g.Attr("name", name),
		g.Attr("id", "seed-"+name),
		g.Attr("value", "1"),
	}
	if checked {
		attrs = append(attrs, g.Attr("checked", ""))
	}
	return h.Div(h.Class("form-check"),
		h.Input(attrs...),
		h.Label(h.Class("form-check-label"), g.Attr("for", "seed-"+name), g.Text(label)),
	)
}
