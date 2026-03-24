package view

import (
	"net/http"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func DataPage(w http.ResponseWriter, r *http.Request) g.Node {
	return DataPageWithMessages(w, r, nil)
}

func DataPageWithMessages(w http.ResponseWriter, r *http.Request, messages []string) g.Node {
	var alerts []g.Node
	for _, msg := range messages {
		alerts = append(alerts, h.Div(h.Class("alert alert-success alert-dismissible fade show"),
			g.Text(msg),
			h.Button(h.Class("btn-close"), g.Attr("type", "button"), g.Attr("data-bs-dismiss", "alert")),
		))
	}

	return BasePage("Datenverwaltung",
		g.Group(alerts),
		h.H1(g.Text("Datenverwaltung")),
		// Export section
		h.Div(h.Class("card mb-4"),
			h.Div(h.Class("card-body"),
				h.H5(h.Class("card-title"), g.Text("Export")),
				h.Div(h.Class("d-flex gap-2"),
					h.A(h.Class("btn btn-outline-primary"), g.Attr("href", "/data/dump"),
						g.Text("SQLite-Dump (Schema + Daten)")),
					h.A(h.Class("btn btn-outline-primary"), g.Attr("href", "/data/export-all"),
						g.Text("Alle Daten (nur INSERTs)")),
				),
			),
		),
		// File import
		h.Div(h.Class("card mb-4"),
			h.Div(h.Class("card-body"),
				h.H5(h.Class("card-title"), g.Text("Datei importieren")),
				h.FormEl(g.Attr("method", "post"), g.Attr("action", "/data/import"),
					g.Attr("enctype", "multipart/form-data"),
					h.Div(h.Class("row g-3 align-items-end"),
						h.Div(h.Class("col-auto"),
							h.Label(h.Class("form-label"), g.Text("Format")),
							formatSelect(),
						),
						h.Div(h.Class("col-auto entity-group d-none"),
							h.Label(h.Class("form-label"), g.Text("Entity")),
							entitySelect(),
						),
						h.Div(h.Class("col"),
							h.Label(h.Class("form-label"), g.Text("Datei")),
							h.Input(h.Class("form-control"),
								g.Attr("type", "file"), g.Attr("name", "file"),
								g.Attr("accept", ".sql,.csv,.txt")),
						),
						h.Div(h.Class("col-auto"),
							h.Button(h.Class("btn btn-primary"), g.Attr("type", "submit"),
								g.Text("Importieren")),
						),
					),
				),
			),
		),
		// Textarea import
		h.Div(h.Class("card mb-4"),
			h.Div(h.Class("card-body"),
				h.H5(h.Class("card-title"), g.Text("Direkt einfügen")),
				h.FormEl(g.Attr("method", "post"), g.Attr("action", "/data/paste"),
					h.Div(h.Class("row g-3 mb-3"),
						h.Div(h.Class("col-auto"),
							h.Label(h.Class("form-label"), g.Text("Format")),
							formatSelect(),
						),
						h.Div(h.Class("col-auto entity-group d-none"),
							h.Label(h.Class("form-label"), g.Text("Entity")),
							entitySelect(),
						),
					),
					h.Textarea(h.Class("form-control font-monospace"),
						g.Attr("name", "content"), g.Attr("rows", "12"),
						g.Attr("placeholder", "SQL-Statements oder CSV-Daten hier einfügen...")),
					h.Div(h.Class("mt-3"),
						h.Button(h.Class("btn btn-primary"), g.Attr("type", "submit"),
							g.Text("Ausführen")),
					),
				),
			),
		),
		// JS: toggle entity select based on format
		h.Script(g.Raw(`
document.querySelectorAll('.format-toggle').forEach(function(sel) {
    var form = sel.closest('form');
    var entity = form.querySelector('.entity-group');
    function toggle() { entity.classList.toggle('d-none', sel.value !== 'csv'); }
    sel.addEventListener('change', toggle);
    toggle();
});
`)),
	)
}

func formatSelect() g.Node {
	return h.Select(h.Class("form-select format-toggle"), g.Attr("name", "format"),
		h.Option(g.Attr("value", "sql"), g.Text("SQL")),
		h.Option(g.Attr("value", "csv"), g.Text("CSV")),
	)
}

func entitySelect() g.Node {
	return h.Select(h.Class("form-select"), g.Attr("name", "entity"),
		h.Option(g.Attr("value", "games"), g.Text("Spiele")),
		h.Option(g.Attr("value", "teams"), g.Text("Teams")),
		h.Option(g.Attr("value", "leagues"), g.Text("Ligen")),
	)
}
