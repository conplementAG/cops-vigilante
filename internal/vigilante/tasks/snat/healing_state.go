package snat

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/clock"
	"time"
)

type NodeHealingState struct {
	HealingStartedAt  time.Time
	NumberOfErrorRuns int
}

func (s *NodeHealingState) IsHealingNecessary(clock clock.Clock) bool {
	return clock.Now().Before(s.HealingStartedAt.Add(DefaultHealingDurationPerNode))
}
