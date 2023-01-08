package snat

import (
	"github.com/sirupsen/logrus"
	"time"
)

type snatTask struct{}

func (s *snatTask) Run() error {
	start := time.Now()
	logrus.Info("[SNAT Task] running")

	logrus.Info("[SNAT Task] finished in " + time.Since(start).String())
	return nil
}
