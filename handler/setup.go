package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/store"
	"github.com/axelrhd/referee-dashboard/store/seed"
	"github.com/axelrhd/referee-dashboard/view"
)

type SetupHandler struct {
	s *store.Store
}

func NewSetupHandler(s *store.Store) *SetupHandler {
	return &SetupHandler{s: s}
}

func (h *SetupHandler) Routes(r chi.Router) {
	r.Get("/setup", h.Page)
	r.Post("/setup/seed", h.Seed)
	r.Post("/setup/restore", h.Restore)
	r.Post("/setup/skip", h.Skip)
}

// NeedsSetup checks if the database has any positions (basic setup indicator)
func (h *SetupHandler) NeedsSetup() bool {
	positions, err := h.s.ListPositions()
	return err != nil || len(positions) == 0
}

func (sh *SetupHandler) Page(w http.ResponseWriter, r *http.Request) {
	if !sh.NeedsSetup() {
		http.Redirect(w, r, "/dashboard/", http.StatusFound)
		return
	}
	page := view.SetupPage(nil, false)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (sh *SetupHandler) Seed(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var messages []string

	seedFiles := []struct {
		param string
		file  string
		label string
	}{
		{"positions", "positions.json", "Positionen"},
		{"leagues", "leagues.json", "Ligen"},
		{"teams", "teams.json", "Teams"},
		{"venues", "venues.json", "Spielorte"},
	}

	for _, sf := range seedFiles {
		if r.FormValue(sf.param) != "1" {
			continue
		}
		data, err := seed.Files.ReadFile(sf.file)
		if err != nil {
			messages = append(messages, fmt.Sprintf("Fehler beim Lesen von %s: %v", sf.file, err))
			continue
		}
		count, errs := sh.seedFromJSON(sf.param, data)
		messages = append(messages, errs...)
		if count > 0 {
			messages = append(messages, fmt.Sprintf("%s: %d Einträge importiert.", sf.label, count))
		}
	}

	if len(messages) == 0 {
		messages = append(messages, "Keine Daten ausgewählt.")
	}

	page := view.SetupPage(messages, !sh.NeedsSetup())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (sh *SetupHandler) Restore(w http.ResponseWriter, r *http.Request) {
	var reader io.Reader

	// Try file upload first
	file, _, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		reader = file
	}

	// Fall back to paste content
	if reader == nil {
		text := strings.TrimSpace(r.FormValue("content"))
		if text == "" {
			page := view.SetupPage([]string{"Keine Daten eingegeben."}, false)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			page.Render(w)
			return
		}
		reader = strings.NewReader(text)
	}

	if err := sh.s.ImportJSON(reader); err != nil {
		page := view.SetupPage([]string{fmt.Sprintf("Import-Fehler: %v", err)}, false)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	page := view.SetupPage([]string{"Daten erfolgreich wiederhergestellt."}, !sh.NeedsSetup())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (sh *SetupHandler) Skip(w http.ResponseWriter, r *http.Request) {
	// Insert a placeholder position so NeedsSetup() returns false
	sh.s.PutPosition(&store.Position{Position: "_SKIP", Long: "Setup übersprungen", Sorter: 0})
	http.Redirect(w, r, "/dashboard/", http.StatusSeeOther)
}

func (sh *SetupHandler) seedFromJSON(entity string, data []byte) (int, []string) {
	count, err := sh.s.SeedBatch(entity, data)
	if err != nil {
		return 0, []string{fmt.Sprintf("Fehler: %v", err)}
	}
	return count, nil
}
