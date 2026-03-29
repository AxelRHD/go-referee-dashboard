package store

import (
	"encoding/json"
	"sort"

	bolt "go.etcd.io/bbolt"
)

func (s *Store) ListTeams() ([]Team, error) {
	var teams []Team
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketTeams)
		return b.ForEach(func(k, v []byte) error {
			var t Team
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			teams = append(teams, t)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Name < teams[j].Name
	})
	return teams, nil
}

func (s *Store) GetTeam(id string) (Team, error) {
	var t Team
	err := s.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketTeams).Get([]byte(id))
		if v == nil {
			return ErrNotFound
		}
		return json.Unmarshal(v, &t)
	})
	return t, err
}

func (s *Store) PutTeam(t *Team) error {
	if t.ID == "" {
		t.ID = newID()
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(t)
		if err != nil {
			return err
		}
		return tx.Bucket(bucketTeams).Put([]byte(t.ID), data)
	})
}

func (s *Store) DeleteTeam(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketTeams).Delete([]byte(id))
	})
}
