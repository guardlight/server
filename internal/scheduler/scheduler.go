package scheduler

import (
	"time"

	"github.com/go-co-op/gocron/v2"
)

type Scheduler struct {
	Gos gocron.Scheduler
}

func NewScheduler(loc *time.Location) (*Scheduler, error) {
	gos, err := gocron.NewScheduler(
		gocron.WithLocation(loc),
		gocron.WithLimitConcurrentJobs(1, gocron.LimitModeReschedule),
	)
	if err != nil {
		return nil, err
	}

	gos.Start()

	s := &Scheduler{
		Gos: gos,
	}

	return s, nil
}
