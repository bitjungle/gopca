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

package core

import (
	"fmt"
	"math"
)

// Advisories about the response that the fit itself cannot express.
//
// These live here, rather than in either front end, because there are two front
// ends. The categorical-response caution was written for the CLI alone, so
// GoPCA Desktop offered the same column in its response picker and said nothing
// — and the desktop is where a reader is likelier to make the mistake, because
// picking from a dropdown takes no thought about what the column holds. A
// caution that only half the product gives is close to no caution at all.
//
// Keeping the text here also means the two cannot drift apart into two
// differently worded, differently scoped versions of the same warning.

// discreteResponseLimit is the largest number of distinct values a response may
// take before it stops looking like a class code. Ten covers the usual encodings
// while leaving genuinely coarse measurements alone.
const discreteResponseLimit = 10

// ResponseAdvisories returns cautions about using y as a regression response,
// as plain unwrapped sentences for the caller to present in its own medium.
//
// They are advisory rather than refusals throughout: each describes something
// the data cannot distinguish from a legitimate case, and the reader may well
// know something the file does not record.
//
// The result is nil when there is nothing to say, so a caller can range over it
// unconditionally.
func ResponseAdvisories(name string, y []float64) []string {
	var out []string
	if a := categoricalResponseAdvisory(name, y); a != "" {
		out = append(out, a)
	}
	return out
}

// categoricalResponseAdvisory flags a response that is probably a class label
// stored as a number.
//
// A column holding 0, 1 and 2 for three species parses as numeric and regresses
// without complaint, but the fit asserts that the classes are ordered and
// equally spaced, which is false. testdata/iris/iris.csv ships exactly this as
// species#target, so it is the first thing a reader is likely to try.
//
// Two conditions must hold: few enough distinct values to look like codes, and
// enough rows per value that the coarseness is a property of the column rather
// than of a tiny dataset. Ten rows per distinct value is the threshold; below it
// a genuine measurement on twenty samples would be flagged constantly.
func categoricalResponseAdvisory(name string, y []float64) string {
	distinct := make(map[float64]struct{}, discreteResponseLimit+1)
	observed := 0
	for _, v := range y {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		observed++
		if len(distinct) <= discreteResponseLimit {
			distinct[v] = struct{}{}
		}
	}

	if observed == 0 || len(distinct) > discreteResponseLimit || len(distinct)*10 > observed {
		return ""
	}

	return fmt.Sprintf(
		"%q takes only %d distinct values across %d rows, which is what a class label "+
			"encoded as a number looks like. Regression treats those values as ordered and "+
			"equally spaced. If they are categories the fit is meaningless; predicting a "+
			"category is classification, which this tool does not do.",
		name, len(distinct), observed)
}
