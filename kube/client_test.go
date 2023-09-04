package kube

import (
	"context"
	"testing"
)

func TestGetDeploymentPods(t *testing.T) {
	ctx := context.Background()
	config, err := LoadConfig(ConfigParameters{})
	if err != nil {
		t.Error(err)
	}
	kc := NewKubeClient(config)

	kc.SetNamespace("kube-system")
	deployments := kc.GetDeployments(ctx)
	for _, d := range deployments {
		pods := kc.GetPods(ctx, d)
		if len(pods) != 1 {
			t.Errorf("Got %v pods expected 1", len(pods))
		}
	}
}

func TestGetNamespaces(t *testing.T) {
	ctx := context.Background()
	config, err := LoadConfig(ConfigParameters{})
	if err != nil {
		t.Error(err)
	}
	kc := NewKubeClient(config)
	namespaces := kc.GetNamespaces(ctx)
	if len(namespaces) < 1 {
		t.Errorf("expected at least 1 namespace")
	}
}
