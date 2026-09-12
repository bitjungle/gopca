package dataquality

import (
	"strings"
	"testing"
)

// grid builds a rows x cols grid where every cell is filled.
func grid(rows, cols int) [][]string {
	out := make([][]string, rows)
	for i := range out {
		out[i] = make([]string, cols)
		for j := range out[i] {
			out[i][j] = "v"
		}
	}
	return out
}

func TestFindSparseRowsIsolatesAStrayRow(t *testing.T) {
	// The shape that prompted this: one row holding a single value in an
	// otherwise fully populated table.
	data := grid(20, 30)
	for j := range data[7] {
		data[7][j] = ""
	}
	data[7][2] = "compression test"

	got := findSparseRows(data, 20, 30)
	if len(got) != 1 || got[0] != 8 {
		t.Fatalf("findSparseRows = %v, want [8] (1-based)", got)
	}
}

func TestFindSparseRowsIgnoresAUniformlySparseDataset(t *testing.T) {
	// The control that matters. A dataset where everything is thin has no
	// stray row, and a rule measuring against a fixed count rather than the
	// dataset's own median would report every row here.
	data := grid(20, 30)
	for i := range data {
		for j := 6; j < 30; j++ {
			data[i][j] = ""
		}
	}
	if got := findSparseRows(data, 20, 30); len(got) != 0 {
		t.Errorf("findSparseRows = %v, want none: every row is equally thin", got)
	}
}

func TestFindSparseRowsReportsCompletelyEmptyRows(t *testing.T) {
	// An empty row is worth reporting whatever the median says.
	data := grid(10, 30)
	for j := range data[4] {
		data[4][j] = ""
	}
	got := findSparseRows(data, 10, 30)
	if len(got) != 1 || got[0] != 5 {
		t.Fatalf("findSparseRows = %v, want [5]", got)
	}
}

func TestFindSparseRowsLeavesSlightlyShortRowsAlone(t *testing.T) {
	// One missing value is missing data, not a stray row. In the dataset that
	// prompted this, 15 rows held 29 of 30 fields and none should be reported.
	data := grid(20, 30)
	for i := 0; i < 15; i++ {
		data[i][29] = ""
	}
	if got := findSparseRows(data, 20, 30); len(got) != 0 {
		t.Errorf("findSparseRows = %v, want none", got)
	}
}

func TestFindSparseRowsSkipsNarrowDatasets(t *testing.T) {
	// Below a few columns "half the typical row" is a cell or two and says
	// nothing, so the check stands down rather than guessing.
	data := [][]string{{"a", "b", "c"}, {"", "", ""}, {"d", "e", "f"}}
	if got := findSparseRows(data, 3, 3); got != nil {
		t.Errorf("findSparseRows = %v, want nil for a 3-column dataset", got)
	}
}

func TestSparseRowsSurfaceAsAQualityIssue(t *testing.T) {
	report := &DataQualityReport{DataProfile: DataProfile{Rows: 100, Columns: 30}}
	issues := generateQualityIssues(report, nil, []int{42})

	var found *QualityIssue
	for i := range issues {
		if issues[i].Category == "structure" {
			found = &issues[i]
		}
	}
	if found == nil {
		t.Fatal("no structure issue raised for a sparse row")
	}
	if !strings.Contains(found.Description, "42") {
		t.Errorf("the message must name the row so the user can find it, got: %q", found.Description)
	}
	if found.Severity != "warning" {
		t.Errorf("severity = %q, want warning", found.Severity)
	}
}

func TestDescribeRowNumbersAbbreviatesLongRuns(t *testing.T) {
	got := describeRowNumbers([]int{1, 2, 3, 4, 5, 6, 7})
	if !strings.Contains(got, "and 2 more") {
		t.Errorf("describeRowNumbers = %q, want an abbreviated tail", got)
	}
	if strings.Contains(got, "7") {
		t.Errorf("describeRowNumbers = %q, should not list past the cap", got)
	}
}

func TestNoSparseRowsRaisesNoIssue(t *testing.T) {
	report := &DataQualityReport{DataProfile: DataProfile{Rows: 100, Columns: 30}}
	for _, issue := range generateQualityIssues(report, nil, nil) {
		if issue.Category == "structure" {
			t.Error("raised a structure issue with no sparse rows")
		}
	}
}

func TestFindSparseRowsLeavesModeratelyThinRowsAlone(t *testing.T) {
	// Guards the *margin*, not just the direction. A row two-thirds populated
	// is a row with missing values, not a stray line, and must not be reported.
	// A rule that flagged everything below the median would catch these.
	data := grid(20, 30)
	for i := 0; i < 6; i++ {
		for j := 20; j < 30; j++ {
			data[i][j] = "" // 20 of 30 filled, median is 30
		}
	}
	if got := findSparseRows(data, 20, 30); len(got) != 0 {
		t.Errorf("findSparseRows = %v, want none: two-thirds populated is missing data, not a stray row", got)
	}
}
