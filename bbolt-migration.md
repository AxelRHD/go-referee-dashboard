# Migration: SQLite → bbolt

## Übersicht

Ersatz von SQLite + goose + sqlc durch bbolt als eingebettete Key-Value-Datenbank.
Dokumente werden als JSON gespeichert, IDs als ULIDs generiert.

## Neue Dependencies

```
go.etcd.io/bbolt          # Embedded KV-Store
github.com/oklog/ulid/v2  # Zeitbasiert sortierbare IDs
```

Entfallen:
```
modernc.org/sqlite        # SQLite-Treiber
github.com/pressly/goose  # Migrationen
```

sqlc wird nicht mehr benötigt — `sqlc.yaml`, `db/queries/`, `model/` (generiert) werden entfernt.

---

## Neue Projektstruktur

```
store/
  store.go        # DB öffnen/schließen, Bucket-Init, Backup
  types.go        # Domain-Structs (Game, League, Team, Venue, Position)
  leagues.go      # LeagueStore: List, Get, Put, Delete
  teams.go        # TeamStore: List, Get, Put, Delete
  venues.go       # VenueStore: List, Get, Put, Delete + UpdateCoords
  games.go        # GameStore: List, Get, Put, Delete, ListBySeason, ListSeasons
  positions.go    # PositionStore: List (read-only, seed-only)
  export.go       # JSON-Export/Import der gesamten DB
```

Ersetzt:
```
model/            # komplett entfernen (generierter sqlc-Code)
db/queries/       # komplett entfernen
db/migrations/    # komplett entfernen
db/embed.go       # komplett entfernen
sqlc.yaml         # komplett entfernen
```

---

## Domain-Typen (store/types.go)

### Stammdaten

```go
type League struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    ShortName string `json:"short_name"`
    Sorter    int    `json:"sorter"`
    Remarks   string `json:"remarks,omitempty"`
}

type Team struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    State    string `json:"state"`
    IsActive bool   `json:"is_active"`
    Remarks  string `json:"remarks,omitempty"`
}

type Venue struct {
    ID        string  `json:"id"`
    City      string  `json:"city"`
    ShortName string  `json:"short_name"`
    Stadium   string  `json:"stadium,omitempty"`
    Lat       float64 `json:"lat,omitempty"`
    Lon       float64 `json:"lon,omitempty"`
}

type Position struct {
    Position string `json:"position"`
    Long     string `json:"long"`
    Sorter   int    `json:"sorter"`
}
```

### Eingebettete Referenzen

```go
type Ref struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    ShortName string `json:"short_name,omitempty"`
}
```

### Spiel (denormalisiert)

```go
type Game struct {
    ID          string  `json:"id"`
    GameDate    string  `json:"game_date"`
    GameTime    string  `json:"game_time,omitempty"`
    HomeTeam    Ref     `json:"home_team"`
    AwayTeam    Ref     `json:"away_team"`
    Venue       Ref     `json:"venue"`
    League      Ref     `json:"league"`
    Position    string  `json:"position"`
    RefereeFee  float64 `json:"referee_fee"`
    TravelCosts float64 `json:"travel_costs"`
    KmDriven    int     `json:"km_driven"`
    Exhibition  bool    `json:"exhibition"`
    Remarks     string  `json:"remarks,omitempty"`
}
```

Kein `created_at`/`updated_at` — ULID enthält bereits den Erstellungszeitpunkt.
Falls gewünscht, kann `updated_at` optional ergänzt werden.

---

## Bucket-Struktur

```
leagues     → Key: ULID, Value: League JSON
teams       → Key: ULID, Value: Team JSON
venues      → Key: ULID, Value: Venue JSON
positions   → Key: Position-Kürzel ("R", "SL", ...), Value: Position JSON
games       → Key: ULID, Value: Game JSON (denormalisiert)
```

Positions nutzen kein ULID — der Position-String selbst ist der Key (wie bisher Primary Key).

---

## Store-Layer (store/store.go)

```go
type Store struct {
    db *bbolt.DB
}

func Open(path string) (*Store, error)   // Öffnet DB, erstellt Buckets
func (s *Store) Close() error
func (s *Store) Backup(w io.Writer) error // Konsistenter Online-Backup
```

### Generisches Pattern pro Entity

Jeder Entity-Store ist eine Methoden-Sammlung auf `*Store`:

```go
// Beispiel: Leagues
func (s *Store) ListLeagues() ([]League, error)
func (s *Store) GetLeague(id string) (League, error)
func (s *Store) PutLeague(l *League) error    // Create + Update (ULID wenn id leer)
func (s *Store) DeleteLeague(id string) error
```

`Put` erzeugt eine neue ULID wenn `l.ID` leer ist (Create), sonst überschreibt es (Update).
Ein einziger `*Store` ersetzt das bisherige `*model.Queries` — gleiche Injection-Stelle.

---

## Handler-Anpassungen

### Vor (SQLite/sqlc)
```go
type LeagueHandler struct {
    q *model.Queries
}

func (lh *LeagueHandler) Create(w http.ResponseWriter, r *http.Request) {
    _, err := lh.q.CreateLeague(r.Context(), model.CreateLeagueParams{
        Name: data.Name,
        ...
    })
}
```

### Nach (bbolt)
```go
type LeagueHandler struct {
    s *store.Store
}

func (lh *LeagueHandler) Create(w http.ResponseWriter, r *http.Request) {
    err := lh.s.PutLeague(&store.League{
        Name: data.Name,
        ...
    })
}
```

Änderungen pro Handler:
- `*model.Queries` → `*store.Store`
- `model.CreateXParams{...}` → `store.X{...}` direkt
- `r.Context()` entfällt (bbolt braucht keinen Context)
- `sql.ErrNoRows` → eigener `store.ErrNotFound`
- ID-Typ: `int64` → `string` (ULID)
- `chi.URLParam` liefert bereits String — kein `strconv.Atoi` mehr nötig

### GameHandler: Spezialfall Denormalisierung

Beim Erstellen/Bearbeiten eines Spiels müssen die Referenzen aufgelöst werden:

```go
func (gh *GameHandler) Create(w http.ResponseWriter, r *http.Request) {
    // Stammdaten für Einbettung holen
    homeTeam, _ := gh.s.GetTeam(r.FormValue("home_team_id"))
    awayTeam, _ := gh.s.GetTeam(r.FormValue("away_team_id"))
    venue, _ := gh.s.GetVenue(r.FormValue("venue_id"))
    league, _ := gh.s.GetLeague(r.FormValue("league_id"))

    err := gh.s.PutGame(&store.Game{
        HomeTeam: store.Ref{ID: homeTeam.ID, Name: homeTeam.Name},
        AwayTeam: store.Ref{ID: awayTeam.ID, Name: awayTeam.Name},
        Venue:    store.Ref{ID: venue.ID, Name: venue.City, ShortName: venue.ShortName},
        League:   store.Ref{ID: league.ID, Name: league.Name, ShortName: league.ShortName},
        ...
    })
}
```

### DashboardHandler

- `ListGamesFull` → `s.ListGames()` — Daten sind bereits denormalisiert
- `ListGamesBySeason` → `s.ListGamesBySeason(year)` — filtert per Iteration
- Kein JOIN nötig, jedes Game-Dokument enthält alles

### DataHandler

- SQL-Dump/Import entfällt komplett
- Ersetzt durch JSON-Export/Import (gesamte DB)
- `s.ExportJSON(w)` / `s.ImportJSON(r)`

### SetupHandler

- Seed-Logik liest weiterhin eingebettete SQL-Dateien ODER wird auf eingebettete JSON-Dateien umgestellt
- Prüfung "DB leer?" per Bucket-Check statt SQL-Query

---

## server/server.go Änderungen

### Vor
```go
conn, _ := sql.Open("sqlite", cfg.DBPath)
conn.Exec("PRAGMA journal_mode=WAL")
conn.Exec("PRAGMA foreign_keys=ON")
queries := model.New(conn)
```

### Nach
```go
s, _ := store.Open(cfg.DBPath)
defer s.Close()
// Kein Pragma, keine Migrationen
```

Alle Handler erhalten `s *store.Store` statt `queries *model.Queries`.

---

## CLI Änderungen (cli/cli.go)

- **`serve`**: `openDB` + `runMigrations` → `store.Open(path)` (Buckets werden automatisch erstellt)
- **`migrate`**: Komplett entfernen — nicht mehr nötig
- **`seed`**: Optional beibehalten, liest JSON-Seed-Dateien statt SQL
- **`health`**: Keine Änderung

Justfile-Recipes:
- `just migrate` → entfernen
- `just generate` → entfernen
- `just seed` → anpassen (JSON statt SQL)

---

## JSON-Export/Import

### Export (store/export.go)

```go
type DatabaseExport struct {
    ExportedAt string     `json:"exported_at"`
    Leagues    []League   `json:"leagues"`
    Teams      []Team     `json:"teams"`
    Venues     []Venue    `json:"venues"`
    Positions  []Position `json:"positions"`
    Games      []Game     `json:"games"`
}

func (s *Store) ExportJSON(w io.Writer) error  // Gesamte DB als JSON
func (s *Store) ImportJSON(r io.Reader) error   // Gesamte DB aus JSON ersetzen
```

- Immer die gesamte DB exportieren/importieren — keine Einzel-Buckets
- Menschenlesbar, in jedem Editor inspizierbar
- Ideal als Backup-Mechanismus auf der Data-Management-Seite

---

## Seed-Dateien

Seed-Daten von SQL auf eingebettete JSON-Dateien umstellen:

```
db/seed/
  positions.json
  leagues.json
  teams.json
  venues.json
```

Beim Setup-Wizard werden diese direkt per `Store.Put*` eingespielt.

---

## Datenmigration (einmalig)

Einmaliges Skript oder CLI-Command zum Übertragen bestehender SQLite-Daten:

1. Bestehende SQLite-DB öffnen
2. Alle Stammdaten lesen, ULIDs vergeben, in bbolt schreiben
3. ID-Mapping aufbauen (alte int64-ID → neue ULID)
4. Spiele lesen, Referenzen über Mapping auflösen, denormalisiert in bbolt schreiben
5. Ergebnis verifizieren

Kann als temporäres `just migrate-to-bbolt`-Recipe implementiert werden und nach erfolgreicher Migration entfernt werden.

---

## Phasenplan

### Phase 1: Store-Layer
1. `store/` Package anlegen mit `store.go`, `types.go`
2. CRUD für Leagues implementieren und testen
3. Pattern auf Teams, Venues, Positions, Games übertragen
4. `ExportJSON` / `ImportJSON` implementieren

### Phase 2: Handler umstellen
1. `server.go`: `store.Open()` statt `sql.Open()` + `model.New()`
2. Handler einzeln umstellen (Leagues → Teams → Venues → Games → Dashboard → Data → Setup)
3. Pro Handler: Typ ändern, Methoden anpassen, testen

### Phase 3: Seed & Setup
1. Seed-Dateien auf JSON umstellen
2. Setup-Wizard anpassen (Bucket-Check statt SQL)
3. Data-Management-Seite: SQL-Export/Import → JSON-Export/Import

### Phase 4: Aufräumen
1. `model/` komplett entfernen
2. `db/queries/`, `db/migrations/`, `db/embed.go` entfernen
3. `sqlc.yaml` entfernen
4. goose + modernc/sqlite aus `go.mod` entfernen
5. Justfile bereinigen (`migrate`, `generate` entfernen)
6. CLAUDE.md aktualisieren (neue Projektstruktur, Workflow)

### Phase 5: Datenmigration (optional)
1. Migrationsskript SQLite → bbolt schreiben
2. Bestehende Daten übertragen
3. Skript nach Verifizierung entfernen

---

## Verifizierung

- [ ] Alle CRUD-Operationen für Leagues, Teams, Venues, Games funktionieren
- [ ] Dashboard zeigt Daten korrekt an (Übersicht + Jahresansicht)
- [ ] Seed-Wizard funktioniert mit JSON-Dateien
- [ ] JSON-Export erzeugt lesbare, vollständige Datei
- [ ] JSON-Import stellt DB korrekt wieder her
- [ ] Geocoding (Venue-Koordinaten) funktioniert
- [ ] Cache-Refresh auf Data-Management-Seite funktioniert
- [ ] Docker-Build + Deployment funktioniert (Volume für .db-Datei)
