package schedule

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/sphiecoh/apimonitor/conf"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/notification"
)

//Scheduler mantains the jobs and crons
type Scheduler struct {
	Cron   *cron.Cron
	Jobs   []*RunnerJob
	Store  *db.Store
	Config *conf.Config
}

//RunnerJob  represents the job to run by cron
type RunnerJob struct {
	db     *db.Store
	target *db.ApiTest
	Next   time.Time
	Prev   time.Time
	Config *conf.Config
}

// ToJob converts a test to a job
func ToJob(test *db.ApiTest, store *db.Store, conf *conf.Config) *RunnerJob {
	job := &RunnerJob{
		target: test,
		db:     store,
		Config: conf,
	}
	return job
}

//New creates a schedular
func New(tests []*db.ApiTest, store *db.Store, conf *conf.Config) *Scheduler {
	jobs := make([]*RunnerJob, 0)
	for _, test := range tests {
		job := &RunnerJob{
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
	logrus.Info("Started job scheduler")
	return nil
}

//Run runs the cron job
func (job RunnerJob) Run() {
	result := job.target.Run()
	if err := job.db.SaveResult(result); err != nil {
		logrus.WithField("test", job.target.Name).Errorf("failed to save result %v", err)
	}

	if result.Status != 200 {
		logrus.WithField("test", job.target.Name).Errorf("Test failed %v", result.Status)
		notification.NotifySlack(result.Error, "Test failed", job.Config)
		return
	}
	logrus.WithField("test", job.target.Name).Infof("Test succeeded ,status %v", result.Status)
}
