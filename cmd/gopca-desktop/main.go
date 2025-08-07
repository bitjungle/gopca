// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
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
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Parse command-line flags
	openFile := flag.String("open", "", "CSV file to open on startup")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Println(version.Get().Short())
		os.Exit(0)
	}

	// Create an instance of the app structure
	app := NewApp()

	// Pass the file to open if provided
	if *openFile != "" {
		app.SetFileToOpen(*openFile)
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "GoPCA Desktop",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
