package monitor

import (
	"net/http"
	"time"

	"encoding/json"

	"fmt"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/sphiecoh/apimonitor/db"
)

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
type Monitor struct {
	Store            *db.Store
	ResultBucketName string
	TestBucketName   string
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
		res := make([]byte, 0)
		response.Body.Read(res)
		result.Error = errors.New(string(res))
	}
	result.Status = response.StatusCode
	return result
}

// GetAllTests retrives all tests
func (m *Monitor) GetAllTests() ([]*ApiTest, error) {
	result := make([]*ApiTest, 0)
	err := m.Store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(m.TestBucketName))

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
