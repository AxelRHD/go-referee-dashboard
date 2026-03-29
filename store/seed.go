package store

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// SeedBatch schreibt alle Items eines Entity-Typs in einer einzigen Transaction.
func (s *Store) SeedBatch(entity string, data []byte) (int, error) {
	switch entity {
	case "positions":
		var items []Position
		if err := json.Unmarshal(data, &items); err != nil {
			return 0, err
		}
		return len(items), s.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketPositions)
			for _, p := range items {
				d, err := json.Marshal(&p)
				if err != nil {
					return err
				}
				if err := b.Put([]byte(p.Position), d); err != nil {
					return err
				}
			}
			return nil
		})

	case "leagues":
		var items []League
		if err := json.Unmarshal(data, &items); err != nil {
			return 0, err
		}
		return len(items), s.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketLeagues)
			for i := range items {
				if items[i].ID == "" {
					items[i].ID = newID()
				}
				d, err := json.Marshal(&items[i])
				if err != nil {
					return err
				}
				if err := b.Put([]byte(items[i].ID), d); err != nil {
					return err
				}
			}
			return nil
		})

	case "teams":
		var items []Team
		if err := json.Unmarshal(data, &items); err != nil {
			return 0, err
		}
		return len(items), s.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketTeams)
			for i := range items {
				if items[i].ID == "" {
					items[i].ID = newID()
				}
				d, err := json.Marshal(&items[i])
				if err != nil {
					return err
				}
				if err := b.Put([]byte(items[i].ID), d); err != nil {
					return err
				}
			}
			return nil
		})

	case "venues":
		var items []Venue
		if err := json.Unmarshal(data, &items); err != nil {
			return 0, err
		}
		return len(items), s.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketVenues)
			for i := range items {
				if items[i].ID == "" {
					items[i].ID = newID()
				}
				d, err := json.Marshal(&items[i])
				if err != nil {
					return err
				}
				if err := b.Put([]byte(items[i].ID), d); err != nil {
					return err
				}
			}
			return nil
		})

	default:
		return 0, fmt.Errorf("unbekannte Entity: %s", entity)
	}
}
