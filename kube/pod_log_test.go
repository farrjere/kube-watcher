package kube

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestGetLog(t *testing.T) {
	ctx := context.Background()
	config, err := LoadConfig(ConfigParameters{})
	if err != nil {
		t.Error(err)
	}
	kc := NewKubeClient(config)

	kc.SetNamespace("kube-system")

	pl := NewPodLog("kube-apiserver-minikube", kc, ctx)
	logs := pl.GetLogs(10)
	if len(logs) != 10 {
		t.Errorf("Expected logs to be 10 lines long but got %v", len(logs))
	}
}

func TestStreamLogs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//ctx := context.Background()
	config, err := LoadConfig(ConfigParameters{})
	if err != nil {
		t.Error(err)
	}

	kc := NewKubeClient(config)
	kc.SetNamespace("kube-system")
	pl := NewPodLog("kube-apiserver-minikube", kc, ctx)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		pl.StreamLogs()
	}()

	wg.Wait()

	if len(pl.Messages) != 10 {
		t.Errorf("Expected there to be 10 log lines waiting, found %v", len(pl.Messages))
	}
}
