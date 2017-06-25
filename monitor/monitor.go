package monitor

import "time"
import "github.com/robfig/cron"
import "net/http"

type ApiResult struct {
	Status    int
	Name      string
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}
type ApiTest struct {
	Url  string
	Cron *cron.Cron
	Name string
}

func (test *ApiTest) Run() *ApiResult {
	start := time.Now()
	result := &ApiResult{
		Name:      test.Name,
		Timestamp: time.Now().UTC(),
		Status:    500,
	}
	response, err := http.DefaultClient.Get(test.Url)
	result.Duration = time.Since(start)
	if err != nil {
		result.Error = err
		return result
	}
	result.Status = response.StatusCode
	return result
}
