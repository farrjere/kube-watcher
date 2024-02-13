package kube

import (
	"context"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
)

type KubeClient struct {
	client    *kubernetes.Clientset
	config    *rest.Config
	namespace string
}

func NewKubeClient(config *rest.Config) *KubeClient {
	client := KubeClient{config: config}
	fmt.Println(config)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panicMsg := fmt.Sprintf("Unable to setup client %v", err)
		panic(panicMsg)
	}
	client.namespace = "default"
	client.client = clientset
	return &client
}

func (kc *KubeClient) GetNamespaces(ctx context.Context) []string {
	options := metav1.ListOptions{}
	namespaceList, err := kc.client.CoreV1().Namespaces().List(ctx, options)
	if err != nil {
		fmt.Printf("Error getting namespaces %v\n", err)
		return []string{}
	}

	var namespaces = make([]string, len(namespaceList.Items))
	for i, namespace := range namespaceList.Items {
		namespaces[i] = namespace.Name
	}
	return namespaces
}

func (kc *KubeClient) SetNamespace(namespace string) {
	kc.namespace = namespace
}

func (kc *KubeClient) GetDeployments(ctx context.Context) []string {
	options := metav1.ListOptions{}
	deploymentList, err := kc.client.AppsV1().Deployments(kc.namespace).List(ctx, options)
	if err != nil {
		fmt.Printf("Error getting deployments %v", err)
		return []string{}
	}
	var deployments = make([]string, len(deploymentList.Items))
	for i, deployment := range deploymentList.Items {
		deployments[i] = deployment.Name
	}
	return deployments
}

type Pod struct {
	Name       string
	Containers []string
	State      string
}

func (kc *KubeClient) GetPods(ctx context.Context, deploymentName string) []Pod {
	options := metav1.GetOptions{}
	deployment, err := kc.client.AppsV1().Deployments(kc.namespace).Get(ctx, deploymentName, options)

	labelSelector := metav1.LabelSelector{MatchLabels: deployment.Labels}
	matchLabels := labels.Set(labelSelector.MatchLabels).String()
	listOptions := metav1.ListOptions{LabelSelector: matchLabels}
	podsList, err := kc.client.CoreV1().Pods(kc.namespace).List(ctx, listOptions)
	if err != nil {
		fmt.Printf("Error getting pods for deployment %v: %v", deploymentName, err)
		return []Pod{}
	}
	var pods = make([]Pod, len(podsList.Items))
	for i, pod := range podsList.Items {
		containers := make([]string, len(pod.Spec.Containers))
		for j, container := range pod.Spec.Containers {
			containers[j] = container.Name
		}
		p := Pod{Name: pod.Name, Containers: containers, State: string(pod.Status.Phase)}
		pods[i] = p
	}
	return pods
}

func (kc *KubeClient) GetContainerLogs(ctx context.Context, podName string, options v1.PodLogOptions) io.ReadCloser {
	logsRq := kc.client.CoreV1().Pods(kc.namespace).GetLogs(podName, &options)
	logs, err := logsRq.Stream(ctx)
	if err != nil {
		fmt.Printf("Failed to get logs for %v- with error %v", podName, err)
		return nil
	}
	return logs
}
