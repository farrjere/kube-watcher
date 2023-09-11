package kube

import (
	"bufio"
	"context"
	v1 "k8s.io/api/core/v1"
)

/*
Things we want to do with a pod log
 1. Watch them until we Cancel
 2. Search:
    Possibly by
    - A container
    - Datetime
    - With a specific phrase
    - Cpu usage
    - Memory usage
*/
type PodLog struct {
	Messages <-chan string
	messages chan string
	PodName  string
	client   *KubeClient
	context  context.Context
}

func NewPodLog(name string, client *KubeClient, context context.Context, args ...int) *PodLog {
	buffer_length := 10
	if len(args) > 0 {
		buffer_length = args[0]
	}
	pl := PodLog{PodName: name, context: context, client: client}
	pl.messages = make(chan string, buffer_length)
	pl.Messages = pl.messages
	return &pl
}

func (pl *PodLog) GetLogs(lines int64) []string {
	options := v1.PodLogOptions{Timestamps: true}
	if lines > 0 {
		options.TailLines = &lines
	}
	return pl.GetLogsWithOpt(options)
}

func (pl *PodLog) GetLogsWithOpt(opts v1.PodLogOptions) []string {
	logs := pl.client.GetContainerLogs(pl.context, pl.PodName, opts)
	logLines := make([]string, 0)
	reader := bufio.NewScanner(logs)
	for {
		if !reader.Scan() {
			break
		}
		line := reader.Text()
		logLines = append(logLines, line)
	}
	return logLines
}

func (pl *PodLog) StreamLogs() {
	defer close(pl.messages)
	options := v1.PodLogOptions{Timestamps: true, Follow: true}
	logs := pl.client.GetContainerLogs(pl.context, pl.PodName, options)
	defer logs.Close()
	reader := bufio.NewScanner(logs)
	var line string

	for {
		for reader.Scan() {
			line = reader.Text()
			select {
			case <-pl.context.Done():
				return
			case pl.messages <- line:
				continue
			}
		}
	}
}
