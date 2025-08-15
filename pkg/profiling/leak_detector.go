// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package profiling

import (
	"fmt"
	"runtime"
	"time"
)

// LeakDetector detects goroutine and memory leaks
type LeakDetector struct {
	initialGoroutines int
	initialMem        runtime.MemStats
	checkpoints       []LeakCheckpoint
}

// LeakCheckpoint represents a leak detection checkpoint
type LeakCheckpoint struct {
	Label      string
	Timestamp  time.Time
	Goroutines int
	MemAlloc   uint64
	Leaked     bool
	Message    string
}

// NewLeakDetector creates a new leak detector
func NewLeakDetector() *LeakDetector {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &LeakDetector{
		initialGoroutines: runtime.NumGoroutine(),
		initialMem:        m,
		checkpoints:       make([]LeakCheckpoint, 0),
	}
}

// CheckGoroutines checks for goroutine leaks
func (ld *LeakDetector) CheckGoroutines(label string) LeakCheckpoint {
	current := runtime.NumGoroutine()
	leaked := current > ld.initialGoroutines

	checkpoint := LeakCheckpoint{
		Label:      label,
		Timestamp:  time.Now(),
		Goroutines: current,
		Leaked:     leaked,
	}

	if leaked {
		checkpoint.Message = fmt.Sprintf("Goroutine leak detected: %d goroutines (started with %d)",
			current, ld.initialGoroutines)
	}

	ld.checkpoints = append(ld.checkpoints, checkpoint)
	return checkpoint
}

// CheckMemory checks for memory leaks after GC
func (ld *LeakDetector) CheckMemory(label string, threshold float64) LeakCheckpoint {
	// Force multiple GC cycles to ensure cleanup
	runtime.GC()
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate memory growth
	growth := float64(m.Alloc) / float64(ld.initialMem.Alloc)
	leaked := growth > threshold

	checkpoint := LeakCheckpoint{
		Label:     label,
		Timestamp: time.Now(),
		MemAlloc:  m.Alloc,
		Leaked:    leaked,
	}

	if leaked {
		checkpoint.Message = fmt.Sprintf("Memory leak suspected: %.2fx growth (%.2f MB -> %.2f MB)",
			growth,
			float64(ld.initialMem.Alloc)/1024/1024,
			float64(m.Alloc)/1024/1024)
	}

	ld.checkpoints = append(ld.checkpoints, checkpoint)
	return checkpoint
}

// GetReport returns a summary of all checkpoints
func (ld *LeakDetector) GetReport() LeakReport {
	var leaks []LeakCheckpoint
	for _, cp := range ld.checkpoints {
		if cp.Leaked {
			leaks = append(leaks, cp)
		}
	}

	return LeakReport{
		InitialGoroutines: ld.initialGoroutines,
		InitialMemory:     ld.initialMem.Alloc,
		Checkpoints:       ld.checkpoints,
		Leaks:             leaks,
		HasLeaks:          len(leaks) > 0,
	}
}

// LeakReport summarizes leak detection results
type LeakReport struct {
	InitialGoroutines int
	InitialMemory     uint64
	Checkpoints       []LeakCheckpoint
	Leaks             []LeakCheckpoint
	HasLeaks          bool
}

// DetectLeaksInFunc runs a function and checks for leaks
func DetectLeaksInFunc(name string, fn func()) LeakReport {
	detector := NewLeakDetector()

	// Run the function
	fn()

	// Wait for goroutines to finish
	time.Sleep(100 * time.Millisecond)

	// Check for leaks
	detector.CheckGoroutines(name + "_goroutines")
	detector.CheckMemory(name+"_memory", 2.0) // 2x growth threshold

	return detector.GetReport()
}

// MonitorGoroutines monitors goroutine count over time
func MonitorGoroutines(duration time.Duration, interval time.Duration) []int {
	counts := make([]int, 0)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeout := time.After(duration)

	for {
		select {
		case <-ticker.C:
			counts = append(counts, runtime.NumGoroutine())
		case <-timeout:
			return counts
		}
	}
}
