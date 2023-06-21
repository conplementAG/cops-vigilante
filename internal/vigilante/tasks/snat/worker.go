package snat

import (
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/conplementag/cops-vigilante/internal/vigilante/clock"
	"github.com/conplementag/cops-vigilante/internal/vigilante/services"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/consts"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/metrics"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"time"
)

type snatTask struct {
	kubernetesService services.KubernetesService
	state             map[string]interface{}
	metrics           metrics.SnatMetrics
	clock             clock.Clock
}

var DefaultHealingDurationPerNode = time.Minute * 30
var NumberOfErrorsToleratedThreshold = 10

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
		logrus.Error("GetAllNodes produced an error (will be logged as the next log output line). This heal() run will be stopped, " +
			"in hope that the next one works without an error.")
		logrus.Error(err)
		return
	}

	// We only want to process ready non-healed windows nodes here, other ones are of no interest since:
	// - in case of non-ready nodes, we cannot cover all edge cases if the "heal" pod is possible to schedule etc.
	// - linux nodes are of no interest in SNAT
	// - healed nodes are simply considered healthy, except perhaps if the annotation is removed manually
	readyNonHealedWindowsNodes := s.filterForReadyNonHealedWindowsNodes(allNodes)

	// we need to "update" our state for the new result, for example, perhaps some nodes are missing because removed / turned not-ready
	// in which case we will remove then from state.
	s.reconcileState(readyNonHealedWindowsNodes)

	for _, node := range readyNonHealedWindowsNodes {
		s.initializeHealingStateIfRequired(node.Name)
		healingState := s.state[node.Name].(*NodeHealingState)

		if healingState.NumberOfErrorRuns >= NumberOfErrorsToleratedThreshold {
			logrus.Debugf("Skipping the healing for node %s to to number of errors reached: %d", node.Name, healingState.NumberOfErrorRuns)
			continue
		}

		if healingState.IsHealingNecessary(s.clock) {
			logrus.Info("[SNAT Task] Found a ready non healed windows node: " + node.Name)
			s.metrics.IncHealAttemptsCounter()

			err := s.deleteHealingPodIfAlreadyScheduled(node.Name)
			if err != nil {
				s.handleError(node.Name, err)
				continue
			}

			err = s.createHealingPod(node.Name)
			if err != nil {
				s.handleError(node.Name, err)
				continue
			}
		} else {
			err := s.deleteHealingPodIfAlreadyScheduled(node.Name)
			if err != nil {
				s.handleError(node.Name, err)
				continue
			}

			err = s.markNodeHealed(node.Name)
			if err != nil {
				s.handleError(node.Name, err)
				continue
			}
		}
	}
}

func (s *snatTask) filterForReadyNonHealedWindowsNodes(nodes []*corev1.Node) []*corev1.Node {
	var results []*corev1.Node

	for _, node := range nodes {
		if val, ok := node.Labels["kubernetes.io/os"]; ok {
			if !strings.EqualFold(val, "windows") {
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

func (s *snatTask) reconcileState(readyWindowsNodes []*corev1.Node) {
	// If the healing is recorded in our state, and the node becomes un-ready or with unknown state, then we should remove
	// the healing record and forget about the node. Once it becomes ready again, the healing process will
	// effectively restart because the new state will be written for that node. In this case, we should also
	// not attempt any.
	var itemKeysToRemoveFromState []string

	for key, _ := range s.state {
		readyNode := linq.From(readyWindowsNodes).WhereT(func(node *corev1.Node) bool {
			return node.Name == key
		}).Single()

		if readyNode == nil {
			itemKeysToRemoveFromState = append(itemKeysToRemoveFromState, key)
		}
	}

	// We delete outside the loop above to prevent modifying the state in the same loop (removing items from
	// a "collection" while iterating the same collection is never a good idea).
	for _, key := range itemKeysToRemoveFromState {
		delete(s.state, key)
	}
}

func (s *snatTask) initializeHealingStateIfRequired(nodeName string) {
	if s.state[nodeName] == nil {
		s.state[nodeName] = &NodeHealingState{
			HealingStartedAt: s.clock.Now(),
		}
	}
}

func (s *snatTask) deleteHealingPodIfAlreadyScheduled(nodeName string) error {
	err := s.kubernetesService.DeletePod(consts.NodeHealerNamespace, consts.NodeHealerPodNamePrefix+nodeName)

	// as the method name suggests, this error can be ignored
	if apierrors.IsNotFound(err) {
		return nil
	} else {
		return err
	}
}

func (s *snatTask) createHealingPod(nodeName string) error {
	return s.kubernetesService.CreatePod(GetHealingPodDefinition(nodeName))
}

// handleError handles errors that occur during a node healing step
func (s *snatTask) handleError(nodeName string, err error) {
	logrus.Warnf("[SNAT Task] error occured, will be recorded. Error: %v", err)
	healingState := s.state[nodeName].(*NodeHealingState)

	if healingState == nil {
		panic(fmt.Sprintf("Got an error during healing of a node %s, but the state for that node was not found. "+
			"This should not be possible.", nodeName))
	}

	healingState.NumberOfErrorRuns++

	if healingState.NumberOfErrorRuns >= NumberOfErrorsToleratedThreshold {
		logrus.Errorf("Healing of node %s reached the maximum ammount of errors that are allowed to occur: %d", nodeName, NumberOfErrorsToleratedThreshold)
	}
}

func (s *snatTask) markNodeHealed(nodeName string) error {
	return s.kubernetesService.AddNodeAnnotation(nodeName, consts.NodeHealedAnnotation, "true")
}
