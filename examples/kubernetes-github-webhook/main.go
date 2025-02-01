package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var (
		client           *kubernetes.Clientset
		deploymentLabels map[string]string
		err              error
		expectedPods     int32
	)

	ctx := context.Background()

	if client, err = getClient(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
	if deploymentLabels, expectedPods, err = deploy(ctx, client); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
	if err = waitForPods(ctx, client, deploymentLabels, expectedPods); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("deployed finished with labels: %v\n", deploymentLabels)
}

func getClient() (*kubernetes.Clientset, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func deploy(ctx context.Context, client *kubernetes.Clientset) (map[string]string, int32, error) {
	var deployment *v1.Deployment

	appFile, err := os.ReadFile("deployment.yaml")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read deployment file: %w", err)
	}

	obj, groupVersionKind, err := scheme.Codecs.UniversalDeserializer().Decode(appFile, nil, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode deployment file: %w", err)
	}

	switch obj := obj.(type) {
	case *v1.Deployment:
		deployment = obj
	default:
		return nil, 0, fmt.Errorf("unrecognized object: %s", groupVersionKind)
	}

	_, err = client.AppsV1().Deployments("default").Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		deploymentResponse, err := client.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create deployment: %w", err)
		}
		return deploymentResponse.Spec.Template.Labels, 0, nil
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, 0, fmt.Errorf("failed to get deployment: %w", err)
	}

	deploymentResponse, err := client.AppsV1().Deployments("default").Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to update deployment: %w", err)
	}
	return deploymentResponse.Spec.Template.Labels, *deploymentResponse.Spec.Replicas, nil
}

func waitForPods(ctx context.Context, client *kubernetes.Clientset, deploymentLabels map[string]string, expectedPods int32) error {
	for {
		validatedLabels, err := labels.ValidatedSelectorFromSet(deploymentLabels)
		if err != nil {
			return fmt.Errorf("failed to validate labels: %w", err)
		}
		podList, err := client.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
			LabelSelector: validatedLabels.String(),
		})
		if err != nil {
			return fmt.Errorf("failed to list pods: %w", err)
		}
		podsRunning := 0
		for _, pod := range podList.Items {
			if pod.Status.Phase == "Running" {
				podsRunning++
			}
		}
		if podsRunning > 0 && podsRunning == len(podList.Items) && podsRunning == int(expectedPods) {
			break
		}

		fmt.Printf("waiting for pods to be running: %d/%d\n", podsRunning, len(podList.Items))
		time.Sleep(1 * time.Second)
	}
	return nil
}
