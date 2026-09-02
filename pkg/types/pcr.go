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

package types

// Selection rules for turning a cross-validated error curve into a choice of
// component count. See SelectionConfig.Rule.
const (
	// SelectMin takes the candidate with the lowest estimated error. It is the
	// obvious rule and usually not the best one: the minimum of a noisy curve is
	// frequently reached by a model far more complex than the data supports.
	SelectMin = "min"

	// SelectOneSE takes the simplest candidate whose error is within one standard
	// error of the minimum. This is the default because it is the rule that most
	// often yields the model a practitioner would actually deploy.
	SelectOneSE = "one-se"

	// SelectTolerance takes the simplest candidate whose error is within a
	// prespecified margin of the minimum, expressed in response units. Easier to
	// justify than one-se when an engineering tolerance is already known.
	SelectTolerance = "tolerance"

	// SelectWold stops at the first candidate where the error ratio between
	// successive component counts exceeds WoldR, formalising "stop when the curve
	// flattens".
	SelectWold = "wold"
)

// Cross-validation schemes. Grouping is orthogonal to the scheme and is set
// through CVConfig.GroupBy.
const (
	// CVRandom shuffles before cutting folds, giving ordinary K-fold. Valid only
	// when observations are exchangeable.
	CVRandom = "random"

	// CVContiguous cuts folds without shuffling, giving contiguous blocks. The
	// conservative choice when row order carries meaning.
	CVContiguous = "contiguous"

	// CVForwardChaining trains on a prefix of the series and tests on the block
	// that follows, never training on observations later than those predicted.
	CVForwardChaining = "forward-chaining"
)

// CVConfig describes a resampling design.
//
// The design must represent how future predictions will actually be made. A
// random split of data that arrives in batches, or that contains replicate
// measurements of the same object, reports an error lower than deployment will
// deliver, because the model gets to interpolate between near-duplicates.
//
// There is no separate leave-one-out scheme. Leave-one-out is Folds = 0, meaning
// as many folds as there are groups, and with the default grouping of one row per
// group that is exactly K-fold at K = n.
type CVConfig struct {
	Scheme string `json:"scheme"` // CVRandom, CVContiguous or CVForwardChaining

	// Folds is K. Zero means "as many folds as there are groups", which yields
	// leave-one-out under the default grouping and leave-one-group-out otherwise.
	// K may not exceed the number of groups.
	Folds int `json:"folds"`

	// GroupBy names a categorical column whose levels define the groups, so that
	// all rows of one group land in the same fold. Empty means one row per group.
	// Set it whenever rows are replicates, batches, assets or sites.
	//
	// This is the human-readable record of the design. The engine reads Groups,
	// not this name, because resolving a column name into group identifiers means
	// knowing about CSV structure and the core engine deliberately does not.
	GroupBy string `json:"group_by,omitempty"`

	// Groups assigns a group identifier to every row, indexed by row. Nil means
	// one group per row. Callers that set GroupBy must resolve it into this slice;
	// the engine treats a nil Groups with a non-empty GroupBy as a caller error
	// rather than silently validating an ungrouped design under a grouped label.
	Groups []int `json:"-"`

	// Repeats re-runs the design with fresh partitions to assess sensitivity to
	// the split. It is meaningless when Folds equals the group count, because the
	// partition is then unique, and is rejected rather than silently ignored.
	Repeats int `json:"repeats,omitempty"`

	// Seed makes the design reproducible. Recorded in CVReport so that a result
	// can be regenerated from its report alone.
	Seed int64 `json:"seed"`
}

// SelectionConfig controls how the number of retained components is chosen.
type SelectionConfig struct {
	// Mode is "fixed" to use Fixed directly, or "cv" to choose by
	// cross-validation.
	Mode  string `json:"mode"`
	Fixed int    `json:"fixed,omitempty"`

	// Metric is the primary selection criterion, always in response units.
	// "rmse" or "mae".
	Metric string `json:"metric"`

	// Rule turns the error curve into a choice; one of the Select* constants.
	Rule      string  `json:"rule"`
	Tolerance float64 `json:"tolerance,omitempty"` // for SelectTolerance, in response units
	WoldR     float64 `json:"wold_r,omitempty"`    // for SelectWold, conventionally 0.90 to 1.00

	CV CVConfig `json:"cv"`
}

// PCRConfig holds configuration for principal component regression.
//
// The predictor side reuses PCAConfig verbatim, so every preprocessing and method
// option available to PCA is available to PCR. PCA.Components sets the largest
// number of components computed; the retained set is chosen from those.
type PCRConfig struct {
	PCA PCAConfig `json:"pca"`

	// Response is the name of the numeric target column to predict.
	Response string `json:"response"`

	Selection SelectionConfig `json:"selection"`
}

// CVReport records a cross-validation sweep in full, so that any number taken
// from it can be traced back to the design that produced it.
//
// All the per-candidate slices are parallel to Candidates.
type CVReport struct {
	Scheme   string `json:"scheme"`
	Design   string `json:"design"` // human-readable splitter description
	Folds    int    `json:"folds"`
	Repeats  int    `json:"repeats"`
	GroupBy  string `json:"group_by,omitempty"`
	Seed     int64  `json:"seed"`
	NSamples int    `json:"n_samples"` // labelled rows entering the sweep

	// Candidates are the component counts evaluated, always including 0, the
	// intercept-only baseline against which any positive count must justify itself.
	Candidates []int `json:"candidates"`

	// RMSECV pools every out-of-fold residual before taking one square root, so
	// each observation carries equal weight regardless of fold size. This is the
	// figure to report.
	RMSECV []float64 `json:"rmsecv"`

	// RMSECVMean averages the per-fold RMSE instead. It is never larger than the
	// pooled value and is kept because the one-standard-error rule needs the
	// per-fold spread. At Folds = group count the two coincide.
	RMSECVMean []float64 `json:"rmsecv_mean"`
	RMSECVSE   []float64 `json:"rmsecv_se"`

	// Bias and SEP decompose the error: RMSECV^2 = Bias^2 + (n-1)/n * SEP^2.
	// A large Bias with a small SEP is a precise model with a systematic offset,
	// which is repairable, and a different failure from simple imprecision.
	Bias []float64 `json:"bias"`
	SEP  []float64 `json:"sep"`

	// MAE is carried as a divergence check rather than as a second primary
	// metric. RMSE is driven by the largest residuals and MAE by the typical one,
	// so when the two select different counts a handful of samples is deciding
	// the model and that is worth surfacing.
	MAE []float64 `json:"mae"`

	// Q2 is the cross-validated coefficient of determination.
	Q2 []float64 `json:"q2"`

	Selected int    `json:"selected"`
	Rule     string `json:"rule"`

	// SelectedByMAE is what the MAE curve would have chosen. When it differs from
	// Selected, the choice is sensitive to a few extreme residuals.
	SelectedByMAE int `json:"selected_by_mae"`

	// OutOfFold holds one held-out prediction per labelled row at the selected
	// component count, indexed as LabelledRows.
	OutOfFold []float64 `json:"out_of_fold_predictions"`
}

// PCRResult contains the results of principal component regression.
type PCRResult struct {
	// PCA is the decomposition the regression was fitted on, refitted on all
	// available rows after selection.
	PCA *PCAResult `json:"pca"`

	Response string `json:"response"`

	// Components is the number of leading components retained.
	Components int `json:"components"`

	// ScoreCoefficients is gamma, the regression coefficients in score space,
	// one per retained component. Their signs follow the component signs and
	// therefore carry no meaning on their own.
	ScoreCoefficients []float64 `json:"score_coefficients"`

	// Intercept is the score-space intercept. Fitting it explicitly rather than
	// relying on a centred response keeps the estimator correct for every
	// column-wise preprocessing combination.
	Intercept float64 `json:"intercept"`

	// Coefficients and InterceptOriginal describe the collapsed original-scale
	// form y = InterceptOriginal + x . Coefficients, which is the form to deploy.
	// Both are invalid when row-wise preprocessing (SNV, vector normalisation) is
	// enabled, because that map depends on the sample itself and no fixed
	// coefficient vector reproduces its predictions. OriginalScaleValid says which
	// case applies; the fields are omitted when it is false rather than filled
	// with a plausible but wrong number.
	Coefficients       []float64 `json:"coefficients,omitempty"`
	InterceptOriginal  float64   `json:"intercept_original,omitempty"`
	OriginalScaleValid bool      `json:"original_scale_valid"`

	ResponseMean float64 `json:"response_mean"`

	// Fitted holds the model's prediction for each labelled row, and Residuals
	// the measured value minus that prediction, both indexed against
	// LabelledRows. Note the sign: a positive residual means the model
	// under-predicted. RegressionMetrics.Bias uses the opposite convention,
	// predicted minus measured, because a positive bias conventionally means
	// over-prediction.
	Fitted    []float64 `json:"fitted"`
	Residuals []float64 `json:"residuals"`

	// RMSEC is the root-mean-square error of the training residuals. It describes
	// the fit and is NOT an estimate of future performance: the model has seen
	// every row it is scored on here. Use CV.RMSECV for that, and an independent
	// test set for RMSEP.
	RMSEC float64 `json:"rmsec"`
	R2C   float64 `json:"r2c"`

	CV *CVReport `json:"cv,omitempty"`

	// LabelledRows lists the 0-based rows carrying an observed response, in order.
	// Fitted, Residuals and CV.OutOfFold are indexed against this slice, not
	// against the full data matrix.
	LabelledRows []int `json:"labelled_rows,omitempty"`

	// ExcludedRows lists the 0-based rows dropped from the regression because the
	// response was missing. Those rows may still have informed the decomposition,
	// since PCA does not use the response.
	ExcludedRows []int `json:"excluded_rows,omitempty"`
}

// PCREngine defines the interface for principal component regression, mirroring
// PCAEngine.
type PCREngine interface {
	// Fit builds a model predicting y from data. Entries of y that are not finite
	// mark rows without an observed response; those rows are excluded from the
	// regression but may still inform the decomposition.
	Fit(data Matrix, y []float64, config PCRConfig) (*PCRResult, error)

	// Predict applies the fitted model to new data.
	Predict(data Matrix) ([]float64, error)

	// FitPredict fits the model and returns its result in one step.
	FitPredict(data Matrix, y []float64, config PCRConfig) (*PCRResult, error)
}
