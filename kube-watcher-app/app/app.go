package app

import (
	"context"
	"github.com/farrjere/kube_watcher/kube"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	kubeClient *kube.KubeClient
	watcher    *kube.DeploymentWatcher
	cancelFunc context.CancelFunc
}

type PodLogMessage struct {
	Message string `json:"message"`
	Pod     string `json:"pod"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
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

func (a *App) Stream() {
	wailsRuntime.LogInfo(a.ctx, "Stream called")
	podContexts := a.watcher.StreamLogs()
	for {
		for name, pc := range podContexts {
			select {
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
