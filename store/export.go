package store

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	bolt "go.etcd.io/bbolt"
)

type DatabaseExport struct {
	ExportedAt string     `json:"exported_at"`
	Leagues    []League   `json:"leagues"`
	Teams      []Team     `json:"teams"`
	Venues     []Venue    `json:"venues"`
	Positions  []Position `json:"positions"`
	Games      []Game     `json:"games"`
}

func (s *Store) ExportJSON(w io.Writer) error {
	leagues, err := s.ListLeagues()
	if err != nil {
		return fmt.Errorf("export leagues: %w", err)
	}
	teams, err := s.ListTeams()
	if err != nil {
		return fmt.Errorf("export teams: %w", err)
	}
	venues, err := s.ListVenues()
	if err != nil {
		return fmt.Errorf("export venues: %w", err)
	}
	positions, err := s.ListPositions()
	if err != nil {
		return fmt.Errorf("export positions: %w", err)
	}
	games, err := s.ListGames()
	if err != nil {
		return fmt.Errorf("export games: %w", err)
	}

	export := DatabaseExport{
		ExportedAt: time.Now().Format(time.RFC3339),
		Leagues:    leagues,
		Teams:      teams,
		Venues:     venues,
		Positions:  positions,
		Games:      games,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(export)
}

func (s *Store) ImportJSON(r io.Reader) error {
	var data DatabaseExport
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		// Alle Buckets leeren
		for _, name := range allBuckets {
			b := tx.Bucket(name)
			c := b.Cursor()
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				if err := b.Delete(k); err != nil {
					return err
				}
			}
		}

		// Positions
		bp := tx.Bucket(bucketPositions)
		for _, p := range data.Positions {
			d, err := json.Marshal(&p)
			if err != nil {
				return err
			}
			if err := bp.Put([]byte(p.Position), d); err != nil {
				return err
			}
		}

		// Leagues
		bl := tx.Bucket(bucketLeagues)
		for _, l := range data.Leagues {
			if l.ID == "" {
				l.ID = newID()
			}
			d, err := json.Marshal(&l)
			if err != nil {
				return err
			}
			if err := bl.Put([]byte(l.ID), d); err != nil {
				return err
			}
		}

		// Teams
		bt := tx.Bucket(bucketTeams)
		for _, t := range data.Teams {
			if t.ID == "" {
				t.ID = newID()
			}
			d, err := json.Marshal(&t)
			if err != nil {
				return err
			}
			if err := bt.Put([]byte(t.ID), d); err != nil {
				return err
			}
		}

		// Venues
		bv := tx.Bucket(bucketVenues)
		for _, v := range data.Venues {
			if v.ID == "" {
				v.ID = newID()
			}
			d, err := json.Marshal(&v)
			if err != nil {
				return err
			}
			if err := bv.Put([]byte(v.ID), d); err != nil {
				return err
			}
		}

		// Games
		bg := tx.Bucket(bucketGames)
		for _, g := range data.Games {
			if g.ID == "" {
				g.ID = newID()
			}
			d, err := json.Marshal(&g)
			if err != nil {
				return err
			}
			if err := bg.Put([]byte(g.ID), d); err != nil {
				return err
			}
		}

		return nil
	})
}
