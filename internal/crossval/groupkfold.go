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
	"math/rand/v2"
)

// GroupKFold partitions rows into K folds, keeping every member of a group
// together in the same fold.
//
// This one splitter covers the whole family of partition designs; see the
// package documentation for the parameter combinations. In particular
// leave-one-out is K equal to the number of groups, not a separate algorithm.
//
// Groups assigns a group identifier to every row, indexed by absolute row index.
// A nil Groups means each row is its own group, which yields ordinary K-fold and,
// at K equal to the row count, leave-one-out.
//
// Shuffle controls how groups are laid out before they are cut into folds. With
// Shuffle false the groups are taken in ascending identifier order and cut into
// contiguous blocks, which is the conservative design for data whose order
// carries meaning. With Shuffle true the group order is permuted first, giving
// ordinary random K-fold.
//
// Note that groups are cut into contiguous blocks rather than dealt out
// round-robin. Round-robin assignment ("venetian blinds") is common in
// chemometrics software but is actively misleading when rows are sorted by the
// response or acquired in a designed sequence, because every fold then spans the
// full range and the model is never asked to extrapolate.
type GroupKFold struct {
	// K is the number of folds. Zero means "as many folds as there are groups",
	// which is leave-one-out under the default grouping and leave-one-group-out
	// otherwise.
	K int

	// Groups holds a group identifier per row, indexed by absolute row index.
	// Nil means one group per row.
	Groups []int

	// Shuffle permutes the group order before folds are cut.
	Shuffle bool

	// Seed makes the permutation reproducible. It is ignored when Shuffle is false.
	Seed int64
}

// Name identifies the design, distinguishing the configurations that share this
// implementation so that a report says what was actually done.
func (g *GroupKFold) Name() string {
	grouped := "row"
	if g.Groups != nil {
		grouped = "group"
	}
	if g.K == 0 {
		return fmt.Sprintf("leave-one-%s-out", grouped)
	}
	order := "contiguous"
	if g.Shuffle {
		order = "shuffled"
	}
	return fmt.Sprintf("%d-fold by %s (%s)", g.K, grouped, order)
}

// Split divides indices into folds, keeping each group intact.
//
// Every index appears in exactly one test fold, so the union of the test folds
// is the input and the folds are pairwise disjoint. Train is the complement of
// Test within indices.
//
// Algorithm complexity: O(n log n) in the number of indices, dominated by
// ordering the group identifiers.
func (g *GroupKFold) Split(indices []int) ([]Fold, error) {
	if err := validateIndices(indices); err != nil {
		return nil, err
	}

	ids, members, err := collectGroups(indices, g.Groups)
	if err != nil {
		return nil, err
	}
	nGroups := len(ids)

	k, err := g.resolveK(nGroups)
	if err != nil {
		return nil, err
	}

	if g.Shuffle {
		// A locally seeded generator, never the global source, so that concurrent
		// callers cannot perturb one another's designs.
		r := rand.New(rand.NewPCG(uint64(g.Seed), 0x9E3779B97F4A7C15))
		r.Shuffle(nGroups, func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
	}

	folds := make([]Fold, 0, k)
	for i := 0; i < k; i++ {
		// Contiguous blocks sized as evenly as integer arithmetic allows: with
		// nGroups not divisible by k the earlier folds take the extra groups.
		start := i * nGroups / k
		end := (i + 1) * nGroups / k

		test := make([]int, 0, len(indices)/k+1)
		for _, id := range ids[start:end] {
			test = append(test, members[id]...)
		}

		folds = append(folds, Fold{
			Train: complement(indices, test),
			Test:  test,
		})
	}

	return folds, nil
}

// resolveK turns the requested fold count into a concrete one, refusing the
// requests that cannot be honoured rather than clamping them.
//
// Silently reducing K to the group count would report a design the caller did
// not ask for: someone requesting 10-fold on 4 batches is reasoning about the
// wrong effective sample size, and a quiet clamp hides that from them.
func (g *GroupKFold) resolveK(nGroups int) (int, error) {
	if nGroups < 2 {
		return 0, fmt.Errorf("cross-validation needs at least 2 groups, found %d: "+
			"with a single group there is nothing to hold out", nGroups)
	}
	if g.K == 0 {
		return nGroups, nil
	}
	if g.K < 0 {
		return 0, fmt.Errorf("fold count must not be negative, got %d", g.K)
	}
	if g.K == 1 {
		return 0, fmt.Errorf("fold count must be at least 2, got 1: " +
			"a single fold leaves no data to train on")
	}
	if g.K > nGroups {
		unit := "rows"
		if g.Groups != nil {
			unit = "groups"
		}
		return 0, fmt.Errorf("cannot make %d folds from %d %s: "+
			"the effective sample size is the number of %s, not the number of rows. "+
			"Use at most %d folds, or 0 for leave-one-out",
			g.K, nGroups, unit, unit, nGroups)
	}
	return g.K, nil
}
