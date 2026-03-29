package view

import (
	"fmt"
	"net/http"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"github.com/axelrhd/referee-dashboard/store"
)

func DataPage(w http.ResponseWriter, r *http.Request, positions []store.Position, messages []string) g.Node {
	var alerts []g.Node
	for _, msg := range messages {
		alerts = append(alerts, h.Div(h.Class("alert alert-success alert-dismissible fade show"),
			g.Text(msg),
			h.Button(h.Class("btn-close"), g.Attr("type", "button"), g.Attr("data-bs-dismiss", "alert")),
		))
	}

	return BasePage("Datenverwaltung",
		FlashAlert(w, r),
		g.Group(alerts),
		h.H1(g.Text("Datenverwaltung")),

		// Export section
		h.Div(h.Class("card mb-4"),
			h.Div(h.Class("card-body"),
				h.H5(h.Class("card-title"), g.Text("Export")),
				h.P(h.Class("text-muted"), g.Text("Exportiert die gesamte Datenbank als JSON-Datei.")),
				h.A(h.Class("btn btn-outline-primary"), g.Attr("href", "/data/export"),
					g.Text("JSON-Export")),
			),
		),

		// Cache refresh
		h.Div(h.Class("card mb-4"),
			h.Div(h.Class("card-body"),
				h.H5(h.Class("card-title"), g.Text("Dashboard")),
				h.P(h.Class("text-muted"), g.Text("Dashboard-Daten werden beim Laden gecacht. Nach Änderungen an Spielen, Teams oder Ligen hier den Cache leeren.")),
				h.A(h.Class("btn btn-outline-secondary"), g.Attr("href", "/dashboard/?reload=1"),
					h.I(h.Class("bi bi-arrow-clockwise me-1")),
					g.Text("Dashboard-Cache leeren"),
				),
			),
		),

		// JSON import
		h.Div(h.Class("card mb-4"),
			h.Div(h.Class("card-body"),
				h.H5(h.Class("card-title"), g.Text("JSON-Import")),
				h.P(h.Class("text-muted"), g.Text("Importiert eine zuvor exportierte JSON-Datei. Achtung: Überschreibt alle bestehenden Daten!")),
				h.FormEl(g.Attr("method", "post"), g.Attr("action", "/data/import"),
					g.Attr("enctype", "multipart/form-data"),
					h.Div(h.Class("row g-3 align-items-end"),
						h.Div(h.Class("col"),
							h.Label(h.Class("form-label"), g.Text("Datei")),
							h.Input(h.Class("form-control"),
								g.Attr("type", "file"), g.Attr("name", "file"),
								g.Attr("accept", ".json")),
						),
						h.Div(h.Class("col-auto"),
							h.Button(h.Class("btn btn-primary"), g.Attr("type", "submit"),
								g.Text("Importieren")),
						),
					),
				),
			),
		),

		// Positions
		positionsCard(positions),
	)
}

func DataPageWithMessages(w http.ResponseWriter, r *http.Request, messages []string) g.Node {
	return DataPage(w, r, nil, messages)
}

func positionsCard(positions []store.Position) g.Node {
	var rows []g.Node
	for _, p := range positions {
		rows = append(rows, h.Tr(
			h.Td(g.Text(p.Position)),
			h.Td(g.Text(p.Long)),
			h.Td(h.Class("text-end"), g.Text(fmt.Sprintf("%d", p.Sorter))),
			h.Td(h.Class("text-end"),
				// Inline edit form
				h.FormEl(h.Class("d-inline"), g.Attr("method", "post"),
					g.Attr("action", fmt.Sprintf("/data/positions/%s/edit", p.Position)),
					h.Input(g.Attr("type", "hidden"), g.Attr("name", "long"), g.Attr("value", p.Long)),
					h.Input(g.Attr("type", "hidden"), g.Attr("name", "sorter"), g.Attr("value", fmt.Sprintf("%d", p.Sorter))),
					h.Button(h.Class("btn btn-sm btn-outline-primary me-1"),
						g.Attr("type", "button"),
						g.Attr("onclick", fmt.Sprintf("editPosition(this, '%s', '%s', %d)", p.Position, p.Long, p.Sorter)),
						g.Text("Bearbeiten"),
					),
				),
				h.FormEl(h.Class("d-inline"), g.Attr("method", "post"),
					g.Attr("action", fmt.Sprintf("/data/positions/%s/delete", p.Position)),
					g.Attr("onsubmit", fmt.Sprintf("return confirm('Position %s wirklich löschen?')", p.Position)),
					h.Button(h.Class("btn btn-sm btn-outline-danger"), g.Attr("type", "submit"),
						g.Text("Löschen"),
					),
				),
			),
		))
	}

	return h.Div(h.Class("card mb-4"),
		h.Div(h.Class("card-body"),
			h.H5(h.Class("card-title"), g.Text("Positionen")),
			h.P(h.Class("text-muted"), g.Text("Schiedsrichter-Positionen verwalten. Diese werden selten geändert.")),
			h.Table(h.Class("table table-sm"),
				h.THead(
					h.Tr(
						h.Th(g.Text("Kürzel")),
						h.Th(g.Text("Bezeichnung")),
						h.Th(h.Class("text-end"), g.Text("Sortierung")),
						h.Th(h.Class("text-end"), g.Text("Aktionen")),
					),
				),
				h.TBody(g.Group(rows)),
			),
			// Add new position form
			h.FormEl(g.Attr("method", "post"), g.Attr("action", "/data/positions/new"),
				h.Div(h.Class("row g-2 align-items-end"),
					h.Div(h.Class("col-auto"),
						h.Label(h.Class("form-label"), g.Text("Kürzel")),
						h.Input(h.Class("form-control form-control-sm"),
							g.Attr("type", "text"), g.Attr("name", "position"),
							g.Attr("placeholder", "z.B. SJ"), g.Attr("required", ""),
							g.Attr("style", "width:80px")),
					),
					h.Div(h.Class("col"),
						h.Label(h.Class("form-label"), g.Text("Bezeichnung")),
						h.Input(h.Class("form-control form-control-sm"),
							g.Attr("type", "text"), g.Attr("name", "long"),
							g.Attr("placeholder", "z.B. Side Judge")),
					),
					h.Div(h.Class("col-auto"),
						h.Label(h.Class("form-label"), g.Text("Sortierung")),
						h.Input(h.Class("form-control form-control-sm"),
							g.Attr("type", "number"), g.Attr("name", "sorter"),
							g.Attr("value", "0"), g.Attr("style", "width:80px")),
					),
					h.Div(h.Class("col-auto"),
						h.Button(h.Class("btn btn-sm btn-success"), g.Attr("type", "submit"),
							g.Text("Hinzufügen")),
					),
				),
			),
		),
		// Inline edit JS
		h.Script(g.Raw(`
function editPosition(btn, key, long, sorter) {
    var row = btn.closest('tr');
    var cells = row.querySelectorAll('td');
    var origLong = cells[1].textContent;
    var origSorter = cells[2].textContent.trim();

    cells[1].innerHTML = '<input class="form-control form-control-sm" name="long" value="' + long + '">';
    cells[2].innerHTML = '<input class="form-control form-control-sm text-end" name="sorter" value="' + sorter + '" style="width:80px">';

    var form = btn.closest('form');
    btn.textContent = 'Speichern';
    btn.className = 'btn btn-sm btn-primary me-1';
    btn.setAttribute('type', 'submit');
    btn.removeAttribute('onclick');

    // Update hidden fields on submit
    form.addEventListener('submit', function() {
        form.querySelector('input[name="long"]').value = cells[1].querySelector('input').value;
        form.querySelector('input[name="sorter"]').value = cells[2].querySelector('input').value;
    });
}
`)),
	)
}
