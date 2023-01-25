package services

import (
	"golang.org/x/net/context"
	core1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type KubernetesService interface {
	GetAllNodes() ([]*core1.Node, error)
	AddNodeAnnotation(nodeName string, annotationKey string, annotationValue string) error
	CreatePod(pod *core1.Pod) error
	DeletePod(namespace string, name string) error
}

func NewKubernetesService() (KubernetesService, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &kubernetesService{
		clientSet: clientset,
	}, nil
}

type kubernetesService struct {
	clientSet *kubernetes.Clientset
}

func (k *kubernetesService) GetAllNodes() ([]*core1.Node, error) {
	nodesList, err := k.clientSet.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})

	if err != nil {
		return nil, err
	}

	var results []*core1.Node

	for _, node := range nodesList.Items {
		results = append(results, &node)
	}

	return results, nil
}

func (k *kubernetesService) AddNodeAnnotation(nodeName string, annotationKey string, annotationValue string) error {
	node, err := k.clientSet.CoreV1().Nodes().Get(context.TODO(), nodeName, v1.GetOptions{})

	if err != nil {
		return err
	}

	node.Annotations[annotationKey] = annotationValue

	_, err = k.clientSet.CoreV1().Nodes().Update(context.TODO(), node, v1.UpdateOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (k *kubernetesService) CreatePod(pod *core1.Pod) error {
	_, err := k.clientSet.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, v1.CreateOptions{})
	return err
}

func (k *kubernetesService) DeletePod(namespace string, name string) error {
	return k.clientSet.CoreV1().Pods(namespace).Delete(context.TODO(), name, v1.DeleteOptions{})
}

func getConfig() (*rest.Config, error) {
	var config *rest.Config

	kubeConfigPath := os.Getenv("KUBECONFIG")

	if kubeConfigPath == "" {
		kubeConfigPath = os.Getenv("HOME") + "/.kube/config"
	}

	if _, err := os.Stat(kubeConfigPath); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
