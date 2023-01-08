package scheduler

import (
	"context"
	"fmt"
	"github.com/conplementag/cops-vigilante/internal/vigilante/errors"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat"
	"github.com/procyon-projects/chrono"
	"github.com/sirupsen/logrus"
	"time"
)

type Scheduler struct {
	taskScheduler chrono.TaskScheduler
}

var (
	instance *Scheduler
)

func InitializeAndStart(scheduleIntervalInSeconds int) *Scheduler {
	logrus.Info("[Scheduler] starting...")
	if instance == nil {
		instance = &Scheduler{
			taskScheduler: chrono.NewDefaultTaskScheduler(),
		}
	}

	instance.start(scheduleIntervalInSeconds)
	logrus.Info("[Scheduler] successfully started.")

	return instance
}

func (s *Scheduler) start(scheduleIntervalInSeconds int) {
	duration, _ := time.ParseDuration(fmt.Sprintf("%ds", scheduleIntervalInSeconds))

	_, err := s.taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		snat.NewSnatTask().Run()
	}, duration)

	if err != nil {
		errors.PanicOnError(err)
	}
}
