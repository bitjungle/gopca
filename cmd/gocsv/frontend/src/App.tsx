import React, { useState, useRef, useEffect } from 'react';
import './App.css';
import { ThemeToggle, CSVGrid, ValidationResults } from './components';
import { ThemeProvider } from './contexts/ThemeContext';
import logo from './assets/images/GoCSV-logo-1024-transp.png';
import { LoadCSV, SaveCSV, SaveExcel, ValidateForGoPCA } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { main } from '../wailsjs/go/models';

type FileData = main.FileData;

function AppContent() {
    const [fileLoaded, setFileLoaded] = useState(false);
    const [fileName, setFileName] = useState<string | null>(null);
    const [fileData, setFileData] = useState<FileData | null>(null);
    const [isDragging, setIsDragging] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [validationResult, setValidationResult] = useState<{ isValid: boolean; messages: string[] } | null>(null);
    const [isValidating, setIsValidating] = useState(false);
    
    // Listen for file-loaded events from backend
    useEffect(() => {
        const unsubscribe = EventsOn('file-loaded', (filename: string) => {
            setFileName(filename);
        });
        
        return () => {
            unsubscribe();
        };
    }, []);
    
    // Scroll to top function
    const scrollToTop = () => {
        window.scrollTo({ top: 0, behavior: 'smooth' });
    };
    
    // Handle file selection
    const handleFile = async (file: File) => {
        // For drag-and-drop and file input, we can't pass the File object directly to Go
        // Instead, we need to trigger the file dialog in the backend
        // This is a limitation of the current Wails file handling
        // For now, just inform the user to use the Browse button
        alert(`File "${file.name}" selected. Please use the "Browse for File" button to load files.`);
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
    const handleDataChange = (rowIndex: number, colIndex: number, newValue: string) => {
        if (fileData) {
            const newData = [...fileData.data];
            newData[rowIndex][colIndex] = newValue;
            setFileData({ ...fileData, data: newData });
            // Clear validation when data changes
            setValidationResult(null);
        }
    };
    
    // Handle header changes
    const handleHeaderChange = (colIndex: number, newHeader: string) => {
        if (fileData) {
            const newHeaders = [...fileData.headers];
            newHeaders[colIndex] = newHeader;
            setFileData({ ...fileData, headers: newHeaders });
            // Clear validation when headers change
            setValidationResult(null);
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
    
    // Drag and drop handlers
    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
    };
    
    const handleDragEnter = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(true);
    };
    
    const handleDragLeave = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(false);
    };
    
    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(false);
        
        const files = e.dataTransfer.files;
        if (files && files.length > 0) {
            handleFile(files[0]);
        }
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
                            <p className="text-sm text-gray-600 dark:text-gray-400">CSV Editor for GoPCA</p>
                        </div>
                    </div>
                    <div className="flex items-center gap-4">
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
                                className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
                                    isDragging 
                                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' 
                                        : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500'
                                }`}
                                onDragOver={handleDragOver}
                                onDragEnter={handleDragEnter}
                                onDragLeave={handleDragLeave}
                                onDrop={handleDrop}
                            >
                                <input
                                    type="file"
                                    accept=".csv,.tsv,.xlsx,.xls"
                                    className="hidden"
                                    id="file-upload"
                                    onChange={(e) => {
                                        console.log('File input changed');
                                        const file = e.target.files?.[0];
                                        if (file) {
                                            console.log('File selected:', file.name);
                                            // Reset the input so the same file can be selected again
                                            e.target.value = '';
                                            handleFile(file);
                                        }
                                    }}
                                />
                                <label 
                                    htmlFor="file-upload"
                                    className="cursor-pointer"
                                >
                                    <svg className={`mx-auto h-12 w-12 ${isDragging ? 'text-blue-500' : 'text-gray-400'} transition-colors`} stroke="currentColor" fill="none" viewBox="0 0 48 48" aria-hidden="true">
                                        <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                                    </svg>
                                    <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                                        {isDragging ? (
                                            <span className="font-medium text-blue-600 dark:text-blue-400">
                                                Drop file here
                                            </span>
                                        ) : (
                                            <>
                                                <span className="font-medium text-gray-900 dark:text-white hover:text-blue-600 dark:hover:text-blue-400">
                                                    Click to upload
                                                </span>{' '}
                                                or drag and drop
                                            </>
                                        )}
                                    </p>
                                    <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                                        CSV, TSV, Excel files supported
                                    </p>
                                </label>
                            </div>
                            
                            <div className="text-center">
                                <span className="text-gray-500 dark:text-gray-400 text-sm">or</span>
                            </div>
                            
                            <button
                                onClick={handleLoadFromDialog}
                                disabled={isLoading}
                                className="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                            >
                                {isLoading ? 'Loading...' : 'Browse for File'}
                            </button>
                            
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
                        <div className="bg-white dark:bg-gray-800 rounded-xl shadow-md p-6 animate-fadeIn">
                            <div className="flex items-center justify-between mb-4">
                                <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
                                    Step 2: Edit Data
                                </h2>
                                <div className="text-sm text-gray-600 dark:text-gray-400">
                                    {fileData.rows} rows × {fileData.columns} columns
                                </div>
                            </div>
                            
                            <div className="h-[600px] w-full">
                                <CSVGrid 
                                    data={fileData.data}
                                    headers={fileData.headers}
                                    rowNames={fileData.rowNames}
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
                                        onClick={() => {
                                            // TODO: Implement Open in GoPCA
                                            alert('Open in GoPCA coming soon!');
                                        }}
                                        className="flex-1 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                                    >
                                        Open in GoPCA
                                    </button>
                                </div>
                                
                                {validationResult && (
                                    <ValidationResults 
                                        isValid={validationResult.isValid}
                                        messages={validationResult.messages}
                                        onClose={() => setValidationResult(null)}
                                    />
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