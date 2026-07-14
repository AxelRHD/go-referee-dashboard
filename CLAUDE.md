# Referee Dashboard (Go)

Schiedsrichter-Dashboard für American Football. Ursprünglich Python/Flask, seit v0.1 in Go.

## Tech-Stack

- **Router:** chi v5
- **HTML:** gomponents (maragu.dev/gomponents)
- **DB:** bbolt (embedded Key-Value Store, JSON-Dokumente)
- **IDs:** ULIDs (zeitbasiert sortierbar, via oklog/ulid)
- **Config:** envconfig + godotenv (.env)
- **CLI:** urfave/cli v3
- **Frontend:** Bootstrap 5 + Nord-Theme, HTMX, Alpine.js, ECharts (CDN)
- **Build:** justfile

## Projektstruktur

```
cmd/main.go          Entry Point (Config laden, CLI aufrufen)
cli/cli.go           urfave/cli v3 Commands (serve, health)
server/server.go     chi-Router, Middleware, Route-Registrierung
config/config.go     envconfig Struct + godotenv
store/               bbolt-basierter Datenzugriff
  store.go           DB öffnen/schließen, Bucket-Init, Backup
  types.go           Domain-Structs + eingebettete Ref-Typen
  leagues.go         LeagueStore: List, Get, Put, Delete
  teams.go           TeamStore: List, Get, Put, Delete
  venues.go          VenueStore: List, Get, Put, Delete, UpdateCoords
  games.go           GameStore: List, Get, Put, Delete, ListBySeason, ListSeasons
  positions.go       PositionStore: List, Put
  export.go          JSON-Export/Import der gesamten DB
  seed/              Eingebettete JSON-Seed-Dateien
handler/             HTTP-Handler pro Modul
view/                Gomponents-Views pro Modul
validation/          Form-Validierung
static/              CSS, Icons
```

## Commands

Task-Runner ist **mise** (Tasks als Skripte in `.mise-tasks/`, `mise tasks` listet alle).

```sh
mise run dev       # Dev-Server starten (air, Hot-Reload)
mise run build     # Binary bauen (mit Git-Version)
mise run fmt       # Code formatieren
mise run vet       # Static Analysis
mise run test      # Tests ausführen
mise run deploy    # Full deploy (build:image + deploy:image)
```

Gruppierte Tasks über Unterordner in `.mise-tasks/`: `build:image`, `deploy:image`,
`deploy:logs`, `deploy:status`. Der bloße `build`/`deploy`-Task liegt als
`build.sh`/`deploy.sh` (mise strippt die `.sh`-Endung), damit er neben dem
gleichnamigen Gruppen-Ordner existieren kann. Toolchain (Go) wird in `mise.toml`
unter `[tools]` gepinnt.

## Datenmodell

### Stammdaten (eigene Buckets)
- **leagues** — Liga mit Name, ShortName, Sorter
- **teams** — Team mit Name, State, IsActive
- **venues** — Spielort mit City, ShortName, Stadium, Lat/Lon
- **positions** — Position mit Kürzel, Langname, Sorter (Key = Position-String)

### Spiele (denormalisiert)
Spiel-Dokumente enthalten eingebettete Referenzen auf Stammdaten:
- `TeamRef` — ID, Name
- `VenueRef` — ID, City, ShortName, Lat, Lon
- `LeagueRef` — ID, Name, ShortName

Beim Erstellen/Bearbeiten eines Spiels werden die Refs aus den Stammdaten aufgelöst und eingebettet. Kein Join nötig beim Lesen.

### IDs
- Stammdaten + Spiele: ULIDs (string)
- Positions: Position-String als Key (kein ULID)

## Workflow: Schema-Änderungen

Keine Migrationen nötig. Struct anpassen → fertig. Unbekannte JSON-Felder werden beim Lesen ignoriert.

## Export/Import

- JSON-Export/Import der gesamten DB (alle Buckets auf einmal)
- Immer die gesamte DB exportieren/importieren, nie einzelne Buckets
- Backup = Dateikopie der `.db`-Datei oder `store.Backup(w)`

## API-Konventionen

- JSON API unter `/api/...`
- HTML-Seiten auf Root-Level (`/leagues`, `/games`, etc.)
- HTMX-Partials wo sinnvoll

## Python-Referenz

Originalprojekt: `~/python-projects/referee-dashboard/src/referee_dashboard/`

## Dokumentation (Antora)

`docs/` ist eine **Antora-Doku-Komponente** (`docs/antora.yml` + `docs/modules/ROOT/`),
die NICHT in diesem Repo gebaut wird. Sie wird von der zentralen Übersichts-Site
**„Hobby Projekte"** aggregiert und dort gebaut:
`~/antora-documentation/hobby-projects` (Playbook + `mise run build` / `mise run dev`).

- Antora publiziert nur Dateien unter `docs/modules/ROOT/pages/`. Screenshots liegen
  in `docs/modules/ROOT/images/` — dieselben Bilder referenziert auch das README.
- Änderungen an der Projektdoku laufen üblicherweise über die Session im
  hobby-projects-Repo (dort liegen die Antora-Konventionen). Hier nichts „bauen".
