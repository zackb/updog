package job

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zackb/updog/db"
)

type Job struct {
	id       cron.EntryID
	Func     func()
	CronExpr string
}

type Scheduler struct {
	c    *cron.Cron
	Jobs []*Job
}

func NewScheduler() *Scheduler {
	c := cron.New(cron.WithLocation(time.UTC))
	return &Scheduler{
		c:    c,
		Jobs: make([]*Job, 0),
	}
}

func (s *Scheduler) AddJob(job *Job) error {
	id, err := s.c.AddFunc(job.CronExpr, job.Func)
	job.id = id
	if err != nil {
		return err
	}
	s.Jobs = append(s.Jobs, job)
	return nil
}

func (s *Scheduler) AddDefaultJobs(store *db.DB) {
	// rollup job runs 2 minutes after midnight UTC
	rollupJob := &Job{
		Func: func() {

			log.Println("Starting daily pageview rollup job...")
			day := time.Now().UTC().AddDate(0, 0, -1)
			err := store.PageviewStorage().RunDailyRollup(context.Background(), day)
			if err != nil {
				log.Println("Error running daily rollup job:", err)
			} else {
				log.Println("Daily pageview rollup job completed successfully.")
			}

		},
		CronExpr: "2 0 * * *",
	}

	err := s.AddJob(rollupJob)
	if err != nil {
		log.Println("Error adding daily rollup job to scheduler:", err)
	}

	// test job runs every 15 minutes
	testJob := &Job{
		Func: func() {
			log.Println("Test job executed at", time.Now().UTC())
		},
		CronExpr: "*/15 * * * *",
	}

	err = s.AddJob(testJob)
	if err != nil {
		log.Println("Error adding test job to scheduler:", err)
	}
}

func (s *Scheduler) Start() {
	s.c.Start()
}

func (s *Scheduler) Stop() {
	s.c.Stop()
}
