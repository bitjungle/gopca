import React, { useState, useRef, useEffect } from 'react';
import './App.css';
import { CSVGrid, ValidationResults, MissingValueSummary, MissingValueDialog, DataQualityDashboard, UndoRedoControls, ImportWizard, DataTransformDialog, DocumentationViewer, ConfirmDialog } from './components';
import { ThemeProvider, ThemeToggle } from '@gopca/ui-components';
import logo from './assets/images/GoCSV-logo-1024-transp.png';
import { LoadCSV, SaveCSV, SaveExcel, ValidateForGoPCA, AnalyzeMissingValues, FillMissingValues, AnalyzeDataQuality, CheckGoPCAStatus, OpenInGoPCA, DownloadGoPCA, ExecuteCellEdit, ExecuteHeaderEdit, ExecuteFillMissingValues, ClearHistory } from '../wailsjs/go/main/App';
import { EventsOn, OnFileDrop, OnFileDropOff } from '../wailsjs/runtime/runtime';
import { main } from '../wailsjs/go/models';

type FileData = main.FileData;

function AppContent() {
    const [fileLoaded, setFileLoaded] = useState(false);
    const [fileName, setFileName] = useState<string | null>(null);
    const [fileData, setFileData] = useState<FileData | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [validationResult, setValidationResult] = useState<{ isValid: boolean; messages: string[] } | null>(null);
    const [isValidating, setIsValidating] = useState(false);
    const [missingValueStats, setMissingValueStats] = useState<main.MissingValueStats | null>(null);
    const [showMissingValueSummary, setShowMissingValueSummary] = useState(false);
    const [showMissingValueDialog, setShowMissingValueDialog] = useState(false);
    const [dataQualityReport, setDataQualityReport] = useState<main.DataQualityReport | null>(null);
    const [showDataQualityReport, setShowDataQualityReport] = useState(false);
    const [isAnalyzingQuality, setIsAnalyzingQuality] = useState(false);
    const [gopcaStatus, setGopcaStatus] = useState<main.GoPCAStatus | null>(null);
    const [isCheckingGoPCA, setIsCheckingGoPCA] = useState(false);
    const [showImportWizard, setShowImportWizard] = useState(false);
    const [showTransformDialog, setShowTransformDialog] = useState(false);
    const [showDocumentation, setShowDocumentation] = useState(false);
    const [showDownloadConfirm, setShowDownloadConfirm] = useState(false);
    
    // Ref for scrolling to Step 2
    const step2Ref = useRef<HTMLDivElement>(null);
    
    // Listen for file-loaded events from backend
    useEffect(() => {
        const unsubscribe = EventsOn('file-loaded', (filename: string) => {
            setFileName(filename);
        });
        
        // Check GoPCA status on startup
        checkGoPCAInstallation();
        
        // Set up drag and drop listener
        OnFileDrop(async (x: number, y: number, paths: string[]) => {
            if (paths.length > 0) {
                const filePath = paths[0];
                // Only accept supported file types
                if (filePath.match(/\.(csv|tsv|xlsx|xls)$/i)) {
                    await handleDroppedFile(filePath);
                } else {
                    alert('Please drop a supported file type: CSV, TSV, or Excel files');
                }
            }
        }, false);
        
        return () => {
            unsubscribe();
            OnFileDropOff();
        };
    }, []);
    
    // Auto-scroll to Step 2 when file is loaded
    useEffect(() => {
        if (fileLoaded && fileData && step2Ref.current) {
            // Small delay to ensure DOM is updated
            setTimeout(() => {
                step2Ref.current?.scrollIntoView({ 
                    behavior: 'smooth', 
                    block: 'start' 
                });
            }, 100);
        }
    }, [fileLoaded, fileData]);
    
    // Scroll to top function
    const scrollToTop = () => {
        window.scrollTo({ top: 0, behavior: 'smooth' });
    };
    
    // Check GoPCA installation status
    const checkGoPCAInstallation = async () => {
        setIsCheckingGoPCA(true);
        try {
            const status = await CheckGoPCAStatus();
            setGopcaStatus(status);
        } catch (error) {
            console.error('Error checking GoPCA status:', error);
            setGopcaStatus({
                installed: false,
                path: '',
                version: '',
                error: 'Failed to check GoPCA status'
            });
        } finally {
            setIsCheckingGoPCA(false);
        }
    };
    
    // Handle file selection
    // Handle files dropped via Wails drag and drop
    const handleDroppedFile = async (filePath: string) => {
        setIsLoading(true);
        try {
            const result = await LoadCSV(filePath);
            if (result && result.data && result.data.length > 0) {
                setFileData(result);
                setFileLoaded(true);
                // Clear history when loading new file
                ClearHistory();
                // Clear any previous analysis
                setMissingValueStats(null);
                setDataQualityReport(null);
                setValidationResult(null);
            } else {
                alert('Failed to load file or file is empty');
            }
        } catch (error) {
            console.error('Error loading dropped file:', error);
            alert(`Error loading file: ${error}`);
        } finally {
            setIsLoading(false);
        }
    };
    
    // Load file from dialog
    const handleLoadFromDialog = async () => {
        setIsLoading(true);
        try {
            const result = await LoadCSV('');
            console.log('Loaded file data:', result);
            if (result && result.data && result.data.length > 0) {
                console.log('Setting file data:', {
                    headers: result.headers?.length,
                    rows: result.rows,
                    columns: result.columns,
                    dataLength: result.data?.length
                });
                setFileData(result);
                setFileLoaded(true);
                // Filename will be set by the event from backend
                // Clear history when loading new file
                ClearHistory();
            } else {
                console.error('Invalid file data received:', result);
                throw new Error('No data found in file');
            }
        } catch (error: any) {
            console.error('Error loading file:', error);
            const errorMsg = error?.message || error?.toString() || 'Unknown error';
            alert('Error loading file: ' + errorMsg);
            setFileLoaded(false);
            setFileName(null);
        } finally {
            setIsLoading(false);
        }
    };
    
    // Handle data changes
    const handleDataChange = async (rowIndex: number, colIndex: number, newValue: string) => {
        if (fileData && fileData.data[rowIndex][colIndex] !== newValue) {
            try {
                const oldValue = fileData.data[rowIndex][colIndex];
                const updatedData = await ExecuteCellEdit(fileData, rowIndex, colIndex, oldValue, newValue);
                setFileData(updatedData);
                setValidationResult(null);
            } catch (error) {
                console.error('Error updating cell:', error);
                // Optionally show an error message to the user
            }
        }
    };
    
    // Handle header changes
    const handleHeaderChange = async (colIndex: number, newHeader: string) => {
        if (fileData && fileData.headers[colIndex] !== newHeader) {
            try {
                const oldHeader = fileData.headers[colIndex];
                const updatedData = await ExecuteHeaderEdit(fileData, colIndex, oldHeader, newHeader);
                setFileData(updatedData);
                setValidationResult(null);
            } catch (error) {
                console.error('Error updating header:', error);
            }
        }
    };
    
    // Handle validation
    const handleValidate = async () => {
        if (!fileData) return;
        
        setIsValidating(true);
        try {
            const result = await ValidateForGoPCA(fileData);
            if (result) {
                setValidationResult({
                    isValid: result.isValid,
                    messages: result.messages || []
                });
            }
        } catch (error) {
            console.error('Validation error:', error);
            setValidationResult({
                isValid: false,
                messages: ['ERROR: Failed to validate data - ' + error]
            });
        } finally {
            setIsValidating(false);
        }
    };
    
    // Handle missing value analysis
    const handleAnalyzeMissingValues = async () => {
        if (!fileData) return;
        
        try {
            const stats = await AnalyzeMissingValues(fileData);
            setMissingValueStats(stats);
            setShowMissingValueSummary(true);
        } catch (error) {
            console.error('Error analyzing missing values:', error);
            alert('Error analyzing missing values: ' + error);
        }
    };
    
    // Handle missing value fill
    const handleFillMissingValues = async (strategy: string, column: string, value?: string) => {
        if (!fileData) return;
        
        try {
            const request = {
                strategy,
                column,
                value: value || ''
            };
            const result = await FillMissingValues(fileData, request);
            if (result) {
                setFileData(result);
                setValidationResult(null);
                // Re-analyze missing values
                const stats = await AnalyzeMissingValues(result);
                setMissingValueStats(stats);
            }
        } catch (error) {
            console.error('Error filling missing values:', error);
            alert('Error filling missing values: ' + error);
        }
    };
    
    // Handle import completion from wizard
    const handleImportComplete = (data: FileData) => {
        setFileData(data);
        setFileLoaded(true);
        setShowImportWizard(false);
        setValidationResult(null);
        setMissingValueStats(null);
        // Filename will be set by the event from backend
    };
    
    // Handle transform completion
    const handleTransformComplete = (data: FileData) => {
        setFileData(data);
        setValidationResult(null);
        setShowTransformDialog(false);
    };
    
    
    return (
        <div className="flex flex-col h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white transition-colors duration-200">
            {/* Header - matching GoPCA Desktop exactly */}
            <header className="sticky top-0 z-50 bg-white dark:bg-gray-800 shadow-lg backdrop-blur-sm bg-opacity-95 dark:bg-opacity-95">
                <div className="flex items-center justify-between max-w-7xl mx-auto px-4 py-3 h-20">
                    <div className="flex items-center gap-4">
                        <img 
                            src={logo} 
                            alt="GoCSV - GoPCA CSV Editor" 
                            className="h-12 cursor-pointer hover:opacity-90 transition-opacity flex-shrink-0"
                            onClick={scrollToTop}
                        />
                        <div>
                            <p className="text-sm text-gray-600 dark:text-gray-400">Data Editor for GoPCA</p>
                        </div>
                    </div>
                    <div className="flex items-center gap-4">
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
            
            {/* Main content area */}
            <main className="flex-1 overflow-y-auto p-4 md:p-6 max-w-7xl mx-auto w-full">
                <div className="space-y-6">
                    {/* Step 1: Load Data - matching GoPCA's card style */}
                    <div className="bg-white dark:bg-gray-800 rounded-xl shadow-md p-6 animate-fadeIn">
                        <h2 className="text-lg font-semibold mb-4 text-gray-800 dark:text-gray-200">
                            Step 1: Load Data
                        </h2>
                        
                        <div className="space-y-4">
                            <div 
                                className="border-2 border-dashed rounded-lg p-8 text-center transition-colors border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500"
                            >
                                <svg className="mx-auto h-12 w-12 text-gray-400 transition-colors" stroke="currentColor" fill="none" viewBox="0 0 48 48" aria-hidden="true">
                                    <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                                </svg>
                                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                                    <span className="font-medium text-gray-900 dark:text-gray-100">
                                        Drag and drop your file here
                                    </span>
                                    <br />
                                    <span className="text-xs">CSV, TSV, or Excel files</span>
                                </p>
                            </div>
                            
                            <div className="text-center">
                                <span className="text-gray-500 dark:text-gray-400 text-sm">or</span>
                            </div>
                            
                            <div className="space-y-3">
                                <div className="grid grid-cols-2 gap-2">
                                    <button
                                        onClick={handleLoadFromDialog}
                                        disabled={isLoading}
                                        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                    >
                                        {isLoading ? 'Loading...' : 'Browse for File'}
                                    </button>
                                    <button
                                        onClick={() => setShowImportWizard(true)}
                                        disabled={isLoading}
                                        className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                    >
                                        Import with Wizard
                                    </button>
                                </div>
                                <div className="text-xs text-gray-500 dark:text-gray-400 space-y-1">
                                    <p><span className="font-medium">Browse for File:</span> Quick file picker for standard CSV/TSV files with automatic format detection</p>
                                    <p><span className="font-medium">Import with Wizard:</span> Advanced options for Excel sheets, custom delimiters, header rows, and column selection</p>
                                </div>
                            </div>
                            
                            {fileName && (
                                <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-3 flex items-center justify-between">
                                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                        {fileName}
                                    </span>
                                    <button
                                        onClick={() => {
                                            setFileName(null);
                                            setFileLoaded(false);
                                            setFileData(null);
                                            setValidationResult(null);
                                        }}
                                        className="text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
                                    >
                                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                                        </svg>
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>
                    
                    {/* Step 2: Edit Data */}
                    {fileLoaded && fileData && (
                        <div ref={step2Ref} className="bg-white dark:bg-gray-800 rounded-xl shadow-md p-6 animate-fadeIn">
                            <div className="flex items-center justify-between mb-4">
                                <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200 text-center flex-1">
                                    Step 2: Edit Data
                                </h2>
                                <div className="text-sm text-gray-600 dark:text-gray-400">
                                    {fileData.rows} rows × {fileData.columns} columns
                                </div>
                            </div>
                            
                            {/* Data Quality Toolbar */}
                            <div className="flex items-center justify-between mb-4 p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                                <div className="flex items-center gap-4">
                                    <UndoRedoControls onDataUpdate={setFileData} />
                                    <div className="w-px h-6 bg-gray-300 dark:bg-gray-600" />
                                    <button
                                        onClick={async () => {
                                            if (!fileData) return;
                                            setIsAnalyzingQuality(true);
                                            try {
                                                const report = await AnalyzeDataQuality(fileData);
                                                setDataQualityReport(report);
                                                setShowDataQualityReport(true);
                                            } catch (error) {
                                                console.error('Error analyzing data quality:', error);
                                                alert('Error analyzing data quality: ' + error);
                                            } finally {
                                                setIsAnalyzingQuality(false);
                                            }
                                        }}
                                        disabled={isAnalyzingQuality}
                                        className="px-3 py-1.5 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                    >
                                        <span className="flex items-center gap-2">
                                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                            </svg>
                                            {isAnalyzingQuality ? 'Analyzing...' : 'Data Quality Report'}
                                        </span>
                                    </button>
                                    <button
                                        onClick={handleAnalyzeMissingValues}
                                        className="px-3 py-1.5 text-sm bg-white dark:bg-gray-600 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-100 dark:hover:bg-gray-500 transition-colors border border-gray-300 dark:border-gray-500"
                                    >
                                        <span className="flex items-center gap-2">
                                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                                            </svg>
                                            Analyze Missing Values
                                        </span>
                                    </button>
                                    <button
                                        onClick={() => setShowMissingValueDialog(true)}
                                        className="px-3 py-1.5 text-sm bg-white dark:bg-gray-600 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-100 dark:hover:bg-gray-500 transition-colors border border-gray-300 dark:border-gray-500"
                                    >
                                        <span className="flex items-center gap-2">
                                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                                            </svg>
                                            Fill Missing Values
                                        </span>
                                    </button>
                                    <button
                                        onClick={() => setShowTransformDialog(true)}
                                        className="px-3 py-1.5 text-sm bg-purple-600 text-white rounded hover:bg-purple-700 transition-colors"
                                    >
                                        <span className="flex items-center gap-2">
                                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
                                            </svg>
                                            Transform Data
                                        </span>
                                    </button>
                                </div>
                                {missingValueStats && (
                                    <div className="text-sm text-gray-600 dark:text-gray-400">
                                        Missing: {missingValueStats.missingCells} cells ({missingValueStats.missingPercent?.toFixed(1)}%)
                                    </div>
                                )}
                            </div>
                            
                            <div className="h-[600px] w-full">
                                <CSVGrid 
                                    data={fileData.data}
                                    headers={fileData.headers}
                                    rowNames={fileData.rowNames}
                                    fileData={fileData}
                                    onDataChange={handleDataChange}
                                    onHeaderChange={handleHeaderChange}
                                    onRowNameChange={(rowIndex, newRowName) => {
                                        if (fileData && fileData.rowNames) {
                                            const newRowNames = [...fileData.rowNames];
                                            newRowNames[rowIndex] = newRowName;
                                            setFileData({ ...fileData, rowNames: newRowNames });
                                            setValidationResult(null);
                                        }
                                    }}
                                    onRefresh={async (updatedData?: any) => {
                                        if (updatedData) {
                                            // Use the updated data returned from backend
                                            setFileData(updatedData);
                                            setValidationResult(null);
                                        } else if (fileData) {
                                            // Fallback: Force React to re-render with updated data
                                            setFileData({
                                                ...fileData,
                                                headers: [...fileData.headers],
                                                data: fileData.data.map(row => [...row])
                                            });
                                            setValidationResult(null);
                                        }
                                    }}
                                />
                            </div>
                        </div>
                    )}
                    
                    {/* Step 3: Validate & Export */}
                    {fileLoaded && (
                        <div className="bg-white dark:bg-gray-800 rounded-xl shadow-md p-6 animate-fadeIn">
                            <h2 className="text-lg font-semibold mb-4 text-gray-800 dark:text-gray-200">
                                Step 3: Validate & Export
                            </h2>
                            
                            <div className="space-y-4">
                                <div className="flex gap-4">
                                    <button 
                                        onClick={handleValidate}
                                        disabled={isValidating}
                                        className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                    >
                                        {isValidating ? 'Validating...' : 'Validate for GoPCA'}
                                    </button>
                                    <button 
                                        onClick={async () => {
                                            if (!fileData) return;
                                            
                                            // Check if GoPCA is installed
                                            if (!gopcaStatus?.installed) {
                                                setShowDownloadConfirm(true);
                                                return;
                                            }
                                            
                                            try {
                                                await OpenInGoPCA(fileData);
                                            } catch (error) {
                                                console.error('Error opening in GoPCA:', error);
                                                alert('Error opening in GoPCA: ' + error);
                                            }
                                        }}
                                        disabled={!gopcaStatus || isCheckingGoPCA}
                                        className="flex-1 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                    >
                                        {isCheckingGoPCA ? 'Checking...' : 
                                         !gopcaStatus?.installed ? 'Install GoPCA' : 
                                         'Open in GoPCA'}
                                    </button>
                                </div>
                                
                                {validationResult && (
                                    <ValidationResults 
                                        isValid={validationResult.isValid}
                                        messages={validationResult.messages}
                                        onClose={() => setValidationResult(null)}
                                    />
                                )}
                                
                                {/* GoPCA Status */}
                                {gopcaStatus && (
                                    <div className={`p-3 rounded-lg text-sm ${
                                        gopcaStatus.installed 
                                            ? 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300' 
                                            : 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-700 dark:text-yellow-300'
                                    }`}>
                                        <div className="flex items-center justify-between">
                                            <div className="flex items-center gap-2">
                                                <svg className={`w-4 h-4 ${gopcaStatus.installed ? 'text-green-600' : 'text-yellow-600'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                    {gopcaStatus.installed ? (
                                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                                                    ) : (
                                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                                                    )}
                                                </svg>
                                                <span>
                                                    {gopcaStatus.installed 
                                                        ? `GoPCA Desktop ${gopcaStatus.version || 'detected'}` 
                                                        : 'GoPCA Desktop not found'}
                                                </span>
                                            </div>
                                            {!gopcaStatus.installed && (
                                                <button
                                                    onClick={() => checkGoPCAInstallation()}
                                                    className="text-xs text-yellow-600 hover:text-yellow-700 dark:text-yellow-400 dark:hover:text-yellow-300"
                                                >
                                                    Refresh
                                                </button>
                                            )}
                                        </div>
                                    </div>
                                )}
                                
                                <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
                                    <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                        Export Options
                                    </h3>
                                    <div className="grid grid-cols-2 gap-2">
                                        <button 
                                            onClick={async () => {
                                                if (fileData) {
                                                    try {
                                                        await SaveCSV(fileData);
                                                    } catch (error) {
                                                        console.error('Error saving file:', error);
                                                        alert('Error saving file: ' + error);
                                                    }
                                                }
                                            }}
                                            className="px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                                        >
                                            Export as CSV
                                        </button>
                                        <button 
                                            onClick={async () => {
                                                if (fileData) {
                                                    try {
                                                        await SaveExcel(fileData);
                                                    } catch (error) {
                                                        console.error('Error saving Excel file:', error);
                                                        alert('Error saving Excel file: ' + error);
                                                    }
                                                }
                                            }}
                                            className="px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                                        >
                                            Export as Excel
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            </main>
            
            {/* Missing Value Summary Dialog */}
            <MissingValueSummary 
                stats={missingValueStats}
                isOpen={showMissingValueSummary}
                onClose={() => setShowMissingValueSummary(false)}
            />
            
            {/* Missing Value Fill Dialog */}
            <MissingValueDialog
                isOpen={showMissingValueDialog}
                onClose={() => setShowMissingValueDialog(false)}
                onFill={handleFillMissingValues}
                columns={fileData?.headers || []}
                columnTypes={fileData?.columnTypes || {}}
            />
            
            {/* Data Quality Report Dashboard */}
            <DataQualityDashboard
                report={dataQualityReport}
                isOpen={showDataQualityReport}
                onClose={() => setShowDataQualityReport(false)}
            />
            
            {/* Import Wizard */}
            <ImportWizard
                isOpen={showImportWizard}
                onClose={() => setShowImportWizard(false)}
                onImportComplete={handleImportComplete}
            />
            
            {/* Data Transform Dialog */}
            {fileData && (
                <DataTransformDialog
                    isOpen={showTransformDialog}
                    onClose={() => setShowTransformDialog(false)}
                    fileData={fileData}
                    onTransformComplete={handleTransformComplete}
                />
            )}
            
            {/* Documentation Viewer */}
            <DocumentationViewer 
                isOpen={showDocumentation}
                onClose={() => setShowDocumentation(false)}
            />
            
            {/* Download GoPCA Confirmation */}
            <ConfirmDialog
                isOpen={showDownloadConfirm}
                onClose={() => setShowDownloadConfirm(false)}
                onConfirm={async () => {
                    setShowDownloadConfirm(false);
                    try {
                        await DownloadGoPCA();
                    } catch (error) {
                        console.error('Error downloading GoPCA:', error);
                    }
                }}
                title="GoPCA Not Installed"
                message="GoPCA Desktop is not installed. Would you like to download it?"
                confirmText="Download"
                cancelText="Cancel"
            />
        </div>
    );
}

function App() {
    return (
        <ThemeProvider>
            <AppContent />
        </ThemeProvider>
    );
}

export default App;