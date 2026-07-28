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

package profiling

import (
	"runtime"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes uint64
		want  string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"just below KB", 1023, "1023 B"},
		{"exactly 1 KB", 1024, "1.00 KB"},
		{"1.5 KB", 1536, "1.50 KB"},
		{"just below MB", 1024*1024 - 1, "1024.00 KB"},
		{"exactly 1 MB", 1024 * 1024, "1.00 MB"},
		{"exactly 1 GB", 1024 * 1024 * 1024, "1.00 GB"},
		{"2.5 GB", 2560 * 1024 * 1024, "2.50 GB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatBytes(tt.bytes); got != tt.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestEstimateMatrixMemory(t *testing.T) {
	tests := []struct {
		name       string
		rows, cols int
		want       uint64
	}{
		// dataSize = rows*cols*8, sliceOverhead = rows*24
		{"empty", 0, 0, 0},
		{"1x1", 1, 1, 8 + 24},
		{"10x10", 10, 10, 10*10*8 + 10*24},
		{"100x50", 100, 50, 100*50*8 + 100*24},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EstimateMatrixMemory(tt.rows, tt.cols); got != tt.want {
				t.Errorf("EstimateMatrixMemory(%d, %d) = %d, want %d", tt.rows, tt.cols, got, tt.want)
			}
		})
	}
}

func TestMemoryProfilerDisabled(t *testing.T) {
	// With profiling disabled (default), the lifecycle is a no-op and Stop
	// returns a zero-valued summary.
	t.Setenv("GOPCA_PROFILE", "0")
	mp := NewMemoryProfiler()
	mp.Start("op")
	mp.Checkpoint("mid")
	summary := mp.Stop()
	if len(summary.Measurements) != 0 {
		t.Errorf("disabled profiler recorded %d measurements, want 0", len(summary.Measurements))
	}
}

func TestMemoryProfilerLifecycle(t *testing.T) {
	// enabled is read from the environment at construction time.
	t.Setenv("GOPCA_PROFILE", "1")
	mp := NewMemoryProfiler()

	mp.Start("op")
	// Allocate something so there is measurable activity.
	sink := make([]byte, 1<<20)
	for i := range sink {
		sink[i] = byte(i)
	}
	mp.Checkpoint("after-alloc")
	summary := mp.Stop()

	// start + checkpoint snapshots must be recorded.
	if len(summary.Measurements) < 2 {
		t.Errorf("got %d measurements, want at least 2", len(summary.Measurements))
	}
	if summary.PeakAlloc == 0 {
		t.Error("PeakAlloc should be non-zero when profiling is enabled")
	}
	// Keep sink alive until after Stop so the allocation is observable.
	runtime.KeepAlive(sink)
}

func TestProfileFunc(t *testing.T) {
	t.Setenv("GOPCA_PROFILE", "1")
	called := false
	summary := ProfileFunc("work", func() {
		called = true
		_ = make([]byte, 1<<16)
	})
	if !called {
		t.Error("ProfileFunc did not invoke the supplied function")
	}
	if len(summary.Measurements) == 0 {
		t.Error("ProfileFunc returned no measurements with profiling enabled")
	}
}
