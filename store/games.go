package store

import (
	"encoding/json"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func (s *Store) ListGames() ([]Game, error) {
	var games []Game
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGames)
		return b.ForEach(func(k, v []byte) error {
			var g Game
			if err := json.Unmarshal(v, &g); err != nil {
				return err
			}
			games = append(games, g)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(games, func(i, j int) bool {
		if games[i].GameDate != games[j].GameDate {
			return games[i].GameDate > games[j].GameDate
		}
		return games[i].GameTime < games[j].GameTime
	})
	return games, nil
}

func (s *Store) ListGamesBySeason(year string) ([]Game, error) {
	var games []Game
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGames)
		return b.ForEach(func(k, v []byte) error {
			var g Game
			if err := json.Unmarshal(v, &g); err != nil {
				return err
			}
			if strings.HasPrefix(g.GameDate, year) {
				games = append(games, g)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(games, func(i, j int) bool {
		return games[i].GameDate < games[j].GameDate
	})
	return games, nil
}

func (s *Store) ListSeasons() ([]string, error) {
	seen := make(map[string]bool)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGames)
		return b.ForEach(func(k, v []byte) error {
			var g Game
			if err := json.Unmarshal(v, &g); err != nil {
				return err
			}
			if len(g.GameDate) >= 4 {
				seen[g.GameDate[:4]] = true
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	seasons := make([]string, 0, len(seen))
	for y := range seen {
		seasons = append(seasons, y)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(seasons)))
	return seasons, nil
}

func (s *Store) GetGame(id string) (Game, error) {
	var g Game
	err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketGames).Get([]byte(id))
		if data == nil {
			return ErrNotFound
		}
		return json.Unmarshal(data, &g)
	})
	return g, err
}

func (s *Store) PutGame(g *Game) error {
	if g.ID == "" {
		g.ID = newID()
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(g)
		if err != nil {
			return err
		}
		return tx.Bucket(bucketGames).Put([]byte(g.ID), data)
	})
}

func (s *Store) DeleteGame(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketGames).Delete([]byte(id))
	})
}

// IsReferenced prüft ob eine Stammdaten-ID in irgendeinem Game referenziert wird.
func (s *Store) IsReferenced(check func(g *Game) bool) bool {
	found := false
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGames)
		return b.ForEach(func(k, v []byte) error {
			if found {
				return nil
			}
			var g Game
			if err := json.Unmarshal(v, &g); err != nil {
				return err
			}
			if check(&g) {
				found = true
			}
			return nil
		})
	})
	return found
}

// UpdateGameRefs aktualisiert eingebettete Referenzen in allen Games.
// Die update-Funktion erhält jedes Game und gibt true zurück wenn es geändert wurde.
func (s *Store) UpdateGameRefs(update func(g *Game) bool) (int, error) {
	count := 0
	return count, s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGames)
		return b.ForEach(func(k, v []byte) error {
			var g Game
			if err := json.Unmarshal(v, &g); err != nil {
				return err
			}
			if update(&g) {
				data, err := json.Marshal(&g)
				if err != nil {
					return err
				}
				if err := b.Put(k, data); err != nil {
					return err
				}
				count++
			}
			return nil
		})
	})
}
