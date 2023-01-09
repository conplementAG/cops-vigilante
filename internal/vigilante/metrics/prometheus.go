package metrics

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat"
	"github.com/prometheus/client_golang/prometheus"
)

func Init() {
	prometheus.MustRegister(snat.SnatHealAttemptsCounter)
}
