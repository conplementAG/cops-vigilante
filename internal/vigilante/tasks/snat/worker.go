package snat

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/conplementag/cops-vigilante/internal/vigilante/clock"
	"github.com/conplementag/cops-vigilante/internal/vigilante/database"
	"github.com/conplementag/cops-vigilante/internal/vigilante/services"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/consts"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/metrics"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apimachinerymetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// heal runs the healing logic and explicitly cannot error out, since all errors should be handled already. In case of issue
// a panic would occur.
func (s *snatTask) heal() {
	allNodes, err := s.kubernetesService.GetAllNodes()

	if err != nil {
		// Errors during node fetch should be logged and basically ignored, in hope the next run would fix the issue.
		// This should be the only error which should go on for infinity. Perhaps consider some backoff logic here in
		// the future.
		logrus.Error(err)
		return
	}

	// We only want to process ready non-healed windows nodes here, other ones are of no interest since:
	// - we cannot cover all edge cases if the "heal" pod is possible to schedule etc.
	// - linux nodes are of no interest in SNAT
	// - healed nodes are simply considered healthy until the annotation is removed manually
	readyNonHealedWindowsNodes := s.filterForReadyNonHealedWindowsNodes(allNodes)

	// Some cases, like nodes becoming unready, require us to restart the healing process.
	s.updateStateDatabase(readyNonHealedWindowsNodes)

	for _, node := range readyNonHealedWindowsNodes {
		logrus.Info("[SNAT Task] Found a ready non healed windows node: " + node.Name)

		s.initializeHealingStateIfRequired(node.Name)
		healingState := s.stateDatabase.Get(node.Name).(*NodeHealingState)

		if healingState.IsHealingNecessary(s.clock) {
			err := s.deleteHealingPodIfAlreadyScheduled(node.Name)

			// TODO error handling
			if err != nil {
				logrus.Error(err)
				panic(err)
			}

			err = s.createHealingPod(node.Name)

			// TODO error handling
			if err != nil {
				logrus.Error(err)
				panic(err)
			}

			s.metrics.IncHealAttemptsCounter()
		} else {
			err := s.deleteHealingPodIfAlreadyScheduled(node.Name)

			// TODO error handling
			if err != nil {
				logrus.Error(err)
				panic(err)
			}

			err = s.markNodeHealed(node.Name)

			// TODO error handling
			if err != nil {
				logrus.Error(err)
				panic(err)
			}
		}
	}
}

func (s *snatTask) filterForReadyNonHealedWindowsNodes(nodes []*corev1.Node) []*corev1.Node {
	var results []*corev1.Node

	for _, node := range nodes {
		if val, ok := node.Labels["kubernetes.io/os"]; ok {
			if val != "windows" && val != "Windows" {
				continue // other node types are of no interest
			}
		}

		conditionSearchResult := linq.From(node.Status.Conditions).WhereT(func(condition corev1.NodeCondition) bool {
			return condition.Type == corev1.NodeReady
		}).Single()

		if conditionSearchResult == nil {
			logrus.Debugf("[SNAT Task] Found a windows node without the ready condition: %v", node.Name)
			s.metrics.IncNumberOfNotReadyNodesSeen()

			continue // node does not need to be processed anymore since not ready
		}

		readyCondition := conditionSearchResult.(corev1.NodeCondition)

		if readyCondition.Status != corev1.ConditionTrue {
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

func (s *snatTask) updateStateDatabase(readyWindowsNodes []*corev1.Node) {
	// If the healing is recorded in our state, and the node becomes un-ready or with unknown state, then we should remove
	// the healing record and forget about the node. Once it becomes ready again, the healing process will
	// effectively restart because the new state will be written for that node. In this case, we should also
	// not attempt any.
	var itemKeysToRemoveFromState []string

	for key, _ := range s.stateDatabase.GetAll() {
		readyNode := linq.From(readyWindowsNodes).WhereT(func(node *corev1.Node) bool {
			return node.Name == key
		}).Single()

		if readyNode == nil {
			itemKeysToRemoveFromState = append(itemKeysToRemoveFromState, key)
		}
	}

	// We delete outside the loop above to prevent modifying the state database in the same loop (removing items from
	// a "collection" while iterating the same collection is never a good idea).
	for _, key := range itemKeysToRemoveFromState {
		s.stateDatabase.Delete(key)
	}
}

func (s *snatTask) initializeHealingStateIfRequired(nodeName string) {
	if s.stateDatabase.Get(nodeName) == nil {
		s.stateDatabase.Set(nodeName, &NodeHealingState{
			HealingStartedAt: s.clock.Now(),
		})
	}
}

func (s *snatTask) deleteHealingPodIfAlreadyScheduled(nodeName string) error {
	return s.kubernetesService.DeletePod(consts.NodeHealerNamespace, consts.NodeHealerPodNamePrefix+nodeName)
}

func (s *snatTask) createHealingPod(nodeName string) error {
	terminationGracePeriod := int64(0)

	return s.kubernetesService.CreatePod(&corev1.Pod{
		ObjectMeta: apimachinerymetav1.ObjectMeta{
			Name:      consts.NodeHealerPodNamePrefix + nodeName,
			Namespace: consts.NodeHealerNamespace,
		},
		Spec: corev1.PodSpec{
			NodeSelector: map[string]string{
				"kubernetes.io/os":       "windows",
				"kubernetes.io/hostname": nodeName,
			},
			Containers: []corev1.Container{
				{
					Name:  "pause-container",
					Image: "mcr.microsoft.com/oss/kubernetes/pause:3.6",
				},
			},
			TerminationGracePeriodSeconds: &terminationGracePeriod,
		},
	})
}

func (s *snatTask) markNodeHealed(nodeName string) error {
	return s.kubernetesService.AddNodeAnnotation(nodeName, consts.NodeHealedAnnotation, "true")
}
