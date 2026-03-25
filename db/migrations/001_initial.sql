-- +goose Up

CREATE TABLE IF NOT EXISTS positions (
    position TEXT PRIMARY KEY,
    long     TEXT NOT NULL,
    sorter   INTEGER NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS leagues (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    short_name TEXT    NOT NULL DEFAULT '',
    sorter     INTEGER NOT NULL DEFAULT 0,
    remarks    TEXT    NOT NULL DEFAULT '',
    created_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS teams (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    state      TEXT    NOT NULL DEFAULT 'Baden-Württemberg',
    is_active  INTEGER NOT NULL DEFAULT 1,
    remarks    TEXT    NOT NULL DEFAULT '',
    created_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS venues (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    city       TEXT    NOT NULL,
    short_name TEXT    NOT NULL DEFAULT '',
    stadium    TEXT    NOT NULL DEFAULT '',
    lat        REAL    NOT NULL DEFAULT 0.0,
    lon        REAL    NOT NULL DEFAULT 0.0,
    created_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at TEXT    NOT NULL DEFAULT (datetime('now', 'localtime'))
);

CREATE TABLE IF NOT EXISTS games (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    game_date     TEXT    NOT NULL,
    game_time     TEXT    NOT NULL DEFAULT '',
    home_team_id  INTEGER NOT NULL REFERENCES teams(id),
    away_team_id  INTEGER NOT NULL REFERENCES teams(id),
    venue_id      INTEGER NOT NULL DEFAULT 0 REFERENCES venues(id),
    league_id     INTEGER NOT NULL REFERENCES leagues(id),
    position      TEXT    NOT NULL REFERENCES positions(position),
    referee_fee   REAL    NOT NULL DEFAULT 0.0,
    travel_costs  REAL    NOT NULL DEFAULT 0.0,
    km_driven     INTEGER NOT NULL DEFAULT 0,
    exhibition    INTEGER NOT NULL DEFAULT 0,
    remarks       TEXT    NOT NULL DEFAULT '',
    created_at    TEXT    NOT NULL DEFAULT (datetime('now', 'localtime')),
    updated_at    TEXT    NOT NULL DEFAULT (datetime('now', 'localtime'))
);

-- +goose Down

DROP TABLE IF EXISTS games;
DROP TABLE IF EXISTS venues;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS leagues;
DROP TABLE IF EXISTS positions;
