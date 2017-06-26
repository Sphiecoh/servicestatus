package schedule

import (
	"time"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/sphiecoh/apimonitor/conf"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/monitor"
	"github.com/sphiecoh/apimonitor/notification"
)

type Scheduler struct {
	Cron   *cron.Cron
	Jobs   []*testJob
	Store  *db.Store
	Config *conf.Config
}

type testJob struct {
	db     *db.Store
	target *monitor.ApiTest
	Next   time.Time
	Prev   time.Time
	Config *conf.Config
}

func ToJob(test *monitor.ApiTest, store *db.Store, conf *conf.Config) *testJob {
	job := &testJob{
		target: test,
		db:     store,
		Config: conf,
	}
	return job
}

//New creates a schedular
func New(tests []*monitor.ApiTest, store *db.Store, conf *conf.Config) *Scheduler {
	jobs := make([]*testJob, 0)
	for _, test := range tests {
		job := &testJob{
			db:     store,
			target: test,
			Config: conf,
		}
		jobs = append(jobs, job)
		logrus.Infof("Scheduling test %s", test.Name)
	}
	s := &Scheduler{
		Cron:   cron.New(),
		Jobs:   jobs,
		Store:  store,
		Config: conf,
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
	job.db.Put(job.target.Name, []byte("results"), data)
	if result.Error != nil {
		logrus.WithField("test", job.target.Name).Errorf("Test failed %v", result.Status)
		notification.NotifySlack(result.Error.Error(), "Test failed", job.Config)
		return
	}
	logrus.WithField("test", job.target.Name).Infof("Test succeeded ,status %v", result.Status)
}
