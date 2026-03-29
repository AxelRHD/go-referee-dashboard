package store

import (
	"encoding/json"
	"sort"

	bolt "go.etcd.io/bbolt"
)

func (s *Store) ListLeagues() ([]League, error) {
	var leagues []League
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketLeagues)
		return b.ForEach(func(k, v []byte) error {
			var l League
			if err := json.Unmarshal(v, &l); err != nil {
				return err
			}
			leagues = append(leagues, l)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(leagues, func(i, j int) bool {
		if leagues[i].Sorter != leagues[j].Sorter {
			return leagues[i].Sorter < leagues[j].Sorter
		}
		return leagues[i].Name < leagues[j].Name
	})
	return leagues, nil
}

func (s *Store) GetLeague(id string) (League, error) {
	var l League
	err := s.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketLeagues).Get([]byte(id))
		if v == nil {
			return ErrNotFound
		}
		return json.Unmarshal(v, &l)
	})
	return l, err
}

func (s *Store) PutLeague(l *League) error {
	if l.ID == "" {
		l.ID = newID()
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(l)
		if err != nil {
			return err
		}
		return tx.Bucket(bucketLeagues).Put([]byte(l.ID), data)
	})
}

func (s *Store) DeleteLeague(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketLeagues).Delete([]byte(id))
	})
}
