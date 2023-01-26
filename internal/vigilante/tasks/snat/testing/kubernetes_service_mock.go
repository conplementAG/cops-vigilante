package testing

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/clock"
	"github.com/conplementag/cops-vigilante/internal/vigilante/services"
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/consts"
	apicorev1 "k8s.io/api/core/v1"
	apimachinerymetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubernetesServiceMock struct {
	services.KubernetesService

	// This is a fake in-memory representation of the cluster state, which can also be used in tests. This Mock
	// service will manipulate this state so that the tests can verify more on behaviours and effects, and less by checking
	// expected method calls. Set this before
	TestData_Nodes []*apicorev1.Node

	// TestData_Pods is the current collection of tracked pods created through this mock service. Key is the pod name.
	// Please do not set this field value yourself.
	TestData_Pods map[string]*apicorev1.Pod

	// TestData_DeletedPods is the collection of pod names deleted through this mock service. Key and the value are both the pod name.
	// Please do not set this field value yourself.
	TestData_DeletedPods map[string]string

	// Clock can be used to influence timestamps set in this mock service, which can be helpful to determine when test data was created / updated
	Clock clock.Clock

	// Error_CreatePod can be set to make CreatePod always return errors
	Error_CreatePod error
}

func NewKubernetesServiceMock(clock clock.Clock) *KubernetesServiceMock {
	return &KubernetesServiceMock{
		Clock:                clock,
		TestData_Pods:        map[string]*apicorev1.Pod{},
		TestData_DeletedPods: map[string]string{},
	}
}

func (m *KubernetesServiceMock) GetAllNodes() ([]*apicorev1.Node, error) {
	return m.TestData_Nodes, nil
}

func (m *KubernetesServiceMock) AddNodeAnnotation(nodeName string, annotationKey string, annotationValue string) error {
	node := services.FindNodeByName(m.TestData_Nodes, nodeName)
	node.Annotations[annotationKey] = annotationValue

	return nil
}

func (m *KubernetesServiceMock) CreatePod(pod *apicorev1.Pod) error {
	if m.Error_CreatePod != nil {
		return m.Error_CreatePod
	}

	pod.CreationTimestamp = apimachinerymetav1.NewTime(m.Clock.Now())
	m.TestData_Pods[pod.Name] = pod

	return nil
}

func (m *KubernetesServiceMock) DeletePod(namespace string, name string) error {
	// we want to simulate a real cluster here, so we should really delete only what is possible to delete,
	_, valueFound := m.TestData_Pods[name]

	if valueFound {
		m.TestData_DeletedPods[name] = name
		delete(m.TestData_Pods, name)
	} else {
		return nil // TODO error like not found etc.
	}

	return nil
}

func GenerateNodesSimilarToInitialNonHealedCluster() []*apicorev1.Node {
	return []*apicorev1.Node{
		GenerateNode(AksLinuxNode1, false, false, false, true),
		GenerateNode(AksLinuxNode2, false, false, false, true),
		GenerateNode(AksLinuxNode3, false, false, false, true),
		GenerateNode(AksWindowsNode1, true, false, false, true),
		GenerateNode(AksWindowsNode2, true, false, false, true),
		GenerateNode(AksWindowsNode3NeverReady, true, false, false, false),
	}
}

func GenerateNodesLikeInHealedCluster() []*apicorev1.Node {
	return []*apicorev1.Node{
		GenerateNode(AksLinuxNode1, false, false, false, true),
		GenerateNode(AksLinuxNode2, false, false, false, true),
		GenerateNode(AksLinuxNode3, false, false, false, true),
		GenerateNode(AksWindowsNode1, true, true, true, true),
		GenerateNode(AksWindowsNode2, true, true, true, true),
		GenerateNode(AksWindowsNode3NeverReady, true, false, false, false),
	}
}

func GenerateNode(name string, isWindows bool, isHealedAnnotationPresent bool, isHealed bool, isReady bool) *apicorev1.Node {
	var labels = map[string]string{}
	labels["random-label"] = "random-value" // making the test data more realistic

	var annotations = map[string]string{}
	annotations["random-annotation"] = "random-value"

	if isWindows {
		labels["kubernetes.io/os"] = "windows"
	} else {
		labels["kubernetes.io/os"] = "linux"
	}

	if isHealedAnnotationPresent {
		if isHealed {
			annotations[consts.NodeHealedAnnotation] = "true"
		} else {
			annotations[consts.NodeHealedAnnotation] = "false"
		}
	}

	result := apicorev1.Node{
		ObjectMeta: apimachinerymetav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
		Status: apicorev1.NodeStatus{},
	}

	UpdateNodeCondition(&result, isReady)

	return &result
}

func UpdateNodeCondition(node *apicorev1.Node, isReady bool) {
	node.Status.Conditions = []apicorev1.NodeCondition{
		{
			Type:    apicorev1.NodeDiskPressure, // making the test data more realistic
			Status:  apicorev1.ConditionTrue,
			Reason:  "disk ok",
			Message: "disk ok",
		},
	}

	readyCondition := apicorev1.NodeCondition{
		Type:    apicorev1.NodeReady,
		Status:  apicorev1.ConditionTrue,
		Reason:  "kubelet up and running",
		Message: "all green",
	}

	notReadyCondition := apicorev1.NodeCondition{
		Type:    apicorev1.NodeReady,
		Status:  apicorev1.ConditionFalse,
		Reason:  "kubelet not running",
		Message: "red red red",
	}

	if isReady {
		node.Status.Conditions = append(node.Status.Conditions, readyCondition)
	} else {
		node.Status.Conditions = append(node.Status.Conditions, notReadyCondition)
	}
}
