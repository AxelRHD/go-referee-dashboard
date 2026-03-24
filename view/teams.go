package view

import (
	"fmt"
	"net/http"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"github.com/axelrhd/referee-dashboard/model"
	"github.com/axelrhd/referee-dashboard/validation"
)

func TeamList(w http.ResponseWriter, r *http.Request, teams []model.Team) g.Node {
	var rows []g.Node
	for _, t := range teams {
		active := "Nein"
		if t.IsActive == 1 {
			active = "Ja"
		}
		rows = append(rows, h.Tr(
			h.Td(g.Text(t.Name)),
			h.Td(g.Text(t.State)),
			h.Td(g.Text(active)),
			h.Td(g.Text(t.Remarks)),
			ActionLinks(
				fmt.Sprintf("/teams/%d/edit", t.ID),
				fmt.Sprintf("/teams/%d/delete", t.ID),
			),
		))
	}

	return BasePage("Teams",
		FlashAlert(w, r),
		h.H1(g.Text("Teams")),
		h.Div(h.Class("d-flex justify-content-between mb-3"),
			h.Div(
				h.A(h.Class("btn btn-success"), g.Attr("href", "/teams/new"), g.Text("Neues Team")),
			),
			h.Div(
				h.A(h.Class("btn btn-sm btn-outline-secondary me-1"), g.Attr("href", "/teams/export/csv"), g.Text("CSV")),
				h.A(h.Class("btn btn-sm btn-outline-secondary"), g.Attr("href", "/teams/export/sql"), g.Text("SQL")),
			),
		),
		DataTable(
			[]string{"Name", "Bundesland", "Aktiv", "Bemerkungen", "Aktionen"},
			rows,
		),
	)
}

func TeamForm(team *model.Team, errors map[string]string, data map[string]string) g.Node {
	title := "Neues Team"
	action := "/teams/new"
	if team != nil {
		title = "Team bearbeiten"
		action = fmt.Sprintf("/teams/%d/edit", team.ID)
	}

	val := func(field string) string {
		if data != nil {
			if v, ok := data[field]; ok {
				return v
			}
		}
		if team == nil {
			switch field {
			case "state":
				return "Baden-Württemberg"
			case "is_active":
				return "1"
			}
			return ""
		}
		switch field {
		case "name":
			return team.Name
		case "state":
			return team.State
		case "is_active":
			return fmt.Sprintf("%d", team.IsActive)
		case "remarks":
			return team.Remarks
		}
		return ""
	}

	errFor := func(field string) string {
		if errors != nil {
			return errors[field]
		}
		return ""
	}

	// Datalist options for Bundesland
	var datalistOpts []g.Node
	for _, bl := range validation.Bundeslaender {
		datalistOpts = append(datalistOpts, h.Option(g.Attr("value", bl)))
	}

	stateInput := h.Div(h.Class("mb-3"),
		h.Label(h.Class("form-label"), g.Attr("for", "state"), g.Text("Bundesland")),
		h.Input(
			h.Class("form-control"),
			g.Attr("type", "text"),
			g.Attr("name", "state"),
			g.Attr("id", "state"),
			g.Attr("value", val("state")),
			g.Attr("list", "bundeslaender"),
		),
		h.DataList(g.Attr("id", "bundeslaender"), g.Group(datalistOpts)),
	)

	isChecked := val("is_active") == "1"

	checkboxAttrs := []g.Node{
		h.Class("form-check-input"),
		g.Attr("type", "checkbox"),
		g.Attr("name", "is_active"),
		g.Attr("id", "is_active"),
		g.Attr("value", "1"),
	}
	if isChecked {
		checkboxAttrs = append(checkboxAttrs, g.Attr("checked", ""))
	}

	return BasePage(title,
		h.H1(g.Text(title)),
		h.Form(g.Attr("method", "post"), g.Attr("action", action),
			FormRow(
				FormField("Name", TextInput("name", val("name"), true, errFor("name")), errFor("name")),
				stateInput,
			),
			FormField("Bemerkungen", TextareaInput("remarks", val("remarks"), 1, ""), ""),
			h.Div(h.Class("mb-3 form-check"),
				h.Input(checkboxAttrs...),
				h.Label(h.Class("form-check-label"), g.Attr("for", "is_active"), g.Text("Aktiv")),
			),
			SubmitButton("Speichern", "/teams"),
		),
	)
}
