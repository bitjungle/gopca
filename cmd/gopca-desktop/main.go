// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/bitjungle/gopca/internal/version"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/linux/icon.png
var icon []byte

func main() {
	// Parse command-line flags
	openFile := flag.String("open", "", "CSV file to open on startup")
	showVersion := flag.Bool("version", false, "Show version information")
	tutorialDataset := flag.String("tutorial", "", "Open a tutorial window for a sample dataset (iris|wine|corn|swiss_roll|stocks)")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Println(version.Get().Short())
		os.Exit(0)
	}

	var err error

	// The same App struct serves both the main window and tutorial windows.
	// In tutorial mode, SetTutorialDataset() is called so that GetAppMode()
	// returns the correct mode to the frontend.
	app := NewApp()

	if *tutorialDataset != "" {
		// Tutorial mode: launched by OpenTutorial() with --tutorial <dataset>.
		// A smaller window shows only the TutorialViewer component.
		app.SetTutorialDataset(*tutorialDataset)
		err = wails.Run(&options.App{
			Title:            fmt.Sprintf("GoPCA Tutorial — %s", *tutorialDataset),
			Width:            900,
			Height:           750,
			WindowStartState: options.Normal,
			AssetServer: &assetserver.Options{
				Assets: assets,
			},
			BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
			OnStartup:        app.startup,
			Bind:             []interface{}{app},
			Mac: &mac.Options{
				TitleBar: &mac.TitleBar{
					HideTitleBar:    false,
					HideTitle:       false,
					FullSizeContent: false,
					UseToolbar:      false,
				},
			},
			Linux: &linux.Options{
				Icon:                icon,
				WindowIsTranslucent: false,
				WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways,
				ProgramName:         "GoPCA Tutorial",
			},
		})
	} else {
		// Normal mode: full GoPCA Desktop application.
		if *openFile != "" {
			app.SetFileToOpen(*openFile)
		}

		err = wails.Run(&options.App{
			Title:            "GoPCA Desktop",
			Width:            1200,
			Height:           800,
			WindowStartState: options.Normal,
			AssetServer: &assetserver.Options{
				Assets: assets,
			},
			BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
			OnStartup:        app.startup,
			Bind: []interface{}{
				app,
			},
			Mac: &mac.Options{
				TitleBar: &mac.TitleBar{
					HideTitleBar:    false,
					HideTitle:       false,
					FullSizeContent: false,
					UseToolbar:      false,
				},
			},
			Linux: &linux.Options{
				Icon:                icon,
				WindowIsTranslucent: false,
				WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways,
				ProgramName:         "GoPCA",
			},
		})
	}

	if err != nil {
		println("Error:", err.Error())
	}
}
