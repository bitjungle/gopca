// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React, { useState, useCallback, useEffect, lazy, Suspense } from 'react';
import './App.css';
import { ExportPCAModel } from '../wailsjs/go/main/App';
import { Copy, Check } from 'lucide-react';
import { DataTable, SelectionTable, MatrixIllustration, HelpWrapper, DocumentationViewer, ModelOverview, AboutDialog } from './components';
import { ThemeProvider, ThemeToggle, ConfirmDialog, CustomSelect, ErrorBoundary, ErrorAlert } from '@gopca/ui-components';
import { HelpProvider, useHelp } from './contexts/HelpContext';
import { PaletteProvider, usePalette } from './contexts/PaletteContext';
import { HelpDisplay } from './components/HelpDisplay';
import { PaletteSelector } from './components/PaletteSelector';
import { FontSizeControl } from './components/FontSizeControl';
import logo from './assets/images/GoPCA-logo-1024-transp.png';
import { generateCLICommand as generateCLICommandUtil } from './utils/cliCommandGenerator';
import { logger } from './utils/logger';

// Hooks
import { useAppInit } from './hooks/useAppInit';
import { useFileData } from './hooks/useFileData';
import { useGoCSVIntegration } from './hooks/useGoCSVIntegration';
import { usePCAConfig } from './hooks/usePCAConfig';
import { usePCARunner } from './hooks/usePCARunner';
import { useVisualization, PlotType } from './hooks/useVisualization';
import { useUIState } from './hooks/useUIState';

// Lazy-loaded visualization components
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

function AppContent() {
    const { currentHelp, currentHelpKey } = useHelp();
    const { setMode } = usePalette();

    // ── File data ────────────────────────────────────────────────────────────
    const {
        fileData, fileName, filePath, fileError, datasetId, loading: fileLoading,
        loadDataset, handleNativeFileSelect, setFileDataDirect, clearFileError,
    } = useFileData();

    // ── PCA configuration & exclusions ───────────────────────────────────────
    const {
        config, setConfig,
        excludedRows, excludedColumns,
        setExcludedRows, setExcludedColumns,
        updateGammaForData, resetExclusions,
    } = usePCAConfig();

    // ── selectedGroupColumn lifted here to break the circular dep between
    //    usePCARunner (needs it to build requests) and useVisualization
    //    (needs it to drive palette mode). Both hooks receive it as a param.
    const [selectedGroupColumn, setSelectedGroupColumn] = useState<string | null>(null);

    // ── PCA runner ───────────────────────────────────────────────────────────
    const {
        pcaResponse, pcaError, loading: pcaLoading,
        pcaHasExclusions, pcaResultsRef, pcaErrorRef,
        runPCA, clearPcaError, clearPcaResponse,
    } = usePCARunner(fileData, config, excludedRows, excludedColumns, selectedGroupColumn);

    // ── Visualization state ───────────────────────────────────────────────────
    const {
        selectedPlot, setSelectedPlot,
        selectedXComponent, setSelectedXComponent,
        selectedYComponent, setSelectedYComponent,
        selectedZComponent, setSelectedZComponent,
        selectedLoadingComponent, setSelectedLoadingComponent,
        showEllipses, setShowEllipses,
        confidenceLevel, setConfidenceLevel,
        showRowLabels, setShowRowLabels,
        maxLabelsToShow, setMaxLabelsToShow,
        loadingsPlotType, setLoadingsPlotType,
        plotFontScale, setPlotFontScale,
        getColumnData, handlePlotSelectionChange,
    } = useVisualization(pcaResponse, fileData, fileLoading || pcaLoading, selectedGroupColumn, setExcludedRows);

    // Combined loading state
    const loading = fileLoading || pcaLoading;

    // Alert the user when a new Kernel PCA result arrives and the current plot
    // is incompatible. useVisualization already switches the plot via its own
    // effect; this effect adds the user-visible alert.
    useEffect(() => {
        if (pcaResponse?.result?.method === 'kernel') {
            if (selectedPlot === 'correlations' || selectedPlot === 'biplot' || selectedPlot === 'biplot3d') {
                alert('The selected visualization is not supported for Kernel PCA. Switching to Scores Plot.');
            }
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pcaResponse]);

    // ── Row / column selection handlers ───────────────────────────────────────
    const handleRowSelectionChange = useCallback((selectedRows: number[]) => {
        if (!fileData) return;
        const allIndices = Array.from({ length: fileData.data.length }, (_, i) => i);
        setExcludedRows(allIndices.filter(i => !selectedRows.includes(i)));
    }, [fileData, setExcludedRows]);

    const handleColumnSelectionChange = useCallback((selectedColumns: number[]) => {
        if (!fileData) return;
        const allIndices = Array.from({ length: fileData.headers.length }, (_, i) => i);
        setExcludedColumns(allIndices.filter(i => !selectedColumns.includes(i)));
    }, [fileData, setExcludedColumns]);

    // ── Dataset loading helpers ───────────────────────────────────────────────
    const handleLoadDataset = useCallback(async (filename: string, defaultGroupColumn?: string) => {
        const result = await loadDataset(filename, defaultGroupColumn);
        if (!result) return;
        const { data, defaultGroupColumn: groupCol } = result;
        resetExclusions();
        clearPcaResponse();

        if (groupCol && data) {
            const isCategorical = data.categoricalColumns && groupCol in data.categoricalColumns;
            const isContinuous = data.numericTargetColumns && groupCol in data.numericTargetColumns;
            if (isCategorical || isContinuous) {
                setSelectedGroupColumn(groupCol);
                setMode(isCategorical ? 'categorical' : 'continuous');
            } else {
                logger.warn(`Column "${groupCol}" not found in ${filename}, setting group column to null`);
                setSelectedGroupColumn(null);
                setMode('none');
            }
        } else {
            setSelectedGroupColumn(null);
            setMode('none');
        }

        updateGammaForData(data);
    }, [loadDataset, resetExclusions, clearPcaResponse, setSelectedGroupColumn, setMode, updateGammaForData]);

    const handleNativeFileSelectWithReset = useCallback(async () => {
        const data = await handleNativeFileSelect();
        if (!data) return;
        resetExclusions();
        clearPcaResponse();
        setSelectedGroupColumn(null);
        setMode('none');
        updateGammaForData(data);
    }, [handleNativeFileSelect, resetExclusions, clearPcaResponse, setSelectedGroupColumn, setMode, updateGammaForData]);

    // ── Run PCA — also resets PC selectors ────────────────────────────────────
    const handleRunPCA = useCallback(async () => {
        await runPCA();
        setSelectedXComponent(0);
        setSelectedYComponent(1);
    }, [runPCA, setSelectedXComponent, setSelectedYComponent]);

    // ── Export model ──────────────────────────────────────────────────────────
    const handleExportModel = useCallback(async () => {
        if (!pcaResponse?.success || !pcaResponse.result || !fileData) return;

        try {
            const { ExportPCAModelRequest } = await import('../wailsjs/go/models').then(m => m.main);
            const request = new ExportPCAModelRequest({
                data: fileData.data,
                headers: fileData.headers,
                rowNames: fileData.rowNames,
                pcaResult: pcaResponse.result,
                config,
                excludedRows,
                excludedColumns,
                filename: fileName,
            });
            await ExportPCAModel(request);
        } catch (err) {
            logger.error('Failed to export model:', err);
            alert(`Failed to export model: ${err}`);
        }
    }, [pcaResponse, fileData, config, excludedRows, excludedColumns, fileName]);

    // ── CLI command ───────────────────────────────────────────────────────────
    const generateCLICommand = useCallback((): string => {
        return generateCLICommandUtil({
            fileName, filePath,
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
            excludedRows,
        });
    }, [fileName, filePath, config, excludedColumns, excludedRows]);

    // ── App init (version, guiConfig, Plotly–Wails, startup file) ────────────
    const onStartupFile = useCallback((data: import('./types').FileData) => {
        setFileDataDirect(data, 'Startup File', '');
        resetExclusions();
        clearPcaResponse();
        setSelectedGroupColumn(null);
        setMode('none');
        updateGammaForData(data);
    }, [setFileDataDirect, resetExclusions, clearPcaResponse, setSelectedGroupColumn, setMode, updateGammaForData]);

    const { version, guiConfig } = useAppInit(onStartupFile);

    // ── GoCSV integration ─────────────────────────────────────────────────────
    const {
        goCSVStatus, isCheckingGoCSV,
        showGoCSVDownloadDialog, setShowGoCSVDownloadDialog,
        handleGoCSVAction, handleGoCSVDownload,
    } = useGoCSVIntegration();

    // ── UI state (modals, clipboard, scroll ref) ──────────────────────────────
    const {
        showDocumentation, setShowDocumentation,
        showAboutDialog, setShowAboutDialog,
        showCopied, mainScrollRef,
        handleLogoClick, copyToClipboard,
    } = useUIState();

    // ── JSX ───────────────────────────────────────────────────────────────────
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
                    {/* Step 1: Load Data */}
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
                                        onClick={handleNativeFileSelectWithReset}
                                        disabled={loading}
                                        className="w-full px-4 py-2 text-sm font-semibold text-white bg-blue-600 rounded-full hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                    >
                                        {loading ? 'Loading...' : 'Choose File'}
                                    </button>
                                </HelpWrapper>

                                <div className="mt-4">
                                    <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                                        Or Use the Data Editor
                                    </p>
                                    <HelpWrapper helpKey="gocsv-integration">
                                        <button
                                            onClick={() => handleGoCSVAction(fileData)}
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
                                <HelpWrapper helpKey="data-table-format">
                                    <MatrixIllustration />
                                </HelpWrapper>
                            </div>

                            {/* Column 3: Sample Datasets */}
                            <div className="flex flex-col justify-center md:col-span-2 lg:col-span-1">
                                <label className="block text-sm font-medium mb-3">
                                    Or Try Sample Datasets
                                </label>
                                <div className="space-y-2">
                                    <HelpWrapper helpKey="sample-dataset-corn">
                                        <button
                                            onClick={() => handleLoadDataset('corn.csv')}
                                            className="w-full px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                            disabled={loading}
                                        >
                                            Corn (NIR)
                                        </button>
                                    </HelpWrapper>
                                    <HelpWrapper helpKey="sample-dataset-iris">
                                        <button
                                            onClick={() => handleLoadDataset('iris.csv', 'species')}
                                            className="w-full px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                            disabled={loading}
                                        >
                                            Iris
                                        </button>
                                    </HelpWrapper>
                                    <HelpWrapper helpKey="sample-dataset-wine">
                                        <button
                                            onClick={() => handleLoadDataset('wine.csv', 'target')}
                                            className="w-full px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                            disabled={loading}
                                        >
                                            Wine
                                        </button>
                                    </HelpWrapper>
                                    <HelpWrapper helpKey="sample-dataset-swiss-roll">
                                        <button
                                            onClick={() => handleLoadDataset('swiss_roll.csv', 'color #target')}
                                            className="w-full px-4 py-2 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                            disabled={loading}
                                        >
                                            Swiss Roll
                                        </button>
                                    </HelpWrapper>
                                    <HelpWrapper helpKey="sample-dataset-stocks">
                                        <button
                                            onClick={() => handleLoadDataset('stocks.csv')}
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

                    {/* File Error */}
                    {fileError && (
                        <ErrorAlert
                            type="error"
                            title="File Error"
                            message={fileError}
                            onDismiss={clearFileError}
                        />
                    )}

                    {/* Loaded Data Table */}
                    {fileData && (
                        <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-200 dark:border-gray-700">
                            <h2 className="text-xl font-semibold mb-4">Loaded Data</h2>
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
                                        logger.error('DataTable Error:', error, errorInfo);
                                    }}
                                >
                                    <HelpWrapper helpKey="data-table-format">
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
                                    </HelpWrapper>
                                </ErrorBoundary>
                            )}
                        </div>
                    )}

                    {/* Step 2: Configure PCA */}
                    {fileData && (
                        <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-200 dark:border-gray-700">
                            <h2 className="text-xl font-semibold mb-6">Step 2: Configure PCA</h2>

                            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                                {/* Left Column - Core PCA Configuration */}
                                <div className="space-y-4">
                                    <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">PCA Options</h3>

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

                                                    if (newMethod === 'kernel') {
                                                        if (newConfig.meanCenter || newConfig.standardScale || newConfig.robustScale) {
                                                            newConfig.meanCenter = false;
                                                            newConfig.standardScale = false;
                                                            newConfig.robustScale = false;
                                                            newConfig.scaleOnly = false;
                                                        }
                                                    } else if (oldMethod === 'kernel' && newMethod !== 'kernel') {
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

                                    {config.method === 'SVD' && (
                                        <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg space-y-3">
                                            <h4 className="font-medium text-sm text-blue-900 dark:text-blue-100">SVD Method</h4>
                                            <div className="space-y-2 text-sm text-blue-800 dark:text-blue-200">
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Gold standard for PCA using Singular Value Decomposition</span></p>
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Fast and numerically stable for complete datasets</span></p>
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Computes all components simultaneously</span></p>
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Best choice for most applications</span></p>
                                            </div>
                                        </div>
                                    )}

                                    {config.method === 'NIPALS' && (
                                        <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg space-y-3">
                                            <h4 className="font-medium text-sm text-green-900 dark:text-green-100">NIPALS Method</h4>
                                            <div className="space-y-2 text-sm text-green-800 dark:text-green-200">
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Nonlinear Iterative Partial Least Squares algorithm</span></p>
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Handles missing data gracefully</span></p>
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Computes components sequentially</span></p>
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Ideal for large datasets when only few components needed</span></p>
                                            </div>
                                        </div>
                                    )}

                                    {config.method === 'kernel' && (
                                        <div className="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg space-y-4">
                                            <h4 className="font-medium text-sm">Kernel PCA Options</h4>

                                            {fileData.data.length > 5000 && (
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
                                                    <label className="block text-sm font-medium mb-1">Kernel Type</label>
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
                                                    <label className="block text-sm font-medium mb-1">Gamma</label>
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
                                                            <label className="block text-sm font-medium mb-1">Degree</label>
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
                                                            <label className="block text-sm font-medium mb-1">Coef0</label>
                                                            <input
                                                                type="number"
                                                                value={config.kernelCoef0}
                                                                step="0.1"
                                                                onChange={(e) => {
                                                                    const value = parseFloat(e.target.value);
                                                                    setConfig({ ...config, kernelCoef0: isNaN(value) ? 1.0 : value });
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

                                    {config.method === 'temporal' && (
                                        <div className="p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg space-y-4">
                                            <h4 className="font-medium text-sm text-purple-900 dark:text-purple-100">Temporal PCA Options</h4>
                                            <div className="space-y-2 text-sm text-purple-800 dark:text-purple-200">
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Time-Delay PCA for time-series analysis</span></p>
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Captures temporal dynamics and dependencies</span></p>
                                                <p className="flex items-start"><span className="mr-2">•</span><span>Based on SSA (Singular Spectrum Analysis) methodology</span></p>
                                            </div>
                                            <div className="space-y-4">
                                                <HelpWrapper helpKey="temporal-lags">
                                                    <label className="block text-sm font-medium mb-1">Number of Time Lags</label>
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
                                                {config.temporalLags >= fileData.data.length && (
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

                                    <HelpWrapper helpKey="row-preprocessing" className="p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                                        <label className="block text-sm font-medium mb-2">
                                            Step 1: Row-wise Preprocessing (optional)
                                        </label>
                                        <CustomSelect
                                            value={config.snv ? 'snv' : config.vectorNorm ? 'vector-norm' : 'none'}
                                            onChange={(value) => {
                                                setConfig({ ...config, snv: value === 'snv', vectorNorm: value === 'vector-norm' });
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
                                                if (config.method === 'kernel' && !['none', 'scale-only'].includes(value)) return;
                                                setConfig({
                                                    ...config,
                                                    meanCenter: value === 'center' || value === 'standard',
                                                    standardScale: value === 'standard',
                                                    robustScale: value === 'robust',
                                                    scaleOnly: value === 'scale-only',
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
                                                { value: 'zero', label: 'Impute with Zero' },
                                                { value: 'native', label: 'Native NIPALS Handling (NIPALS only)' }
                                            ]}
                                            className="w-full"
                                        />
                                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                            Choose how to handle missing values (NaN) in your data
                                        </p>
                                    </HelpWrapper>
                                </div>
                            </div>

                            {/* Go PCA! button */}
                            <div className="mt-6 flex justify-center">
                                <HelpWrapper helpKey="go-pca-button">
                                    <button
                                        onClick={handleRunPCA}
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

                    {/* PCA Error */}
                    {pcaError && fileData && (
                        <div ref={pcaErrorRef}>
                            <ErrorAlert
                                type="error"
                                title="Analysis Error"
                                message={pcaError}
                                suggestion="Please check your data and parameters, then try again"
                                onDismiss={clearPcaError}
                            />
                        </div>
                    )}

                    {/* Step 3: Results */}
                    {pcaResponse?.success && pcaResponse.result && (
                        <div ref={pcaResultsRef} className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-200 dark:border-gray-700">
                            <h2 className="text-xl font-semibold mb-4">Step 3: Interpret PCA Model</h2>

                            {pcaResponse.info && (
                                <div className="mb-4 p-3 bg-blue-100 dark:bg-blue-800 border border-blue-300 dark:border-blue-600 rounded-lg">
                                    <p className="text-blue-700 dark:text-blue-200 text-sm">
                                        <span className="font-semibold">Note:</span> {pcaResponse.info}
                                    </p>
                                </div>
                            )}

                            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch">
                                <HelpWrapper helpKey="explained-variance" className="h-full">
                                    <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4 h-full flex flex-col">
                                        <div className="mb-2">
                                            <h3 className="text-lg font-semibold">Explained Variance</h3>
                                        </div>
                                        <div className="space-y-2 flex-grow">
                                            {pcaResponse.result.explained_variance_ratio.map((percentage, i) => (
                                                <div key={i} className="flex justify-between">
                                                    <span>{pcaResponse.result?.component_labels?.[i] || `PC${i+1}`}:</span>
                                                    <span>{percentage.toFixed(2)}%</span>
                                                </div>
                                            ))}
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

                                <ModelOverview
                                    pcaResult={pcaResponse.result}
                                    selectedPC={selectedXComponent}
                                    standardScale={config.standardScale}
                                    originalData={fileData?.data}
                                />
                            </div>

                            {/* Visualizations */}
                            <div className="mt-6">
                                {/* Tier 1: Primary Controls */}
                                <div className="flex items-center justify-between mb-3 pb-3 border-b border-gray-200 dark:border-gray-600">
                                    <div className="flex items-center gap-4">
                                        <h3 className="text-lg font-semibold">Visualizations</h3>
                                        <HelpWrapper helpKey={`${selectedPlot}-plot`}>
                                            <CustomSelect
                                                value={selectedPlot}
                                                onChange={(value) => setSelectedPlot(value as PlotType)}
                                                options={[
                                                    { value: 'scores', label: 'Scores Plot' },
                                                    { value: 'scores3d', label: '3D Scores Plot' },
                                                    { value: 'scree', label: 'Scree Plot' },
                                                    ...(pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? [{ value: 'loadings', label: 'Loadings Plot' }] : []),
                                                    ...(pcaResponse.result.method === 'temporal' ? [{ value: 'temporal-loadings', label: 'Temporal Loadings' }] : []),
                                                    ...(pcaResponse.result.method === 'temporal' ? [{ value: 'temporal-variable-importance', label: 'Variable Importance' }] : []),
                                                    ...(pcaResponse.result.preprocessing_applied && pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? [{ value: 'biplot', label: 'Biplot' }] : []),
                                                    ...(pcaResponse.result.preprocessing_applied && pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? [{ value: 'biplot3d', label: '3D Biplot' }] : []),
                                                    ...(pcaResponse.result.preprocessing_applied && pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? [{ value: 'correlations', label: 'Circle of Correlations' }] : []),
                                                    ...(pcaResponse.result.method !== 'kernel' && pcaResponse.result.method !== 'temporal' ? [{ value: 'diagnostics', label: 'Diagnostic Plot' }] : []),
                                                    ...(pcaResponse.result.eigencorrelations && pcaResponse.result.method !== 'kernel' ? [{ value: 'eigencorrelation', label: 'Eigencorrelation Plot' }] : []),
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
                                        {(selectedPlot === 'scores' || selectedPlot === 'scores3d' || selectedPlot === 'biplot' || selectedPlot === 'biplot3d' || selectedPlot === 'diagnostics') && fileData && (
                                            <div className="flex items-center gap-3 px-3 py-2 bg-gray-50 dark:bg-gray-800 rounded-lg">
                                                <HelpWrapper helpKey="group-coloring" className="flex items-center gap-2">
                                                    <label className="text-sm text-gray-600 dark:text-gray-400">Color by:</label>
                                                    <CustomSelect
                                                        value={selectedGroupColumn || ''}
                                                        onChange={(value) => {
                                                            const selectedValue = value || null;
                                                            setSelectedGroupColumn(selectedValue);
                                                            if (!selectedValue) {
                                                                setMode('none');
                                                                setShowEllipses(false);
                                                            } else if (selectedValue === 'Row Index') {
                                                                setMode('continuous');
                                                                setShowEllipses(false);
                                                            } else if (fileData.numericTargetColumns && selectedValue in fileData.numericTargetColumns) {
                                                                setMode('continuous');
                                                                setShowEllipses(false);
                                                            } else if (fileData.categoricalColumns && selectedValue in fileData.categoricalColumns) {
                                                                setMode('categorical');
                                                            }
                                                        }}
                                                        options={[
                                                            { value: '', label: 'None' },
                                                            { value: 'Row Index', label: '📊 Row Index', group: 'Continuous' },
                                                            ...(fileData.categoricalColumns && Object.keys(fileData.categoricalColumns).length > 0
                                                                ? Object.keys(fileData.categoricalColumns).map((colName) => ({
                                                                    value: colName,
                                                                    label: `🏷️ ${colName}`,
                                                                    group: 'Categorical'
                                                                }))
                                                                : []),
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
                                                {selectedPlot === 'diagnostics' && (
                                                    <div className="flex items-center gap-2 ml-3">
                                                        <HelpWrapper helpKey="diagnostic-threshold">
                                                            <label className="text-sm text-gray-600 dark:text-gray-400">Threshold:</label>
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
                                                    onChange={(value) => setLoadingsPlotType(value as 'bar' | 'line')}
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

                                {/* Plot area */}
                                <div className="bg-gray-50 dark:bg-gray-700 rounded-lg" style={{ height: '500px' }}>
                                    <ErrorBoundary
                                        onError={(error, errorInfo) => {
                                            logger.error('Visualization Error:', error, errorInfo);
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

                                {/* Export Model */}
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

            <DocumentationViewer
                isOpen={showDocumentation}
                onClose={() => setShowDocumentation(false)}
            />

            <AboutDialog
                isOpen={showAboutDialog}
                onClose={() => setShowAboutDialog(false)}
                version={version}
            />

            <ConfirmDialog
                isOpen={showGoCSVDownloadDialog}
                onClose={() => setShowGoCSVDownloadDialog(false)}
                onConfirm={handleGoCSVDownload}
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
                            logger.error('App Error Boundary:', error, errorInfo);
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
