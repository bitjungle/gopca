// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

/**
 * Describes the preprocessing that will actually be applied, in the order it is
 * applied.
 *
 * The panel used to convey that order by numbering its controls "Step 1", "Step
 * 2", "Step 3". That was read as a sequence the user had to work through --
 * pick a row-wise method first, and only then a filter -- when in fact each
 * control is independently optional and any of them can be left at None.
 *
 * Listing the order as a sentence says the same thing without implying an
 * obligation, and has the advantage of describing what is selected rather than
 * what could be. An empty pipeline is worth stating too: raw data going
 * straight into the decomposition is a choice, and one worth seeing.
 */

export interface PreprocessingSummaryConfig {
    snv: boolean;
    vectorNorm: boolean;
    savgolWindow: number;
    savgolPolyOrder: number;
    savgolDeriv: number;
    meanCenter: boolean;
    standardScale: boolean;
    robustScale: boolean;
    scaleOnly: boolean;
}

/**
 * Returns the preprocessing steps in application order.
 *
 * The order here is the engine's, not the panel's: row-wise normalisation,
 * then Savitzky-Golay, then column statistics. If the two ever disagree the
 * summary is the one that is wrong, so it is derived from the same fields the
 * request carries rather than from the panel's layout.
 */
export function preprocessingPipeline(config: PreprocessingSummaryConfig): string[] {
    return [...rowStagePipeline(config), ...columnStagePipeline(config)];
}

/**
 * The steps applied along each row: normalisation, then Savitzky-Golay.
 *
 * Separate from the column steps because the preview plot shows exactly this
 * much and no more. A caption listing "Mean center" beside curves that were
 * never centred would describe a picture the reader is not looking at.
 */
export function rowStagePipeline(config: PreprocessingSummaryConfig): string[] {
    const steps: string[] = [];

    if (config.snv) {
        steps.push('SNV');
    } else if (config.vectorNorm) {
        steps.push('L2 normalization');
    }

    if (config.savgolWindow > 0) {
        const what
            = config.savgolDeriv === 1
                ? '1st derivative'
                : config.savgolDeriv === 2
                    ? '2nd derivative'
                    : config.savgolDeriv > 2
                        ? `derivative ${config.savgolDeriv}`
                        : 'smoothing';
        steps.push(`Savitzky-Golay ${what} (window ${config.savgolWindow}, order ${config.savgolPolyOrder})`);
    }

    return steps;
}

/**
 * The steps applied down each column, after the row stage.
 *
 * Precedence matches the column-wise selector, so the summary cannot disagree
 * with the control above it.
 */
export function columnStagePipeline(config: PreprocessingSummaryConfig): string[] {
    if (config.scaleOnly) {
        return ['Variance scale'];
    }
    if (config.robustScale) {
        return ['Robust scale'];
    }
    if (config.standardScale) {
        return ['Standard scale'];
    }
    if (config.meanCenter) {
        return ['Mean center'];
    }
    return [];
}

/**
 * Whether a preview of the preprocessed data is worth drawing.
 *
 * Two conditions, and both are about whether the plot could mean anything:
 * the variables must form an axis worth drawing a curve along, and something
 * must actually happen to them, or the "preprocessed" view would be the raw one
 * under a different name.
 *
 * Note that Savitzky-Golay is *not* required. Row-wise normalisation alone
 * changes the spectra, and seeing what SNV did is as reasonable a question as
 * seeing what a derivative did. This is exported and tested rather than written
 * inline because getting it wrong is invisible: the preview simply does not
 * appear, and nothing says why.
 */
export function shouldShowPreview(
    isContinuousAxis: boolean,
    config: Pick<PreprocessingSummaryConfig, 'snv' | 'vectorNorm' | 'savgolWindow'>
): boolean {
    if (!isContinuousAxis) {
        return false;
    }
    return config.snv || config.vectorNorm || config.savgolWindow > 0;
}
