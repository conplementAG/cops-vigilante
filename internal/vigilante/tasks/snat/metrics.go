package snat

import "github.com/prometheus/client_golang/prometheus"

var SnatHealAttemptsCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "vigilante_snat_node_healing_attempts",
		Help: "Count of node healing attempts for the SNAT issue.",
	},
)
