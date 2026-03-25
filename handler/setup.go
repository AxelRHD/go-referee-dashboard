package handler

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/db/seed"
	"github.com/axelrhd/referee-dashboard/view"
)

type SetupHandler struct {
	db *sql.DB
}

func NewSetupHandler(db *sql.DB) *SetupHandler {
	return &SetupHandler{db: db}
}

func (h *SetupHandler) Routes(r chi.Router) {
	r.Get("/setup", h.Page)
	r.Post("/setup/seed", h.Seed)
	r.Post("/setup/restore", h.Restore)
	r.Post("/setup/skip", h.Skip)
}

// NeedsSetup checks if the database has any positions (basic setup indicator)
func (h *SetupHandler) NeedsSetup() bool {
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM positions").Scan(&count)
	return err != nil || count == 0
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
		{"positions", "positions.sql", "Positionen"},
		{"leagues", "leagues.sql", "Ligen"},
		{"teams", "teams.sql", "Teams"},
		{"venues", "venues.sql", "Spielorte"},
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
		count, errs := sh.executeSQLStatements(string(data))
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
	var text string

	// Try file upload first
	file, _, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		data, _ := io.ReadAll(file)
		text = strings.TrimPrefix(string(data), "\xEF\xBB\xBF")
	}

	// Fall back to paste content
	if text == "" {
		text = strings.TrimSpace(r.FormValue("content"))
	}

	if text == "" {
		page := view.SetupPage([]string{"Keine Daten eingegeben."}, false)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page.Render(w)
		return
	}

	// Disable FK checks for restore
	sh.db.Exec("PRAGMA foreign_keys=OFF")
	defer sh.db.Exec("PRAGMA foreign_keys=ON")

	statements, parseErrors := sanitizeSQL(text)
	var messages []string
	messages = append(messages, parseErrors...)

	if len(statements) > 0 {
		count, execErrors := sh.executeSQLStatements(strings.Join(statements, ";\n"))
		messages = append(messages, execErrors...)
		if count > 0 {
			messages = append(messages, fmt.Sprintf("%d SQL-Statement(s) erfolgreich ausgeführt.", count))
		}
	} else if len(parseErrors) == 0 {
		messages = append(messages, "Keine gültigen SQL-Statements gefunden.")
	}

	page := view.SetupPage(messages, !sh.NeedsSetup())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (sh *SetupHandler) Skip(w http.ResponseWriter, r *http.Request) {
	// Insert a placeholder position so NeedsSetup() returns false
	sh.db.Exec("INSERT OR IGNORE INTO positions (position, long, sorter) VALUES ('_SKIP', 'Setup übersprungen', 0)")
	http.Redirect(w, r, "/dashboard/", http.StatusSeeOther)
}

func (sh *SetupHandler) executeSQLStatements(text string) (int, []string) {
	var errors []string
	count := 0

	tx, err := sh.db.Begin()
	if err != nil {
		return 0, []string{fmt.Sprintf("Fehler: %v", err)}
	}

	for _, stmt := range strings.Split(text, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := tx.Exec(stmt); err != nil {
			errors = append(errors, fmt.Sprintf("Fehler: %v — %s...", err, truncate(stmt, 80)))
		} else {
			count++
		}
	}

	if err := tx.Commit(); err != nil {
		errors = append(errors, fmt.Sprintf("Fehler beim Commit: %v", err))
	}

	return count, errors
}
