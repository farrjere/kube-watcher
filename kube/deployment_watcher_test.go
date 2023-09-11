package kube

import (
	"context"
	"os"
	"testing"
)

func TestLogAllPodsToDisk(t *testing.T) {
	ctx := context.Background()
	config, err := LoadConfig(ConfigParameters{})
	if err != nil {
		t.Error(err)
	}
	kc := NewKubeClient(config)
	kc.SetNamespace("default")
	dl := NewDeploymentWatcher("test", kc, ctx)
	tempDir, err := os.MkdirTemp("", "sampledir")
	defer os.RemoveAll(tempDir)
	dl.LogAllPodsToDisk(tempDir, 15)
	d, e := os.ReadDir(tempDir)
	if e != nil {
		panic(e)
	}
	if len(d) != 10 {
		t.Errorf("expected there to be 3 files instead found %v", len(d))
	}
}

func TestSearchLogs(t *testing.T) {
	ctx := context.Background()
	config, err := LoadConfig(ConfigParameters{})
	if err != nil {
		t.Error(err)
	}
	kc := NewKubeClient(config)
	kc.SetNamespace("default")
	dl := NewDeploymentWatcher("test", kc, ctx)
	searchParams := SearchParameters{Query: "hel"}
	results := dl.SearchLogs(searchParams)
	if len(results) < 1 {
		t.Errorf("Expected some results")
	}
	for _, r := range results {
		if len(r.Matches) == 0 {
			t.Errorf("Expected there to be at least one (likely a lot of matches) %v - %v", r.PodName, len(r.Matches))
		}
	}
}
