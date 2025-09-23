// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState, useRef, useEffect, lazy, Suspense } from 'react';
import './App.css';
import { RunPCA, LoadDatasetFile, GetVersion, GetGUIConfig, LoadCSVFile, CheckGoCSVStatus, OpenInGoCSV, LaunchGoCSV, DownloadGoCSV, SaveFile, SelectCSVFile } from '../wailsjs/go/main/App';
import { Copy, Check } from 'lucide-react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { DataTable, SelectionTable, MatrixIllustration, HelpWrapper, DocumentationViewer, ModelOverview, AboutDialog } from './components';
import { setupPlotlyWailsIntegration } from '@gopca/ui-components';

// Lazy load visualization components for better performance
const ScoresPlot = lazy(() => import('./components/visualizations/ScoresPlot').then(m => ({ default: m.ScoresPlot })));
const Scores3DPlot = lazy(() => import('./components/visualizations/Scores3DPlot').then(m => ({ default: m.Scores3DPlot })));
const ScreePlot = lazy(() => import('./components/visualizations/ScreePlot').then(m => ({ default: m.ScreePlot })));
const LoadingsPlot = lazy(() => import('./components/visualizations/LoadingsPlot').then(m => ({ default: m.LoadingsPlot })));
const Biplot = lazy(() => import('./components/visualizations/Biplot').then(m => ({ default: m.Biplot })));
const Biplot3D = lazy(() => import('./components/visualizations/Biplot3D').then(m => ({ default: m.Biplot3D })));
const CircleOfCorrelations = lazy(() => import('./components/visualizations/CircleOfCorrelations').then(m => ({ default: m.CircleOfCorrelations })));
const DiagnosticScatterPlot = lazy(() => import('./components/visualizations/DiagnosticScatterPlot').then(m => ({ default: m.DiagnosticScatterPlot })));
const EigencorrelationPlot = lazy(() => import('./components/visualizations/EigencorrelationPlot').then(m => ({ default: m.EigencorrelationPlot })));
const TemporalLoadingsPlot = lazy(() => import('./components/visualizations/TemporalLoadingsPlot').then(m => ({ default: m.TemporalLoadingsPlot })));
const TemporalVariableImportancePlot = lazy(() => import('./components/visualizations/TemporalVariableImportancePlot').then(m => ({ default: m.TemporalVariableImportancePlot })));
const KernelMatrixHeatmap = lazy(() => import('./components/visualizations/KernelMatrixHeatmap').then(m => ({ default: m.KernelMatrixHeatmap })));
const SampleContributionPlot = lazy(() => import('./components/visualizations/SampleContributionPlot').then(m => ({ default: m.SampleContributionPlot })));
import { FileData, PCARequest, PCAResponse } from './types';
import { ThemeProvider, ThemeToggle, ConfirmDialog, CustomSelect, ErrorBoundary, ErrorAlert } from '@gopca/ui-components';
import { HelpProvider, useHelp } from './contexts/HelpContext';
import { PaletteProvider, usePalette } from './contexts/PaletteContext';
import { HelpDisplay } from './components/HelpDisplay';
import { PaletteSelector } from './components/PaletteSelector';
import { FontSizeControl } from './components/FontSizeControl';
import { config } from '../wailsjs/go/models';
import logo from './assets/images/GoPCA-logo-1024-transp.png';
import { generateCLICommand as generateCLICommandUtil } from './utils/cliCommandGenerator';

// Plot-specific palette configuration
type PlotPaletteConfig = {
    hasPalette: boolean;
    paletteType: 'categorical' | 'continuous' | 'dynamic'; // dynamic means depends on Color by selection
};

const PLOT_PALETTE_CONFIG: Record<string, PlotPaletteConfig> = {
    'scores': { hasPalette: true, paletteType: 'dynamic' },
    'scores3d': { hasPalette: true, paletteType: 'dynamic' },
    'scree': { hasPalette: true, paletteType: 'categorical' },
    'loadings': { hasPalette: true, paletteType: 'categorical' },
    'biplot': { hasPalette: true, paletteType: 'dynamic' },
    'biplot3d': { hasPalette: true, paletteType: 'dynamic' },
    'correlations': { hasPalette: true, paletteType: 'categorical' },
    'diagnostics': { hasPalette: true, paletteType: 'dynamic' },
    'eigencorrelation': { hasPalette: true, paletteType: 'continuous' },
    'temporal-loadings': { hasPalette: true, paletteType: 'categorical' },
    'temporal-variable-importance': { hasPalette: true, paletteType: 'continuous' },
    'kernel-matrix': { hasPalette: false, paletteType: 'continuous' }, // Uses fixed colorscale
    'sample-contributions': { hasPalette: true, paletteType: 'categorical' } // Actually uses categorical
};

function AppContent() {
    const { currentHelp, currentHelpKey } = useHelp();
    const { setMode } = usePalette();
    const [fileData, setFileData] = useState<FileData | null>(null);
    const [fileName, setFileName] = useState<string>('');
    const [pcaResponse, setPcaResponse] = useState<PCAResponse | null>(null);
    const [loading, setLoading] = useState(false);
    const [fileError, setFileError] = useState<string | null>(null);
    const [pcaError, setPcaError] = useState<string | null>(null);
    const [version, setVersion] = useState<string>('');
    const [guiConfig, setGuiConfig] = useState<config.GUIConfig | null>(null);

    // Selection state
    const [excludedRows, setExcludedRows] = useState<number[]>([]);
    const [excludedColumns, setExcludedColumns] = useState<number[]>([]);
    const [selectedPlot, setSelectedPlot] = useState<'scores' | 'scores3d' | 'scree' | 'loadings' | 'biplot' | 'biplot3d' | 'correlations' | 'diagnostics' | 'eigencorrelation' | 'temporal-loadings' | 'temporal-variable-importance' | 'kernel-matrix' | 'sample-contributions'>('scores');
    const [selectedXComponent, setSelectedXComponent] = useState(0);
    const [selectedYComponent, setSelectedYComponent] = useState(1);
    const [selectedZComponent, setSelectedZComponent] = useState(2);
    const [selectedLoadingComponent, setSelectedLoadingComponent] = useState(0);
    const [selectedGroupColumn, setSelectedGroupColumn] = useState<string | null>(null);
    const [showEllipses, setShowEllipses] = useState(false);
    const [confidenceLevel, setConfidenceLevel] = useState<0.90 | 0.95 | 0.99>(0.95);
    const [showRowLabels, setShowRowLabels] = useState(false);
    const [maxLabelsToShow, setMaxLabelsToShow] = useState(10);
    const [showDocumentation, setShowDocumentation] = useState(false);
    const [showAboutDialog, setShowAboutDialog] = useState(false);
    const [datasetId, setDatasetId] = useState(0); // Force DataTable re-render on dataset change
    const [showCopied, setShowCopied] = useState(false);
    const [loadingsPlotType, setLoadingsPlotType] = useState<'bar' | 'line' | null>(null); // null means auto
    const [plotFontScale, setPlotFontScale] = useState(1.0); // Font scale factor for all plots

    // Refs for smooth scrolling
    const pcaErrorRef = useRef<HTMLDivElement>(null);
    const pcaResultsRef = useRef<HTMLDivElement>(null);
    const mainScrollRef = useRef<HTMLDivElement>(null);

    // PCA configuration
    const [config, setConfig] = useState({
        components: 5,
        meanCenter: true,
        standardScale: false,
        robustScale: false,
        scaleOnly: false,
        snv: false,
        vectorNorm: false,
        method: 'SVD',
        missingStrategy: 'error',
        // Kernel PCA parameters
        kernelType: 'rbf',
        kernelGamma: 1.0,
        kernelDegree: 3,
        kernelCoef0: 0.0,
        // Temporal PCA parameters
        temporalLags: 10,
        varianceExplained: 0.0
        // Confidence ellipse parameters
    });

    // GoCSV integration state
    const [goCSVStatus, setGoCSVStatus] = useState<{installed: boolean, path?: string, error?: string} | null>(null);
    const [isCheckingGoCSV, setIsCheckingGoCSV] = useState(false);
    const [showGoCSVDownloadDialog, setShowGoCSVDownloadDialog] = useState(false);

    // Track if current PCA result was generated with exclusions applied
    const [pcaHasExclusions, setPcaHasExclusions] = useState(false);

    const updateGammaForData = (data: FileData) => {
        if (data && data.data && data.data[0]) {
            const numFeatures = data.data[0].length;
            setConfig(prev => ({
                ...prev,
                kernelGamma: 1.0 / numFeatures,
                components: Math.min(5, numFeatures)  // Default to 5 or number of features if less
            }));
        }
    };

    // Fetch version and GUI config on mount
    useEffect(() => {
        // Make SaveFile available globally for Plotly integration
        if (typeof SaveFile !== 'undefined') {
            (window as any).SaveFile = SaveFile;
            // SaveFile made available globally for Plotly integration
        }

        // Setup Plotly-Wails integration for export functionality
        setupPlotlyWailsIntegration();

        GetVersion().then((v) => {
            setVersion(v);
        }).catch((err) => {
            console.error('Failed to get version:', err);
        });

        GetGUIConfig().then((config) => {
            setGuiConfig(config);
        }).catch((err) => {
            console.error('Failed to get GUI config:', err);
        });

        // Check GoCSV installation status on startup
        CheckGoCSVStatus().then((status) => {
            setGoCSVStatus(status);
        }).catch((err) => {
            console.error('Failed to check GoCSV status:', err);
        });

        // Listen for file to load on startup
        const unsubscribe = EventsOn('load-file-on-startup', async (filePath: string) => {
            setLoading(true);
            setFileError(null);
            setPcaError(null);

            try {
                const result = await LoadCSVFile(filePath);
                setFileData(result);
                setPcaResponse(null);
                setExcludedRows([]);
                setExcludedColumns([]);
                setSelectedGroupColumn(null);
                setMode('none'); // Reset palette mode
                setDatasetId(prev => prev + 1); // Force DataTable re-render
                updateGammaForData(result);
            } catch (err) {
                setFileError(`Failed to load file: ${err}`);
            } finally {
                setLoading(false);
            }
        });

        // Cleanup event listener on unmount
        return () => {
            unsubscribe();
        };
    }, []);

    // Auto-switch visualization when Kernel PCA is selected and incompatible plot is active
    useEffect(() => {
        if (pcaResponse && pcaResponse.result && pcaResponse.result.method === 'kernel') {
            if (selectedPlot === 'loadings' || selectedPlot === 'diagnostics' || selectedPlot === 'eigencorrelation') {
                // Switch to scores plot as a safe default
                setSelectedPlot('scores');
            }
        }
    }, [pcaResponse, selectedPlot]);

    // Update palette mode based on selected plot type
    useEffect(() => {
        // Only update palette mode if we have a PCA result
        if (!pcaResponse || !pcaResponse.result) {
            return;
        }

        const plotConfig = PLOT_PALETTE_CONFIG[selectedPlot];
        if (!plotConfig || !plotConfig.hasPalette) {
            setMode('none');
            return;
        }

        // For plots with fixed palette types, set the mode immediately
        if (plotConfig.paletteType === 'categorical') {
            setMode('categorical');
        } else if (plotConfig.paletteType === 'continuous') {
            setMode('continuous');
        } else if (plotConfig.paletteType === 'dynamic') {
            // For dynamic plots, the mode is controlled by the Color by selection
            // Don't change the mode here to avoid conflicts with the Color by control
            // If no group column is selected and plot needs a palette, set a default
            if (!selectedGroupColumn) {
                // For dynamic plots without a selection, we still want to show the palette control
                // Set to categorical as a reasonable default that works for most plots
                setMode('categorical');
            }
            // Otherwise, let the Color by control handle the mode
        }
    }, [selectedPlot, pcaResponse, setMode, selectedGroupColumn]);

    // Helper function to get column data and type
    const getColumnData = (columnName: string | null): { values?: string[] | number[], type?: 'categorical' | 'continuous' } => {
        if (!columnName || !fileData) {
            return {};
        }

        // Handle special "Row Index" column
        if (columnName === 'Row Index') {
            // Generate 1-based row indices for the current data
            // Use the actual number of samples in the PCA result if available
            const numSamples = pcaResponse?.result?.scores?.length || fileData.data.length;
            const indices = Array.from({ length: numSamples }, (_, i) => i + 1);
            return { values: indices, type: 'continuous' };
        }

        // CRITICAL FIX: Use filtered data from PCA response when rows have been dropped
        // This ensures categorical/numeric column data aligns with the reduced scores matrix
        if (pcaResponse?.filteredCategoricalColumns && columnName in pcaResponse.filteredCategoricalColumns) {
            return { values: pcaResponse.filteredCategoricalColumns[columnName], type: 'categorical' };
        }

        if (pcaResponse?.filteredNumericTargetColumns && columnName in pcaResponse.filteredNumericTargetColumns) {
            return { values: pcaResponse.filteredNumericTargetColumns[columnName], type: 'continuous' };
        }

        // Fall back to original data if no filtered data is available
        if (fileData.categoricalColumns && columnName in fileData.categoricalColumns) {
            return { values: fileData.categoricalColumns[columnName], type: 'categorical' };
        }

        if (fileData.numericTargetColumns && columnName in fileData.numericTargetColumns) {
            return { values: fileData.numericTargetColumns[columnName], type: 'continuous' };
        }

        return {};
    };

    // Centralized dataset loading function
    const loadDataset = async (filename: string, defaultGroupColumn?: string) => {
        setLoading(true);
        setFileError(null);
        setPcaError(null);

        try {
            const result = await LoadDatasetFile(filename);
            setFileData(result);
            setFileName(filename); // Store the sample dataset filename
            setPcaResponse(null);
            setExcludedRows([]);
            setExcludedColumns([]);
            setDatasetId(prev => prev + 1); // Force DataTable re-render

            // Validate group column exists before setting
            if (defaultGroupColumn && result) {
                const isCategorical = result.categoricalColumns && defaultGroupColumn in result.categoricalColumns;
                const isContinuous = result.numericTargetColumns && defaultGroupColumn in result.numericTargetColumns;
                const isValid = isCategorical || isContinuous;

                if (isValid) {
                    setSelectedGroupColumn(defaultGroupColumn);
                    // Set the appropriate palette mode based on column type
                    if (isCategorical) {
                        setMode('categorical');
                    } else if (isContinuous) {
                        setMode('continuous');
                    }
                } else {
                    console.warn(`Column "${defaultGroupColumn}" not found in ${filename}, setting group column to null`);
                    setSelectedGroupColumn(null);
                    setMode('none');
                }
            } else {
                setSelectedGroupColumn(null);
                setMode('none');
            }

            updateGammaForData(result);
        } catch (err) {
            setFileError(`Failed to load ${filename}: ${err}`);
        } finally {
            setLoading(false);
        }
    };

    const handleNativeFileSelect = async () => {
        setLoading(true);
        setFileError(null);
        setPcaError(null); // Clear any previous PCA errors

        try {
            const parseResult = await SelectCSVFile();

            // User cancelled
            if (!parseResult) {
                setLoading(false);
                return;
            }

            // Set a generic filename for display purposes
            setFileName('Selected File');
            setFileData(parseResult);
            setPcaResponse(null);
            // Reset exclusions and selections when loading new data
            setExcludedRows([]);
            setExcludedColumns([]);
            setSelectedGroupColumn(null);
            setMode('none'); // Reset palette mode
            setDatasetId(prev => prev + 1); // Force DataTable re-render

            // Calculate and set default gamma for kernel PCA
            updateGammaForData(parseResult);
        } catch (err) {
            console.error('File selection failed:', err);
            setFileError(`Failed to load file: ${err}`);
            setFileData(null);
        } finally {
            setLoading(false);
        }
    };

    const handleGoCSVAction = async () => {
        setIsCheckingGoCSV(true);

        try {
            // Check if GoCSV is installed
            const status = await CheckGoCSVStatus();
            setGoCSVStatus(status);

            if (status.installed) {
                // If installed and we have data, open it in GoCSV
                if (fileData) {
                    await OpenInGoCSV(fileData);
                } else {
                    // If no data, just launch GoCSV without a file
                    await LaunchGoCSV();
                }
            } else {
                // If not installed, show download dialog
                setShowGoCSVDownloadDialog(true);
            }
        } catch (err) {
            console.error('GoCSV action failed:', err);
            alert(`Failed to perform GoCSV action: ${err}`);
        } finally {
            setIsCheckingGoCSV(false);
        }
    };

    const handleRowSelectionChange = React.useCallback((selectedRows: number[]) => {
        // Convert selected indices to excluded indices
        if (fileData) {
            const allIndices = Array.from({ length: fileData.data.length }, (_, i) => i);
            const excluded = allIndices.filter(i => !selectedRows.includes(i));
            setExcludedRows(excluded);
        }
    }, [fileData]);

    // Handle plot selection changes
    const handlePlotSelectionChange = React.useCallback((indices: number[]) => {
        if (loading || !fileData || indices.length === 0) {
            return;
        }

        // Toggle selected indices in excluded rows
        setExcludedRows(prev => {
            const newExcluded = new Set(prev);
            indices.forEach(idx => {
                if (newExcluded.has(idx)) {
                    newExcluded.delete(idx);
                } else {
                    newExcluded.add(idx);
                }
            });
            return Array.from(newExcluded);
        });
    }, [fileData, loading]);

    const handleColumnSelectionChange = React.useCallback((selectedColumns: number[]) => {
        // Convert selected indices to excluded indices
        if (fileData) {
            const allIndices = Array.from({ length: fileData.headers.length }, (_, i) => i);
            const excluded = allIndices.filter(i => !selectedColumns.includes(i));
            setExcludedColumns(excluded);
        }
    }, [fileData]);

    const handleLogoClick = () => {
        setShowAboutDialog(true);
    };

    const generateCLICommand = (): string => {
        return generateCLICommandUtil({
            fileName,
            components: config.components,
            method: config.method,
            kernelType: config.kernelType,
            kernelGamma: config.kernelGamma,
            kernelDegree: config.kernelDegree,
            kernelCoef0: config.kernelCoef0,
            temporalLags: config.temporalLags,
            varianceExplained: config.varianceExplained,
            snv: config.snv,
            vectorNorm: config.vectorNorm,
            standardScale: config.standardScale,
            robustScale: config.robustScale,
            meanCenter: config.meanCenter,
            scaleOnly: config.scaleOnly,
            missingStrategy: config.missingStrategy,
            excludedColumns,
            excludedRows
        });
    };

    const copyToClipboard = async (text: string) => {
        try {
            await navigator.clipboard.writeText(text);
            setShowCopied(true);
            setTimeout(() => setShowCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy to clipboard:', err);
        }
    };

    const runPCA = async () => {
        if (!fileData) {
return;
}

        setLoading(true);
        setPcaError(null);

        try {
            const request: PCARequest = {
                data: fileData.data,
                missingMask: fileData.missingMask,
                headers: fileData.headers,
                rowNames: fileData.rowNames,
                ...config,
                excludedRows,
                excludedColumns,
                // Add group information if a group column is selected
                ...(selectedGroupColumn && fileData.categoricalColumns && {
                    groupColumn: selectedGroupColumn,
                    groupLabels: fileData.categoricalColumns[selectedGroupColumn]
                }),
                // Add metadata for eigencorrelations if available
                metadataNumeric: fileData.numericTargetColumns || {},
                metadataCategorical: fileData.categoricalColumns || {},
                calculateEigencorrelations: (fileData.numericTargetColumns && Object.keys(fileData.numericTargetColumns).length > 0) ||
                                          (fileData.categoricalColumns && Object.keys(fileData.categoricalColumns).length > 0)
            };
            const result = await RunPCA(request);
            if (result.success) {
                setPcaResponse(result);
                // Reset PC selections to default
                setSelectedXComponent(0);
                setSelectedYComponent(1);
                // Clear any previous errors
                setPcaError(null);

                // If we had excluded rows, update fileData to reflect the filtered dataset
                if (excludedRows.length > 0) {
                    setPcaHasExclusions(true);  // Mark that this PCA was run with exclusions
                    const includedIndices = fileData.data
                        .map((_, i) => i)
                        .filter(i => !excludedRows.includes(i));

                    const filteredData = includedIndices.map(i => fileData.data[i]);
                    const filteredRowNames = includedIndices.map(i => fileData.rowNames[i]);

                    // Update categorical and numeric columns if they exist
                    const filteredCategorical: Record<string, string[]> = {};
                    const filteredNumeric: Record<string, number[]> = {};

                    if (fileData.categoricalColumns) {
                        Object.keys(fileData.categoricalColumns).forEach(col => {
                            filteredCategorical[col] = includedIndices.map(i =>
                                fileData.categoricalColumns![col][i]
                            );
                        });
                    }

                    if (fileData.numericTargetColumns) {
                        Object.keys(fileData.numericTargetColumns).forEach(col => {
                            filteredNumeric[col] = includedIndices.map(i =>
                                fileData.numericTargetColumns![col][i]
                            );
                        });
                    }

                    // Clear excluded rows before updating fileData
                    setExcludedRows([]);

                    // Then update fileData with filtered dataset
                    setFileData({
                        ...fileData,
                        data: filteredData,
                        rowNames: filteredRowNames,
                        categoricalColumns: Object.keys(filteredCategorical).length > 0 ? filteredCategorical : undefined,
                        numericTargetColumns: Object.keys(filteredNumeric).length > 0 ? filteredNumeric : undefined
                    });
                    // Force table components to reset their selection state for the new dataset
                    setDatasetId(prev => prev + 1);
                } else {
                    setPcaHasExclusions(false);  // No exclusions in this PCA
                }

                // Check if Kernel PCA is selected with unsupported visualization
                if (config.method === 'kernel' &&
                    (selectedPlot === 'correlations' || selectedPlot === 'biplot' || selectedPlot === 'biplot3d')) {
                    // Switch to scores plot
                    setSelectedPlot('scores');
                    // Alert user about the automatic switch
                    alert('The selected visualization is not supported for Kernel PCA. Switching to Scores Plot.');
                }

                // Smooth scroll to results
                setTimeout(() => {
                    pcaResultsRef.current?.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }, 100);
            } else {
                setPcaError(result.error || 'PCA analysis failed');
                setPcaResponse(null);
                // Smooth scroll to error
                setTimeout(() => {
                    pcaErrorRef.current?.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }, 100);
            }
        } catch (err) {
            setPcaError(`Failed to run PCA: ${err}`);
        } finally {
            setLoading(false);
        }
    };

    const handleExportModel = async () => {
        if (!pcaResponse?.success || !pcaResponse.result || !fileData) {
return;
}

        try {
            const { ExportPCAModel } = await import('../wailsjs/go/main/App');
            const { ExportPCAModelRequest } = await import('../wailsjs/go/models').then(m => m.main);

            const request = new ExportPCAModelRequest({
                data: fileData.data,
                headers: fileData.headers,
                rowNames: fileData.rowNames,
                pcaResult: pcaResponse.result,
                config,
                excludedRows,
                excludedColumns,
                filename: fileName  // Add the original filename
            });

            await ExportPCAModel(request);
        } catch (err) {
            console.error('Failed to export model:', err);
            alert(`Failed to export model: ${err}`);
        }
    };

    return (
        <div className="flex flex-col h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white transition-colors duration-200">
            <header className="sticky top-0 z-50 bg-white dark:bg-gray-800 shadow-lg backdrop-blur-sm bg-opacity-95 dark:bg-opacity-95">
                <div className="flex items-center justify-between max-w-7xl mx-auto px-4 py-3 h-20">
                    <div className="flex items-center gap-4">
                        <HelpWrapper helpKey="logo-about">
                            <img
                                src={logo}
                                alt="GoPCA - Principal Component Analysis Tool"
                                className="h-12 cursor-pointer hover:opacity-90 transition-opacity flex-shrink-0"
                                onClick={handleLogoClick}
                            />
                        </HelpWrapper>
                    </div>
                    <div className="flex-1 mx-8 overflow-hidden">
                        <HelpDisplay
                            helpKey={currentHelpKey}
                            title={currentHelp?.title || ''}
                            text={currentHelp?.text || ''}
                        />
                    </div>
                    <div className="flex items-center gap-2">
                        <HelpWrapper helpKey="documentation-button">
                            <button
                                onClick={() => setShowDocumentation(true)}
                                className="p-2 rounded-lg bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors duration-200"
                                aria-label="Open documentation"
                            >
                                {/* Book icon */}
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                    strokeWidth={1.5}
                                    stroke="currentColor"
                                    className="w-5 h-5 text-gray-700 dark:text-gray-300"
                                >
                                    <path
                                        strokeLinecap="round"
                                        strokeLinejoin="round"
                                        d="M12 6.042A8.967 8.967 0 006 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 016 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 016-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0018 18c-2.305 0-4.408.867-6 2.292m0-14.25v14.25"
                                    />
                                </svg>
                            </button>
                        </HelpWrapper>
                        <HelpWrapper helpKey="theme-toggle">
                            <ThemeToggle />
                        </HelpWrapper>
                    </div>
                </div>
            </header>

            <main ref={mainScrollRef} className="flex-1 overflow-auto p-6">
                <div className="max-w-7xl mx-auto space-y-6">
                    {/* File Upload Section */}
                    <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-200 dark:border-gray-700">
                        <h2 className="text-xl font-semibold mb-6">Step 1: Load Data</h2>
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-[1fr_2fr_1fr] gap-6">
                            {/* Column 1: File Upload */}
                            <HelpWrapper helpKey="data-upload" className="flex flex-col justify-center">
                                <label className="block text-sm font-medium mb-3">
                                    Upload Your CSV File
                                </label>
                                <HelpWrapper helpKey="choose-file">
                                    <button
                                        onClick={handleNativeFileSelect}
                                        disabled={loading}
                                        className="w-full px-4 py-2 text-sm font-semibold text-white bg-blue-600 rounded-full hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                    >
                                        {loading ? 'Loading...' : 'Choose File'}
                                    </button>
                                </HelpWrapper>
                                <p className="mt-2 text-xs text-gray-500 dark:text-gray-400">
                                    Accepts CSV files with headers
                                </p>

                                {/* GoCSV Integration Button */}
                                <div className="mt-4">
                                    <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                                        Or Use the Data Editor
                                    </p>
                                    <HelpWrapper helpKey="gocsv-integration">
                                        <button
                                            onClick={handleGoCSVAction}
                                            disabled={isCheckingGoCSV}
                                            className="w-full px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                        >
                                            {isCheckingGoCSV ? 'Checking...' :
                                             !goCSVStatus?.installed ? 'Install GoCSV' :
                                             'Open GoCSV'}
                                        </button>
                                    </HelpWrapper>
                                </div>
                            </HelpWrapper>

                            {/* Column 2: Matrix Illustration */}
                            <div className="flex items-center justify-center border-0 md:border-x lg:border-x border-gray-200 dark:border-gray-700 px-4 py-6 md:py-0">
                                <MatrixIllustration />
                            </div>

                            {/* Column 3: Sample Datasets */}
                            <div className="flex flex-col justify-center md:col-span-2 lg:col-span-1">
                                <label className="block text-sm font-medium mb-3">
                                    Or Try Sample Datasets
                                </label>
                                <div className="space-y-2">
                                    <HelpWrapper helpKey="sample-dataset-corn">
                                        <button
                                            onClick={() => loadDataset('corn.csv')}
                                            className="w-full px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                            disabled={loading}
                                        >
                                            Corn (NIR)
                                        </button>
                                    </HelpWrapper>
                                    <HelpWrapper helpKey="sample-dataset-iris">
                                        <button
                                            onClick={() => loadDataset('iris.csv', 'species')}
                                            className="w-full px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                            disabled={loading}
                                        >
                                            Iris
                                        </button>
                                    </HelpWrapper>
                                    <HelpWrapper helpKey="sample-dataset-wine">
                                        <button
                                            onClick={() => loadDataset('wine.csv', 'target')}
                                            className="w-full px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                            disabled={loading}
                                        >
                                            Wine
                                        </button>
                                    </HelpWrapper>
                                    <HelpWrapper helpKey="sample-dataset-swiss-roll">
                                        <button
                                            onClick={() => loadDataset('swiss_roll.csv', 'color #target')}
                                            className="w-full px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                            disabled={loading}
                                        >
                                            Swiss Roll
                                        </button>
                                    </HelpWrapper>
                                    <HelpWrapper helpKey="sample-dataset-stocks">
                                        <button
                                            onClick={() => loadDataset('stocks.csv')}
                                            className="w-full px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                            disabled={loading}
                                        >
                                            Stocks
                                        </button>
                                    </HelpWrapper>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* File Error Display */}
                    {fileError && (
                        <ErrorAlert
                            type="error"
                            title="File Error"
                            message={fileError}
                            onDismiss={() => setFileError(null)}
                        />
                    )}

                    {/* Data Display */}
                    {fileData && (
                        <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-200 dark:border-gray-700">
                            <h2 className="text-xl font-semibold mb-4">Loaded Data</h2>
                            {/* Check if dataset is large (>10,000 cells) */}
                            {fileData.data.length * fileData.headers.length > 10000 ? (
                                <SelectionTable
                                    key={`dataset-${datasetId}`}
                                    headers={fileData.headers}
                                    rowNames={fileData.rowNames}
                                    data={fileData.data}
                                    title="Input Data"
                                    onRowSelectionChange={handleRowSelectionChange}
                                    onColumnSelectionChange={handleColumnSelectionChange}
                                    externalSelectedRows={fileData.data.map((_, i) => i).filter(i => !excludedRows.includes(i))}
                                    highlightExternalSelections={true}
                                />
                            ) : (
                                <ErrorBoundary
                                    onError={(error, errorInfo) => {
                                        console.error('DataTable Error:', error, errorInfo);
                                    }}
                                >
                                    <DataTable
                                        key={`dataset-${datasetId}`}
                                        headers={fileData.headers}
                                        rowNames={fileData.rowNames}
                                        data={fileData.data}
                                        title="Input Data"
                                    enableRowSelection={true}
                                    enableColumnSelection={true}
                                    onRowSelectionChange={handleRowSelectionChange}
                                    onColumnSelectionChange={handleColumnSelectionChange}
                                    externalSelectedRows={fileData.data.map((_, i) => i).filter(i => !excludedRows.includes(i))}
                                    highlightExternalSelections={true}
                                />
                                </ErrorBoundary>
                            )}
                        </div>
                    )}

                    {/* Configuration Section */}
                    {fileData && (
                        <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-200 dark:border-gray-700">
                            <h2 className="text-xl font-semibold mb-6">Step 2: Configure PCA</h2>

                            {/* Two-column layout */}
                            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                                {/* Left Column - Core PCA Configuration */}
                                <div className="space-y-4">
                                    <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">PCA Options</h3>

                                    {/* Basic Configuration */}
                                    <div className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg space-y-4">
                                        <HelpWrapper helpKey="num-components">
                                            <label className="block text-sm font-medium mb-2">
                                                Number of Components
                                            </label>
                                            <input
                                                type="number"
                                                min="1"
                                                max={Math.min(fileData.headers.length, fileData.data.length)}
                                                value={config.components}
                                                onChange={(e) => setConfig({ ...config, components: parseInt(e.target.value) || 5 })}
                                                className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                            />
                                        </HelpWrapper>

                                        <HelpWrapper helpKey="pca-method">
                                            <label className="block text-sm font-medium mb-2">
                                                Method
                                            </label>
                                            <CustomSelect
                                            value={config.method}
                                            onChange={(value) => {
                                                const newMethod = value;
                                                const oldMethod = config.method;
                                                const newConfig = { ...config, method: newMethod };

                                                // If switching to kernel PCA and current preprocessing is invalid
                                                if (newMethod === 'kernel') {
                                                    // Check if current preprocessing is invalid for kernel PCA
                                                    // Valid options for kernel PCA are: none (all false) or scale-only
                                                    if (newConfig.meanCenter || newConfig.standardScale || newConfig.robustScale) {
                                                        // Reset to "None" - the default valid option
                                                        newConfig.meanCenter = false;
                                                        newConfig.standardScale = false;
                                                        newConfig.robustScale = false;
                                                        newConfig.scaleOnly = false;
                                                    }
                                                    // scaleOnly is valid, so we keep it as-is
                                                } else if (oldMethod === 'kernel' && newMethod !== 'kernel') {
                                                    // When switching FROM kernel PCA to other methods, restore default preprocessing
                                                    // This prevents the bug where SVD runs on uncentered data after using kernel PCA
                                                    // Restore default preprocessing for standard PCA methods
                                                    // Mean centering is the default and most important for SVD/NIPALS
                                                    newConfig.meanCenter = true;
                                                    newConfig.standardScale = false;
                                                    newConfig.robustScale = false;
                                                    newConfig.scaleOnly = false;
                                                }

                                                setConfig(newConfig);
                                            }}
                                            options={[
                                                { value: 'SVD', label: 'SVD' },
                                                { value: 'NIPALS', label: 'NIPALS' },
                                                { value: 'kernel', label: 'Kernel PCA' },
                                                { value: 'temporal', label: 'Temporal PCA' }
                                            ]}
                                            className="w-full"
                                        />
                                        </HelpWrapper>
                                    </div>

                                    {/* Method-specific information */}
                                    {config.method === 'SVD' && (
                                        <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg space-y-3">
                                            <h4 className="font-medium text-sm text-blue-900 dark:text-blue-100">SVD Method</h4>
                                            <div className="space-y-2 text-sm text-blue-800 dark:text-blue-200">
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Gold standard for PCA using Singular Value Decomposition</span>
                                                </p>
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Fast and numerically stable for complete datasets</span>
                                                </p>
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Computes all components simultaneously</span>
                                                </p>
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Best choice for most applications</span>
                                                </p>
                                            </div>
                                        </div>
                                    )}

                                    {config.method === 'NIPALS' && (
                                        <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg space-y-3">
                                            <h4 className="font-medium text-sm text-green-900 dark:text-green-100">NIPALS Method</h4>
                                            <div className="space-y-2 text-sm text-green-800 dark:text-green-200">
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Nonlinear Iterative Partial Least Squares algorithm</span>
                                                </p>
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Handles missing data gracefully</span>
                                                </p>
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Computes components sequentially</span>
                                                </p>
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Ideal for large datasets when only few components needed</span>
                                                </p>
                                            </div>
                                        </div>
                                    )}

                                    {/* Kernel PCA Options */}
                                    {config.method === 'kernel' && (
                                        <div className="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg space-y-4">
                                            <h4 className="font-medium text-sm">Kernel PCA Options</h4>

                                            {/* Memory warning for large datasets */}
                                            {fileData && fileData.data && fileData.data.length > 5000 && (
                                                <div className="p-3 bg-yellow-100 dark:bg-yellow-900/50 border border-yellow-300 dark:border-yellow-700 rounded-lg">
                                                    <p className="text-sm text-yellow-800 dark:text-yellow-200">
                                                        <strong>⚠️ Warning:</strong> Kernel PCA with {fileData.data.length.toLocaleString()} samples
                                                        may require significant memory (~{Math.round(fileData.data.length * fileData.data.length * 24 / 1024 / 1024 / 1024 * 10) / 10} GB).
                                                        {fileData.data.length > 10000 && (
                                                            <span className="block mt-1 font-semibold">
                                                                Maximum supported: 10,000 samples. Consider using SVD or NIPALS instead.
                                                            </span>
                                                        )}
                                                    </p>
                                                </div>
                                            )}

                                            <div className="space-y-4">
                                                <HelpWrapper helpKey="kernel-type">
                                                    <label className="block text-sm font-medium mb-1">
                                                        Kernel Type
                                                    </label>
                                                    <CustomSelect
                                                        value={config.kernelType}
                                                        onChange={(value) => setConfig({ ...config, kernelType: value })}
                                                        options={[
                                                            { value: 'rbf', label: 'RBF (Gaussian)' },
                                                            { value: 'linear', label: 'Linear' },
                                                            { value: 'poly', label: 'Polynomial' }
                                                        ]}
                                                        className="w-full"
                                                    />
                                                </HelpWrapper>
                                                <HelpWrapper helpKey="kernel-gamma">
                                                    <label className="block text-sm font-medium mb-1">
                                                        Gamma
                                                    </label>
                                                    <input
                                                        type="number"
                                                        value={config.kernelGamma}
                                                        step="0.01"
                                                        min="0.001"
                                                        onChange={(e) => {
                                                            const value = parseFloat(e.target.value);
                                                            setConfig({ ...config, kernelGamma: isNaN(value) ? 1.0 : value });
                                                        }}
                                                        className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                                    />
                                                </HelpWrapper>
                                                {config.kernelType === 'poly' && (
                                                    <>
                                                        <HelpWrapper helpKey="kernel-degree">
                                                            <label className="block text-sm font-medium mb-1">
                                                                Degree
                                                            </label>
                                                            <input
                                                                type="number"
                                                                value={config.kernelDegree}
                                                                min="1"
                                                                max="10"
                                                                onChange={(e) => setConfig({ ...config, kernelDegree: parseInt(e.target.value) || 3 })}
                                                                className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                                            />
                                                        </HelpWrapper>
                                                        <HelpWrapper helpKey="kernel-coef0">
                                                            <label className="block text-sm font-medium mb-1">
                                                                Coef0
                                                            </label>
                                                            <input
                                                                type="number"
                                                                value={config.kernelCoef0}
                                                                step="0.1"
                                                                onChange={(e) => {
                                                                    const value = parseFloat(e.target.value);
                                                                    setConfig({ ...config, kernelCoef0: isNaN(value) ? 0.0 : value });
                                                                }}
                                                                className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                                            />
                                                        </HelpWrapper>
                                                    </>
                                                )}
                                            </div>
                                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                                                Note: Kernel PCA uses its own centering in kernel space.
                                            </p>
                                        </div>
                                    )}

                                    {/* Temporal PCA Options */}
                                    {config.method === 'temporal' && (
                                        <div className="p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg space-y-4">
                                            <h4 className="font-medium text-sm text-purple-900 dark:text-purple-100">Temporal PCA Options</h4>

                                            <div className="space-y-2 text-sm text-purple-800 dark:text-purple-200">
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Time-Delay PCA for time-series analysis</span>
                                                </p>
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Captures temporal dynamics and dependencies</span>
                                                </p>
                                                <p className="flex items-start">
                                                    <span className="mr-2">•</span>
                                                    <span>Based on SSA (Singular Spectrum Analysis) methodology</span>
                                                </p>
                                            </div>

                                            <div className="space-y-4">
                                                <HelpWrapper helpKey="temporal-lags">
                                                    <label className="block text-sm font-medium mb-1">
                                                        Number of Time Lags
                                                    </label>
                                                    <input
                                                        type="number"
                                                        value={config.temporalLags}
                                                        min="2"
                                                        max="100"
                                                        onChange={(e) => {
                                                            const value = parseInt(e.target.value);
                                                            setConfig({ ...config, temporalLags: isNaN(value) || value < 2 ? 2 : value });
                                                        }}
                                                        className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                                    />
                                                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                                        Number of time points to include in lag matrix. Use 24 for daily cycles in hourly data, 7 for weekly patterns in daily data.
                                                    </p>
                                                </HelpWrapper>

                                                <div className="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
                                                    <p className="text-xs text-purple-800 dark:text-purple-200">
                                                        <strong>Lag Selection Guidelines:</strong><br/>
                                                        • Hourly data with daily patterns: L = 24<br/>
                                                        • Daily data with weekly patterns: L = 7<br/>
                                                        • Monthly data with annual patterns: L = 12<br/>
                                                        • General exploration: Start with T/4 (where T = number of samples)
                                                    </p>
                                                </div>

                                                {fileData && fileData.data && config.temporalLags >= fileData.data.length && (
                                                    <div className="p-3 bg-yellow-100 dark:bg-yellow-900/50 border border-yellow-300 dark:border-yellow-700 rounded-lg">
                                                        <p className="text-sm text-yellow-800 dark:text-yellow-200">
                                                            <strong>⚠️ Warning:</strong> Number of lags ({config.temporalLags}) should be less than the number of samples ({fileData.data.length}).
                                                            Recommended: {Math.floor(fileData.data.length / 4)} lags or less.
                                                        </p>
                                                    </div>
                                                )}
                                            </div>

                                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                                                Creates a lag matrix where each row contains L consecutive observations, enabling capture of temporal patterns.
                                            </p>
                                        </div>
                                    )}
                                </div>

                                {/* Right Column - Preprocessing Options */}
                                <div className="space-y-4">
                                    <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">Preprocessing Options</h3>

                                    {/* Step 1: Row-wise preprocessing */}
                                    <HelpWrapper helpKey="row-preprocessing" className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                                        <label className="block text-sm font-medium mb-2">
                                            Step 1: Row-wise Preprocessing (optional)
                                        </label>
                                        <CustomSelect
                                            value={
                                                config.snv ? 'snv' :
                                                config.vectorNorm ? 'vector-norm' :
                                                'none'
                                            }
                                            onChange={(value) => {
                                                setConfig({
                                                    ...config,
                                                    snv: value === 'snv',
                                                    vectorNorm: value === 'vector-norm'
                                                });
                                            }}
                                            options={[
                                                { value: 'none', label: 'None' },
                                                { value: 'snv', label: 'SNV (Standard Normal Variate)' },
                                                { value: 'vector-norm', label: 'L2 Vector Normalization' }
                                            ]}
                                            className="w-full"
                                        />
                                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                            Normalizes each row/sample independently (useful for spectral data)
                                        </p>
                                    </HelpWrapper>

                                    {/* Step 2: Column-wise preprocessing */}
                                    <HelpWrapper helpKey="column-preprocessing" className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                                        <label className="block text-sm font-medium mb-2">
                                            Step 2: Column-wise Preprocessing
                                        </label>
                                        <CustomSelect
                                            value={
                                                config.scaleOnly ? 'scale-only' :
                                                config.robustScale ? 'robust' :
                                                config.standardScale ? 'standard' :
                                                config.meanCenter ? 'center' : 'none'
                                            }
                                            onChange={(value) => {
                                                // For kernel PCA, only allow none or scale-only
                                                if (config.method === 'kernel' && !['none', 'scale-only'].includes(value)) {
                                                    return;
                                                }
                                                setConfig({
                                                    ...config,
                                                    meanCenter: value === 'center' || value === 'standard',
                                                    standardScale: value === 'standard',
                                                    robustScale: value === 'robust',
                                                    scaleOnly: value === 'scale-only'
                                                });
                                            }}
                                            options={[
                                                { value: 'none', label: 'None (Raw Data)' },
                                                { value: 'center', label: 'Mean Center Only', disabled: config.method === 'kernel' },
                                                { value: 'standard', label: 'Standard Scale (Mean + Std Dev)', disabled: config.method === 'kernel' },
                                                { value: 'robust', label: 'Robust Scale (Median + MAD)', disabled: config.method === 'kernel' },
                                                { value: 'scale-only', label: 'Variance Scale (Std Dev Only)' }
                                            ]}
                                            className="w-full"
                                        />
                                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                            {config.method === 'kernel'
                                                ? config.scaleOnly
                                                    ? 'Variance scaling divides by standard deviation without centering - suitable for Kernel PCA'
                                                    : 'Kernel PCA performs centering in kernel space. Consider Variance Scale if features have different scales.'
                                                : 'Normalizes each column/feature across all samples'}
                                        </p>
                                    </HelpWrapper>

                                    {/* Missing Data Strategy */}
                                    <HelpWrapper helpKey="missing-strategy" className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                                        <label className="block text-sm font-medium mb-2">
                                            Missing Data Strategy
                                        </label>
                                        <CustomSelect
                                            value={config.missingStrategy}
                                            onChange={(value) => setConfig({ ...config, missingStrategy: value })}
                                            options={[
                                                { value: 'error', label: 'Show Error (default)' },
                                                { value: 'drop', label: 'Drop Rows with Missing Values' },
                                                { value: 'mean', label: 'Impute with Column Mean' },
                                                { value: 'median', label: 'Impute with Column Median' },
                                                { value: 'native', label: 'Native NIPALS Handling (NIPALS only)' }
                                            ]}
                                            className="w-full"
                                        />
                                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                            Choose how to handle missing values (NaN) in your data
                                        </p>
                                    </HelpWrapper>

                                    {/* Diagnostic Metrics Option */}
                                </div>
                            </div>

                            {/* Go PCA! button - centered and spanning both columns */}
                            <div className="mt-6 flex justify-center">
                                <HelpWrapper helpKey="go-pca-button">
                                    <button
                                        onClick={runPCA}
                                        disabled={loading}
                                        className="px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 dark:disabled:bg-gray-600 rounded-lg font-medium text-white"
                                    >
                                        {loading ? 'Running...' : 'Go PCA!'}
                                    </button>
                                </HelpWrapper>
                            </div>

                            {/* CLI Command Preview */}
                            {fileData && fileName && (
                                <div className="mt-4 bg-gray-900 dark:bg-gray-950 rounded-lg p-4 border border-gray-700">
                                    <div className="flex items-center justify-between gap-3">
                                        <div className="flex items-center gap-3 flex-1">
                                            <span className="text-sm font-medium text-gray-300">Command line:</span>
                                            <HelpWrapper helpKey="cli-command-preview">
                                                <div className="flex-1 bg-black rounded px-3 py-2 font-mono text-xs text-green-400 overflow-x-auto">
                                                    {generateCLICommand()}
                                                </div>
                                            </HelpWrapper>
                                        </div>
                                        <button
                                            onClick={() => copyToClipboard(generateCLICommand())}
                                            className="px-2 py-1 bg-gray-700 hover:bg-gray-600 rounded text-white transition-colors flex-shrink-0"
                                            title="Copy command"
                                        >
                                            {showCopied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}

                    {/* PCA Error Display - shown between Step 2 and Results */}
                    {pcaError && fileData && (
                        <div ref={pcaErrorRef}>
                            <ErrorAlert
                                type="error"
                                title="Analysis Error"
                                message={pcaError}
                                suggestion="Please check your data and parameters, then try again"
                                onDismiss={() => setPcaError(null)}
                            />
                        </div>
                    )}

                    {/* PCA Results */}
                    {pcaResponse?.success && pcaResponse.result && (
                        <div ref={pcaResultsRef} className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-200 dark:border-gray-700">
                            <h2 className="text-xl font-semibold mb-4">Step 3: Interpret PCA Model</h2>

                            {/* Info message about missing data handling */}
                            {pcaResponse.info && (
                                <div className="mb-4 p-3 bg-blue-100 dark:bg-blue-800 border border-blue-300 dark:border-blue-600 rounded-lg">
                                    <p className="text-blue-700 dark:text-blue-200 text-sm">
                                        <span className="font-semibold">Note:</span> {pcaResponse.info}
                                    </p>
                                </div>
                            )}

                            {/* Explained Variance and Model Overview Grid */}
                            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch">
                                {/* Explained Variance */}
                                <HelpWrapper helpKey="explained-variance" className="h-full">
                                    <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4 h-full flex flex-col">
                                        <div className="mb-2">
                                            <h3 className="text-lg font-semibold">Explained Variance</h3>
                                        </div>
                                        <div className="space-y-2 flex-grow">
                                            {pcaResponse.result.explained_variance_ratio.map((percentage, i) => {
                                                return (
                                                    <div key={i} className="flex justify-between">
                                                        <span>{pcaResponse.result?.component_labels?.[i] || `PC${i+1}`}:</span>
                                                        <span>{percentage.toFixed(2)}%</span>
                                                    </div>
                                                );
                                            })}
                                            <div className="border-t border-gray-300 dark:border-gray-600 pt-2 font-semibold">
                                                <div className="flex justify-between">
                                                    <span>Cumulative:</span>
                                                    <span>
                                                        {pcaResponse.result.cumulative_variance[pcaResponse.result.cumulative_variance.length - 1].toFixed(2)}%
                                                    </span>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </HelpWrapper>

                                {/* Model Overview */}
                                <ModelOverview
                                    pcaResult={pcaResponse.result}
                                    selectedPC={selectedXComponent}
                                    standardScale={config.standardScale}
                                    originalData={fileData?.data}
                                />
                            </div>

                            {/* Plot Selector and Visualization */}
                            <div className="mt-6">
                                {/* Tier 1: Primary Controls */}
                                <div className="flex items-center justify-between mb-3 pb-3 border-b border-gray-200 dark:border-gray-600">
                                    <div className="flex items-center gap-4">
                                        <h3 className="text-lg font-semibold">Visualizations</h3>
                                        <HelpWrapper helpKey={`${selectedPlot}-plot`}>
                                            <CustomSelect
                                                value={selectedPlot}
                                                onChange={(value) => setSelectedPlot(value as 'scores' | 'scores3d' | 'scree' | 'loadings' | 'biplot' | 'biplot3d' | 'correlations' | 'diagnostics' | 'eigencorrelation' | 'temporal-loadings' | 'temporal-variable-importance' | 'kernel-matrix' | 'sample-contributions')}
                                                options={[
                                                    { value: 'scores', label: 'Scores Plot' },
                                                    { value: 'scores3d', label: '3D Scores Plot' },
                                                    { value: 'scree', label: 'Scree Plot' },
                                                    // Loadings plot not available for kernel PCA (different space) or temporal PCA (dimension mismatch)
                                                    ...(pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? [{ value: 'loadings', label: 'Loadings Plot' }] : []),
                                                    // Temporal loadings pattern - only for temporal PCA
                                                    ...(pcaResponse.result.method === 'temporal' ? [{ value: 'temporal-loadings', label: 'Temporal Loadings' }] : []),
                                                    // Variable importance plot - only for temporal PCA
                                                    ...(pcaResponse.result.method === 'temporal' ? [{ value: 'temporal-variable-importance', label: 'Variable Importance' }] : []),
                                                    // Biplot - available for standard PCA with preprocessing (not for kernel PCA or temporal PCA)
                                                    ...(pcaResponse.result.preprocessing_applied && pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? [{ value: 'biplot', label: 'Biplot' }] : []),
                                                    // 3D Biplot and Circle of Correlations - not available for kernel PCA or temporal PCA
                                                    ...(pcaResponse.result.preprocessing_applied && pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? [{ value: 'biplot3d', label: '3D Biplot' }] : []),
                                                    ...(pcaResponse.result.preprocessing_applied && pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? [{ value: 'correlations', label: 'Circle of Correlations' }] : []),
                                                    // Diagnostic plot not available for:
                                                    // - Kernel PCA: Works in transformed feature space, RSS calculation not meaningful
                                                    // - Temporal PCA: Dimension mismatch - scores have n-lags+1 samples while original data has n samples
                                                    ...(pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? [{ value: 'diagnostics', label: 'Diagnostic Plot' }] : []),
                                                    ...(pcaResponse.result.eigencorrelations && pcaResponse.result.method !== 'kernel' ? [{ value: 'eigencorrelation', label: 'Eigencorrelation Plot' }] : []),
                                                    // Kernel PCA specific visualizations
                                                    ...(pcaResponse.result.method === 'kernel' ? [
                                                        { value: 'kernel-matrix', label: 'Kernel Matrix Heatmap' },
                                                        { value: 'sample-contributions', label: 'Sample Contributions' }
                                                    ] : [])
                                                ]}
                                                className="min-w-[200px]"
                                            />
                                        </HelpWrapper>
                                    </div>
                                    <div className="flex-shrink-0">
                                        <HelpWrapper helpKey="font-size-control">
                                            <FontSizeControl value={plotFontScale} onChange={setPlotFontScale} />
                                        </HelpWrapper>
                                    </div>
                                    <div className="flex-shrink-0">
                                        <HelpWrapper helpKey="palette-selector">
                                            <PaletteSelector />
                                        </HelpWrapper>
                                    </div>
                                </div>

                                {/* Tier 2: Context-Sensitive Controls */}
                                <div className="mb-4">
                                    <div className="flex flex-wrap items-center gap-4">
                                        {/* Data Display Group - Always show Color by control since Row Index is always available */}
                                        {(selectedPlot === 'scores' || selectedPlot === 'scores3d' || selectedPlot === 'biplot' || selectedPlot === 'biplot3d' || selectedPlot === 'diagnostics') && fileData && (
                                            <div className="flex items-center gap-3 px-3 py-2 bg-gray-50 dark:bg-gray-800 rounded-lg">
                                                <HelpWrapper helpKey="group-coloring" className="flex items-center gap-2">
                                                    <label className="text-sm text-gray-600 dark:text-gray-400">Color by:</label>
                                                    <CustomSelect
                                                        value={selectedGroupColumn || ''}
                                                        onChange={(value) => {
                                                            const selectedValue = value || null;
                                                            setSelectedGroupColumn(selectedValue);

                                                            // Auto-switch palette mode based on column type
                                                            if (!selectedValue) {
                                                                setMode('none');
                                                                setShowEllipses(false);
                                                            } else if (selectedValue === 'Row Index') {
                                                                // Row Index is always continuous
                                                                setMode('continuous');
                                                                setShowEllipses(false); // Ellipses only work with categorical data
                                                            } else if (fileData.numericTargetColumns && selectedValue in fileData.numericTargetColumns) {
                                                                setMode('continuous');
                                                                setShowEllipses(false); // Ellipses only work with categorical data
                                                            } else if (fileData.categoricalColumns && selectedValue in fileData.categoricalColumns) {
                                                                setMode('categorical');
                                                                // Keep current showEllipses state for categorical columns
                                                            }
                                                        }}
                                                        options={[
                                                            { value: '', label: 'None' },
                                                            // Row Index - always available as a continuous option
                                                            { value: 'Row Index', label: '📊 Row Index', group: 'Continuous' },
                                                            // Categorical columns
                                                            ...(fileData.categoricalColumns && Object.keys(fileData.categoricalColumns).length > 0
                                                                ? Object.keys(fileData.categoricalColumns).map((colName) => ({
                                                                    value: colName,
                                                                    label: `🏷️ ${colName}`,
                                                                    group: 'Categorical'
                                                                }))
                                                                : []),
                                                            // Continuous columns
                                                            ...(fileData.numericTargetColumns && Object.keys(fileData.numericTargetColumns).length > 0
                                                                ? Object.keys(fileData.numericTargetColumns).map((colName) => ({
                                                                    value: colName,
                                                                    label: `📊 ${colName}`,
                                                                    group: 'Continuous'
                                                                }))
                                                                : [])
                                                        ]}
                                                        className="min-w-[150px]"
                                                    />
                                                </HelpWrapper>
                                            </div>
                                        )}

                                        {/* Plot Options Group - For Scores Plot, 3D Scores Plot, Biplot, and Diagnostic Plot */}
                                        {(selectedPlot === 'scores' || selectedPlot === 'scores3d' || selectedPlot === 'biplot' || selectedPlot === 'biplot3d' || selectedPlot === 'diagnostics') && (
                                            <div className="flex items-center gap-3 px-3 py-2 bg-gray-50 dark:bg-gray-800 rounded-lg">
                                                <HelpWrapper helpKey="row-labels" className="flex items-center gap-2">
                                                    <label className="text-sm text-gray-600 dark:text-gray-400">
                                                        <input
                                                            type="checkbox"
                                                            checked={showRowLabels}
                                                            onChange={(e) => setShowRowLabels(e.target.checked)}
                                                            className="mr-1"
                                                        />
                                                        Show labels
                                                    </label>
                                                </HelpWrapper>
                                                {showRowLabels && (
                                                    <div className="flex items-center gap-2">
                                                        <label className="text-sm text-gray-600 dark:text-gray-400">Max:</label>
                                                        <input
                                                            type="number"
                                                            min="5"
                                                            max="50"
                                                            value={maxLabelsToShow}
                                                            onChange={(e) => setMaxLabelsToShow(parseInt(e.target.value) || 10)}
                                                            className="w-12 px-1 py-0.5 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white"
                                                        />
                                                    </div>
                                                )}
                                                {/* Confidence Level for Diagnostic Plot Thresholds */}
                                                {selectedPlot === 'diagnostics' && (
                                                    <div className="flex items-center gap-2 ml-3">
                                                        <HelpWrapper helpKey="diagnostic-threshold">
                                                            <label className="text-sm text-gray-600 dark:text-gray-400">
                                                                Threshold:
                                                            </label>
                                                        </HelpWrapper>
                                                        <CustomSelect
                                                            value={confidenceLevel.toFixed(2)}
                                                            onChange={(value) => setConfidenceLevel(parseFloat(value) as 0.90 | 0.95 | 0.99)}
                                                            options={[
                                                                { value: '0.95', label: '95%' },
                                                                { value: '0.99', label: '99%' }
                                                            ]}
                                                            className="min-w-[80px]"
                                                        />
                                                    </div>
                                                )}
                                                {fileData?.categoricalColumns &&
                                                 Object.keys(fileData.categoricalColumns).length > 0 &&
                                                 selectedGroupColumn &&
                                                 getColumnData(selectedGroupColumn).type === 'categorical' &&
                                                 (selectedPlot === 'scores' || selectedPlot === 'biplot' || selectedPlot === 'biplot3d') && (
                                                    <>
                                                        <div className="w-px h-5 bg-gray-300 dark:bg-gray-600 mx-1" />
                                                        <HelpWrapper helpKey="confidence-ellipses" className="flex items-center gap-2">
                                                            <label className="text-sm text-gray-600 dark:text-gray-400">
                                                                <input
                                                                    type="checkbox"
                                                                    checked={showEllipses}
                                                                    onChange={(e) => setShowEllipses(e.target.checked)}
                                                                    className="mr-1"
                                                                />
                                                                Ellipses
                                                            </label>
                                                        </HelpWrapper>
                                                        {showEllipses && (
                                                            <CustomSelect
                                                                value={confidenceLevel.toFixed(2)}
                                                                onChange={(value) => setConfidenceLevel(parseFloat(value) as 0.90 | 0.95 | 0.99)}
                                                                options={[
                                                                    { value: '0.90', label: '90%' },
                                                                    { value: '0.95', label: '95%' },
                                                                    { value: '0.99', label: '99%' }
                                                                ]}
                                                                className="min-w-[80px]"
                                                            />
                                                        )}
                                                    </>
                                                )}
                                            </div>
                                        )}

                                        {/* Component Selectors Group */}
                                        {(selectedPlot === 'scores' || selectedPlot === 'scores3d' || selectedPlot === 'biplot' || selectedPlot === 'biplot3d' || selectedPlot === 'correlations') && pcaResponse.result.scores[0]?.length > 2 && (
                                            <div className="flex items-center gap-3 px-3 py-2 bg-gray-50 dark:bg-gray-800 rounded-lg">
                                                <div className="flex items-center gap-2">
                                                    <label className="text-sm text-gray-600 dark:text-gray-400">X:</label>
                                                    <CustomSelect
                                                        value={selectedXComponent.toString()}
                                                        onChange={(value) => setSelectedXComponent(parseInt(value))}
                                                        options={pcaResponse.result.component_labels?.map((label, i) => ({
                                                            value: i.toString(),
                                                            label: `${label} (${pcaResponse.result!.explained_variance_ratio[i].toFixed(1)}%)`
                                                        })) || []}
                                                        className="min-w-[150px]"
                                                    />
                                                </div>
                                                <div className="flex items-center gap-2">
                                                    <label className="text-sm text-gray-600 dark:text-gray-400">Y:</label>
                                                    <CustomSelect
                                                        value={selectedYComponent.toString()}
                                                        onChange={(value) => setSelectedYComponent(parseInt(value))}
                                                        options={pcaResponse.result.component_labels?.map((label, i) => ({
                                                            value: i.toString(),
                                                            label: `${label} (${pcaResponse.result!.explained_variance_ratio[i].toFixed(1)}%)`
                                                        })) || []}
                                                        className="min-w-[150px]"
                                                    />
                                                </div>
                                                {(selectedPlot === 'scores3d' || selectedPlot === 'biplot3d') && pcaResponse.result.scores[0]?.length > 2 && (
                                                    <div className="flex items-center gap-2">
                                                        <label className="text-sm text-gray-600 dark:text-gray-400">Z:</label>
                                                        <CustomSelect
                                                            value={selectedZComponent.toString()}
                                                            onChange={(value) => setSelectedZComponent(parseInt(value))}
                                                            options={pcaResponse.result.component_labels?.map((label, i) => ({
                                                                value: i.toString(),
                                                                label: `${label} (${pcaResponse.result!.explained_variance_ratio[i].toFixed(1)}%)`
                                                            })) || []}
                                                            className="min-w-[150px]"
                                                        />
                                                    </div>
                                                )}
                                            </div>
                                        )}

                                        {/* Loadings Plot Component Selector */}
                                        {selectedPlot === 'loadings' && pcaResponse.result?.method !== 'kernel' && (
                                            <div className="flex items-center gap-3 px-3 py-2 bg-gray-50 dark:bg-gray-800 rounded-lg">
                                                <label className="text-sm text-gray-600 dark:text-gray-400">Component:</label>
                                                <CustomSelect
                                                    value={selectedLoadingComponent.toString()}
                                                    onChange={(value) => setSelectedLoadingComponent(parseInt(value))}
                                                    options={pcaResponse.result?.component_labels?.map((label, i) => ({
                                                        value: i.toString(),
                                                        label: `${label} (${pcaResponse.result!.explained_variance_ratio[i].toFixed(1)}%)`
                                                    })) || []}
                                                    className="min-w-[150px]"
                                                />
                                                <div className="w-px h-5 bg-gray-300 dark:bg-gray-600 mx-1" />
                                                <label className="text-sm text-gray-600 dark:text-gray-400">Plot type:</label>
                                                <CustomSelect
                                                    value={loadingsPlotType || (pcaResponse.result?.loadings[0]?.length > (guiConfig?.visualization?.loadings_variable_threshold || 100) ? 'line' : 'bar')}
                                                    onChange={(value) => {
                                                        setLoadingsPlotType(value as 'bar' | 'line');
                                                    }}
                                                    options={[
                                                        { value: 'bar', label: 'Bar Chart' },
                                                        { value: 'line', label: 'Line Chart' }
                                                    ]}
                                                    className="min-w-[120px]"
                                                />
                                            </div>
                                        )}
                                    </div>
                                </div>

                                <div className="bg-gray-50 dark:bg-gray-700 rounded-lg" style={{ height: '500px' }}>
                                    <ErrorBoundary
                                        onError={(error, errorInfo) => {
                                            console.error('Visualization Error:', error, errorInfo);
                                        }}
                                    >
                                        <Suspense fallback={
                                            <div className="w-full h-full flex items-center justify-center">
                                                <div className="flex flex-col items-center gap-4">
                                                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
                                                    <p className="text-gray-600 dark:text-gray-400">Loading visualization...</p>
                                                </div>
                                            </div>
                                        }>
                                        {selectedPlot === 'scores' && pcaResponse.result.scores.length > 0 && pcaResponse.result.scores[0].length >= 2 ? (
                                            <ScoresPlot
                                                key={`scores-${pcaResponse.result.scores.length}-${excludedRows.length}`}
                                                pcaResult={pcaResponse.result}
                                                rowNames={fileData?.rowNames || []}
                                                xComponent={selectedXComponent}
                                                yComponent={selectedYComponent}
                                                groupColumn={selectedGroupColumn}
                                                fontScale={plotFontScale}
                                                groupLabels={getColumnData(selectedGroupColumn).type === 'categorical' ? getColumnData(selectedGroupColumn).values as string[] : undefined}
                                                groupValues={getColumnData(selectedGroupColumn).type === 'continuous' ? getColumnData(selectedGroupColumn).values as number[] : undefined}
                                                groupType={getColumnData(selectedGroupColumn).type}
                                                groupEllipses={
                                                    confidenceLevel === 0.90 ? pcaResponse.groupEllipses90 :
                                                    confidenceLevel === 0.95 ? pcaResponse.groupEllipses95 :
                                                    pcaResponse.groupEllipses99
                                                }
                                                showEllipses={showEllipses && !!selectedGroupColumn && getColumnData(selectedGroupColumn).type === 'categorical'}
                                                confidenceLevel={confidenceLevel}
                                                showRowLabels={showRowLabels}
                                                maxLabelsToShow={maxLabelsToShow}
                                                onSelectionChange={handlePlotSelectionChange}
                                                excludedRows={pcaHasExclusions ? [] : excludedRows}
                                            />
                                        ) : selectedPlot === 'scores3d' && pcaResponse.result.scores.length > 0 && pcaResponse.result.scores[0].length >= 3 ? (
                                            <Scores3DPlot
                                                pcaResult={pcaResponse.result}
                                                rowNames={fileData?.rowNames || []}
                                                xComponent={selectedXComponent}
                                                yComponent={selectedYComponent}
                                                zComponent={selectedZComponent}
                                                fontScale={plotFontScale}
                                                groupColumn={selectedGroupColumn}
                                                groupLabels={getColumnData(selectedGroupColumn).type === 'categorical' ? getColumnData(selectedGroupColumn).values as string[] : undefined}
                                                groupValues={getColumnData(selectedGroupColumn).type === 'continuous' ? getColumnData(selectedGroupColumn).values as number[] : undefined}
                                                groupType={getColumnData(selectedGroupColumn).type}
                                                showRowLabels={showRowLabels}
                                                maxLabelsToShow={maxLabelsToShow}
                                            />
                                        ) : selectedPlot === 'scree' ? (
                                            <ScreePlot
                                                pcaResult={pcaResponse.result}
                                                showCumulative={true}
                                                elbowThreshold={80}
                                                fontScale={plotFontScale}
                                            />
                                        ) : selectedPlot === 'loadings' && pcaResponse.result.method !== 'kernel' ? (
                                            <LoadingsPlot
                                                pcaResult={pcaResponse.result}
                                                selectedComponent={selectedLoadingComponent}
                                                variableThreshold={guiConfig?.visualization?.loadings_variable_threshold || 100}
                                                plotType={loadingsPlotType || undefined}
                                                fontScale={plotFontScale}
                                            />
                                        ) : selectedPlot === 'biplot' ? (
                                            <Biplot
                                                pcaResult={pcaResponse.result}
                                                rowNames={fileData?.rowNames || []}
                                                xComponent={selectedXComponent}
                                                yComponent={selectedYComponent}
                                                showRowLabels={showRowLabels}
                                                fontScale={plotFontScale}
                                                maxLabelsToShow={maxLabelsToShow}
                                                groupColumn={selectedGroupColumn}
                                                groupLabels={getColumnData(selectedGroupColumn).type === 'categorical' ? getColumnData(selectedGroupColumn).values as string[] : undefined}
                                                groupValues={getColumnData(selectedGroupColumn).type === 'continuous' ? getColumnData(selectedGroupColumn).values as number[] : undefined}
                                                groupType={getColumnData(selectedGroupColumn).type}
                                                maxVariables={guiConfig?.visualization?.biplot_max_variables || 100}
                                                groupEllipses={
                                                    confidenceLevel === 0.90 ? pcaResponse.groupEllipses90 :
                                                    confidenceLevel === 0.95 ? pcaResponse.groupEllipses95 :
                                                    pcaResponse.groupEllipses99
                                                }
                                                showEllipses={showEllipses && !!selectedGroupColumn && getColumnData(selectedGroupColumn).type === 'categorical'}
                                                confidenceLevel={confidenceLevel}
                                                onSelectionChange={handlePlotSelectionChange}
                                                excludedRows={pcaHasExclusions ? [] : excludedRows}
                                            />
                                        ) : selectedPlot === 'biplot3d' ? (
                                            <Biplot3D
                                                pcaResult={pcaResponse.result}
                                                rowNames={fileData?.rowNames || []}
                                                xComponent={selectedXComponent}
                                                yComponent={selectedYComponent}
                                                zComponent={selectedZComponent}
                                                fontScale={plotFontScale}
                                                showRowLabels={showRowLabels}
                                                maxLabelsToShow={maxLabelsToShow}
                                                groupColumn={selectedGroupColumn}
                                                groupLabels={getColumnData(selectedGroupColumn).type === 'categorical' ? getColumnData(selectedGroupColumn).values as string[] : undefined}
                                                groupValues={getColumnData(selectedGroupColumn).type === 'continuous' ? getColumnData(selectedGroupColumn).values as number[] : undefined}
                                                groupType={getColumnData(selectedGroupColumn).type}
                                                maxVariables={guiConfig?.visualization?.biplot_max_variables || 50}
                                                groupEllipses={
                                                    confidenceLevel === 0.90 ? pcaResponse.groupEllipses90 :
                                                    confidenceLevel === 0.95 ? pcaResponse.groupEllipses95 :
                                                    pcaResponse.groupEllipses99
                                                }
                                                showEllipses={showEllipses && !!selectedGroupColumn && getColumnData(selectedGroupColumn).type === 'categorical'}
                                                confidenceLevel={confidenceLevel}
                                            />
                                        ) : selectedPlot === 'correlations' ? (
                                            <CircleOfCorrelations
                                                pcaResult={pcaResponse.result}
                                                xComponent={selectedXComponent}
                                                yComponent={selectedYComponent}
                                                fontScale={plotFontScale}
                                            />
                                        ) : selectedPlot === 'diagnostics' && pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? (
                                            <DiagnosticScatterPlot
                                                pcaResult={pcaResponse.result}
                                                rowNames={fileData?.rowNames || []}
                                                showRowLabels={showRowLabels}
                                                maxLabelsToShow={maxLabelsToShow}
                                                confidenceLevel={confidenceLevel === 0.90 ? 0.95 : confidenceLevel}
                                                fontScale={plotFontScale}
                                                groupColumn={selectedGroupColumn}
                                                groupLabels={getColumnData(selectedGroupColumn).type === 'categorical' ? getColumnData(selectedGroupColumn).values as string[] : undefined}
                                                groupValues={getColumnData(selectedGroupColumn).type === 'continuous' ? getColumnData(selectedGroupColumn).values as number[] : undefined}
                                                groupType={getColumnData(selectedGroupColumn).type as 'categorical' | 'continuous' | undefined}
                                                onSelectionChange={handlePlotSelectionChange}
                                                excludedRows={pcaHasExclusions ? [] : excludedRows}
                                            />
                                        ) : selectedPlot === 'eigencorrelation' ? (
                                            <EigencorrelationPlot
                                                pcaResult={pcaResponse.result}
                                                fontScale={plotFontScale}
                                            />
                                        ) : selectedPlot === 'temporal-loadings' ? (
                                            <TemporalLoadingsPlot
                                                pcaResult={pcaResponse.result}
                                                fontScale={plotFontScale}
                                            />
                                        ) : selectedPlot === 'temporal-variable-importance' ? (
                                            <TemporalVariableImportancePlot
                                                pcaResult={pcaResponse.result}
                                                fontScale={plotFontScale}
                                            />
                                        ) : selectedPlot === 'kernel-matrix' ? (
                                            <KernelMatrixHeatmap
                                                pcaResult={pcaResponse.result}
                                                rowNames={fileData?.rowNames || []}
                                                fontScale={plotFontScale}
                                            />
                                        ) : selectedPlot === 'sample-contributions' ? (
                                            <SampleContributionPlot
                                                pcaResult={pcaResponse.result}
                                                rowNames={fileData?.rowNames || []}
                                                selectedComponent={selectedLoadingComponent}
                                                fontScale={plotFontScale}
                                            />
                                        ) : (
                                            <div className="w-full h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
                                                <p>Not enough components for scores plot (minimum 2 required)</p>
                                            </div>
                                        )}
                                        </Suspense>
                                    </ErrorBoundary>
                                </div>

                                {/* Export Model button - centered below plot */}
                                <div className="mt-6 flex justify-center">
                                    <HelpWrapper helpKey="export-model">
                                        <button
                                            onClick={handleExportModel}
                                            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg font-medium text-white"
                                        >
                                            Export Model
                                        </button>
                                    </HelpWrapper>
                                </div>
                            </div>

                        </div>
                    )}
                </div>
            </main>

            {/* Documentation Viewer */}
            <DocumentationViewer
                isOpen={showDocumentation}
                onClose={() => setShowDocumentation(false)}
            />

            {/* About Dialog */}
            <AboutDialog
                isOpen={showAboutDialog}
                onClose={() => setShowAboutDialog(false)}
                version={version}
            />

            {/* GoCSV Download Confirmation Dialog */}
            <ConfirmDialog
                isOpen={showGoCSVDownloadDialog}
                onClose={() => setShowGoCSVDownloadDialog(false)}
                onConfirm={async () => {
                    setShowGoCSVDownloadDialog(false);
                    try {
                        await DownloadGoCSV();
                    } catch (error) {
                        console.error('Error downloading GoCSV:', error);
                        alert('Failed to open download page: ' + error);
                    }
                }}
                title="GoCSV Not Installed"
                message="GoCSV is not installed. Would you like to download it?"
                confirmText="Download"
                cancelText="Cancel"
            />
        </div>
    );
}

function App() {
    return (
        <ThemeProvider>
            <PaletteProvider>
                <HelpProvider>
                    <ErrorBoundary
                        onError={(error, errorInfo) => {
                            console.error('App Error Boundary:', error, errorInfo);
                        }}
                    >
                        <AppContent />
                    </ErrorBoundary>
                </HelpProvider>
            </PaletteProvider>
        </ThemeProvider>
    );
}

export default App;