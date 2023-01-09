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
	GetAllNodes() ([]core1.Node, error)
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

func (a *kubernetesService) GetAllNodes() ([]core1.Node, error) {
	ingressList, err := a.clientSet.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})

	if err != nil {
		return nil, err
	}

	return ingressList.Items, nil
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
