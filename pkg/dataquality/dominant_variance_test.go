package dataquality

import (
	"strings"
	"testing"
)

func col(name string, sd float64) ColumnAnalysis {
	return ColumnAnalysis{Name: name, Type: "numeric", Stats: ColumnStatistics{StdDev: &sd}}
}

func TestDominantVarianceColumnNamesTheOffender(t *testing.T) {
	// The shape that prompted this: a category code stored as an integer beside
	// columns of weight fractions.
	r := &DataQualityReport{ColumnAnalysis: []ColumnAnalysis{
		col("proc_num", 3.54), col("Al", 0.041), col("Si", 0.027), col("Zn", 0.025),
	}}
	name, share := dominantVarianceColumn(r)
	if name != "proc_num" {
		t.Fatalf("name = %q, want proc_num", name)
	}
	if share < 99 {
		t.Errorf("share = %.2f, want above 99", share)
	}
}

func TestNoColumnDominatesWhenVarianceIsShared(t *testing.T) {
	// The control that matters most, and it has to use realistic proportions. On
	// this dataset once the code column is held out, aluminium carries 45.4% of
	// the variance -- high, and legitimate chemistry rather than a mistake. An
	// earlier version of this fixture used only four columns, which inflated
	// aluminium to 50.04% and made the test fail against correct code.
	r := &DataQualityReport{ColumnAnalysis: []ColumnAnalysis{
		col("Al", 6.7387),
		col("Si", 4.4609),
		col("Zn", 4.1413),
		col("Cu", 3.0299),
		col("Mg", 2.0248),
		col("rest", 2.0640),
	}}
	if name, share := dominantVarianceColumn(r); name != "" {
		t.Errorf("named %q at %.1f%%, want nothing: no column holds half the variance", name, share)
	}
}

func TestEqualVarianceNamesNothing(t *testing.T) {
	var cols []ColumnAnalysis
	for _, n := range []string{"a", "b", "c", "d", "e"} {
		cols = append(cols, col(n, 1.0))
	}
	if name, _ := dominantVarianceColumn(&DataQualityReport{ColumnAnalysis: cols}); name != "" {
		t.Errorf("named %q on a perfectly balanced dataset", name)
	}
}

func TestTargetAndCategoryColumnsAreNotConsidered(t *testing.T) {
	// They are already held out, so their variance cannot dominate the analysis.
	// One held-out column, not two: two of equal size split the variance exactly
	// in half and neither crosses the threshold, so the test would pass whether
	// or not the filter was there. Found by mutation.
	huge := 1000.0
	r := &DataQualityReport{ColumnAnalysis: []ColumnAnalysis{
		{Name: "code", Type: "categorical", Stats: ColumnStatistics{StdDev: &huge}},
		col("a", 1.0), col("b", 1.0), col("c", 1.0),
	}}
	if name, _ := dominantVarianceColumn(r); name != "" {
		t.Errorf("named %q, but held-out columns are not analysis variables", name)
	}
}

func TestSingleColumnIsNotDominant(t *testing.T) {
	// One column is trivially 100% of the variance and there is nothing to
	// advise: it is the whole analysis by definition.
	r := &DataQualityReport{ColumnAnalysis: []ColumnAnalysis{col("only", 5)}}
	if name, _ := dominantVarianceColumn(r); name != "" {
		t.Errorf("named %q in a one-column dataset", name)
	}
}

func TestZeroVarianceEverywhereNamesNothing(t *testing.T) {
	r := &DataQualityReport{ColumnAnalysis: []ColumnAnalysis{col("a", 0), col("b", 0)}}
	if name, _ := dominantVarianceColumn(r); name != "" {
		t.Errorf("named %q when no column varies at all", name)
	}
}

func TestRecommendationNamesTheColumnAndQuotesTheShare(t *testing.T) {
	r := &DataQualityReport{
		DataProfile:    DataProfile{Rows: 100, Columns: 4, NumericColumns: 4},
		ColumnAnalysis: []ColumnAnalysis{col("proc_num", 3.54), col("Al", 0.041), col("Si", 0.027)},
	}
	var found *Recommendation
	for i, rec := range generateRecommendations(r) {
		if rec.Category == "scaling" && len(rec.Columns) == 1 && rec.Columns[0] == "proc_num" {
			found = &generateRecommendations(r)[i]
		}
	}
	if found == nil {
		t.Fatal("no recommendation named proc_num")
	}
	if !strings.Contains(found.Description, "%") {
		t.Errorf("description does not quote a share: %q", found.Description)
	}
	if !strings.Contains(found.Description, "#category") {
		t.Errorf("description does not say what to do about it: %q", found.Description)
	}
}
