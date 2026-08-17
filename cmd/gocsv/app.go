// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

package main

import (
	"context"

	"github.com/bitjungle/gopca/internal/version"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx               context.Context
	history           *CommandHistory
	currentData       *FileData
	hasUnsavedChanges bool
	pendingZipPath    string // path to a downloaded ZIP awaiting entry selection
}

func (a *App) logInfo(msg string) {
	if a.ctx != nil {
		wailsruntime.LogInfo(a.ctx, msg)
	}
}

func (a *App) logWarning(msg string) {
	if a.ctx != nil {
		wailsruntime.LogWarning(a.ctx, msg)
	}
}

func (a *App) logError(msg string) {
	if a.ctx != nil {
		wailsruntime.LogError(a.ctx, msg)
	}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		history: NewCommandHistory(100), // Keep last 100 commands
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// markDirty marks the app as having unsaved changes and notifies the frontend.
// Only emits an event on the first transition to dirty to avoid noise.
func (a *App) markDirty() {
	if !a.hasUnsavedChanges {
		a.hasUnsavedChanges = true
		if a.ctx != nil {
			wailsruntime.EventsEmit(a.ctx, "unsaved-state-changed", true)
		}
	}
}

// markClean marks the app as having no unsaved changes and notifies the frontend.
// Called after a successful save or when a new file is loaded.
func (a *App) markClean() {
	if a.hasUnsavedChanges {
		a.hasUnsavedChanges = false
		if a.ctx != nil {
			wailsruntime.EventsEmit(a.ctx, "unsaved-state-changed", false)
		}
	}
}

// HasUnsavedChanges returns true if there are unsaved data changes.
// Used by OnBeforeClose to determine whether to prompt the user.
func (a *App) HasUnsavedChanges() bool {
	return a.hasUnsavedChanges
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return version.Get().Short()
}
