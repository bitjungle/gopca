// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package profiling

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// MemoryProfiler tracks memory usage during operations
type MemoryProfiler struct {
	startMem     runtime.MemStats
	currentMem   runtime.MemStats
	peakAlloc    uint64
	measurements []MemorySnapshot
	enabled      bool
}

// MemorySnapshot represents memory state at a point in time
type MemorySnapshot struct {
	Label       string
	Timestamp   time.Time
	AllocBytes  uint64
	TotalAlloc  uint64
	HeapAlloc   uint64
	HeapObjects uint64
	NumGC       uint32
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler() *MemoryProfiler {
	return &MemoryProfiler{
		enabled:      os.Getenv("GOPCA_PROFILE") == "1",
		measurements: make([]MemorySnapshot, 0),
	}
}

// Start begins memory profiling
func (mp *MemoryProfiler) Start(label string) {
	if !mp.enabled {
		return
	}

	runtime.GC() // Force GC to get accurate baseline
	runtime.ReadMemStats(&mp.startMem)
	mp.peakAlloc = mp.startMem.Alloc

	mp.measurements = append(mp.measurements, MemorySnapshot{
		Label:       label + "_start",
		Timestamp:   time.Now(),
		AllocBytes:  mp.startMem.Alloc,
		TotalAlloc:  mp.startMem.TotalAlloc,
		HeapAlloc:   mp.startMem.HeapAlloc,
		HeapObjects: mp.startMem.HeapObjects,
		NumGC:       mp.startMem.NumGC,
	})

	if os.Getenv("GOPCA_DEBUG") == "1" {
		fmt.Printf("[PROFILE] Memory profiling started: %s\n", label)
		fmt.Printf("[PROFILE] Initial memory: %.2f MB\n", float64(mp.startMem.Alloc)/1024/1024)
	}
}

// Checkpoint records memory state at a specific point
func (mp *MemoryProfiler) Checkpoint(label string) {
	if !mp.enabled {
		return
	}

	runtime.ReadMemStats(&mp.currentMem)

	if mp.currentMem.Alloc > mp.peakAlloc {
		mp.peakAlloc = mp.currentMem.Alloc
	}

	mp.measurements = append(mp.measurements, MemorySnapshot{
		Label:       label,
		Timestamp:   time.Now(),
		AllocBytes:  mp.currentMem.Alloc,
		TotalAlloc:  mp.currentMem.TotalAlloc,
		HeapAlloc:   mp.currentMem.HeapAlloc,
		HeapObjects: mp.currentMem.HeapObjects,
		NumGC:       mp.currentMem.NumGC,
	})

	if os.Getenv("GOPCA_DEBUG") == "1" {
		deltaBytes := int64(mp.currentMem.Alloc) - int64(mp.startMem.Alloc)
		fmt.Printf("[PROFILE] Checkpoint %s: %.2f MB (delta: %+.2f MB)\n",
			label,
			float64(mp.currentMem.Alloc)/1024/1024,
			float64(deltaBytes)/1024/1024)
	}
}

// Stop ends profiling and returns summary
func (mp *MemoryProfiler) Stop() MemorySummary {
	if !mp.enabled {
		return MemorySummary{}
	}

	runtime.ReadMemStats(&mp.currentMem)

	summary := MemorySummary{
		InitialAlloc:   mp.startMem.Alloc,
		FinalAlloc:     mp.currentMem.Alloc,
		PeakAlloc:      mp.peakAlloc,
		TotalAllocated: mp.currentMem.TotalAlloc - mp.startMem.TotalAlloc,
		NumGCs:         mp.currentMem.NumGC - mp.startMem.NumGC,
		Measurements:   mp.measurements,
	}

	if os.Getenv("GOPCA_DEBUG") == "1" {
		fmt.Printf("[PROFILE] Memory profiling stopped\n")
		fmt.Printf("[PROFILE] Peak memory: %.2f MB\n", float64(mp.peakAlloc)/1024/1024)
		fmt.Printf("[PROFILE] Total allocated: %.2f MB\n", float64(summary.TotalAllocated)/1024/1024)
		fmt.Printf("[PROFILE] GC runs: %d\n", summary.NumGCs)
	}

	return summary
}

// MemorySummary contains profiling results
type MemorySummary struct {
	InitialAlloc   uint64
	FinalAlloc     uint64
	PeakAlloc      uint64
	TotalAllocated uint64
	NumGCs         uint32
	Measurements   []MemorySnapshot
}

// WriteHeapProfile writes a heap profile to file
func WriteHeapProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create heap profile: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			// Best effort, file may have already been closed
			_ = err
		}
	}()

	runtime.GC() // Force GC before heap profile
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write heap profile: %w", err)
	}

	return nil
}

// ProfileFunc profiles memory usage of a function
func ProfileFunc(name string, fn func()) MemorySummary {
	profiler := NewMemoryProfiler()
	profiler.Start(name)
	fn()
	return profiler.Stop()
}

// EstimateMatrixMemory estimates memory usage for a matrix
func EstimateMatrixMemory(rows, cols int) uint64 {
	// Each float64 is 8 bytes
	// Slice overhead is approximately 24 bytes per slice
	dataSize := uint64(rows * cols * 8)
	sliceOverhead := uint64(rows * 24)
	return dataSize + sliceOverhead
}

// FormatBytes formats bytes in human-readable format
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
