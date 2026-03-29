package store

import (
	"encoding/json"
	"sort"

	bolt "go.etcd.io/bbolt"
)

func (s *Store) ListVenues() ([]Venue, error) {
	var venues []Venue
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketVenues)
		return b.ForEach(func(k, v []byte) error {
			var venue Venue
			if err := json.Unmarshal(v, &venue); err != nil {
				return err
			}
			venues = append(venues, venue)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(venues, func(i, j int) bool {
		if venues[i].ShortName != venues[j].ShortName {
			return venues[i].ShortName < venues[j].ShortName
		}
		return venues[i].City < venues[j].City
	})
	return venues, nil
}

func (s *Store) GetVenue(id string) (Venue, error) {
	var v Venue
	err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketVenues).Get([]byte(id))
		if data == nil {
			return ErrNotFound
		}
		return json.Unmarshal(data, &v)
	})
	return v, err
}

func (s *Store) PutVenue(v *Venue) error {
	if v.ID == "" {
		v.ID = newID()
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return tx.Bucket(bucketVenues).Put([]byte(v.ID), data)
	})
}

func (s *Store) UpdateVenueCoords(id string, lat, lon float64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketVenues)
		data := b.Get([]byte(id))
		if data == nil {
			return ErrNotFound
		}
		var v Venue
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		v.Lat = lat
		v.Lon = lon
		updated, err := json.Marshal(&v)
		if err != nil {
			return err
		}
		return b.Put([]byte(id), updated)
	})
}

func (s *Store) DeleteVenue(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketVenues).Delete([]byte(id))
	})
}
