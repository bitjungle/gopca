// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
// SPDX-License-Identifier: See LICENSE file for details.

import { describe, expect, it } from 'vitest';
import { generateCLICommand, RegressionCLIConfig } from './cliCommandGenerator';

const base = {
    filePath: '/data/corn.csv',
    components: 5,
    method: 'SVD',
    meanCenter: true
};

const regression: RegressionCLIConfig = {
    response: 'Moisture#target',
    components: 0,
    maxComponents: 20,
    cvFolds: 10,
    cvScheme: 'random',
    cvGroupColumn: '',
    cvSeed: 42,
    selectRule: 'one-se',
    metric: 'rmse',
    tolerance: 0,
    woldR: 1
};

describe('generateCLICommand', () => {
    it('emits the analyze verb without a regression block', () => {
        expect(generateCLICommand(base)).toContain('pca analyze');
        expect(generateCLICommand(base)).toContain('--components 5');
    });

    // The preview is presented as a command the reader can paste, so a flag the
    // CLI does not accept is not a cosmetic error. This one was wrong: the flag
    // is --exclude-columns, and --exclude-cols is rejected outright.
    it('names the exclusion flags as the CLI accepts them', () => {
        const cmd = generateCLICommand({ ...base, excludedColumns: [0, 1], excludedRows: [3] });
        expect(cmd).toContain('--exclude-columns');
        expect(cmd).not.toContain('--exclude-cols ');
        expect(cmd).toContain('--exclude-rows');
    });

    describe('regression mode', () => {
        it('emits the regress verb and the chosen response', () => {
            const cmd = generateCLICommand({ ...base, regression });
            expect(cmd).toContain('pca regress');
            expect(cmd).not.toContain('pca analyze');
            expect(cmd).toContain('--response "Moisture#target"');
        });

        it('sweeps when the count is chosen by cross-validation', () => {
            const cmd = generateCLICommand({ ...base, regression });
            expect(cmd).toContain('--max-components 20');
            expect(cmd).toContain('--cv 10');
            // The panel's own default should not be restated as a flag.
            expect(cmd).not.toContain('--select');
            expect(cmd).not.toContain('--cv-seed');
        });

        it('fixes the count when one was given', () => {
            const cmd = generateCLICommand({
                ...base, regression: { ...regression, components: 7 }
            });
            expect(cmd).toContain('--components 7');
            expect(cmd).not.toContain('--max-components');
            expect(cmd).not.toContain('--cv ');
        });

        it('spells leave-one-out the way the CLI does', () => {
            const cmd = generateCLICommand({
                ...base, regression: { ...regression, cvFolds: 0 }
            });
            expect(cmd).toContain('--cv loo');
            expect(cmd).not.toContain('--cv 0');
        });

        it('carries the grouping column and non-default rules', () => {
            const cmd = generateCLICommand({
                ...base,
                regression: {
                    ...regression, cvGroupColumn: 'Product',
                    selectRule: 'first-min', metric: 'mae', cvSeed: 7
                }
            });
            expect(cmd).toContain('--cv-group "Product"');
            expect(cmd).toContain('--select first-min');
            expect(cmd).toContain('--metric mae');
            expect(cmd).toContain('--cv-seed 7');
        });

        // Kernel and temporal decompositions cannot project new data, so the
        // regress command refuses them. Emitting their flags would print a
        // command that fails the moment it is pasted.
        it('omits kernel and temporal flags, which regress does not accept', () => {
            const kernel = generateCLICommand({
                ...base, method: 'kernel', kernelType: 'rbf', kernelGamma: 0.01, regression
            });
            expect(kernel).not.toContain('--kernel-type');

            const temporal = generateCLICommand({
                ...base, method: 'temporal', temporalLags: 10, regression
            });
            expect(temporal).not.toContain('--temporal-lags');
        });

        it('keeps the preprocessing flags, which apply to both modes', () => {
            const cmd = generateCLICommand({
                ...base, snv: true, standardScale: true, missingStrategy: 'drop', regression
            });
            expect(cmd).toContain('--snv');
            expect(cmd).toContain('--scale standard');
            expect(cmd).toContain('--missing-strategy drop');
        });
    });
});
