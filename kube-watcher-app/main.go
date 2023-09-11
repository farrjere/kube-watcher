package main

import (
	"context"
	"embed"
	"github.com/farrjere/kube_watcher/kube-watcher-app/app"
	"github.com/farrjere/kube_watcher/kube-watcher-app/ui"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	ui := ui.New()
	app := app.NewApp(ui)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Kube Watcher",
		Width:  1500,
		Height: 1000,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			ui.Startup(ctx)
			app.Startup(ctx)
		},
		Bind: []interface{}{app, ui},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
