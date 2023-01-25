package snat

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/conplementag/cops-vigilante/internal/vigilante/clock"
	"github.com/conplementag/cops-vigilante/internal/vigilante/database"
	"github.com/conplementag/cops-vigilante/internal/vigilante/services"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/consts"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/metrics"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type snatTask struct {
	kubernetesService services.KubernetesService
	stateDatabase     database.Database
	metrics           metrics.SnatMetrics
	clock             clock.Clock
}

var DefaultHealingDurationPerNode = time.Minute * 30

func (s *snatTask) Run() {
	start := s.clock.Now()
	logrus.Debug("[SNAT Task] running")

	s.heal()

	logrus.Debug("[SNAT Task] finished in " + time.Since(start).String())
}

type NodeHealingState struct {
	HealingStartedAt time.Time
}

func (s *NodeHealingState) IsHealingNecessary(clock clock.Clock) bool {
	return clock.Now().Before(s.HealingStartedAt.Add(DefaultHealingDurationPerNode))
}

func (s *snatTask) heal() {
	allNodes, err := s.kubernetesService.GetAllNodes() // TODO error handling

	if err != nil { // TODO error handling
		logrus.Error(err)
		return
	}

	// we only want to process ready non-healed windows nodes here, other ones are of no interest since:
	// - we cannot cover all edge cases if the "heal" pod is possible to schedule etc.
	// - linux nodes are of no interest in SNAT
	// - healed nodes are simply considered healthy until the annotation is removed manually
	readyNonHealedWindowsNodes := s.filterForReadyNonHealedWindowsNodes(allNodes)

	for _, node := range readyNonHealedWindowsNodes {
		logrus.Debug("[SNAT Task] Found a ready non healed windows node: " + node.Name)

		// initialize healing state if required
		if s.stateDatabase.Get(node.Name) == nil {
			s.stateDatabase.Set(node.Name, &NodeHealingState{
				HealingStartedAt: s.clock.Now(),
			})
		}

		healingState := s.stateDatabase.Get(node.Name).(*NodeHealingState)

		if healingState.IsHealingNecessary(s.clock) {
			err := s.kubernetesService.DeletePod(consts.NodeHealerPodNamePrefix+consts.NodeHealerNamespace, node.Name)

			// TODO error handling
			if err != nil {
				logrus.Error(err)
				panic(err)
			}

			terminationGracePeriod := int64(0)

			err = s.kubernetesService.CreatePod(&v1.Pod{
				ObjectMeta: v12.ObjectMeta{
					Name:      consts.NodeHealerPodNamePrefix + node.Name,
					Namespace: consts.NodeHealerNamespace,
				},
				Spec: v1.PodSpec{
					NodeSelector: map[string]string{
						"kubernetes.io/os":       "windows",
						"kubernetes.io/hostname": node.Name,
					},
					Containers: []v1.Container{
						{
							Name:  "pause-container",
							Image: "mcr.microsoft.com/oss/kubernetes/pause:3.6",
						},
					},
					TerminationGracePeriodSeconds: &terminationGracePeriod,
				},
			})

			// TODO error handling
			if err != nil {
				logrus.Error(err)
				panic(err)
			}

			s.metrics.IncHealAttemptsCounter()
		} else {
			err := s.kubernetesService.DeletePod(consts.NodeHealerPodNamePrefix+consts.NodeHealerNamespace, node.Name)

			// TODO error handling
			if err != nil {
				logrus.Error(err)
				panic(err)
			}

			err = s.kubernetesService.AddNodeAnnotation(node.Name, consts.NodeHealedAnnotation, "true")

			// TODO error handling
			if err != nil {
				logrus.Error(err)
				panic(err)
			}
		}
	}
}

func (s *snatTask) filterForReadyNonHealedWindowsNodes(nodes []*v1.Node) []*v1.Node {
	var results []*v1.Node

	for _, node := range nodes {
		if val, ok := node.Labels["kubernetes.io/os"]; ok {
			if val != "windows" && val != "Windows" {
				continue // other node types are of no interest
			}
		}

		conditionSearchResult := linq.From(node.Status.Conditions).WhereT(func(condition v1.NodeCondition) bool {
			return condition.Type == v1.NodeReady
		}).Single()

		if conditionSearchResult == nil {
			logrus.Debugf("[SNAT Task] Found a windows node without the ready condition: %v", node.Name)
			s.metrics.IncNumberOfNotReadyNodesSeen()

			continue // node does not need to be processed anymore since not ready
		}

		readyCondition := conditionSearchResult.(v1.NodeCondition)

		if readyCondition.Status != v1.ConditionTrue {
			logrus.Debugf("[SNAT Task] Found non-ready windows node %v, reason: %v", node.Name, readyCondition.Reason)
			s.metrics.IncNumberOfNotReadyNodesSeen()

			continue // node does not need to be processed anymore since not ready
		}

		isHealed, keyExists := node.Annotations[consts.NodeHealedAnnotation]

		if keyExists && (isHealed == "true" || isHealed == "True") {
			continue // node already healed, skip
		} else {
			results = append(results, node)
		}
	}

	return results
}
