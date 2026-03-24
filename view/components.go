package view

import (
	"fmt"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

func FormField(label string, input g.Node, errMsg string) g.Node {
	nodes := []g.Node{
		h.Label(h.Class("form-label"), g.Text(label)),
		input,
	}
	if errMsg != "" {
		nodes = append(nodes, h.Div(h.Class("invalid-feedback d-block"), g.Text(errMsg)))
	}
	return h.Div(h.Class("mb-3"), g.Group(nodes))
}

func FormRow(fields ...g.Node) g.Node {
	var cols []g.Node
	for _, f := range fields {
		cols = append(cols, h.Div(h.Class("col"), f))
	}
	return h.Div(h.Class("row"), g.Group(cols))
}

func TextInput(name, value string, required bool, errMsg string) g.Node {
	cls := "form-control"
	if errMsg != "" {
		cls += " is-invalid"
	}
	attrs := []g.Node{
		h.Class(cls),
		g.Attr("type", "text"),
		g.Attr("name", name),
		g.Attr("id", name),
		g.Attr("value", value),
	}
	if required {
		attrs = append(attrs, g.Attr("required", ""))
	}
	return h.Input(attrs...)
}

func NumberInput(name, value string, required bool, errMsg string) g.Node {
	cls := "form-control"
	if errMsg != "" {
		cls += " is-invalid"
	}
	attrs := []g.Node{
		h.Class(cls),
		g.Attr("type", "number"),
		g.Attr("name", name),
		g.Attr("id", name),
		g.Attr("value", value),
		g.Attr("step", "1"),
	}
	if required {
		attrs = append(attrs, g.Attr("required", ""))
	}
	return h.Input(attrs...)
}

func TextareaInput(name, value string, rows int, errMsg string) g.Node {
	cls := "form-control"
	if errMsg != "" {
		cls += " is-invalid"
	}
	return h.Textarea(
		h.Class(cls),
		g.Attr("name", name),
		g.Attr("id", name),
		g.Attr("rows", fmt.Sprintf("%d", rows)),
		g.Text(value),
	)
}

func SubmitButton(label, cancelURL string) g.Node {
	nodes := []g.Node{
		h.Button(h.Class("btn btn-primary"), g.Attr("type", "submit"), g.Text(label)),
	}
	if cancelURL != "" {
		nodes = append(nodes,
			h.A(h.Class("btn btn-outline-secondary ms-2"), g.Attr("href", cancelURL), g.Text("Abbrechen")),
		)
	}
	return h.Div(h.Class("mt-3"), g.Group(nodes))
}

func DataTable(headers []string, rows []g.Node) g.Node {
	var ths []g.Node
	for _, hdr := range headers {
		ths = append(ths, h.Th(g.Text(hdr)))
	}
	return h.Div(h.Class("table-responsive"),
		h.Table(h.Class("table table-striped table-hover table-sm"),
			h.THead(h.Tr(g.Group(ths))),
			h.TBody(g.Group(rows)),
		),
	)
}

func ActionLinks(editURL, deleteURL string) g.Node {
	return h.Td(
		h.A(h.Class("btn btn-sm btn-outline-primary me-1"), g.Attr("href", editURL),
			h.I(h.Class("bi bi-pencil")),
		),
		DeleteButton(deleteURL),
	)
}

func DeleteButton(url string) g.Node {
	return h.Form(
		h.Class("d-inline"),
		g.Attr("method", "post"),
		g.Attr("action", url),
		g.Attr("onsubmit", "return confirm('Wirklich löschen?')"),
		h.Button(h.Class("btn btn-sm btn-outline-danger"), g.Attr("type", "submit"),
			h.I(h.Class("bi bi-trash")),
		),
	)
}
