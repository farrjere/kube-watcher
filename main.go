package main

import (
	"context"
	"fmt"
	"github.com/farrjere/kube_watcher/kube"
	cli "github.com/urfave/cli/v2"
	"k8s.io/client-go/rest"
	"log"
	"os"
)

func setContext(cCtx *cli.Context) *rest.Config {
	path := cCtx.String("path")
	context := cCtx.String("context")
	save := cCtx.Bool("save")

	contextParms := kube.ConfigParameters{Context: context, Path: path, Save: save}

	conf, err := kube.LoadConfig(contextParms)
	if err != nil {
		fmt.Println("unable to properly load config")
		panic(err)
	}
	return conf
}

func saveDeploymentLogs(cCtx *cli.Context) {
	ctx := context.Background()
	config, err := kube.LoadConfig(kube.ConfigParameters{})
	if err != nil {
		fmt.Println("error loading config")
		panic(err)
	}
	kc := kube.NewKubeClient(config)
	namespace := cCtx.String("namespace")
	kc.SetNamespace(namespace)
	deployment := cCtx.String("deployment")
	lines := cCtx.Int64("lines")
	dl := kube.NewDeploymentWatcher(deployment, kc, ctx)
	path := cCtx.Args().Get(0)

	dl.LogAllPodsToDisk(path, lines)
	fmt.Printf("Output logs for %v to %v", deployment, path)
}

func searchDeploymentLogs(cCtx *cli.Context) {
	ctx := context.Background()
	config, err := kube.LoadConfig(kube.ConfigParameters{})
	if err != nil {
		fmt.Println("error loading config")
		panic(err)
	}
	kc := kube.NewKubeClient(config)
	namespace := cCtx.String("namespace")
	kc.SetNamespace(namespace)
	deployment := cCtx.String("deployment")
	query := cCtx.String("query")
	dl := kube.NewDeploymentWatcher(deployment, kc, ctx)
	searchParams := kube.SearchParameters{Query: query, AllContainers: true}
	results := dl.SearchLogs(searchParams)
	fmt.Printf("Found %v results", len(results))
	for _, result := range results {
		fmt.Printf("Results for %v\n", result.PodName)
		fmt.Println("----------------------------------------------------------------")
		for _, match := range result.Matches {
			fmt.Println(match)
		}
		fmt.Println()
	}
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "set_context",
				Aliases: []string{"c"},
				Usage:   "Sets the context that will be used for all other commands",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "save", Value: true},
					&cli.StringFlag{Name: "path", Usage: "the path to your kube conf"},
					&cli.StringFlag{Name: "context", Usage: "the context to use"},
				},
				Action: func(cCtx *cli.Context) error {
					setContext(cCtx)
					return nil
				},
			},
			{
				Name:    "deployment_logs",
				Aliases: []string{"dl"},
				Usage:   "saves all logs to disk, searches logs, or attaches to logs to watch",
				Subcommands: []*cli.Command{
					&cli.Command{
						Name:  "search",
						Usage: "searches a deployment logs for the query",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "namespace", Usage: "the namespace to use"},
							&cli.StringFlag{Name: "deployment", Usage: "deployment"},
							&cli.StringFlag{Name: "query", Usage: "the query to search for"},
						},
						Action: func(cCtx *cli.Context) error {
							searchDeploymentLogs(cCtx)
							return nil
						},
					},
					&cli.Command{
						Name:  "save",
						Usage: "saves all logs for a deployment to disk: dl save -flags path",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "namespace", Usage: "the namespace to use"},
							&cli.StringFlag{Name: "deployment", Usage: "deployment"},
							&cli.Int64Flag{Name: "lines", Usage: "the # of lines to output", Value: 0},
						},
						Action: func(cCtx *cli.Context) error {
							saveDeploymentLogs(cCtx)
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
