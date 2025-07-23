import React, { useState, useRef } from 'react';
import './App.css';
import { ParseCSV, RunPCA, LoadIrisDataset } from "../wailsjs/go/main/App";
import { DataTable, SelectionTable, ThemeToggle } from './components';
import { ScoresPlot, ScreePlot, LoadingsPlot, Biplot } from './components/visualizations';
import { FileData, PCARequest, PCAResponse } from './types';
import { ThemeProvider } from './contexts/ThemeContext';
import logo from './assets/images/GoPCA-logo-1024-transp.png';

function App() {
    const [fileData, setFileData] = useState<FileData | null>(null);
    const [pcaResponse, setPcaResponse] = useState<PCAResponse | null>(null);
    const [loading, setLoading] = useState(false);
    const [fileError, setFileError] = useState<string | null>(null);
    const [pcaError, setPcaError] = useState<string | null>(null);
    
    // Selection state
    const [excludedRows, setExcludedRows] = useState<number[]>([]);
    const [excludedColumns, setExcludedColumns] = useState<number[]>([]);
    const [selectedPlot, setSelectedPlot] = useState<'scores' | 'scree' | 'loadings' | 'biplot'>('scores');
    const [selectedXComponent, setSelectedXComponent] = useState(0);
    const [selectedYComponent, setSelectedYComponent] = useState(1);
    const [selectedLoadingComponent, setSelectedLoadingComponent] = useState(0);
    const [selectedGroupColumn, setSelectedGroupColumn] = useState<string | null>(null);
    
    // Refs for smooth scrolling
    const pcaErrorRef = useRef<HTMLDivElement>(null);
    const pcaResultsRef = useRef<HTMLDivElement>(null);
    
    // PCA configuration
    const [config, setConfig] = useState({
        components: 2,
        meanCenter: true,
        standardScale: true,
        robustScale: false,
        snv: false,
        method: 'SVD',
        missingStrategy: 'error',
        // Kernel PCA parameters
        kernelType: 'rbf',
        kernelGamma: 1.0,
        kernelDegree: 3,
        kernelCoef0: 0.0
    });
    
    const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (!file) return;
        
        setLoading(true);
        setFileError(null);
        setPcaError(null); // Clear any previous PCA errors
        
        try {
            const content = await file.text();
            const result = await ParseCSV(content);
            setFileData(result);
            setPcaResponse(null);
            // Reset exclusions and selections when loading new data
            setExcludedRows([]);
            setExcludedColumns([]);
            setSelectedGroupColumn(null);
        } catch (err) {
            setFileError(`Failed to parse CSV: ${err}`);
        } finally {
            setLoading(false);
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
    
    const runPCA = async () => {
        if (!fileData) return;
        
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
                excludedColumns
            };
            const result = await RunPCA(request);
            if (result.success) {
                setPcaResponse(result);
                // Reset PC selections to default
                setSelectedXComponent(0);
                setSelectedYComponent(1);
                // Clear any previous errors
                setPcaError(null);
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
    
    return (
        <ThemeProvider>
            <div className="flex flex-col h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white transition-colors duration-200">
                <header className="sticky top-0 z-50 bg-white dark:bg-gray-800 p-4 shadow-lg backdrop-blur-sm bg-opacity-95 dark:bg-opacity-95">
                    <div className="flex items-center justify-between max-w-7xl mx-auto">
                        <img src={logo} alt="GoPCA - Principal Component Analysis Tool" className="h-12" />
                        <ThemeToggle />
                    </div>
                </header>
            
            <main className="flex-1 overflow-auto p-6">
                <div className="max-w-7xl mx-auto space-y-6">
                    {/* File Upload Section */}
                    <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-200 dark:border-gray-700">
                        <h2 className="text-xl font-semibold mb-4">Step 1: Load Data</h2>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium mb-2">
                                    Upload CSV File
                                </label>
                                <input
                                    type="file"
                                    accept=".csv"
                                    onChange={handleFileUpload}
                                    className="block w-full text-sm text-gray-700 dark:text-gray-300
                                        file:mr-4 file:py-2 file:px-4
                                        file:rounded-full file:border-0
                                        file:text-sm file:font-semibold
                                        file:bg-blue-600 file:text-white
                                        hover:file:bg-blue-700"
                                />
                            </div>
                            <div className="text-sm text-gray-600 dark:text-gray-400">
                                Or try the example dataset with categorical groups
                            </div>
                            <button
                                onClick={async () => {
                                    setLoading(true);
                                    setFileError(null);
                                    setPcaError(null);
                                    try {
                                        const result = await LoadIrisDataset();
                                        setFileData(result);
                                        setPcaResponse(null);
                                        setExcludedRows([]);
                                        setExcludedColumns([]);
                                        setSelectedGroupColumn('species');
                                    } catch (err) {
                                        setFileError(`Failed to load iris dataset: ${err}`);
                                    } finally {
                                        setLoading(false);
                                    }
                                }}
                                className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
                                disabled={loading}
                            >
                                Load Iris Dataset (with species)
                            </button>
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
                                    headers={fileData.headers}
                                    rowNames={fileData.rowNames}
                                    data={fileData.data}
                                    title="Input Data"
                                    onRowSelectionChange={handleRowSelectionChange}
                                    onColumnSelectionChange={handleColumnSelectionChange}
                                />
                            ) : (
                                <DataTable
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
                            <h2 className="text-xl font-semibold mb-4">Step 2: Configure PCA</h2>
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-sm font-medium mb-2">
                                        Number of Components
                                    </label>
                                    <input
                                        type="number"
                                        min="1"
                                        max={Math.min(fileData.headers.length, fileData.data.length)}
                                        value={config.components}
                                        onChange={(e) => setConfig({...config, components: parseInt(e.target.value) || 2})}
                                        className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium mb-2">
                                        Method
                                    </label>
                                    <select
                                        value={config.method}
                                        onChange={(e) => setConfig({...config, method: e.target.value})}
                                        className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                    >
                                        <option value="SVD">SVD</option>
                                        <option value="NIPALS">NIPALS</option>
                                        <option value="kernel">Kernel PCA</option>
                                    </select>
                                </div>
                            </div>
                            <div className="mt-4">
                                <label className="block text-sm font-medium mb-2">
                                    Preprocessing Method
                                </label>
                                <select
                                    value={
                                        config.method === 'kernel' ? 'none' :
                                        config.robustScale ? 'robust' :
                                        config.standardScale ? 'standard' :
                                        config.meanCenter ? 'center' : 'none'
                                    }
                                    onChange={(e) => {
                                        const value = e.target.value;
                                        setConfig({
                                            ...config,
                                            meanCenter: value === 'center' || value === 'standard',
                                            standardScale: value === 'standard',
                                            robustScale: value === 'robust'
                                        });
                                    }}
                                    disabled={config.method === 'kernel'}
                                    className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    <option value="none">None (Raw Data)</option>
                                    <option value="center">Mean Center Only</option>
                                    <option value="standard">Standard Scale (Mean + Std Dev)</option>
                                    <option value="robust">Robust Scale (Median + MAD)</option>
                                </select>
                                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                    {config.method === 'kernel' 
                                        ? 'Kernel PCA uses its own centering in kernel space'
                                        : 'Choose how to preprocess your data before PCA'}
                                </p>
                            </div>
                            
                            <div className="mt-4">
                                <label className="flex items-center">
                                    <input
                                        type="checkbox"
                                        checked={config.snv}
                                        onChange={(e) => setConfig({...config, snv: e.target.checked})}
                                        disabled={config.method === 'kernel'}
                                        className="mr-2 rounded border-gray-300 dark:border-gray-600 text-indigo-600 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
                                    />
                                    <span className="text-sm font-medium">Apply SNV (Standard Normal Variate)</span>
                                </label>
                                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1 ml-6">
                                    Row-wise normalization, useful for spectral data and other profile-based measurements
                                </p>
                            </div>
                            
                            <div className="mt-4">
                                <label className="block text-sm font-medium mb-2">
                                    Missing Data Strategy
                                </label>
                                <select
                                    value={config.missingStrategy}
                                    onChange={(e) => setConfig({...config, missingStrategy: e.target.value})}
                                    className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                >
                                    <option value="error">Show Error (default)</option>
                                    <option value="drop">Drop Rows with Missing Values</option>
                                    <option value="mean">Impute with Column Mean</option>
                                    <option value="median">Impute with Column Median</option>
                                </select>
                                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                    Choose how to handle missing values (NaN) in your data
                                </p>
                            </div>
                            
                            {/* Kernel PCA Options */}
                            {config.method === 'kernel' && (
                                <div className="mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg space-y-4">
                                    <h4 className="font-medium text-sm">Kernel PCA Options</h4>
                                    <div className="grid grid-cols-2 gap-4">
                                        <div>
                                            <label className="block text-sm font-medium mb-1">
                                                Kernel Type
                                            </label>
                                            <select
                                                value={config.kernelType}
                                                onChange={(e) => setConfig({...config, kernelType: e.target.value})}
                                                className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                            >
                                                <option value="rbf">RBF (Gaussian)</option>
                                                <option value="linear">Linear</option>
                                                <option value="poly">Polynomial</option>
                                            </select>
                                        </div>
                                        <div>
                                            <label className="block text-sm font-medium mb-1">
                                                Gamma
                                            </label>
                                            <input
                                                type="number"
                                                value={config.kernelGamma}
                                                step="0.01"
                                                min="0.001"
                                                onChange={(e) => setConfig({...config, kernelGamma: parseFloat(e.target.value) || 1.0})}
                                                className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                            />
                                        </div>
                                        {config.kernelType === 'poly' && (
                                            <>
                                                <div>
                                                    <label className="block text-sm font-medium mb-1">
                                                        Degree
                                                    </label>
                                                    <input
                                                        type="number"
                                                        value={config.kernelDegree}
                                                        min="1"
                                                        max="10"
                                                        onChange={(e) => setConfig({...config, kernelDegree: parseInt(e.target.value) || 3})}
                                                        className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                                    />
                                                </div>
                                                <div>
                                                    <label className="block text-sm font-medium mb-1">
                                                        Coef0
                                                    </label>
                                                    <input
                                                        type="number"
                                                        value={config.kernelCoef0}
                                                        step="0.1"
                                                        onChange={(e) => setConfig({...config, kernelCoef0: parseFloat(e.target.value) || 0.0})}
                                                        className="w-full px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                                    />
                                                </div>
                                            </>
                                        )}
                                    </div>
                                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                                        Note: Kernel PCA uses its own centering in kernel space. Standard preprocessing options will be ignored.
                                    </p>
                                </div>
                            )}
                            
                            <button
                                onClick={runPCA}
                                disabled={loading}
                                className="mt-4 px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 dark:disabled:bg-gray-600 rounded-lg font-medium text-white"
                            >
                                {loading ? 'Running...' : 'Go PCA!'}
                            </button>
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
                            <h2 className="text-xl font-semibold mb-4">PCA Results</h2>
                            
                            {/* Info message about missing data handling */}
                            {pcaResponse.info && (
                                <div className="mb-4 p-3 bg-blue-100 dark:bg-blue-800 border border-blue-300 dark:border-blue-600 rounded-lg">
                                    <p className="text-blue-700 dark:text-blue-200 text-sm">
                                        <span className="font-semibold">Note:</span> {pcaResponse.info}
                                    </p>
                                </div>
                            )}
                            
                            {/* Explained Variance */}
                            <div className="bg-gray-100 dark:bg-gray-700 rounded-lg p-4">
                                <h3 className="text-lg font-semibold mb-2">Explained Variance</h3>
                                <div className="space-y-2">
                                    {pcaResponse.result.explained_variance.map((variance, i) => {
                                        const percentage = variance;
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
                            
                            {/* Plot Selector and Visualization */}
                            <div className="mt-6">
                                <div className="flex items-center justify-between mb-4">
                                    <h3 className="text-lg font-semibold">Visualizations</h3>
                                    <div className="flex items-center gap-4">
                                        {/* Group selection for color coding */}
                                        {(selectedPlot === 'scores' || selectedPlot === 'biplot') && fileData?.categoricalColumns && Object.keys(fileData.categoricalColumns).length > 0 && (
                                            <div className="flex items-center gap-2">
                                                <label className="text-sm text-gray-600 dark:text-gray-400">Color by:</label>
                                                <select
                                                    value={selectedGroupColumn || ''}
                                                    onChange={(e) => setSelectedGroupColumn(e.target.value || null)}
                                                    className="px-2 py-1 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white"
                                                >
                                                    <option value="">None</option>
                                                    {Object.keys(fileData.categoricalColumns).map((colName) => (
                                                        <option key={colName} value={colName}>
                                                            {colName}
                                                        </option>
                                                    ))}
                                                </select>
                                            </div>
                                        )}
                                        {(selectedPlot === 'scores' || selectedPlot === 'biplot') && pcaResponse.result.scores[0]?.length > 2 && (
                                            <>
                                                <div className="flex items-center gap-2">
                                                    <label className="text-sm text-gray-600 dark:text-gray-400">X-axis:</label>
                                                    <select
                                                        value={selectedXComponent}
                                                        onChange={(e) => setSelectedXComponent(parseInt(e.target.value))}
                                                        className="px-2 py-1 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white"
                                                    >
                                                        {pcaResponse.result.component_labels?.map((label, i) => (
                                                            <option key={i} value={i}>
                                                                {label} ({pcaResponse.result!.explained_variance[i].toFixed(1)}%)
                                                            </option>
                                                        ))}
                                                    </select>
                                                </div>
                                                <div className="flex items-center gap-2">
                                                    <label className="text-sm text-gray-600 dark:text-gray-400">Y-axis:</label>
                                                    <select
                                                        value={selectedYComponent}
                                                        onChange={(e) => setSelectedYComponent(parseInt(e.target.value))}
                                                        className="px-2 py-1 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white"
                                                    >
                                                        {pcaResponse.result.component_labels?.map((label, i) => (
                                                            <option key={i} value={i}>
                                                                {label} ({pcaResponse.result!.explained_variance[i].toFixed(1)}%)
                                                            </option>
                                                        ))}
                                                    </select>
                                                </div>
                                            </>
                                        )}
                                        {selectedPlot === 'loadings' && (
                                            <div className="flex items-center gap-2">
                                                <label className="text-sm text-gray-600 dark:text-gray-400">Component:</label>
                                                <select
                                                    value={selectedLoadingComponent}
                                                    onChange={(e) => setSelectedLoadingComponent(parseInt(e.target.value))}
                                                    className="px-2 py-1 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded text-sm text-gray-900 dark:text-white"
                                                >
                                                    {pcaResponse.result?.component_labels?.map((label, i) => (
                                                        <option key={i} value={i}>
                                                            {label} ({pcaResponse.result!.explained_variance[i].toFixed(1)}%)
                                                        </option>
                                                    ))}
                                                </select>
                                            </div>
                                        )}
                                        <select
                                            value={selectedPlot}
                                            onChange={(e) => setSelectedPlot(e.target.value as 'scores' | 'scree' | 'loadings' | 'biplot')}
                                            className="px-3 py-2 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white"
                                        >
                                            <option value="scores">Scores Plot</option>
                                            <option value="scree">Scree Plot</option>
                                            <option value="loadings">Loadings Plot</option>
                                            <option value="biplot">Biplot</option>
                                        </select>
                                    </div>
                                </div>
                                
                                <div className="bg-gray-50 dark:bg-gray-700 rounded-lg p-4 border border-gray-200 dark:border-gray-600" style={{ height: '500px' }}>
                                    {selectedPlot === 'scores' && pcaResponse.result.scores.length > 0 && pcaResponse.result.scores[0].length >= 2 ? (
                                        <ScoresPlot
                                            pcaResult={pcaResponse.result}
                                            rowNames={fileData?.rowNames || []}
                                            xComponent={selectedXComponent}
                                            yComponent={selectedYComponent}
                                            groupColumn={selectedGroupColumn}
                                            groupLabels={selectedGroupColumn && fileData?.categoricalColumns ? fileData.categoricalColumns[selectedGroupColumn] : undefined}
                                        />
                                    ) : selectedPlot === 'scree' ? (
                                        <ScreePlot
                                            pcaResult={pcaResponse.result}
                                            showCumulative={true}
                                            elbowThreshold={80}
                                        />
                                    ) : selectedPlot === 'loadings' ? (
                                        <LoadingsPlot
                                            pcaResult={pcaResponse.result}
                                            selectedComponent={selectedLoadingComponent}
                                        />
                                    ) : selectedPlot === 'biplot' ? (
                                        <Biplot
                                            pcaResult={pcaResponse.result}
                                            rowNames={fileData?.rowNames || []}
                                            xComponent={selectedXComponent}
                                            yComponent={selectedYComponent}
                                            groupColumn={selectedGroupColumn}
                                            groupLabels={selectedGroupColumn && fileData?.categoricalColumns ? fileData.categoricalColumns[selectedGroupColumn] : undefined}
                                        />
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
                                            <p>Not enough components for scores plot (minimum 2 required)</p>
                                        </div>
                                    )}
                                </div>
                            </div>
                            
                            {/* Scores Matrix */}
                            <div className="mt-6">
                                <DataTable
                                    headers={pcaResponse.result.component_labels || []}
                                    rowNames={fileData?.rowNames || []}
                                    data={pcaResponse.result.scores}
                                    title="Scores Matrix"
                                />
                            </div>
                            
                        </div>
                    )}
                </div>
            </main>
            </div>
        </ThemeProvider>
    );
}

export default App;