package kube

import (
	"bufio"
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type podContext struct {
	podLog  *PodLog
	pod     Pod
	context context.Context
	cancel  context.CancelFunc
}

type DeploymentWatcher struct {
	name        string
	client      *KubeClient
	context     context.Context
	pods        []Pod
	podContexts map[string]podContext
}

type SearchParameters struct {
	Query         string
	Container     string
	Since         time.Time
	AllContainers bool
}

type SearchResult struct {
	PodName string
	Matches []string
}

func NewDeploymentWatcher(name string, client *KubeClient, ctx context.Context) *DeploymentWatcher {
	dl := DeploymentWatcher{name: name, client: client, context: ctx}
	dl.pods = client.GetPods(ctx, name)
	dl.podContexts = make(map[string]podContext)
	for _, p := range dl.pods {
		childContext, cancel := context.WithCancel(ctx)
		dl.podContexts[p.Name] = podContext{pod: p, podLog: NewPodLog(p.Name, client, childContext), context: childContext, cancel: cancel}
	}
	return &dl
}

func (dl *DeploymentWatcher) LogAllPodsToDisk(path string, lines int64) {
	var wg sync.WaitGroup
	for podName, pc := range dl.podContexts {
		wg.Add(1)
		podName := podName
		pc := pc
		go func() {
			defer wg.Done()
			writeLogsToDisk(path, pc, podName, lines)
		}()
	}
	wg.Wait()
}

func writeLogsToDisk(path string, pc podContext, podName string, lines int64) {
	logs := pc.podLog.GetLogs(lines)
	f, err := os.Create(filepath.Join(path, podName+".log"))
	defer f.Close()
	if err != nil {
		fmt.Printf("Unable to write logs for %v\n", podName)
		return
	}
	w := bufio.NewWriter(f)
	for _, line := range logs {
		_, err = w.WriteString(line + "\n")
		if err != nil {
			fmt.Printf("unable to write string %v for pod %v ", line, podName)
			break
		}
	}
	w.Flush()
}

func (dl *DeploymentWatcher) SearchLogs(searchParams SearchParameters) []SearchResult {
	var wg sync.WaitGroup
	finalRes := make([]SearchResult, 0)
	results := make(chan SearchResult, len(dl.pods))
	for podName, pc := range dl.podContexts {
		wg.Add(1)
		go searchPodLogs(&wg, searchParams, podName, pc, results)
	}

	wg.Wait()
	close(results)
	for res := range results {
		if len(res.Matches) > 0 {
			finalRes = append(finalRes, res)
		}
	}
	return finalRes
}

func searchPodLogs(wg *sync.WaitGroup, searchParams SearchParameters, podname string, pc podContext, resultChannel chan<- SearchResult) {
	defer wg.Done()
	matches := make([]string, 0)
	opts := v1.PodLogOptions{Timestamps: true}
	if searchParams.AllContainers {
		for _, c := range pc.pod.Containers {
			matches = append(matches, searchContainerLog(opts, searchParams, pc, c)...)
		}
	} else {
		matches = searchContainerLog(opts, searchParams, pc, searchParams.Container)
	}

	res := SearchResult{PodName: podname, Matches: matches}
	resultChannel <- res
}

func searchContainerLog(opts v1.PodLogOptions, searchParams SearchParameters, pc podContext, container string) []string {
	opts.Container = container
	if !searchParams.Since.IsZero() {
		opts.SinceTime = &metav1.Time{Time: searchParams.Since}
	}
	matches := make([]string, 0)
	logs := pc.podLog.GetLogsWithOpt(opts)
	for _, l := range logs {
		match := strings.Index(l, searchParams.Query)
		if match > -1 {
			matches = append(matches, l)
		}
	}
	return matches
}
