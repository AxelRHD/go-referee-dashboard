package view

import (
	"fmt"
	"net/http"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"github.com/axelrhd/referee-dashboard/store"
)

func LeagueList(w http.ResponseWriter, r *http.Request, leagues []store.League) g.Node {
	var rows []g.Node
	for _, lg := range leagues {
		rows = append(rows, h.Tr(
			h.Td(h.Class("text-end"), g.Text(fmt.Sprintf("%d", lg.Sorter))),
			h.Td(g.Text(lg.Name)),
			h.Td(g.Text(lg.ShortName)),
			h.Td(g.Text(lg.Remarks)),
			ActionLinks(
				fmt.Sprintf("/leagues/%s/edit", lg.ID),
				fmt.Sprintf("/leagues/%s/delete", lg.ID),
			),
		))
	}

	return BasePage("Ligen",
		FlashAlert(w, r),
		h.H1(g.Text("Ligen")),
		h.Div(h.Class("d-flex justify-content-between mb-3"),
			h.Div(
				h.A(h.Class("btn btn-success"), g.Attr("href", "/leagues/new"), g.Text("Neue Liga")),
			),
			h.Div(
				h.A(h.Class("btn btn-sm btn-outline-secondary me-1"), g.Attr("href", "/leagues/export/csv"), g.Text("CSV")),
				h.A(h.Class("btn btn-sm btn-outline-secondary"), g.Attr("href", "/leagues/export/sql"), g.Text("SQL")),
			),
		),
		DataTable(
			[]string{"#", "Name", "Kürzel", "Bemerkungen", "Aktionen"},
			rows,
		),
	)
}

func LeagueForm(league *store.League, errors map[string]string, data map[string]string) g.Node {
	title := "Neue Liga"
	action := "/leagues/new"
	if league != nil {
		title = "Liga bearbeiten"
		action = fmt.Sprintf("/leagues/%s/edit", league.ID)
	}

	val := func(field string) string {
		if data != nil {
			if v, ok := data[field]; ok {
				return v
			}
		}
		if league == nil {
			return ""
		}
		switch field {
		case "name":
			return league.Name
		case "short_name":
			return league.ShortName
		case "sorter":
			return fmt.Sprintf("%d", league.Sorter)
		case "remarks":
			return league.Remarks
		}
		return ""
	}

	errFor := func(field string) string {
		if errors != nil {
			return errors[field]
		}
		return ""
	}

	return BasePage(title,
		h.H1(g.Text(title)),
		h.Form(g.Attr("method", "post"), g.Attr("action", action),
			FormRow(
				FormField("Name", TextInput("name", val("name"), true, errFor("name")), errFor("name")),
				FormField("Kürzel", TextInput("short_name", val("short_name"), false, ""), ""),
				FormField("Sortierung", NumberInput("sorter", val("sorter"), false, errFor("sorter")), errFor("sorter")),
			),
			FormField("Bemerkungen", TextareaInput("remarks", val("remarks"), 2, errFor("remarks")), errFor("remarks")),
			SubmitButton("Speichern", "/leagues"),
		),
	)
}
