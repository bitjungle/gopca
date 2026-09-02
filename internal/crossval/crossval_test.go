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

package crossval

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// seq returns indices 0..n-1, the common case where every row is eligible.
func seq(n int) []int {
	s := make([]int, n)
	for i := range s {
		s[i] = i
	}
	return s
}

// assertPartition checks the property every partition design must have: the test
// folds are pairwise disjoint and together cover exactly the input, and each
// fold's Train is precisely the complement of its Test.
//
// This is the check that goes red if fold arithmetic drops or duplicates rows,
// which is otherwise invisible because an error estimate computed over slightly
// the wrong set still looks entirely reasonable.
func assertPartition(t *testing.T, indices []int, folds []Fold) {
	t.Helper()

	want := make(map[int]struct{}, len(indices))
	for _, i := range indices {
		want[i] = struct{}{}
	}

	seen := make(map[int]int, len(indices))
	for f, fold := range folds {
		for _, row := range fold.Test {
			if prev, dup := seen[row]; dup {
				t.Errorf("row %d appears in test folds %d and %d", row, prev, f)
			}
			seen[row] = f
		}

		inTest := make(map[int]struct{}, len(fold.Test))
		for _, row := range fold.Test {
			inTest[row] = struct{}{}
		}
		if len(fold.Train)+len(fold.Test) != len(indices) {
			t.Errorf("fold %d: train %d + test %d != %d input rows",
				f, len(fold.Train), len(fold.Test), len(indices))
		}
		for _, row := range fold.Train {
			if _, bad := inTest[row]; bad {
				t.Errorf("fold %d: row %d is in both train and test", f, row)
			}
			if _, ok := want[row]; !ok {
				t.Errorf("fold %d: train row %d was not in the input", f, row)
			}
		}
	}

	if len(seen) != len(want) {
		t.Errorf("test folds cover %d rows, want %d", len(seen), len(want))
	}
	for row := range want {
		if _, ok := seen[row]; !ok {
			t.Errorf("row %d never appears in any test fold", row)
		}
	}
}

func TestGroupKFoldPartitions(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		splitter  GroupKFold
		wantFolds int
	}{
		{"5-fold shuffled", 20, GroupKFold{K: 5, Shuffle: true, Seed: 1}, 5},
		{"5-fold contiguous", 20, GroupKFold{K: 5}, 5},
		{"10-fold uneven", 23, GroupKFold{K: 10, Shuffle: true, Seed: 7}, 10},
		{"leave-one-out via K=0", 12, GroupKFold{K: 0}, 12},
		{"leave-one-out via K=n", 12, GroupKFold{K: 12}, 12},
		{"2-fold minimum", 5, GroupKFold{K: 2}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indices := seq(tt.n)
			folds, err := tt.splitter.Split(indices)
			if err != nil {
				t.Fatalf("Split: %v", err)
			}
			if len(folds) != tt.wantFolds {
				t.Errorf("got %d folds, want %d", len(folds), tt.wantFolds)
			}
			assertPartition(t, indices, folds)
		})
	}
}

// TestGroupKFoldLeaveOneOutEquivalence pins the design decision that
// leave-one-out is K-fold at K = n rather than a separate algorithm. If the two
// ever diverge, the package documentation is lying about what LOO means here.
func TestGroupKFoldLeaveOneOutEquivalence(t *testing.T) {
	indices := seq(9)

	byZero, err := (&GroupKFold{K: 0}).Split(indices)
	if err != nil {
		t.Fatalf("K=0: %v", err)
	}
	byN, err := (&GroupKFold{K: len(indices)}).Split(indices)
	if err != nil {
		t.Fatalf("K=n: %v", err)
	}

	if !reflect.DeepEqual(byZero, byN) {
		t.Error("K=0 and K=n produced different designs; leave-one-out must be K-fold at K=n")
	}
	for i, fold := range byZero {
		if len(fold.Test) != 1 {
			t.Errorf("fold %d holds out %d rows, leave-one-out must hold out exactly 1", i, len(fold.Test))
		}
	}
}

// TestGroupKFoldKeepsGroupsIntact is the check that matters for replicate data.
// A group split across folds lets a model interpolate between near-duplicates and
// reports an error lower than deployment will deliver, which no ordinary
// assertion would catch because the numbers merely look good.
func TestGroupKFoldKeepsGroupsIntact(t *testing.T) {
	// 12 rows, 4 objects measured 3 times each.
	groups := make([]int, 12)
	for row := range groups {
		groups[row] = row / 3
	}
	indices := seq(12)

	for _, shuffle := range []bool{false, true} {
		t.Run(fmt.Sprintf("shuffle=%v", shuffle), func(t *testing.T) {
			folds, err := (&GroupKFold{K: 2, Groups: groups, Shuffle: shuffle, Seed: 3}).Split(indices)
			if err != nil {
				t.Fatalf("Split: %v", err)
			}
			assertPartition(t, indices, folds)

			for f, fold := range folds {
				inTest := map[int]bool{}
				for _, row := range fold.Test {
					inTest[groups[row]] = true
				}
				for _, row := range fold.Train {
					if inTest[groups[row]] {
						t.Errorf("fold %d: group %d appears in both train and test (row %d)",
							f, groups[row], row)
					}
				}
			}
		})
	}
}

func TestGroupKFoldLeaveOneGroupOut(t *testing.T) {
	groups := []int{0, 0, 1, 1, 1, 2, 2, 3}
	folds, err := (&GroupKFold{K: 0, Groups: groups}).Split(seq(8))
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if len(folds) != 4 {
		t.Fatalf("got %d folds, want one per group (4)", len(folds))
	}
	assertPartition(t, seq(8), folds)

	for f, fold := range folds {
		g := map[int]struct{}{}
		for _, row := range fold.Test {
			g[groups[row]] = struct{}{}
		}
		if len(g) != 1 {
			t.Errorf("fold %d holds out %d groups, leave-one-group-out must hold out exactly 1", f, len(g))
		}
	}
}

// TestGroupKFoldDeterminism backs the claim that a recorded seed reproduces a
// design exactly, which is what makes a CVReport auditable.
func TestGroupKFoldDeterminism(t *testing.T) {
	indices := seq(30)
	a, err := (&GroupKFold{K: 5, Shuffle: true, Seed: 42}).Split(indices)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	b, err := (&GroupKFold{K: 5, Shuffle: true, Seed: 42}).Split(indices)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Error("same seed produced different folds")
	}

	c, err := (&GroupKFold{K: 5, Shuffle: true, Seed: 43}).Split(indices)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if reflect.DeepEqual(a, c) {
		t.Error("different seeds produced identical folds; the seed is not reaching the shuffle")
	}
}

// TestGroupKFoldContiguousBlocks checks that the unshuffled design really is
// contiguous blocks and not round-robin. The distinction matters: round-robin
// assignment on data sorted by the response gives every fold the full range and
// never asks the model to extrapolate.
func TestGroupKFoldContiguousBlocks(t *testing.T) {
	folds, err := (&GroupKFold{K: 3}).Split(seq(9))
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	want := [][]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}}
	for i, fold := range folds {
		got := append([]int(nil), fold.Test...)
		sort.Ints(got)
		if !reflect.DeepEqual(got, want[i]) {
			t.Errorf("fold %d test = %v, want contiguous block %v", i, got, want[i])
		}
	}
}

// TestGroupKFoldNonContiguousIndices covers the semi-supervised case, where only
// the rows carrying an observed response are eligible for validation. Folds must
// come back in the caller's row space, not as positions within the subset.
func TestGroupKFoldNonContiguousIndices(t *testing.T) {
	labelled := []int{3, 7, 11, 15, 19, 23}
	folds, err := (&GroupKFold{K: 3}).Split(labelled)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	assertPartition(t, labelled, folds)

	allowed := map[int]struct{}{}
	for _, i := range labelled {
		allowed[i] = struct{}{}
	}
	for f, fold := range folds {
		for _, row := range append(append([]int(nil), fold.Train...), fold.Test...) {
			if _, ok := allowed[row]; !ok {
				t.Errorf("fold %d contains row %d, which was not among the eligible indices", f, row)
			}
		}
	}
}

func TestGroupKFoldErrors(t *testing.T) {
	tests := []struct {
		name     string
		indices  []int
		splitter GroupKFold
	}{
		{"more folds than rows", seq(4), GroupKFold{K: 10}},
		{"more folds than groups", seq(8), GroupKFold{K: 5, Groups: []int{0, 0, 0, 0, 1, 1, 1, 1}}},
		{"single fold", seq(6), GroupKFold{K: 1}},
		{"negative folds", seq(6), GroupKFold{K: -2}},
		{"no rows", []int{}, GroupKFold{K: 2}},
		{"one group only", seq(4), GroupKFold{K: 0, Groups: []int{5, 5, 5, 5}}},
		{"duplicate index", []int{0, 1, 1, 2}, GroupKFold{K: 2}},
		{"negative index", []int{0, -1, 2}, GroupKFold{K: 2}},
		{"group index out of range", seq(4), GroupKFold{K: 2, Groups: []int{0, 1}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.splitter.Split(tt.indices); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}
}

// TestForwardChainingNeverTrainsOnTheFuture is the property that makes this
// design worth having. If a training index ever sits later in the series than a
// test index, the estimate describes a situation that cannot occur in deployment.
func TestForwardChainingNeverTrainsOnTheFuture(t *testing.T) {
	indices := seq(20)
	folds, err := (&ForwardChaining{Splits: 4}).Split(indices)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if len(folds) != 4 {
		t.Fatalf("got %d folds, want 4", len(folds))
	}

	position := make(map[int]int, len(indices))
	for pos, row := range indices {
		position[row] = pos
	}

	for f, fold := range folds {
		if len(fold.Train) == 0 || len(fold.Test) == 0 {
			t.Fatalf("fold %d has empty train (%d) or test (%d)", f, len(fold.Train), len(fold.Test))
		}
		earliestTest := len(indices)
		for _, row := range fold.Test {
			if position[row] < earliestTest {
				earliestTest = position[row]
			}
		}
		for _, row := range fold.Train {
			if position[row] >= earliestTest {
				t.Errorf("fold %d trains on position %d, at or after the earliest test position %d",
					f, position[row], earliestTest)
			}
		}
	}
}

// TestForwardChainingNestsTrainingSets checks the defining structure: each
// training prefix contains the previous one.
func TestForwardChainingNestsTrainingSets(t *testing.T) {
	folds, err := (&ForwardChaining{Splits: 3}).Split(seq(16))
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	for i := 1; i < len(folds); i++ {
		prev, cur := folds[i-1].Train, folds[i].Train
		if len(cur) <= len(prev) {
			t.Errorf("fold %d training set (%d) did not grow beyond fold %d (%d)",
				i, len(cur), i-1, len(prev))
		}
		for j, row := range prev {
			if cur[j] != row {
				t.Errorf("fold %d training set is not a superset of fold %d at position %d", i, i-1, j)
				break
			}
		}
	}
}

func TestForwardChainingRespectsMinTrain(t *testing.T) {
	folds, err := (&ForwardChaining{Splits: 4, MinTrain: 10}).Split(seq(20))
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	for f, fold := range folds {
		if len(fold.Train) < 10 {
			t.Errorf("fold %d trains on %d rows, below the requested minimum of 10", f, len(fold.Train))
		}
	}
}

func TestForwardChainingErrors(t *testing.T) {
	tests := []struct {
		name     string
		indices  []int
		splitter ForwardChaining
	}{
		{"zero splits", seq(10), ForwardChaining{Splits: 0}},
		{"negative splits", seq(10), ForwardChaining{Splits: -1}},
		{"more splits than observations", seq(3), ForwardChaining{Splits: 9}},
		{"min train exceeds series", seq(10), ForwardChaining{Splits: 2, MinTrain: 10}},
		{"negative min train", seq(10), ForwardChaining{Splits: 2, MinTrain: -1}},
		{"no rows", []int{}, ForwardChaining{Splits: 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.splitter.Split(tt.indices); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}
}

func TestSplitterNames(t *testing.T) {
	tests := []struct {
		splitter Splitter
		want     string
	}{
		{&GroupKFold{K: 5, Shuffle: true}, "5-fold by row (shuffled)"},
		{&GroupKFold{K: 5}, "5-fold by row (contiguous)"},
		{&GroupKFold{K: 0}, "leave-one-row-out"},
		{&GroupKFold{K: 0, Groups: []int{0}}, "leave-one-group-out"},
		{&ForwardChaining{Splits: 3}, "forward-chaining (3 splits)"},
	}
	for _, tt := range tests {
		if got := tt.splitter.Name(); got != tt.want {
			t.Errorf("Name() = %q, want %q", got, tt.want)
		}
	}
}
