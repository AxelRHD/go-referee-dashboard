# bbolt als SQLite-Ersatz im Referee Dashboard

## Ausgangslage

Das Referee Dashboard nutzt aktuell SQLite mit goose für Migrationen. Das Schema ist noch in aktiver Entwicklung, was regelmäßige Migrationen erfordert – lästig in der Entwicklungsphase und beim Deployment.

## Warum bbolt?

- **Keine Migrationen**: Go-Structs werden als JSON serialisiert. Schemaänderungen bedeuten einfach die Struct anpassen – unbekannte Felder werden beim Lesen ignoriert.
- **Embedded**: Wie SQLite eine einzelne Datei, kein separater Prozess, kein extra Container.
- **Go-nativ**: Erstklassige Go-Integration, passt zur bestehenden Architektur.
- **Performance kein Problem**: Das Dashboard wird nie mehr als ~1000 Einträge haben – Iteration über alle Dokumente erfolgt in Mikrosekunden.

## Warum kein FerretDB?

FerretDB ist **nicht embedded** – es ist ein separater Server-Prozess der MongoDB-Wire-Protocol spricht. Kann nicht als Go-Library eingebettet werden. Würde einen extra Container erfordern.

## Datenfluss (aktuell)

- Go-Backend liefert Daten als JSON via HTMX
- Filterung nach Jahr erfolgt im Backend
- Aggregation (Summen, Durchschnitte für Charts) erfolgt im Frontend via Plotly.js / Alpine.js
- Ansonsten werden alle Daten geholt

## Konsequenz für bbolt

Da die meisten Abfragen ohnehin "gib mir alle Spiele" oder "gib mir alle Spiele für Jahr X" sind, verliert man durch den Wechsel zu bbolt **keine aktiv genutzte SQL-Funktionalität**. Die komplexen Aggregationen passieren eh im Frontend.

## Tradeoff

| | SQLite + goose | bbolt |
|---|---|---|
| Migrationen | Ja, bei jeder Schemaänderung | Keine |
| Query-Sprache | Vollständiges SQL | Nur Key-Value + Iteration |
| Aggregation | SQL möglich, aber nicht genutzt | Im Frontend, wie bisher |
| Embedded | Ja | Ja |
| Container-Größe | Kein Unterschied | Kein Unterschied |
| Performance | Kein Problem | Kein Problem |

## Datenmodell: Referenzen und Denormalisierung

### Bucket-Struktur
Separate Buckets für Stammdaten (CRUD-Verwaltung) und Spiele:
- `spielorte` – Bucket für Spielort-Stammdaten
- `teams` – Bucket für Team-Stammdaten
- `positionen` – Bucket für Positionen
- `spiele` – Hauptbucket

### Denormalisiert mit ID einbetten
Beim Speichern eines Spiels werden die Referenzen **vollständig eingebettet** – also sowohl ID als auch Name:

```json
{
  "id": "spiel-123",
  "datum": "2026-03-29",
  "spielort": { "id": "so-1", "name": "Sportpark Heddesheim" },
  "heimteam": { "id": "t-1", "name": "Heddesheim Sharks" },
  "gastteam": { "id": "t-2", "name": "Mannheim Bulls" },
  "position": { "id": "p-1", "name": "R" },
  "vergütung": 72.00,
  "kilometer": 25
}
```

### IDs: ULID statt numerisch

Statt numerischer IDs werden **ULIDs** verwendet:
- Zeitbasiert sortierbar – neue Einträge landen automatisch am Ende beim Iterieren
- Eindeutig und kollisionssicher
- Lesbarer als UUID

Stammdaten erhalten zusätzlich ein `short_name` Feld als menschenlesbares Kürzel:

```json
{
  "id": "01HX3K7P2QABCDEF",
  "short_name": "SPH",
  "name": "Sportpark Heddesheim"
}
```

Eingebettet im Spiel-Dokument:

```json
{
  "id": "01HX3K7P2QXYZ123",
  "spielort": { 
    "id": "01HX3K7P2QABCDEF",
    "short_name": "SPH",
    "name": "Sportpark Heddesheim" 
  },
  "heimteam": { 
    "id": "01HX3K7P2QHHSXXX",
    "short_name": "HHS",
    "name": "Heddesheim Sharks" 
  }
}
```

- `id` – technischer Schlüssel, stabil, wird nie geändert
- `short_name` – lesbares Kürzel für Debugging und Anzeige
- `name` – vollständiger Name

### Workflow CRUD
1. Stammdaten-Bucket (z.B. `spielorte`) füllt Dropdown beim Erfassen/Bearbeiten
2. User wählt Eintrag aus Dropdown
3. Beim Speichern wird `{id, name}` ins Spiel-Dokument eingebettet
4. Beim Anzeigen kein weiterer Lookup nötig – alle Daten direkt im Dokument

### Vorteile dieses Ansatzes
- **Kein Join nötig** beim Lesen – ein Dokument enthält alles
- **Dropdown-Vorauswahl** beim Bearbeiten möglich über die eingebettete ID
- **Filterung** nach Spielort/Team durch Iteration über `spielort.id` möglich
- **Historische Korrektheit** – Umbenennung eines Teams betrifft nicht alte Spiele
- Optional: Bei Namensänderung eines Stammdatums können alle referenzierenden Spiele mitaktualisiert werden (bei ~300 Einträgen trivial)

## Backup & Restore

### Dateistruktur
bbolt speichert **alles in einer einzigen `.db` Datei** – alle Buckets, alle Dokumente. Wie SQLite, kein separater Prozess nötig.

### Dump / Inspektion
Offizielles CLI-Tool für lesbare Ausgabe aller Buckets und Keys:
```bash
go install go.etcd.io/bbolt/cmd/bbolt@latest
bbolt dump referee.db
```

### Online-Backup via API
Konsistenter Backup während die DB läuft – kein Stoppen nötig:
```go
db.View(func(tx *bbolt.Tx) error {
    _, err := tx.WriteTo(w) // w = io.Writer, z.B. eine Datei
    return err
})
```

Ideal für einen täglichen Snapshot der dann von Borg Backup auf Mimir erfasst wird.

### Wichtig: Immer die gesamte DB sichern und restoren

Import, Export, Backup und Restore einzelner Buckets ist **gefährlich und ergibt keinen Sinn**:
- Stammdaten-Bucket (z.B. `spielorte`) und `spiele`-Bucket müssen immer konsistent zueinander sein
- Ein Spielort-Bucket von Zeitpunkt A + Spiele-Bucket von Zeitpunkt B = inkonsistente Referenzen, da IDs eingebettet sind
- **Backup und Restore immer auf die gesamte `.db` Datei anwenden**
- **Keine Import/Export-Funktionalität für einzelne Buckets implementieren** – das verleitet zu inkonsistenten Zuständen

Die Einfachheit von bbolt (eine Datei = alles) macht das ohnehin natürlich – Backup ist schlicht eine Dateikopie.

## Empfehlung

Wechsel zu bbolt macht Sinn **wenn** das Schema noch aktiv in Entwicklung ist und Migrationen den Workflow stören. Sobald das Schema stabil ist, verliert das Argument an Gewicht.

Zu klären mit Claude Code:
- Welche Abfragen nutzen aktuell tatsächlich SQL-Features (WHERE, GROUP BY, etc.)?
- Werden diese im Backend oder Frontend ausgeführt?
- Ist die Jahresfilterung die einzige Backend-seitige Filterung?
