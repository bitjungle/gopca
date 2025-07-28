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
			name: "column with _target suffix",
			csvContent: `feature1,feature2,category,value_target
1.0,2.0,A,10.5
3.0,4.0,B,20.3
5.0,6.0,A,15.7`,
			targetColumns:  nil,
			wantDataCols:   2, // feature1, feature2
			wantCatCols:    1, // category
			wantTargetCols: 1, // value_target
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
			csvContent: `feature1,feature2,category,manual_target,auto_target
1.0,2.0,A,10.5,100
3.0,4.0,B,20.3,200
5.0,6.0,A,15.7,300`,
			targetColumns:  []string{"manual_target"},
			wantDataCols:   2, // feature1, feature2
			wantCatCols:    1, // category
			wantTargetCols: 2, // manual_target, auto_target
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
					for _, v := range values {
						if math.IsNaN(v) {
							t.Errorf("unexpected NaN in target column %s", colName)
						}
					}
				}
			}
		})
	}
}
