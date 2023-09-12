package app

import (
	"context"
	"github.com/farrjere/kube_watcher/kube"
	"github.com/farrjere/kube_watcher/kube-watcher-app/ui"
	"github.com/skratchdot/open-golang/open"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx           context.Context
	kubeClient    *kube.KubeClient
	watcher       *kube.DeploymentWatcher
	cancelFunc    context.CancelFunc
	CancelChannel chan string
	ui            *ui.UI
}

type PodLogMessage struct {
	Message string `json:"message"`
	Pod     string `json:"pod"`
}

// NewApp creates a new App application struct
func NewApp(ui *ui.UI) *App {
	c := make(chan string)
	return &App{ui: ui, CancelChannel: c}
}

func (a *App) Test() PodLogMessage {
	return PodLogMessage{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetContexts() []string {
	return kube.AvailableContexts("")
}

func (a *App) GetNamespaces() []string {
	ctx := context.Background()
	return a.kubeClient.GetNamespaces(ctx)
}

func (a *App) GetDeployments() []string {
	ctx := context.Background()
	return a.kubeClient.GetDeployments(ctx)
}

func (a *App) SetNamespace(namespace string) {
	wailsRuntime.LogInfo(a.ctx, "Called set namespace")
	a.kubeClient.SetNamespace(namespace)
}

func (a *App) SetDeployment(deployment string) []string {
	wailsRuntime.LogInfo(a.ctx, "Called set deployment")
	ctx := context.Background()
	a.watcher = kube.NewDeploymentWatcher(deployment, a.kubeClient, ctx)
	return a.watcher.GetPods()
}

func (a *App) CancelPodStream(pod string) {
	wailsRuntime.LogInfof(a.ctx, "Called cancel pod stream for pod %v", pod)
	a.CancelChannel <- pod
	wailsRuntime.LogInfof(a.ctx, "Messages in channel %v", len(a.CancelChannel))
}

func (a *App) Save() {
	dir := a.ui.ChooseDir("")
	a.watcher.LogAllPodsToDisk(dir, 0)
	err := open.Run(dir)
	if err != nil {
		wailsRuntime.LogError(a.ctx, err.Error())
	}
	wailsRuntime.LogInfo(a.ctx, "Saved logs to disk")
}

func (a *App) Search(query string, limit int64) []kube.SearchResult {
	wailsRuntime.LogInfo(a.ctx, "Search called")
	params := kube.SearchParameters{Query: query, AllContainers: true, Limit: limit}
	results := a.watcher.SearchLogs(params)
	return results
}

func (a *App) Stream() {
	wailsRuntime.LogInfo(a.ctx, "Stream called")
	podContexts := a.watcher.StreamLogs()
	for {
		for name, pc := range podContexts {
			select {
			case m := <-a.CancelChannel:
				wailsRuntime.LogInfof(a.ctx, "Canceling pod %v", m)
				podContexts[m].Cancel()
			case m := <-pc.PodLog.Messages:
				event := PodLogMessage{Message: m, Pod: name}
				wailsRuntime.EventsEmit(a.ctx, "pod_log", &event)
			}
		}
	}
}

func (a *App) LoadCluster(path string, context string) {
	config := kube.ConfigParameters{Path: path, Context: context}
	restConfig, err := kube.LoadConfig(config)
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "%v - %v - %v", err, path, context)
		return
	}
	a.kubeClient = kube.NewKubeClient(restConfig)
	if err != nil {
		return
	}
}
