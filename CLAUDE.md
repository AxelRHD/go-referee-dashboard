# Referee Dashboard (Go)

Migration eines Schiedsrichter-Dashboards von Python/Flask nach Go.

## Tech-Stack

- **Router:** chi v5
- **HTML:** gomponents (maragu.dev/gomponents)
- **DB:** SQLite via modernc.org/sqlite (pure Go, kein CGO)
- **Schema:** goose (Migrationen) + sqlc (typsicherer Query-Code)
- **Config:** envconfig + godotenv (.env)
- **CLI:** urfave/cli v3
- **Frontend:** Bootstrap 5 + Nord-Theme, HTMX, Alpine.js, Plotly.js (CDN)
- **Build:** justfile

## Projektstruktur

```
cmd/main.go          Entry Point (Config laden, CLI aufrufen)
cli/cli.go           urfave/cli v3 Commands (serve, migrate, seed)
server/server.go     chi-Router, Middleware, Route-Registrierung
config/config.go     envconfig Struct + godotenv
handler/             HTTP-Handler pro Modul
view/                Gomponents-Views pro Modul
model/               sqlc-generierter Code (nicht manuell editieren!)
validation/          Form-Validierung
db/migrations/       Goose SQL-Migrationsdateien
db/queries/          sqlc SQL-Query-Dateien
static/              CSS, Icons
```

## Commands

```sh
just dev          # Dev-Server starten
just build        # Binary bauen (mit Git-Version)
just migrate      # Goose-Migrationen ausführen
just seed         # Positionen seeden
just generate     # sqlc Code generieren
just fmt          # Code formatieren
just vet          # Static Analysis
just test         # Tests ausführen
```

## Workflow: DB-Änderungen

1. Neue Migration in `db/migrations/` anlegen
2. Queries in `db/queries/` schreiben/anpassen
3. `just generate` → model/ wird neu generiert
4. `just migrate` → Schema anwenden

## API-Konventionen

- JSON API unter `/api/...`
- HTML-Seiten auf Root-Level (`/leagues`, `/games`, etc.)
- HTMX-Partials wo sinnvoll

## Python-Referenz

Originalprojekt: `~/python-projects/referee-dashboard/src/referee_dashboard/`

## Migrationsplan

### Phase 0 — Projektgerüst ✅
- Projektstruktur, Dependencies, Config, CLI, chi-Server, Goose-Migration, sqlc, Justfile, Docker

### Phase 1 — Leagues (CRUD-Prototyp)
Etabliert das Pattern für alle weiteren Module:
- sqlc-Queries (List, GetByID, Create, Update, Delete)
- JSON API Endpoints (`/api/leagues/...`)
- HTML-Handler (Liste, Formular)
- Gomponents: Layout (base_page, Navbar) + wiederverwendbare Components
- Validation
- CSV/SQL-Export

### Phase 2 — Teams
- Wie Leagues, plus Bundesland-Datalist und is_active Toggle

### Phase 3 — Venues
- Wie Teams, plus Geocoding (Nominatim API)

### Phase 4 — Games
- Komplexestes Modell: Fremdschlüssel, Filter (Season, Monat, Liga, Position), Pagination, Stats
- HTMX-Partials für Tabelle + Filter

### Phase 5 — Dashboard
- API-Endpoints für Aggregationen (`/api/dashboard/overview`, `/api/dashboard/season/:season`)
- Dashboard-View mit Alpine.js + Plotly.js Charts
- Filter (Season, Liga, Position), localStorage-Persistenz

### Phase 6 — Datenmanagement
- Import/Export-Seite
- SQL-Dump, INSERT-Export, CSV/SQL-Import mit Sanitization

### Phase 7 — Deployment
- Dockerfile + compose.yaml finalisieren
- Static Assets (nord.css, pfeife.png) übernehmen
- Deployment auf mimir via Docker Compose
