package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
type ApiResult struct {
	Status    int
	Name      string
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}
type ApiTest struct {
	Url  string
	Cron string
	Name string
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

func (s *Store) GetAllTests() ([]*ApiTest, error) {
	result := make([]*ApiTest, 0)
	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.TestBucket)

		b.ForEach(func(k, v []byte) error {
			apitest := new(ApiTest)
			err := json.Unmarshal(v, apitest)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to unmashal json for key %s", string(k)))
			}
			result = append(result, apitest)
			return nil
		})
		return nil
	})

	return result, err
}

func (s *Store) GetResultsByTest(name string) ([]*ApiResult, error) {
	result := make([]*ApiResult, 0)
	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.ResultBucket)

		b.ForEach(func(k, v []byte) error {
			apitest := new(ApiResult)
			err := json.Unmarshal(v, apitest)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to unmashal json for key %s", string(k)))
			}
			result = append(result, apitest)
			return nil
		})
		return nil
	})

	return result, err
}

// Run runs the API test
func (test *ApiTest) Run() *ApiResult {
	start := time.Now()
	result := &ApiResult{
		Name:      test.Name,
		Timestamp: time.Now(),
		Status:    500,
	}
	response, err := http.DefaultClient.Get(test.Url)
	result.Duration = time.Since(start)
	if err != nil {
		result.Error = err
		return result
	} else if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		res, _ := ioutil.ReadAll(response.Body)

		result.Error = errors.New(string(res))
	}
	result.Status = response.StatusCode
	return result
}
