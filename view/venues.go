package view

import (
	"fmt"
	"net/http"

	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"

	"github.com/axelrhd/referee-dashboard/model"
)

func VenueList(w http.ResponseWriter, r *http.Request, venues []model.Venue) g.Node {
	var rows []g.Node
	for _, v := range venues {
		lat := ""
		lon := ""
		if v.Lat != 0 {
			lat = fmt.Sprintf("%.4f", v.Lat)
		}
		if v.Lon != 0 {
			lon = fmt.Sprintf("%.4f", v.Lon)
		}
		rows = append(rows, h.Tr(
			h.Td(g.Text(v.ShortName)),
			h.Td(g.Text(v.City)),
			h.Td(g.Text(v.Stadium)),
			h.Td(h.Class("text-muted"), g.Text(lat)),
			h.Td(h.Class("text-muted"), g.Text(lon)),
			ActionLinks(
				fmt.Sprintf("/venues/%d/edit", v.ID),
				fmt.Sprintf("/venues/%d/delete", v.ID),
			),
		))
	}

	return BasePage("Spielorte",
		FlashAlert(w, r),
		h.H1(g.Text("Spielorte")),
		h.Div(h.Class("d-flex justify-content-between mb-3"),
			h.Div(
				h.A(h.Class("btn btn-success"), g.Attr("href", "/venues/new"), g.Text("Neuer Spielort")),
			),
			h.Div(
				h.A(h.Class("btn btn-sm btn-outline-secondary me-1"), g.Attr("href", "/venues/export/csv"), g.Text("CSV")),
				h.A(h.Class("btn btn-sm btn-outline-secondary"), g.Attr("href", "/venues/export/sql"), g.Text("SQL")),
			),
		),
		DataTable(
			[]string{"Kürzel", "Stadt", "Stadion", "Lat", "Lon", "Aktionen"},
			rows,
		),
	)
}

func VenueForm(venue *model.Venue, errors map[string]string, data map[string]string) g.Node {
	title := "Neuer Spielort"
	action := "/venues/new"
	if venue != nil {
		title = "Spielort bearbeiten"
		action = fmt.Sprintf("/venues/%d/edit", venue.ID)
	}

	val := func(field string) string {
		if data != nil {
			if v, ok := data[field]; ok {
				return v
			}
		}
		if venue == nil {
			return ""
		}
		switch field {
		case "city":
			return venue.City
		case "stadium":
			return venue.Stadium
		case "short_name":
			return venue.ShortName
		case "lat":
			if venue.Lat != 0 {
				return fmt.Sprintf("%.6f", venue.Lat)
			}
			return ""
		case "lon":
			if venue.Lon != 0 {
				return fmt.Sprintf("%.6f", venue.Lon)
			}
			return ""
		}
		return ""
	}

	errFor := func(field string) string {
		if errors != nil {
			return errors[field]
		}
		return ""
	}

	latVal := val("lat")
	lonVal := val("lon")

	xData := fmt.Sprintf(`{
		city: '%s', stadium: '%s', lat: '%s', lon: '%s',
		async geocode() {
			if (!this.city) return;
			try {
				let res = await fetch('/api/venues/geocode', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({city: this.city, stadium: this.stadium})
				});
				let data = await res.json();
				if (data.error) { alert(data.error); return; }
				this.lat = data.lat;
				this.lon = data.lon;
				if (data.message) { alert(data.message); }
			} catch(e) { alert('Geocoding-Fehler: ' + e); }
		}
	}`, val("city"), val("stadium"), latVal, lonVal)

	cityInput := h.Input(h.Class("form-control"), g.Attr("type", "text"), g.Attr("name", "city"), g.Attr("id", "city"), g.Attr("x-model", "city"), g.Attr("required", ""))
	if errFor("city") != "" {
		cityInput = h.Input(h.Class("form-control is-invalid"), g.Attr("type", "text"), g.Attr("name", "city"), g.Attr("id", "city"), g.Attr("x-model", "city"), g.Attr("required", ""))
	}
	stadiumInput := h.Input(h.Class("form-control"), g.Attr("type", "text"), g.Attr("name", "stadium"), g.Attr("id", "stadium"), g.Attr("x-model", "stadium"))
	latInput := h.Input(h.Class("form-control"), g.Attr("type", "text"), g.Attr("name", "lat"), g.Attr("id", "lat"), g.Attr("x-model", "lat"))
	lonInput := h.Input(h.Class("form-control"), g.Attr("type", "text"), g.Attr("name", "lon"), g.Attr("id", "lon"), g.Attr("x-model", "lon"))

	return BasePage(title,
		h.H1(g.Text(title)),
		h.Form(
			g.Attr("method", "post"), g.Attr("action", action),
			g.Attr("x-data", xData),
			FormRow(
				FormField("Stadt", cityInput, errFor("city")),
				FormField("Stadion", stadiumInput, ""),
				FormField("Kürzel", TextInput("short_name", val("short_name"), false, ""), ""),
			),
			FormRow(
				FormField("Lat", latInput, ""),
				FormField("Lon", lonInput, ""),
			),
			h.Div(h.Class("mb-3"),
				h.Button(h.Class("btn btn-sm btn-outline-primary"),
					g.Attr("type", "button"),
					g.Attr("@click", "geocode()"),
					g.Text("Koordinaten nachschlagen"),
				),
			),
			SubmitButton("Speichern", "/venues"),
		),
	)
}
