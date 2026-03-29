package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/store"
	"github.com/axelrhd/referee-dashboard/validation"
	"github.com/axelrhd/referee-dashboard/view"
)

type LeagueHandler struct {
	s *store.Store
}

func NewLeagueHandler(s *store.Store) *LeagueHandler {
	return &LeagueHandler{s: s}
}

func (h *LeagueHandler) Routes(r chi.Router) {
	r.Get("/leagues", h.List)
	r.Get("/leagues/new", h.NewForm)
	r.Post("/leagues/new", h.Create)
	r.Get("/leagues/{id}/edit", h.EditForm)
	r.Post("/leagues/{id}/edit", h.Update)
	r.Post("/leagues/{id}/delete", h.Delete)
	r.Get("/leagues/export/csv", h.ExportCSV)
	r.Get("/leagues/export/sql", h.ExportSQL)
}

func (h *LeagueHandler) APIRoutes(r chi.Router) {
	r.Get("/leagues", h.APIList)
	r.Get("/leagues/{id}", h.APIGet)
}

// HTML Handlers

func (lh *LeagueHandler) List(w http.ResponseWriter, r *http.Request) {
	leagues, err := lh.s.ListLeagues()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page := view.LeagueList(w, r, leagues)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (lh *LeagueHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	page := view.LeagueForm(nil, nil, nil)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (lh *LeagueHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	data, errs := validation.ValidateLeague(r.Form)

	if len(errs) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.LeagueForm(nil, errs, formValues(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	err := lh.s.PutLeague(&store.League{
		Name:      data.Name,
		ShortName: data.ShortName,
		Sorter:    int(data.Sorter),
		Remarks:   data.Remarks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Liga wurde erstellt.")
	http.Redirect(w, r, "/leagues", http.StatusSeeOther)
}

func (lh *LeagueHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	league, err := lh.getLeague(w, r)
	if err != nil {
		return
	}
	page := view.LeagueForm(&league, nil, nil)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (lh *LeagueHandler) Update(w http.ResponseWriter, r *http.Request) {
	league, err := lh.getLeague(w, r)
	if err != nil {
		return
	}

	r.ParseForm()
	data, errs := validation.ValidateLeague(r.Form)

	if len(errs) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		page := view.LeagueForm(&league, errs, formValues(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	err = lh.s.PutLeague(&store.League{
		ID:        league.ID,
		Name:      data.Name,
		ShortName: data.ShortName,
		Sorter:    int(data.Sorter),
		Remarks:   data.Remarks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Liga wurde aktualisiert.")
	http.Redirect(w, r, "/leagues", http.StatusSeeOther)
}

func (lh *LeagueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return
	}

	if err := lh.s.DeleteLeague(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.SetFlash(w, "Liga wurde gelöscht.")
	http.Redirect(w, r, "/leagues", http.StatusSeeOther)
}

// JSON API

func (lh *LeagueHandler) APIList(w http.ResponseWriter, r *http.Request) {
	leagues, err := lh.s.ListLeagues()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, leagues)
}

func (lh *LeagueHandler) APIGet(w http.ResponseWriter, r *http.Request) {
	league, err := lh.getLeague(w, r)
	if err != nil {
		return
	}
	jsonResponse(w, league)
}

// Export

func (lh *LeagueHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	leagues, err := lh.s.ListLeagues()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", exportFilename("ligen", "csv"))
	// UTF-8 BOM for Excel
	w.Write([]byte("\xEF\xBB\xBF"))

	cw := csv.NewWriter(w)
	cw.Comma = ';'
	cw.Write([]string{"Name", "Kürzel", "Sortierung", "Bemerkungen"})

	for _, lg := range leagues {
		cw.Write([]string{lg.Name, lg.ShortName, fmt.Sprintf("%d", lg.Sorter), lg.Remarks})
	}
	cw.Flush()
}

func (lh *LeagueHandler) ExportSQL(w http.ResponseWriter, r *http.Request) {
	leagues, err := lh.s.ListLeagues()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", exportFilename("ligen", "sql"))

	for _, lg := range leagues {
		fmt.Fprintf(w, "INSERT INTO leagues (name, short_name, sorter, remarks) VALUES (%s, %s, %d, %s);\n",
			sqlEscape(lg.Name), sqlEscape(lg.ShortName), lg.Sorter, sqlEscape(lg.Remarks))
	}
}

// Helpers

func (lh *LeagueHandler) getLeague(w http.ResponseWriter, r *http.Request) (store.League, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Ungültige ID", http.StatusBadRequest)
		return store.League{}, fmt.Errorf("empty id")
	}
	league, err := lh.s.GetLeague(id)
	if errors.Is(err, store.ErrNotFound) {
		http.Error(w, "Liga nicht gefunden", http.StatusNotFound)
		return store.League{}, err
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return store.League{}, err
	}
	return league, nil
}

func formValues(r *http.Request) map[string]string {
	vals := make(map[string]string)
	for k, v := range r.Form {
		if len(v) > 0 {
			vals[k] = v[0]
		}
	}
	return vals
}

func sqlEscape(val string) string {
	return "'" + strings.ReplaceAll(val, "'", "''") + "'"
}

const exportTimestampFormat = "2006-01-02_150405"

func exportFilename(base, ext string) string {
	ts := time.Now().Format(exportTimestampFormat)
	return fmt.Sprintf(`attachment; filename="%s_%s.%s"`, base, ts, ext)
}

func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
