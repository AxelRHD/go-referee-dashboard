package handler

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/model"
	"github.com/axelrhd/referee-dashboard/validation"
	"github.com/axelrhd/referee-dashboard/view"
)

type VenueHandler struct {
	q *model.Queries
}

func NewVenueHandler(q *model.Queries) *VenueHandler {
	return &VenueHandler{q: q}
}

func (h *VenueHandler) Routes(r chi.Router) {
	r.Get("/venues", h.List)
	r.Get("/venues/new", h.NewForm)
	r.Post("/venues/new", h.Create)
	r.Get("/venues/{id}/edit", h.EditForm)
	r.Post("/venues/{id}/edit", h.Update)
	r.Post("/venues/{id}/delete", h.Delete)
	r.Get("/venues/export/csv", h.ExportCSV)
	r.Get("/venues/export/sql", h.ExportSQL)
}

func (h *VenueHandler) APIRoutes(r chi.Router) {
	r.Get("/venues", h.APIList)
	r.Get("/venues/{id}", h.APIGet)
	r.Post("/venues/geocode", h.Geocode)
}

// HTML Handlers

func (vh *VenueHandler) List(w http.ResponseWriter, r *http.Request) {
	venues, err := vh.q.ListVenues(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page := view.VenueList(w, r, venues)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (vh *VenueHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	page := view.VenueForm(nil, nil, nil)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (vh *VenueHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	data, errors := validation.ValidateVenue(r.Form)

	if len(errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.VenueForm(nil, errors, formValues(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	_, err := vh.q.CreateVenue(r.Context(), model.CreateVenueParams{
		City:    data.City,
		Stadium: data.Stadium,
		Lat:     data.Lat,
		Lon:     data.Lon,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Spielort wurde erstellt.")
	http.Redirect(w, r, "/venues", http.StatusSeeOther)
}

func (vh *VenueHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	venue, err := vh.getVenue(w, r)
	if err != nil {
		return
	}
	page := view.VenueForm(&venue, nil, nil)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (vh *VenueHandler) Update(w http.ResponseWriter, r *http.Request) {
	venue, err := vh.getVenue(w, r)
	if err != nil {
		return
	}

	r.ParseForm()
	data, errors := validation.ValidateVenue(r.Form)

	if len(errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.VenueForm(&venue, errors, formValues(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	err = vh.q.UpdateVenue(r.Context(), model.UpdateVenueParams{
		ID:      venue.ID,
		City:    data.City,
		Stadium: data.Stadium,
		Lat:     data.Lat,
		Lon:     data.Lon,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Spielort wurde aktualisiert.")
	http.Redirect(w, r, "/venues", http.StatusSeeOther)
}

func (vh *VenueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	if err := vh.q.DeleteVenue(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Spielort wurde gelöscht.")
	http.Redirect(w, r, "/venues", http.StatusSeeOther)
}

func (vh *VenueHandler) Geocode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		City    string `json:"city"`
		Stadium string `json:"stadium"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Ungültige Anfrage", http.StatusBadRequest)
		return
	}

	query := body.City + ", Germany"
	if body.Stadium != "" {
		query = body.Stadium + ", " + query
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", "https://nominatim.openstreetmap.org/search", nil)
	q := req.URL.Query()
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("limit", "1")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", "RefereeApp/1.0")

	resp, err := client.Do(req)
	if err != nil {
		jsonError(w, fmt.Sprintf("Geocoding-Fehler: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		jsonError(w, fmt.Sprintf("Geocoding-Fehler: %v", err), http.StatusInternalServerError)
		return
	}

	if len(results) == 0 {
		jsonError(w, "Keine Koordinaten gefunden.", http.StatusNotFound)
		return
	}

	jsonResponse(w, map[string]string{
		"lat": results[0].Lat,
		"lon": results[0].Lon,
	})
}

// JSON API

func (vh *VenueHandler) APIList(w http.ResponseWriter, r *http.Request) {
	venues, err := vh.q.ListVenues(r.Context())
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, venues)
}

func (vh *VenueHandler) APIGet(w http.ResponseWriter, r *http.Request) {
	venue, err := vh.getVenue(w, r)
	if err != nil {
		return
	}
	jsonResponse(w, venue)
}

// Export

func (vh *VenueHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	venues, err := vh.q.ListVenues(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="spielorte.csv"`)
	w.Write([]byte("\xEF\xBB\xBF"))

	cw := csv.NewWriter(w)
	cw.Comma = ';'
	cw.Write([]string{"Stadt", "Stadion", "Lat", "Lon"})

	for _, v := range venues {
		lat := ""
		lon := ""
		if v.Lat != 0 {
			lat = fmt.Sprintf("%f", v.Lat)
		}
		if v.Lon != 0 {
			lon = fmt.Sprintf("%f", v.Lon)
		}
		cw.Write([]string{v.City, v.Stadium, lat, lon})
	}
	cw.Flush()
}

func (vh *VenueHandler) ExportSQL(w http.ResponseWriter, r *http.Request) {
	venues, err := vh.q.ListVenues(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="spielorte.sql"`)

	for _, v := range venues {
		fmt.Fprintf(w, "INSERT INTO venues (city, stadium, lat, lon) VALUES (%s, %s, %f, %f);\n",
			sqlEscape(v.City), sqlEscape(v.Stadium), v.Lat, v.Lon)
	}
}

// Helpers

func (vh *VenueHandler) getVenue(w http.ResponseWriter, r *http.Request) (model.Venue, error) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return model.Venue{}, err
	}
	venue, err := vh.q.GetVenue(r.Context(), id)
	if err == sql.ErrNoRows {
		http.Error(w, "Spielort nicht gefunden", http.StatusNotFound)
		return model.Venue{}, err
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return model.Venue{}, err
	}
	return venue, nil
}
