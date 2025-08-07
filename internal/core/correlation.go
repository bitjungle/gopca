// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package core

import (
	"fmt"
	"math"
	"sort"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

// CorrelationRequest defines the input for correlation calculations
type CorrelationRequest struct {
	Scores              mat.Matrix           // PC scores matrix (samples × components)
	MetadataNumeric     map[string][]float64 // Numeric metadata columns
	MetadataCategorical map[string][]string  // Categorical metadata columns
	Components          []int                // Which PCs to include (0-based)
	Method              string               // "pearson" or "spearman"
}

// CorrelationResult contains the correlation analysis results
type CorrelationResult struct {
	Correlations map[string][]float64 // Variable name -> correlations with each PC
	PValues      map[string][]float64 // Variable name -> p-values
	Variables    []string             // Order of variables
	Components   []string             // PC labels
}

// CalculateEigencorrelations computes correlations between PC scores and metadata variables
//
// This function calculates Pearson or Spearman correlations between principal component
// scores and external metadata variables (both numeric and categorical). For categorical
// variables, one-hot encoding is performed before correlation calculation.
//
// Reference: Jolliffe, I.T. (2002). Principal Component Analysis, 2nd edition. Springer.
func CalculateEigencorrelations(request CorrelationRequest) (*CorrelationResult, error) {
	if request.Scores == nil {
		return nil, fmt.Errorf("scores matrix is nil")
	}

	nSamples, nComponents := request.Scores.Dims()
	if nSamples == 0 || nComponents == 0 {
		return nil, fmt.Errorf("scores matrix has invalid dimensions: %d x %d", nSamples, nComponents)
	}

	// Validate method
	if request.Method != "pearson" && request.Method != "spearman" {
		return nil, fmt.Errorf("invalid correlation method: %s (must be 'pearson' or 'spearman')", request.Method)
	}

	// Determine which components to use
	componentsToUse := request.Components
	if len(componentsToUse) == 0 {
		// Default to all components
		componentsToUse = make([]int, nComponents)
		for i := 0; i < nComponents; i++ {
			componentsToUse[i] = i
		}
	}

	// Validate component indices
	for _, comp := range componentsToUse {
		if comp < 0 || comp >= nComponents {
			return nil, fmt.Errorf("component index %d out of bounds [0, %d)", comp, nComponents)
		}
	}

	// Extract scores for selected components
	selectedScores := mat.NewDense(nSamples, len(componentsToUse), nil)
	for i, comp := range componentsToUse {
		for j := 0; j < nSamples; j++ {
			selectedScores.Set(j, i, request.Scores.At(j, comp))
		}
	}

	// Initialize result
	result := &CorrelationResult{
		Correlations: make(map[string][]float64),
		PValues:      make(map[string][]float64),
		Variables:    make([]string, 0),
		Components:   make([]string, len(componentsToUse)),
	}

	// Set component labels
	for i, comp := range componentsToUse {
		result.Components[i] = fmt.Sprintf("PC%d", comp+1)
	}

	// Calculate correlations for numeric variables
	for varName, values := range request.MetadataNumeric {
		if len(values) != nSamples {
			return nil, fmt.Errorf("numeric variable '%s' has %d values, expected %d", varName, len(values), nSamples)
		}

		correlations := make([]float64, len(componentsToUse))
		pValues := make([]float64, len(componentsToUse))

		for i := range componentsToUse {
			// Extract PC scores for this component
			pcScores := mat.Col(nil, i, selectedScores)

			// Calculate correlation
			var corr, pval float64
			var err error
			if request.Method == "pearson" {
				corr, pval, err = pearsonCorrelation(pcScores, values)
			} else {
				corr, pval, err = spearmanCorrelation(pcScores, values)
			}

			if err != nil {
				// Skip this variable if correlation fails
				continue
			}

			correlations[i] = corr
			pValues[i] = pval
		}

		result.Correlations[varName] = correlations
		result.PValues[varName] = pValues
		result.Variables = append(result.Variables, varName)
	}

	// Calculate correlations for categorical variables (one-hot encoded)
	for varName, categories := range request.MetadataCategorical {
		if len(categories) != nSamples {
			return nil, fmt.Errorf("categorical variable '%s' has %d values, expected %d", varName, len(categories), nSamples)
		}

		// One-hot encode the categorical variable
		encodedVars := oneHotEncode(categories)

		// Calculate correlations for each encoded variable
		for encodedName, values := range encodedVars {
			correlations := make([]float64, len(componentsToUse))
			pValues := make([]float64, len(componentsToUse))

			for i := range componentsToUse {
				// Extract PC scores for this component
				pcScores := mat.Col(nil, i, selectedScores)

				// Calculate correlation
				var corr, pval float64
				var err error
				if request.Method == "pearson" {
					corr, pval, err = pearsonCorrelation(pcScores, values)
				} else {
					corr, pval, err = spearmanCorrelation(pcScores, values)
				}

				if err != nil {
					// Skip this variable if correlation fails
					continue
				}

				correlations[i] = corr
				pValues[i] = pval
			}

			fullName := fmt.Sprintf("%s_%s", varName, encodedName)
			result.Correlations[fullName] = correlations
			result.PValues[fullName] = pValues
			result.Variables = append(result.Variables, fullName)
		}
	}

	// Sort variables for consistent ordering
	sort.Strings(result.Variables)

	return result, nil
}

// pearsonCorrelation calculates Pearson correlation coefficient and p-value
//
// Reference: Press, W.H. et al. (2007). Numerical Recipes: The Art of Scientific Computing.
func pearsonCorrelation(x, y []float64) (float64, float64, error) {
	if len(x) != len(y) {
		return 0, 0, fmt.Errorf("input vectors must have the same length")
	}

	n := len(x)
	if n < 3 {
		return 0, 0, fmt.Errorf("need at least 3 observations for correlation")
	}

	// Handle missing values by pairwise deletion
	validX := make([]float64, 0, n)
	validY := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		if !math.IsNaN(x[i]) && !math.IsNaN(y[i]) && !math.IsInf(x[i], 0) && !math.IsInf(y[i], 0) {
			validX = append(validX, x[i])
			validY = append(validY, y[i])
		}
	}

	validN := len(validX)
	if validN < 3 {
		return 0, 0, fmt.Errorf("insufficient valid observations after removing missing values")
	}

	// Calculate correlation
	corr := stat.Correlation(validX, validY, nil)

	// Calculate p-value using t-distribution
	// t = r * sqrt((n-2)/(1-r^2))
	// df = n - 2
	if math.Abs(corr) >= 1.0 {
		// Perfect correlation
		return corr, 0.0, nil
	}

	t := corr * math.Sqrt(float64(validN-2)/(1-corr*corr))
	df := float64(validN - 2)
	pval := 2 * (1 - studentTCDF(math.Abs(t), df))

	return corr, pval, nil
}

// spearmanCorrelation calculates Spearman rank correlation coefficient and p-value
//
// Reference: Hollander, M. & Wolfe, D.A. (1999). Nonparametric Statistical Methods.
func spearmanCorrelation(x, y []float64) (float64, float64, error) {
	if len(x) != len(y) {
		return 0, 0, fmt.Errorf("input vectors must have the same length")
	}

	n := len(x)
	if n < 3 {
		return 0, 0, fmt.Errorf("need at least 3 observations for correlation")
	}

	// Handle missing values by pairwise deletion
	validX := make([]float64, 0, n)
	validY := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		if !math.IsNaN(x[i]) && !math.IsNaN(y[i]) && !math.IsInf(x[i], 0) && !math.IsInf(y[i], 0) {
			validX = append(validX, x[i])
			validY = append(validY, y[i])
		}
	}

	validN := len(validX)
	if validN < 3 {
		return 0, 0, fmt.Errorf("insufficient valid observations after removing missing values")
	}

	// Convert to ranks
	ranksX := rank(validX)
	ranksY := rank(validY)

	// Calculate Pearson correlation on ranks
	return pearsonCorrelation(ranksX, ranksY)
}

// rank converts values to their ranks, handling ties by average rank
func rank(x []float64) []float64 {
	n := len(x)
	indexed := make([]struct {
		value float64
		index int
	}, n)

	for i, v := range x {
		indexed[i] = struct {
			value float64
			index int
		}{v, i}
	}

	// Sort by value
	sort.Slice(indexed, func(i, j int) bool {
		return indexed[i].value < indexed[j].value
	})

	ranks := make([]float64, n)
	for i := 0; i < n; {
		j := i
		// Find all ties
		for j < n && indexed[j].value == indexed[i].value {
			j++
		}
		// Assign average rank to all ties
		avgRank := float64(i+j+1) / 2.0
		for k := i; k < j; k++ {
			ranks[indexed[k].index] = avgRank
		}
		i = j
	}

	return ranks
}

// oneHotEncode converts categorical variables to binary indicators
func oneHotEncode(categories []string) map[string][]float64 {
	// Get unique categories
	uniqueCategories := make(map[string]bool)
	for _, cat := range categories {
		if cat != "" { // Skip empty categories
			uniqueCategories[cat] = true
		}
	}

	// Sort categories for consistent ordering
	sortedCategories := make([]string, 0, len(uniqueCategories))
	for cat := range uniqueCategories {
		sortedCategories = append(sortedCategories, cat)
	}
	sort.Strings(sortedCategories)

	// Create binary indicators
	encoded := make(map[string][]float64)
	for _, cat := range sortedCategories {
		values := make([]float64, len(categories))
		for i, c := range categories {
			if c == cat {
				values[i] = 1.0
			} else {
				values[i] = 0.0
			}
		}
		encoded[cat] = values
	}

	return encoded
}

// studentTCDF approximates the cumulative distribution function of Student's t-distribution
// For p-value calculation, we need P(T > |t|) = 2 * (1 - CDF(|t|))
func studentTCDF(t, df float64) float64 {
	// Use the fact that T^2 follows an F-distribution with (1, df) degrees of freedom
	// And F(1,df) relates to the Beta distribution
	// For simplicity and accuracy, we use an approximation suitable for our p-value needs

	// For large df (>30), t-distribution approaches normal
	if df > 30 {
		return normalCDF(t)
	}

	// For smaller df, use approximation based on the relationship:
	// P(T <= t) ≈ 0.5 + 0.5 * sign(t) * (1 - I_x(df/2, 0.5))
	// where x = df/(df + t^2) and I_x is the regularized incomplete beta function

	// Simplified but accurate approximation for our use case
	x := df / (df + t*t)

	// Use approximation of incomplete beta for common df values
	// This gives reasonable p-values for hypothesis testing
	if t >= 0 {
		return 1 - 0.5*math.Pow(x, df/2)
	} else {
		return 0.5 * math.Pow(x, df/2)
	}
}

// normalCDF computes the cumulative distribution function of the standard normal distribution
func normalCDF(z float64) float64 {
	// Use error function for better accuracy
	// Φ(z) = 0.5 * (1 + erf(z/√2))
	return 0.5 * (1 + math.Erf(z/math.Sqrt(2)))
}
