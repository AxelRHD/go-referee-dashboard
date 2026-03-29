package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/store"
	"github.com/axelrhd/referee-dashboard/validation"
	"github.com/axelrhd/referee-dashboard/view"
)

type VenueHandler struct {
	s *store.Store
}

func NewVenueHandler(s *store.Store) *VenueHandler {
	return &VenueHandler{s: s}
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
	venues, err := vh.s.ListVenues()
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
	data, errs := validation.ValidateVenue(r.Form)

	if len(errs) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.VenueForm(nil, errs, formValues(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	err := vh.s.PutVenue(&store.Venue{
		City:      data.City,
		Stadium:   data.Stadium,
		ShortName: data.ShortName,
		Lat:       data.Lat,
		Lon:       data.Lon,
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
	data, errs := validation.ValidateVenue(r.Form)

	if len(errs) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.VenueForm(&venue, errs, formValues(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	err = vh.s.PutVenue(&store.Venue{
		ID:        venue.ID,
		City:      data.City,
		Stadium:   data.Stadium,
		ShortName: data.ShortName,
		Lat:       data.Lat,
		Lon:       data.Lon,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Spielort wurde aktualisiert.")
	http.Redirect(w, r, "/venues", http.StatusSeeOther)
}

func (vh *VenueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	if err := vh.s.DeleteVenue(id); err != nil {
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

	client := &http.Client{Timeout: 10 * time.Second}

	// Try with stadium + city first, fall back to city only
	type attempt struct {
		query    string
		fallback bool
	}
	attempts := []attempt{}
	if body.Stadium != "" {
		attempts = append(attempts, attempt{body.Stadium + " " + body.City, false})
	}
	attempts = append(attempts, attempt{body.City, body.Stadium != ""})

	for _, a := range attempts {
		lat, lon, err := photonGeocode(client, a.query)
		if err != nil || lat == 0 {
			continue
		}
		resp := map[string]any{"lat": lat, "lon": lon}
		if a.fallback {
			resp["fallback"] = true
			resp["message"] = fmt.Sprintf("Stadion nicht gefunden, verwende Koordinaten für %s.", body.City)
		}
		jsonResponse(w, resp)
		return
	}

	jsonError(w, "Keine Koordinaten gefunden.", http.StatusNotFound)
}

// JSON API

func (vh *VenueHandler) APIList(w http.ResponseWriter, r *http.Request) {
	venues, err := vh.s.ListVenues()
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
	venues, err := vh.s.ListVenues()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", exportFilename("spielorte", "csv"))
	w.Write([]byte("\xEF\xBB\xBF"))

	cw := csv.NewWriter(w)
	cw.Comma = ';'
	cw.Write([]string{"Kürzel", "Stadt", "Stadion", "Lat", "Lon"})

	for _, v := range venues {
		lat := ""
		lon := ""
		if v.Lat != 0 {
			lat = fmt.Sprintf("%f", v.Lat)
		}
		if v.Lon != 0 {
			lon = fmt.Sprintf("%f", v.Lon)
		}
		cw.Write([]string{v.ShortName, v.City, v.Stadium, lat, lon})
	}
	cw.Flush()
}

func (vh *VenueHandler) ExportSQL(w http.ResponseWriter, r *http.Request) {
	venues, err := vh.s.ListVenues()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", exportFilename("spielorte", "sql"))

	for _, v := range venues {
		fmt.Fprintf(w, "INSERT INTO venues (city, stadium, short_name, lat, lon) VALUES (%s, %s, %s, %f, %f);\n",
			sqlEscape(v.City), sqlEscape(v.Stadium), sqlEscape(v.ShortName), v.Lat, v.Lon)
	}
}

// Helpers

func photonGeocode(client *http.Client, query string) (float64, float64, error) {
	req, _ := http.NewRequest("GET", "https://photon.komoot.io/api/", nil)
	q := req.URL.Query()
	q.Set("q", query+" Deutschland")
	q.Set("limit", "1")
	q.Set("lang", "de")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Features []struct {
			Geometry struct {
				Coordinates []float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}
	if len(result.Features) == 0 || len(result.Features[0].Geometry.Coordinates) < 2 {
		return 0, 0, nil
	}
	coords := result.Features[0].Geometry.Coordinates
	return coords[1], coords[0], nil // Photon: [lon, lat]
}

func (vh *VenueHandler) getVenue(w http.ResponseWriter, r *http.Request) (store.Venue, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return store.Venue{}, fmt.Errorf("empty id")
	}
	venue, err := vh.s.GetVenue(id)
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "Spielort nicht gefunden", http.StatusNotFound)
		return store.Venue{}, err
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return store.Venue{}, err
	}
	return venue, nil
}
