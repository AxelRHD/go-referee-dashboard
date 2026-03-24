package handler

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/axelrhd/referee-dashboard/model"
	"github.com/axelrhd/referee-dashboard/view"
)

type DataHandler struct {
	q  *model.Queries
	db *sql.DB
}

func NewDataHandler(q *model.Queries, db *sql.DB) *DataHandler {
	return &DataHandler{q: q, db: db}
}

func (h *DataHandler) Routes(r chi.Router) {
	r.Get("/data", h.Page)
	r.Get("/data/dump", h.Dump)
	r.Get("/data/export-all", h.ExportAll)
	r.Post("/data/import", h.Import)
	r.Post("/data/paste", h.Paste)
}

func (dh *DataHandler) Page(w http.ResponseWriter, r *http.Request) {
	page := view.DataPage(w, r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

func (dh *DataHandler) Dump(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", exportFilename("referee_dump", "sql"))

	tableOrder := []string{"positions", "leagues", "teams", "venues", "games"}
	fmt.Fprintf(w, "-- Referee Dashboard Dump %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(w, "BEGIN TRANSACTION;")

	for _, table := range tableOrder {
		var sqlStmt string
		err := dh.db.QueryRow(
			"SELECT sql FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&sqlStmt)
		if err == nil && sqlStmt != "" {
			sqlStmt = strings.Replace(sqlStmt, "CREATE TABLE", "CREATE TABLE IF NOT EXISTS", 1)
			fmt.Fprintf(w, "%s;\n", sqlStmt)
		}
	}

	for _, table := range tableOrder {
		rows, err := dh.db.Query(fmt.Sprintf("SELECT * FROM [%s]", table))
		if err != nil {
			continue
		}
		cols, _ := rows.Columns()
		colStr := strings.Join(cols, ", ")
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		for rows.Next() {
			rows.Scan(ptrs...)
			var escaped []string
			for _, v := range vals {
				escaped = append(escaped, sqlEscapeAny(v))
			}
			fmt.Fprintf(w, "INSERT OR IGNORE INTO %s (%s) VALUES (%s);\n",
				table, colStr, strings.Join(escaped, ", "))
		}
		rows.Close()
	}

	fmt.Fprintln(w, "COMMIT;")
}

func (dh *DataHandler) ExportAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", exportFilename("referee_data", "sql"))

	tableOrder := []string{"positions", "leagues", "teams", "venues", "games"}
	fmt.Fprintf(w, "-- Referee Dashboard Data Export %s\n", time.Now().Format("2006-01-02 15:04:05"))

	for _, table := range tableOrder {
		rows, err := dh.db.Query(fmt.Sprintf("SELECT * FROM [%s]", table))
		if err != nil {
			continue
		}
		cols, _ := rows.Columns()
		colStr := strings.Join(cols, ", ")
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		for rows.Next() {
			rows.Scan(ptrs...)
			var escaped []string
			for _, v := range vals {
				escaped = append(escaped, sqlEscapeAny(v))
			}
			fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (%s);\n",
				table, colStr, strings.Join(escaped, ", "))
		}
		rows.Close()
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

	data, _ := io.ReadAll(file)
	text := strings.TrimPrefix(string(data), "\xEF\xBB\xBF") // strip BOM
	format := r.FormValue("format")
	entity := r.FormValue("entity")

	dh.processImport(w, r, text, format, entity)
}

func (dh *DataHandler) Paste(w http.ResponseWriter, r *http.Request) {
	text := strings.TrimSpace(r.FormValue("content"))
	if text == "" {
		view.SetFlash(w, "Keine Daten eingegeben.")
		http.Redirect(w, r, "/data", http.StatusSeeOther)
		return
	}

	format := r.FormValue("format")
	entity := r.FormValue("entity")

	dh.processImport(w, r, text, format, entity)
}

func (dh *DataHandler) processImport(w http.ResponseWriter, r *http.Request, text, format, entity string) {
	var messages []string

	if format == "sql" {
		statements, parseErrors := sanitizeSQL(text)
		messages = append(messages, parseErrors...)
		if len(statements) > 0 {
			count, execErrors := dh.executeSQL(statements)
			messages = append(messages, execErrors...)
			if count > 0 {
				messages = append(messages, fmt.Sprintf("%d SQL-Statement(s) erfolgreich ausgeführt.", count))
			}
		} else if len(parseErrors) == 0 {
			messages = append(messages, "Keine gültigen SQL-Statements gefunden.")
		}
	} else {
		count, csvErrors := dh.importCSV(r.Context(), text, entity)
		messages = append(messages, csvErrors...)
		if count > 0 {
			messages = append(messages, fmt.Sprintf("%d Datensatz/Datensätze importiert.", count))
		} else if len(csvErrors) == 0 {
			messages = append(messages, "Keine Daten zum Importieren gefunden.")
		}
	}

	page := view.DataPageWithMessages(w, r, messages)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.Render(w)
}

// SQL sanitization

var (
	allowedPrefixes = regexp.MustCompile(`(?i)^\s*(INSERT|CREATE\s+TABLE)`)
	blockedKeywords = regexp.MustCompile(`(?i)^\s*(DROP|DELETE|UPDATE|ALTER|TRUNCATE|REPLACE|ATTACH|DETACH|PRAGMA|VACUUM)`)
	skipRe          = regexp.MustCompile(`(?i)^\s*(BEGIN|COMMIT|ROLLBACK)`)
)

func sanitizeSQL(text string) ([]string, []string) {
	var statements, errors []string

	for _, stmt := range strings.Split(text, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") || skipRe.MatchString(stmt) {
			continue
		}
		if blockedKeywords.MatchString(stmt) {
			keyword := strings.ToUpper(strings.Fields(stmt)[0])
			errors = append(errors, fmt.Sprintf("%s-Statements sind nicht erlaubt: %s...", keyword, truncate(stmt, 80)))
		} else if allowedPrefixes.MatchString(stmt) {
			statements = append(statements, stmt)
		} else {
			errors = append(errors, fmt.Sprintf("Statement nicht erlaubt: %s...", truncate(stmt, 80)))
		}
	}
	return statements, errors
}

func (dh *DataHandler) executeSQL(statements []string) (int, []string) {
	var errors []string
	count := 0

	// Disable FK checks for import (dump order may vary)
	dh.db.Exec("PRAGMA foreign_keys=OFF")
	defer dh.db.Exec("PRAGMA foreign_keys=ON")

	tx, err := dh.db.Begin()
	if err != nil {
		return 0, []string{fmt.Sprintf("Fehler beim Starten der Transaktion: %v", err)}
	}

	for _, stmt := range statements {
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

// Helpers

func sqlEscapeAny(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	switch v := val.(type) {
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case []byte:
		return "'" + strings.ReplaceAll(string(v), "'", "''") + "'"
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	default:
		return "'" + strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''") + "'"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
