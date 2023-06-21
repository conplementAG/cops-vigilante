package snat_test

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/clock/testing"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("HealingState", func() {
	state := &snat.NodeHealingState{
		HealingStartedAt:  time.Now(),
		NumberOfErrorRuns: 0,
	}

	It("should heal until the trigger time is reached", func() {
		Expect(state.IsHealingNecessary(&testing.FakeClock{CurrentTime: time.Now().Add(-time.Minute)})).To(BeTrue())
		Expect(state.IsHealingNecessary(&testing.FakeClock{CurrentTime: time.Now().Add(2 * time.Minute)})).To(BeTrue())
		Expect(state.IsHealingNecessary(&testing.FakeClock{CurrentTime: time.Now().Add(snat.DefaultHealingDurationPerNode - time.Second)})).To(BeTrue())
		Expect(state.IsHealingNecessary(&testing.FakeClock{CurrentTime: time.Now().Add(snat.DefaultHealingDurationPerNode + time.Second)})).To(BeFalse())
	})
})
