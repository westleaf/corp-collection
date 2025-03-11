package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientSet, err := getClientSet()
	if err != nil {
		log.Fatal(err)
	}

	nsFoo := createNamespace(ctx, clientSet, "foo")
	defer func() {
		deleteNamespace(ctx, clientSet, nsFoo)
	}()

	deployNginx(ctx, clientSet, nsFoo, "hello-world")
	fmt.Printf("Service is now running here: http://localhost:8080\n\n")

	listenToPodLogs(ctx, clientSet, nsFoo, "hello-world")

	waitForExitSignal()
}

func getClientSet() (*kubernetes.Clientset, error) {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String(
			"kubconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file",
		)
	} else {
		kubeconfig = flag.String(
			"kubeconfig",
			"",
			"absolut path to the kubeconfig file",
		)
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags(
		"",
		*kubeconfig,
	)
	if err != nil {
		return nil, err
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func createNamespace(ctx context.Context, clientSet *kubernetes.Clientset, name string) (*corev1.Namespace, error) {
	fmt.Printf("Creating namespace %q.\n\n", name)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	ns, err := clientSet.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return ns, nil
}

func deployNginx(ctx context.Context, clientSet *kubernetes.Clientset, ns *corev1.Namespace, name string) {
	deployment := createNginxDeployment(ctx, clientSet, ns, name)
	waitForReadyReplicas(ctx, clientSet, deployment)
	createNginxService(ctx, clientSet, ns, name)
	createNginxIngress(ctx, clientset, ns, name)
}

func createNginxDeployment(ctx context.Context, clientSet *kubernetes.Clientset, ns *corev1.Namespace, name string) *appsv1.Deployment {
	var (
		matchLabel = map[string]string{"app": "nginx"}
		objMeta    = metav1.ObjectMeta{
			Name:      name,
			Namespace: ns.Name,
			Labels:    matchLabel,
		}
	)

	template := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: matchLabel,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  name,
					Image: "nginxdemos/hello:latest",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 8080,
						},
					},
				},
			},
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: objMeta,
		Spec: appsv1.DeploymentSpec{
			Replicas: to.int32,
		},
	}
}
