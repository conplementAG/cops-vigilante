package snat

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/clock"
	"time"
)

type NodeHealingState struct {
	HealingStartedAt  time.Time
	NumberOfErrorRuns int
}

// IsHealingNecessary follows the logic that the healing for a node is required until the DefaultHealingDurationPerNode period
// is reached, after which the node is considered "healed"
func (s *NodeHealingState) IsHealingNecessary(clock clock.Clock) bool {
	currentTime := clock.Now()
	triggerTime := s.HealingStartedAt.Add(DefaultHealingDurationPerNode)

	return currentTime.Before(triggerTime)
}
