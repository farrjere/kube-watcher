package kube

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type PodContext struct {
	PodLog  *PodLog
	pod     Pod
	context context.Context
	Cancel  context.CancelFunc
}

type DeploymentWatcher struct {
	name        string
	client      *KubeClient
	context     context.Context
	pods        []Pod
	podContexts map[string]PodContext
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
	dl.podContexts = make(map[string]PodContext)
	for _, p := range dl.pods {
		childContext, cancel := context.WithCancel(ctx)
		dl.podContexts[p.Name] = PodContext{pod: p, PodLog: NewPodLog(p.Name, client, childContext), context: childContext, Cancel: cancel}
	}
	return &dl
}

func (dl *DeploymentWatcher) GetPods() []string {
	podNames := make([]string, len(dl.pods))
	for i, p := range dl.pods {
		podNames[i] = p.Name
	}
	return podNames
}

func (dl *DeploymentWatcher) StreamLogs() map[string]PodContext {
	for _, p := range dl.pods {
		pc := dl.podContexts[p.Name]
		go func() {
			pc.PodLog.StreamLogs()
		}()
	}
	return dl.podContexts
}

func (dl *DeploymentWatcher) StreamLogsConsole() {
	logColors := make(map[string]*color.Color)
	ignoreColors := []int{0, 15, 16, 231}
	for _, p := range dl.pods {
		pc := dl.podContexts[p.Name]
		i := rand.Intn(231)
		for {
			if !slices.Contains(ignoreColors, i) {
				ignoreColors = append(ignoreColors, i)
				break
			}
			i = rand.Intn(231)
		}
		logColor := color.New(color.Attribute(38), color.Attribute(5), color.Attribute(i))
		logColors[p.Name] = logColor
		go func() {
			pc.PodLog.StreamLogs()
		}()
	}

	for {
		for _, p := range dl.pods {
			pc := dl.podContexts[p.Name]
			select {
			case m := <-pc.PodLog.Messages:
				logColor := logColors[p.Name]
				_, err := logColor.Println(p.Name, m)
				if err != nil {
					fmt.Println("unable to print log line")
				}
			case <-dl.context.Done():
				if errors.Is(pc.context.Err(), context.Canceled) {
					break
				}
				pc.Cancel()
			}
		}
	}
}

func (dl *DeploymentWatcher) LogAllPodsToDisk(path string, lines int64) {
	var wg sync.WaitGroup
	for podName, pc := range dl.podContexts {
		wg.Add(1)
		podName := podName
		pc := pc
		go func() {
			defer wg.Done()
			logs := pc.PodLog.GetLogs(lines)
			logPath := filepath.Join(path, podName+".log")
			WriteLinesToDisk(logPath, logs)
		}()
	}
	wg.Wait()
}

func WriteLinesToDisk(path string, lines []string) {
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		fmt.Printf("Unable to write %v\n", path)
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(f)
	for _, line := range lines {
		_, err = w.WriteString(line + "\n")
		if err != nil {
			fmt.Printf("unable to write string %v to %v ", line, path)
			fmt.Println(err)
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

func searchPodLogs(wg *sync.WaitGroup, searchParams SearchParameters, podname string, pc PodContext, resultChannel chan<- SearchResult) {
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

func searchContainerLog(opts v1.PodLogOptions, searchParams SearchParameters, pc PodContext, container string) []string {
	opts.Container = container
	if !searchParams.Since.IsZero() {
		opts.SinceTime = &metav1.Time{Time: searchParams.Since}
	}
	matches := make([]string, 0)
	logs := pc.PodLog.GetLogsWithOpt(opts)
	for _, l := range logs {
		match := strings.Index(l, searchParams.Query)
		if match > -1 {
			matches = append(matches, l)
		}
	}
	return matches
}
