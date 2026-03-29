package store

import (
	"errors"
	"io"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
	bolt "go.etcd.io/bbolt"
)

var (
	ErrNotFound = errors.New("not found")

	bucketLeagues   = []byte("leagues")
	bucketTeams     = []byte("teams")
	bucketVenues    = []byte("venues")
	bucketPositions = []byte("positions")
	bucketGames     = []byte("games")

	allBuckets = [][]byte{
		bucketLeagues,
		bucketTeams,
		bucketVenues,
		bucketPositions,
		bucketGames,
	}

	entropy = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type Store struct {
	db *bolt.DB
}

func Open(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range allBuckets {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Backup(w io.Writer) error {
	return s.db.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
}

func newID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}
