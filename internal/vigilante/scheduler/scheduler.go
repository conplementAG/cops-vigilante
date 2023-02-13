package scheduler

import (
	"context"
	"fmt"
	"github.com/conplementag/cops-vigilante/internal/vigilante/errors"
	"github.com/conplementag/cops-vigilante/internal/vigilante/services"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat"
	snatmetrics "github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/metrics"
	"github.com/procyon-projects/chrono"
	"github.com/sirupsen/logrus"
	"k8s.io/utils/clock"
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

	k8sClient, err := services.NewKubernetesService()
	if err != nil {
		errors.PanicOnError(err)
	}

	// task instances need to be kept "outside" the scheduled loop to preserve state (if any)
	snatTask := snat.NewSnatTask(k8sClient, &snatmetrics.SnatMetricsRecorder{}, &clock.RealClock{})

	_, err = s.taskScheduler.ScheduleAtFixedRate(func(ctx context.Context) {
		snatTask.Run()
	}, duration)

	if err != nil {
		errors.PanicOnError(err)
	}
}
