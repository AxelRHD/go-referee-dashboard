package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/store"
	"github.com/axelrhd/referee-dashboard/view"
)

type DataHandler struct {
	s *store.Store
}

func NewDataHandler(s *store.Store) *DataHandler {
	return &DataHandler{s: s}
}

func (h *DataHandler) Routes(r chi.Router) {
	r.Get("/data", h.Page)
	r.Get("/data/export", h.Export)
	r.Post("/data/import", h.Import)
	r.Post("/data/positions/new", h.CreatePosition)
	r.Post("/data/positions/{key}/edit", h.UpdatePosition)
	r.Post("/data/positions/{key}/delete", h.DeletePosition)
}

func (dh *DataHandler) Page(w http.ResponseWriter, r *http.Request) {
	positions, _ := dh.s.ListPositions()
	games, _ := dh.s.ListGames()
	seasons, _ := dh.s.ListSeasons()
	page := view.DataPage(w, r, positions, len(games) > 0, seasons, nil)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (dh *DataHandler) Export(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", exportFilename("referee_data", "json"))

	if err := dh.s.ExportJSON(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (dh *DataHandler) Import(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		view.SetFlash(w, "Keine Datei ausgewählt.")
		http.Redirect(w, r, "/data", http.StatusSeeOther)
		return
	}
	defer file.Close()

	// Prüfe ob Daten vorhanden und Überschreiben bestätigt
	games, _ := dh.s.ListGames()
	if len(games) > 0 && r.FormValue("overwrite") != "1" {
		view.SetFlash(w, "Import abgebrochen: Es sind bereits Daten vorhanden. Bitte bestätigen Sie das Überschreiben.")
		http.Redirect(w, r, "/data", http.StatusSeeOther)
		return
	}

	if err := dh.s.ImportJSON(file); err != nil {
		positions, _ := dh.s.ListPositions()
		seasons, _ := dh.s.ListSeasons()
		page := view.DataPage(w, r, positions, len(games) > 0, seasons, []string{fmt.Sprintf("Import-Fehler: %v", err)})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	view.SetFlash(w, "Daten erfolgreich importiert.")
	http.Redirect(w, r, "/data", http.StatusSeeOther)
}

func (dh *DataHandler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	pos := strings.TrimSpace(r.FormValue("position"))
	long := strings.TrimSpace(r.FormValue("long"))
	sorter, _ := strconv.Atoi(r.FormValue("sorter"))

	if pos == "" {
		view.SetFlash(w, "Position-Kürzel ist ein Pflichtfeld.")
		http.Redirect(w, r, "/data", http.StatusSeeOther)
		return
	}

	if err := dh.s.PutPosition(&store.Position{Position: pos, Long: long, Sorter: sorter}); err != nil {
		view.SetFlash(w, fmt.Sprintf("Fehler: %v", err))
	} else {
		view.SetFlash(w, fmt.Sprintf("Position %s wurde erstellt.", pos))
	}
	http.Redirect(w, r, "/data", http.StatusSeeOther)
}

func (dh *DataHandler) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	r.ParseForm()
	long := strings.TrimSpace(r.FormValue("long"))
	sorter, _ := strconv.Atoi(r.FormValue("sorter"))

	if err := dh.s.PutPosition(&store.Position{Position: key, Long: long, Sorter: sorter}); err != nil {
		view.SetFlash(w, fmt.Sprintf("Fehler: %v", err))
	} else {
		view.SetFlash(w, fmt.Sprintf("Position %s wurde aktualisiert.", key))
	}
	http.Redirect(w, r, "/data", http.StatusSeeOther)
}

func (dh *DataHandler) DeletePosition(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if err := dh.s.DeletePosition(key); err != nil {
		view.SetFlash(w, fmt.Sprintf("Fehler: %v", err))
	} else {
		view.SetFlash(w, fmt.Sprintf("Position %s wurde gelöscht.", key))
	}
	http.Redirect(w, r, "/data", http.StatusSeeOther)
}

