package snat

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/services"
	"github.com/sirupsen/logrus"
	"time"
)

type snatTask struct {
	kubernetesService services.KubernetesService
}

func (s *snatTask) Run() error {
	start := time.Now()
	logrus.Info("[SNAT Task] running")
	SnatHealAttemptsCounter.Inc()
	logrus.Info("[SNAT Task] finished in " + time.Since(start).String())
	return nil
}
