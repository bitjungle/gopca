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
	"sort"
)

// Fold is one train/test division of a resampling design. Both slices hold
// 0-based indices in the caller's row space, not positions within the subset
// passed to Split, so a caller that validates a subset of its rows (for example
// only those with an observed response) gets indices it can use directly.
type Fold struct {
	Train []int
	Test  []int
}

// Splitter produces a resampling design over a set of row indices.
//
// Implementations must be deterministic: the same indices and the same
// configuration must always yield the same folds, so that a design recorded
// alongside a result can be reproduced from its parameters alone.
type Splitter interface {
	// Split divides the given row indices into folds. The indices need not be
	// contiguous or sorted; they are the rows eligible for validation.
	Split(indices []int) ([]Fold, error)

	// Name identifies the design in reports and error messages.
	Name() string
}

// groupOf returns the group identifier for a row index. A nil groups slice means
// every row is its own group, which is what makes plain K-fold and leave-one-out
// special cases of the grouped design rather than separate implementations.
func groupOf(groups []int, row int) (int, error) {
	if groups == nil {
		return row, nil
	}
	if row < 0 || row >= len(groups) {
		return 0, fmt.Errorf("row index %d is outside the group assignment of length %d", row, len(groups))
	}
	return groups[row], nil
}

// collectGroups returns the distinct group identifiers present in indices, in
// ascending order, together with the member rows of each.
//
// Ordering by identifier rather than by first appearance matters: it makes the
// unshuffled design depend only on the group labels, so contiguous blocks mean
// the same thing however the caller happened to order its rows.
func collectGroups(indices []int, groups []int) ([]int, map[int][]int, error) {
	members := make(map[int][]int, len(indices))
	for _, row := range indices {
		g, err := groupOf(groups, row)
		if err != nil {
			return nil, nil, err
		}
		members[g] = append(members[g], row)
	}

	ids := make([]int, 0, len(members))
	for g := range members {
		ids = append(ids, g)
	}
	sort.Ints(ids)

	return ids, members, nil
}

// validateIndices rejects the input conditions that would otherwise produce a
// silently malformed design: duplicated rows (which would place one observation
// in two folds) and negative indices.
func validateIndices(indices []int) error {
	if len(indices) == 0 {
		return fmt.Errorf("no rows to split")
	}
	seen := make(map[int]struct{}, len(indices))
	for _, row := range indices {
		if row < 0 {
			return fmt.Errorf("negative row index %d", row)
		}
		if _, dup := seen[row]; dup {
			return fmt.Errorf("duplicate row index %d: an observation cannot belong to two folds", row)
		}
		seen[row] = struct{}{}
	}
	return nil
}

// complement returns every index in indices that is not in test, preserving the
// order of indices.
func complement(indices []int, test []int) []int {
	excluded := make(map[int]struct{}, len(test))
	for _, row := range test {
		excluded[row] = struct{}{}
	}
	train := make([]int, 0, len(indices)-len(test))
	for _, row := range indices {
		if _, skip := excluded[row]; !skip {
			train = append(train, row)
		}
	}
	return train
}
