// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState, useRef, useEffect, lazy, Suspense } from 'react';
import './App.css';
import { ParseCSV, RunPCA, LoadIrisDataset, LoadDatasetFile, GetVersion, CalculateEllipses, GetGUIConfig, LoadCSVFile, CheckGoCSVStatus, OpenInGoCSV, LaunchGoCSV, DownloadGoCSV, SaveFile } from '../wailsjs/go/main/App';
import { Copy, Check } from 'lucide-react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { DataTable, SelectionTable, MatrixIllustration, HelpWrapper, DocumentationViewer, ModelOverview } from './components';
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
import { FileData, PCARequest, PCAResponse } from './types';
import { ThemeProvider, ThemeToggle, ConfirmDialog, CustomSelect, SelectOption } from '@gopca/ui-components';
import { HelpProvider, useHelp } from './contexts/HelpContext';
import { PaletteProvider, usePalette } from './contexts/PaletteContext';
import { HelpDisplay } from './components/HelpDisplay';
import { PaletteSelector } from './components/PaletteSelector';
import { FontSizeControl } from './components/FontSizeControl';
import { config } from '../wailsjs/go/models';
import logo from './assets/images/GoPCA-logo-1024-transp.png';

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
    const [selectedPlot, setSelectedPlot] = useState<'scores' | 'scores3d' | 'scree' | 'loadings' | 'biplot' | 'biplot3d' | 'correlations' | 'diagnostics' | 'eigencorrelation'>('scores');
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
        kernelCoef0: 0.0
        // Confidence ellipse parameters
    });

    // GoCSV integration state
    const [goCSVStatus, setGoCSVStatus] = useState<{installed: boolean, path?: string, error?: string} | null>(null);
    const [isCheckingGoCSV, setIsCheckingGoCSV] = useState(false);
    const [showGoCSVDownloadDialog, setShowGoCSVDownloadDialog] = useState(false);

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
            console.info('SaveFile made available globally');
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

    // Helper function to get column data and type
    const getColumnData = (columnName: string | null): { values?: string[] | number[], type?: 'categorical' | 'continuous' } => {
        if (!columnName || !fileData) {
return {};
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

    const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (!file) {
return;
}

        setLoading(true);
        setFileError(null);
        setPcaError(null); // Clear any previous PCA errors
        setFileName(file.name); // Store the file name

        try {
            const content = await file.text();
            const result = await ParseCSV(content);
            setFileData(result);
            setPcaResponse(null);
            // Reset exclusions and selections when loading new data
            setExcludedRows([]);
            setExcludedColumns([]);
            setSelectedGroupColumn(null);
            setMode('none'); // Reset palette mode
            setDatasetId(prev => prev + 1); // Force DataTable re-render

            // Calculate and set default gamma for kernel PCA
            updateGammaForData(result);
        } catch (err) {
            setFileError(`Failed to parse CSV: ${err}`);
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

    const handleColumnSelectionChange = React.useCallback((selectedColumns: number[]) => {
        // Convert selected indices to excluded indices
        if (fileData) {
            const allIndices = Array.from({ length: fileData.headers.length }, (_, i) => i);
            const excluded = allIndices.filter(i => !selectedColumns.includes(i));
            setExcludedColumns(excluded);
        }
    }, [fileData]);

    const scrollToTop = () => {
        if (mainScrollRef.current) {
            mainScrollRef.current.scrollTo({
                top: 0,
                behavior: 'smooth'
            });
        }
    };

    const generateCLICommand = (): string => {
        let cmd = 'pca analyze';

        // Add file path (with quotes if it contains spaces)
        if (fileName) {
            if (fileName.includes(' ')) {
                cmd += ` "${fileName}"`;
            } else {
                cmd += ` ${fileName}`;
            }
        }

        // Add number of components
        cmd += ` --components ${config.components}`;

        // Add method
        cmd += ` --method ${config.method.toLowerCase()}`;

        // Add kernel parameters if using kernel PCA
        if (config.method === 'Kernel') {
            cmd += ` --kernel-type ${config.kernelType}`;
            if (config.kernelType === 'rbf' || config.kernelType === 'laplacian' || config.kernelType === 'sigmoid') {
                cmd += ` --kernel-gamma ${config.kernelGamma}`;
            }
            if (config.kernelType === 'polynomial' || config.kernelType === 'sigmoid') {
                cmd += ` --kernel-degree ${config.kernelDegree}`;
                cmd += ` --kernel-coef0 ${config.kernelCoef0}`;
            }
        }

        // Add row preprocessing (Step 1)
        if (config.snv) {
            cmd += ' --snv';
        } else if (config.vectorNorm) {
            cmd += ' --vector-norm';
        }

        // Add column preprocessing (Step 2)
        if (config.standardScale) {
            cmd += ' --scale standard';
        } else if (config.robustScale) {
            cmd += ' --scale robust';
        } else if (!config.meanCenter) {
            // Mean centering is on by default in CLI, so we need to explicitly disable it
            cmd += ' --no-mean-centering';
        }
        // Note: if only meanCenter is true, we don't need any flag (it's the default)

        // Add scale-only flag if needed
        if (config.scaleOnly) {
            cmd += ' --scale-only';
        }

        // Add missing data strategy
        if (config.missingStrategy && config.missingStrategy !== 'error') {
            cmd += ` --missing-strategy ${config.missingStrategy}`;
        }

        // Add excluded columns if any
        if (excludedColumns.length > 0) {
            // Convert 0-indexed to 1-indexed for CLI
            const columnIndices = excludedColumns.map(c => c + 1).join(',');
            cmd += ` --exclude-cols ${columnIndices}`;
        }

        // Add excluded rows if any
        if (excludedRows.length > 0) {
            // Convert 0-indexed to 1-indexed for CLI
            const rowIndices = excludedRows.map(r => r + 1).join(',');
            cmd += ` --exclude-rows ${rowIndices}`;
        }

        return cmd;
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
                        <img
                            src={logo}
                            alt="GoPCA - Principal Component Analysis Tool"
                            className="h-12 cursor-pointer hover:opacity-90 transition-opacity flex-shrink-0"
                            onClick={scrollToTop}
                        />
                        {version && (
                            <span className="text-xs text-gray-500 dark:text-gray-400">
                                {version}
                            </span>
                        )}
                    </div>
                    <div className="flex-1 mx-8 overflow-hidden">
                        <HelpDisplay
                            helpKey={currentHelpKey}
                            title={currentHelp?.title || ''}
                            text={currentHelp?.text || ''}
                        />
                    </div>
                    <div className="flex items-center gap-2">
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
                        <ThemeToggle />
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
                                    <input
                                        type="file"
                                        accept=".csv"
                                        onChange={handleFileUpload}
                                        className="block w-full text-sm text-gray-700 dark:text-gray-300
                                            file:mr-4 file:py-2 file:px-4
                                            file:rounded-full file:border-0
                                            file:text-sm file:font-semibold
                                            file:bg-blue-600 file:text-white
                                            hover:file:bg-blue-700
                                            file:transition-colors"
                                    />
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
                                    <HelpWrapper helpKey="sample-dataset-met">
                                        <button
                                            onClick={() => loadDataset('met.csv')}
                                            className="w-full px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                            disabled={loading}
                                        >
                                            MET
                                        </button>
                                    </HelpWrapper>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* File Error Display */}
                    {fileError && (
                        <div className="bg-red-100 dark:bg-red-800 border border-red-300 dark:border-red-600 rounded-lg p-4">
                            <p className="text-red-700 dark:text-red-200">{fileError}</p>
                        </div>
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
                                />
                            ) : (
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
                                />
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
                                                }

                                                setConfig(newConfig);
                                            }}
                                            options={[
                                                { value: 'SVD', label: 'SVD' },
                                                { value: 'NIPALS', label: 'NIPALS' },
                                                { value: 'kernel', label: 'Kernel PCA' }
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
                        <div ref={pcaErrorRef} className="bg-red-100 dark:bg-red-800 border border-red-300 dark:border-red-600 rounded-lg p-4">
                            <p className="text-red-700 dark:text-red-200">{pcaError}</p>
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
                            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                                {/* Explained Variance */}
                                <HelpWrapper helpKey="explained-variance">
                                    <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4">
                                        <div className="mb-2">
                                            <h3 className="text-lg font-semibold">Explained Variance</h3>
                                        </div>
                                        <div className="space-y-2">
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
                                                onChange={(value) => setSelectedPlot(value as 'scores' | 'scores3d' | 'scree' | 'loadings' | 'biplot' | 'biplot3d' | 'correlations' | 'diagnostics' | 'eigencorrelation')}
                                                options={[
                                                    { value: 'scores', label: 'Scores Plot' },
                                                    { value: 'scores3d', label: '3D Scores Plot' },
                                                    { value: 'scree', label: 'Scree Plot' },
                                                    ...(pcaResponse.result.method !== 'kernel' ? [{ value: 'loadings', label: 'Loadings Plot' }] : []),
                                                    ...(pcaResponse.result.preprocessing_applied ? [{ value: 'biplot', label: 'Biplot' }] : []),
                                                    ...(pcaResponse.result.preprocessing_applied ? [{ value: 'biplot3d', label: '3D Biplot' }] : []),
                                                    ...(pcaResponse.result.preprocessing_applied ? [{ value: 'correlations', label: 'Circle of Correlations' }] : []),
                                                    ...(pcaResponse.result.method !== 'kernel' ? [{ value: 'diagnostics', label: 'Diagnostic Plot' }] : []),
                                                    ...(pcaResponse.result.eigencorrelations && pcaResponse.result.method !== 'kernel' ? [{ value: 'eigencorrelation', label: 'Eigencorrelation Plot' }] : [])
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
                                        {/* Data Display Group */}
                                        {(selectedPlot === 'scores' || selectedPlot === 'scores3d' || selectedPlot === 'biplot' || selectedPlot === 'biplot3d') && fileData &&
                                         ((fileData.categoricalColumns && Object.keys(fileData.categoricalColumns).length > 0) ||
                                          (fileData.numericTargetColumns && Object.keys(fileData.numericTargetColumns).length > 0)) && (
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
                                                {selectedPlot === 'scores3d' && pcaResponse.result.scores[0]?.length > 2 && (
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
                                        ) : selectedPlot === 'diagnostics' && pcaResponse.result.method !== 'kernel' ? (
                                            <DiagnosticScatterPlot
                                                pcaResult={pcaResponse.result}
                                                rowNames={fileData?.rowNames || []}
                                                showRowLabels={showRowLabels}
                                                maxLabelsToShow={maxLabelsToShow}
                                                confidenceLevel={confidenceLevel === 0.90 ? 0.95 : confidenceLevel}
                                                fontScale={plotFontScale}
                                            />
                                        ) : selectedPlot === 'eigencorrelation' ? (
                                            <EigencorrelationPlot
                                                pcaResult={pcaResponse.result}
                                                fontScale={plotFontScale}
                                            />
                                        ) : (
                                            <div className="w-full h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
                                                <p>Not enough components for scores plot (minimum 2 required)</p>
                                            </div>
                                        )}
                                    </Suspense>
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
                    <AppContent />
                </HelpProvider>
            </PaletteProvider>
        </ThemeProvider>
    );
}

export default App;