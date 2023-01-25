package metrics

import "github.com/prometheus/client_golang/prometheus"

var SnatHealAttemptsCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "cops_vigilante_snat_node_healing_attempts",
		Help: "Count of node healing attempts for the SNAT issue.",
	},
)

var SnatNumberOfNotReadyNodesFound = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "cops_vigilante_snat_number_of_not_ready_nodes_found",
		Help: "Count of nodes in not ready state found during attempting to heal the SNAT issue.",
	},
)
