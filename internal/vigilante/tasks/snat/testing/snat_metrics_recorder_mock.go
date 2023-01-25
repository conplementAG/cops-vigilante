package testing

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/metrics"
	"github.com/stretchr/testify/mock"
)

type SnatMetricsRecorderMock struct {
	mock.Mock
	metrics.SnatMetrics
}

func (m *SnatMetricsRecorderMock) IncHealAttemptsCounter() {
	m.Called()
}

func (m *SnatMetricsRecorderMock) IncNumberOfNotReadyNodesSeen() {
	m.Called()
}
