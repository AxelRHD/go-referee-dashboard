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

### Phase 7 — Deployment ✅
- Dockerfile + compose.yaml finalisieren
- Static Assets (nord.css, pfeife.png) übernehmen
- Deployment auf mimir via Docker Compose

---

## Nächste Schritte

### ECharts Map Migration (Maps only, Plotly bleibt für Rest)
Plotly `scattermapbox` → ECharts `scatter` auf `geo` für bessere Maps ohne Mapbox.

**Dateien:**
- `view/layout.go` — ECharts CDN Script hinzufügen (nach Plotly, koexistieren)
- `view/dashboard.go` — `renderMap()` umschreiben + GeoJSON fetch in `init()`

**Ansatz:** Beide Libraries parallel laden, Toggle-Switch im Dashboard (Alpine.js State `chartLib: 'plotly'|'echarts'`, localStorage). Jede Render-Funktion hat ein Plotly- und ECharts-Pendant. So kann man direkt vergleichen und schrittweise migrieren.

**Schritte:**
1. ECharts v5 via CDN laden: `https://cdn.jsdelivr.net/npm/echarts@5/dist/echarts.min.js`
2. Deutschland GeoJSON in `init()` registrieren: `echarts.registerMap('Germany', geoJson)`
3. `renderMap()` duplizieren als `renderMapECharts()`:
   - `echarts.init()` / `getInstanceByDom()` für Reuse
   - `geo: { map: 'Germany', roam: true }` für Zoom/Pan
   - `series: { type: 'scatter', coordinateSystem: 'geo' }` für Bubbles
   - **Koordinaten: ECharts nutzt [lon, lat], nicht [lat, lon]!**
   - Tooltip: `venue name (count Spiele)`
4. Theme-Wechsel: Chart `dispose()` + neu erstellen bei `theme-changed`
5. Nord-Farben:
   - Dark: `areaColor: '#3B4252'`, `borderColor: '#4C566A'`
   - Light: `areaColor: '#E5E9F0'`, `borderColor: '#D8DEE9'`
   - Scatter: `color: '#5E81AC'`, `opacity: 0.7`

6. Toggle-Switch im Dashboard-Sidebar:
   - Alpine State: `chartLib: localStorage.getItem('db_chartLib') || 'plotly'`
   - Button-Group wie Jahr/Übersicht: `Plotly | ECharts`
   - In jeder Render-Funktion: `if (this.chartLib === 'echarts') renderXxxECharts() else renderXxxPlotly()`
   - Start mit Map, dann schrittweise weitere Charts portieren

**Verifizierung:**
- Toggle zwischen Plotly und ECharts
- Map zeigt Deutschland mit Scatter-Bubbles in beiden Varianten
- Zoom/Pan funktioniert bei ECharts
- Hover zeigt Venue + Anzahl
- Theme-Toggle aktualisiert Farben
- localStorage merkt sich die Auswahl

### Spielorte: Anzeigenamen
Aktuell zeigt die Map den vollen Venue-String ("City, Stadium") als Label. Überlegung: einen separaten `display_name` oder kürzeren Anzeigenamen für die Karte und Tooltips definieren. Z.B. nur Stadt anzeigen wenn eindeutig, oder Stadion-Kürzel. Muss noch entschieden werden.

### Gridstack.js (optional, v0.2+)
Dashboard-Widgets per Drag & Drop anordnen, Layout in localStorage speichern.
