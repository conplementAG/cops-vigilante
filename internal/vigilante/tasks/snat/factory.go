package snat

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/clock"
	"github.com/conplementag/cops-vigilante/internal/vigilante/database"
	"github.com/conplementag/cops-vigilante/internal/vigilante/services"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/metrics"
)

func NewSnatTask(
	kubernetesService services.KubernetesService,
	stateDatabase database.Database,
	metrics metrics.SnatMetrics,
	clock clock.Clock,
) tasks.Task {
	return &snatTask{
		kubernetesService: kubernetesService,
		stateDatabase:     stateDatabase,
		metrics:           metrics,
		clock:             clock,
	}
}
