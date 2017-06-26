package db

import (
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

// Store represent a bolt db datastore
type Store struct {
	*bolt.DB
	TestBucket   []byte
	ResultBucket []byte
}

// NewStore creates a new store.
func NewStore(path string) (*Store, error) {
	config := &bolt.Options{Timeout: 1 * time.Second}
	d, err := bolt.Open(path, 0600, config)
	if err != nil {
		return nil, errors.Wrapf(err, "Opening store %s failed", path)
	}

	return &Store{d, []byte("tests"), []byte("results")}, nil
}

// CreateBuckets creates a new buckets
func (db *Store) CreateBuckets() error {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(db.TestBucket)
		if err != nil {
			return errors.Wrapf(err, "Could not create test bucket")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(db.ResultBucket)
		if err != nil {
			return errors.Wrapf(err, "Could not create result bucket")
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// RemoveBucket deletes bucket by name
func (db *Store) RemoveBucket(name []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(name)
	})
}
func (db *Store) Put(key string, bucket, data []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		return b.Put([]byte(key), data)

	})
}
