package metrics

import (
	snatmetrics "github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func Init() {
	prometheus.MustRegister(snatmetrics.SnatHealAttemptsCounter)
	prometheus.MustRegister(snatmetrics.SnatNumberOfNotReadyNodesFound)
}
