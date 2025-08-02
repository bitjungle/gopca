package types

import (
	"math"
	"strings"
	"testing"
)

func TestParseCSVMixedWithTargets(t *testing.T) {
	tests := []struct {
		name           string
		csvContent     string
		targetColumns  []string
		wantDataCols   int
		wantCatCols    int
		wantTargetCols int
	}{
		{
			name: "column with #target suffix (no space)",
			csvContent: `feature1,feature2,category,value#target
1.0,2.0,A,10.5
3.0,4.0,B,20.3
5.0,6.0,A,15.7`,
			targetColumns:  nil,
			wantDataCols:   2, // feature1, feature2
			wantCatCols:    1, // category
			wantTargetCols: 1, // value#target
		},
		{
			name: "column with #target suffix (with space)",
			csvContent: `feature1,feature2,category,value #target
1.0,2.0,A,10.5
3.0,4.0,B,20.3
5.0,6.0,A,15.7`,
			targetColumns:  nil,
			wantDataCols:   2, // feature1, feature2
			wantCatCols:    1, // category
			wantTargetCols: 1, // value #target
		},
		{
			name: "explicit target column",
			csvContent: `x,y,group,score
1.0,2.0,A,10.5
3.0,4.0,B,20.3
5.0,6.0,A,15.7`,
			targetColumns:  []string{"score"},
			wantDataCols:   2, // x, y
			wantCatCols:    1, // group
			wantTargetCols: 1, // score
		},
		{
			name: "mixed targets",
			csvContent: `feature1,feature2,category,manual_target,auto#target
1.0,2.0,A,10.5,100
3.0,4.0,B,20.3,200
5.0,6.0,A,15.7,300`,
			targetColumns:  []string{"manual_target"},
			wantDataCols:   2, // feature1, feature2
			wantCatCols:    1, // category
			wantTargetCols: 2, // manual_target, auto#target
		},
		{
			name: "target column with empty values in first rows",
			csvContent: `feature1,feature2,sparse#target,dense#target
1.0,2.0,,100
2.0,3.0,,200
3.0,4.0,,300
4.0,5.0,,400
5.0,6.0,,500
6.0,7.0,,600
7.0,8.0,,700
8.0,9.0,,800
9.0,10.0,,900
10.0,11.0,,1000
11.0,12.0,42.5,1100
12.0,13.0,43.5,1200`,
			targetColumns:  nil,
			wantDataCols:   2, // feature1, feature2
			wantCatCols:    0,
			wantTargetCols: 2, // sparse#target, dense#target
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := DefaultCSVFormat()
			format.HasRowNames = false // Disable row names for this test
			data, catData, targetData, err := ParseCSVMixedWithTargets(
				strings.NewReader(tt.csvContent),
				format,
				tt.targetColumns,
			)

			if err != nil {
				t.Fatalf("ParseCSVMixedWithTargets() error = %v", err)
			}

			if data.Columns != tt.wantDataCols {
				t.Errorf("got %d data columns, want %d", data.Columns, tt.wantDataCols)
			}

			if len(catData) != tt.wantCatCols {
				t.Errorf("got %d categorical columns, want %d", len(catData), tt.wantCatCols)
			}

			if len(targetData) != tt.wantTargetCols {
				t.Errorf("got %d target columns, want %d", len(targetData), tt.wantTargetCols)
			}

			// Verify target data values
			if tt.wantTargetCols > 0 {
				for colName, values := range targetData {
					t.Logf("Target column %s: %v", colName, values)
					// Don't check for NaN in sparse columns test
					if tt.name != "target column with empty values in first rows" {
						for _, v := range values {
							if math.IsNaN(v) {
								t.Errorf("unexpected NaN in target column %s", colName)
							}
						}
					}
				}
			}
		})
	}
}
