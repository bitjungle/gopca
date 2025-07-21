import React, { useState } from 'react';
import './App.css';
import { ParseCSV, RunPCA } from "../wailsjs/go/main/App";
import { DataTable } from './components/DataTable';
import { ScoresPlot } from './components/visualizations';
import { FileData, PCARequest, PCAResponse } from './types';
import logo from './assets/images/GoPCA-logo-1024.png';

function App() {
    const [fileData, setFileData] = useState<FileData | null>(null);
    const [pcaResponse, setPcaResponse] = useState<PCAResponse | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    
    // Selection state
    const [excludedRows, setExcludedRows] = useState<number[]>([]);
    const [excludedColumns, setExcludedColumns] = useState<number[]>([]);
    
    // PCA configuration
    const [config, setConfig] = useState({
        components: 2,
        meanCenter: true,
        standardScale: true,
        robustScale: false,
        method: 'NIPALS'
    });
    
    const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (!file) return;
        
        setLoading(true);
        setError(null);
        
        try {
            const content = await file.text();
            const result = await ParseCSV(content);
            setFileData(result);
            setPcaResponse(null);
            // Reset exclusions when loading new data
            setExcludedRows([]);
            setExcludedColumns([]);
        } catch (err) {
            setError(`Failed to parse CSV: ${err}`);
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
        setError(null);
        
        try {
            const request: PCARequest = {
                ...fileData,
                ...config,
                excludedRows,
                excludedColumns
            };
            const result = await RunPCA(request);
            if (result.success) {
                setPcaResponse(result);
            } else {
                setError(result.error || 'PCA analysis failed');
            }
        } catch (err) {
            setError(`Failed to run PCA: ${err}`);
        } finally {
            setLoading(false);
        }
    };
    
    return (
        <div className="flex flex-col h-screen bg-gray-900 text-white">
            <header className="bg-gray-800 p-4 shadow-lg">
                <img src={logo} alt="GoPCA - Principal Component Analysis Tool" className="h-12 mx-auto" />
            </header>
            
            <main className="flex-1 overflow-auto p-6">
                <div className="max-w-7xl mx-auto space-y-6">
                    {/* File Upload Section */}
                    <div className="bg-gray-800 rounded-lg p-6 shadow-lg">
                        <h2 className="text-xl font-semibold mb-4">Step 1: Load Data</h2>
                        <div>
                            <label className="block text-sm font-medium mb-2">
                                Upload CSV File
                            </label>
                            <input
                                type="file"
                                accept=".csv"
                                onChange={handleFileUpload}
                                className="block w-full text-sm text-gray-300
                                    file:mr-4 file:py-2 file:px-4
                                    file:rounded-full file:border-0
                                    file:text-sm file:font-semibold
                                    file:bg-blue-600 file:text-white
                                    hover:file:bg-blue-700"
                            />
                        </div>
                    </div>
                    
                    {/* Data Display */}
                    {fileData && (
                        <div className="bg-gray-800 rounded-lg p-6 shadow-lg">
                            <h2 className="text-xl font-semibold mb-4">Loaded Data</h2>
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
                        </div>
                    )}
                    
                    {/* Configuration Section */}
                    {fileData && (
                        <div className="bg-gray-800 rounded-lg p-6 shadow-lg">
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
                                        className="w-full px-3 py-2 bg-gray-700 rounded-lg"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium mb-2">
                                        Method
                                    </label>
                                    <select
                                        value={config.method}
                                        onChange={(e) => setConfig({...config, method: e.target.value})}
                                        className="w-full px-3 py-2 bg-gray-700 rounded-lg"
                                    >
                                        <option value="NIPALS">NIPALS</option>
                                        <option value="SVD">SVD</option>
                                    </select>
                                </div>
                            </div>
                            <div className="mt-4 space-y-2">
                                <label className="flex items-center">
                                    <input
                                        type="checkbox"
                                        checked={config.meanCenter}
                                        onChange={(e) => setConfig({...config, meanCenter: e.target.checked})}
                                        className="mr-2"
                                    />
                                    Mean Center
                                </label>
                                <label className="flex items-center">
                                    <input
                                        type="checkbox"
                                        checked={config.standardScale}
                                        onChange={(e) => setConfig({...config, standardScale: e.target.checked})}
                                        className="mr-2"
                                    />
                                    Standard Scale
                                </label>
                                <label className="flex items-center">
                                    <input
                                        type="checkbox"
                                        checked={config.robustScale}
                                        onChange={(e) => setConfig({...config, robustScale: e.target.checked})}
                                        className="mr-2"
                                    />
                                    Robust Scale
                                </label>
                            </div>
                            <button
                                onClick={runPCA}
                                disabled={loading}
                                className="mt-4 px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 rounded-lg font-medium"
                            >
                                {loading ? 'Running...' : 'Run PCA Analysis'}
                            </button>
                        </div>
                    )}
                    
                    {/* Error Display */}
                    {error && (
                        <div className="bg-red-800 border border-red-600 rounded-lg p-4">
                            <p className="text-red-200">{error}</p>
                        </div>
                    )}
                    
                    {/* PCA Results */}
                    {pcaResponse?.success && pcaResponse.result && (
                        <div className="bg-gray-800 rounded-lg p-6 shadow-lg">
                            <h2 className="text-xl font-semibold mb-4">PCA Results</h2>
                            
                            {/* Scores Matrix */}
                            <div className="mb-6">
                                <DataTable
                                    headers={pcaResponse.result.component_labels || []}
                                    rowNames={fileData?.rowNames || []}
                                    data={pcaResponse.result.scores}
                                    title="Scores Matrix"
                                />
                            </div>
                            
                            {/* Explained Variance */}
                            <div className="bg-gray-700 rounded-lg p-4">
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
                                    <div className="border-t border-gray-600 pt-2 font-semibold">
                                        <div className="flex justify-between">
                                            <span>Cumulative:</span>
                                            <span>
                                                {pcaResponse.result.cumulative_variance[pcaResponse.result.cumulative_variance.length - 1].toFixed(2)}%
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            
                            {/* Scores Plot */}
                            {pcaResponse.result.scores.length > 0 && pcaResponse.result.scores[0].length >= 2 && (
                                <div className="mt-6">
                                    <h3 className="text-lg font-semibold mb-4">Scores Plot</h3>
                                    <div className="bg-gray-700 rounded-lg p-4" style={{ height: '500px' }}>
                                        <ScoresPlot
                                            pcaResult={pcaResponse.result}
                                            rowNames={fileData?.rowNames || []}
                                        />
                                    </div>
                                </div>
                            )}
                            
                        </div>
                    )}
                </div>
            </main>
        </div>
    );
}

export default App;