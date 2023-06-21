package snat

import (
	"github.com/conplementag/cops-vigilante/internal/vigilante/tasks/snat/consts"
	corev1 "k8s.io/api/core/v1"
	apimachinerymetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetHealingPodDefinition(nodeName string) *corev1.Pod {
	terminationGracePeriod := int64(0)

	return &corev1.Pod{
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
	}
}
