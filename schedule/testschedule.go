package schedule

import (
	"time"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/sphiecoh/servicestatus/db"
	"github.com/sphiecoh/servicestatus/monitor"
)

type Scheduler struct {
	Cron  *cron.Cron
	Jobs  []*testJob
	Store *db.Store
}

type testJob struct {
	db     *db.Store
	target *monitor.ApiTest
	Next   time.Time
	Prev   time.Time
}

func ToJob(test *monitor.ApiTest, store *db.Store) *testJob {
	job := &testJob{
		target: test,
		db:     store,
	}
	return job
}

//New creates a schedular
func New(tests []*monitor.ApiTest, store *db.Store) *Scheduler {
	jobs := make([]*testJob, 0)
	for _, test := range tests {
		job := &testJob{
			db:     store,
			target: test,
		}
		jobs = append(jobs, job)
	}
	s := &Scheduler{
		Cron:  cron.New(),
		Jobs:  jobs,
		Store: store,
	}
	return s
}

//Start starts the scheduler
func (s *Scheduler) Start() error {
	for _, job := range s.Jobs {
		schedule, err := cron.ParseStandard(job.target.Cron)
		if err != nil {
			return errors.Wrapf(err, "Invalid cron %v for test %v", job.target.Cron, job.target.Name)
		}
		s.Cron.Schedule(schedule, job)
	}

	s.Cron.Start()
	return nil
}

func (job testJob) Run() {
	result := job.target.Run()
	data, err := json.Marshal(result)
	if err != nil {
		logrus.WithField("test", job.target.Name).Errorf("failed to marshal result json %v", err)
	}
	job.db.Put("results", job.target.Name, data)
	if result.Error != nil {
		logrus.WithField("test", job.target.Name).Errorf("Test failed %v", result.Status)
		//TODO notify
		return
	}
	logrus.WithField("test", job.target.Name).Infof("Test succeeded ,status %v", result.Status)
}
