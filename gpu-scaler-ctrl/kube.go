package main

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func newK8sClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback on local `kubectl` configuration, to test.
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
