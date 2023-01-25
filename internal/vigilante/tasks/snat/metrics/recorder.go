package metrics

type SnatMetrics interface {
	IncHealAttemptsCounter()
	IncNumberOfNotReadyNodesSeen()
}

type SnatMetricsRecorder struct{}

func (r *SnatMetricsRecorder) IncHealAttemptsCounter() {
	SnatHealAttemptsCounter.Inc()
}

func (r *SnatMetricsRecorder) IncNumberOfNotReadyNodesSeen() {
	SnatNumberOfNotReadyNodesFound.Inc()
}
