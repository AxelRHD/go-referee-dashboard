package store

import (
	"encoding/json"
	"sort"

	bolt "go.etcd.io/bbolt"
)

func (s *Store) ListPositions() ([]Position, error) {
	var positions []Position
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPositions)
		return b.ForEach(func(k, v []byte) error {
			var p Position
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}
			positions = append(positions, p)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].Sorter < positions[j].Sorter
	})
	return positions, nil
}

func (s *Store) PutPosition(p *Position) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(p)
		if err != nil {
			return err
		}
		return tx.Bucket(bucketPositions).Put([]byte(p.Position), data)
	})
}

func (s *Store) DeletePosition(key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPositions).Delete([]byte(key))
	})
}
