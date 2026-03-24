# Migration Guide: Python → Go (Referee Dashboard)

## Flask → Chi

Chi ist ein leichtgewichtiger HTTP-Router. Routen, Middleware und Handler-Struktur sind ähnlich zu Flask-Blueprints. Der größte Unterschied: statt Dekoratoren registrierst du Routen explizit per `r.Get(...)`, `r.Post(...)` etc.

---

## htpy → Gomponents

Beide erzeugen HTML programmatisch aus Code heraus – das mentale Modell ist identisch. Gomponents nutzt Go-Funktionen statt Python-Callables. Komponenten sind einfache Go-Funktionen, die `gomponents.Node` zurückgeben.

---

## SQLAlchemy → sqlc + goose

Das ist die größte konzeptionelle Änderung:

- **Goose** übernimmt die Migrationen. Du schreibst plain SQL-Migrationsdateien (`001_initial.sql`, `002_add_column.sql` etc.), goose führt sie in Reihenfolge aus und trackt den Stand in der DB.
- **sqlc** liest deine SQL-Queries aus `.sql`-Dateien und generiert daraus typsicheren Go-Code. Du schreibst kein ORM-Mapping mehr – stattdessen plain SQL mit Annotationen, sqlc macht den Rest.

Workflow: Schema in Goose-Migration → Queries in `.sql`-Dateien → `sqlc generate` → typsichere Go-Funktionen nutzen.

---

## Plotly → Plotly.js

Kein Änderungsaufwand am Konzept. Plotly.js per CDN einbinden, Chart-Daten als JSON vom Go-Backend liefern, im Frontend rendern. Die JSON-Struktur ist dieselbe wie in Python.

---

## HTMX & Alpine

Identisch – keine Migration nötig.

---

## Empfohlene Reihenfolge

1. Goose-Schema (Migrationen)
2. sqlc-Queries
3. Chi-Handler
4. Gomponents-Templates
5. Plotly.js (parallel zu jedem Schritt integrierbar)

---

## Deployment auf mimir (OMV)

OMV hat keine UI für beliebige systemd-Services. Es gibt zwei sinnvolle Optionen:

### Option 1: systemd Service

Binary auf den Server kopieren und als systemd-Service registrieren. Verwaltung ausschließlich per CLI.

```ini
# /etc/systemd/system/referee-dashboard.service
[Unit]
Description=Referee Dashboard
After=network.target

[Service]
ExecStart=/opt/referee-dashboard/referee-dashboard
WorkingDirectory=/opt/referee-dashboard
Restart=on-failure
User=axel

[Install]
WantedBy=multi-user.target
```

```sh
systemctl enable --now referee-dashboard
```

SQLite-Datei liegt direkt im `WorkingDirectory`, kein Volume-Mapping nötig.

**Nachteil:** Kein OMV-UI, kein zentrales Log-Management mit den anderen Services.

---

### Option 2: Docker via OMV Compose Plugin (empfohlen)

Go-Binary in ein minimales Image packen. Mit `FROM scratch` ist das Image nur wenige MB groß.

```dockerfile
# Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o referee-dashboard .

FROM scratch
COPY --from=builder /build/referee-dashboard /referee-dashboard
CMD ["/referee-dashboard"]
```

```yaml
# compose.yaml
services:
  referee-dashboard:
    build: .
    image: referee-dashboard:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - DB_PATH=/data/referee.db
```

Die SQLite-Datei liegt im gemounteten `./data`-Verzeichnis und bleibt beim Neustart erhalten.

**Vorteil:** Start/Stop/Logs direkt in der OMV-UI, konsistent mit dem Rest der mimir-Services.

---

### Empfehlung

Docker via Compose Plugin, da du es auf mimir ohnehin bereits nutzt. Der Container-Overhead ist bei `FROM scratch` minimal, dafür gewinnst du zentrale Verwaltung in der OMV-UI.

### Python Projekt
Das Python Projekt liegt unter ~/python-projects/referee-dashboard
