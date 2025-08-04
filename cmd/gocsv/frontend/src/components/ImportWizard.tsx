import React, { useState, useEffect } from 'react';
import { SelectFileForImport, GetFileInfo, PreviewFile, ImportFile } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';
import { FileSelector } from './FileSelector';
import { FormatOptions } from './FormatOptions';
import { DataPreview } from './DataPreview';
import { ImportProgress } from './ImportProgress';

type ImportFileInfo = main.ImportFileInfo;
type ImportOptions = main.ImportOptions;
type FilePreview = main.FilePreview;
type FileData = main.FileData;

interface ImportWizardProps {
    isOpen: boolean;
    onClose: () => void;
    onImportComplete: (data: FileData) => void;
}

type WizardStep = 'file-selection' | 'format-options' | 'data-preview' | 'importing';

export const ImportWizard: React.FC<ImportWizardProps> = ({ isOpen, onClose, onImportComplete }) => {
    const [currentStep, setCurrentStep] = useState<WizardStep>('file-selection');
    const [selectedFile, setSelectedFile] = useState<string | null>(null);
    const [fileInfo, setFileInfo] = useState<ImportFileInfo | null>(null);
    const [importOptions, setImportOptions] = useState<ImportOptions>({
        format: 'csv',
        delimiter: ',',
        hasHeaders: true,
        headerRow: 0,
        sheet: '',
        range: '',
        rowNameColumn: -1,
        skipRows: 0,
        maxRows: 0,
        selectedColumns: []
    });
    const [preview, setPreview] = useState<FilePreview | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [importProgress, setImportProgress] = useState(0);

    // Reset when closed
    useEffect(() => {
        if (!isOpen) {
            setCurrentStep('file-selection');
            setSelectedFile(null);
            setFileInfo(null);
            setPreview(null);
            setError(null);
            setImportProgress(0);
        }
    }, [isOpen]);

    const handleFileSelect = async (filePath: string) => {
        setIsLoading(true);
        setError(null);
        try {
            const info = await GetFileInfo(filePath);
            setSelectedFile(filePath);
            setFileInfo(info);
            
            // Update format based on file info
            setImportOptions(prev => ({
                ...prev,
                format: info.fileFormat,
                delimiter: info.fileFormat === 'tsv' ? '\t' : ','
            }));
            
            setCurrentStep('format-options');
        } catch (err) {
            setError(`Failed to get file info: ${err}`);
        } finally {
            setIsLoading(false);
        }
    };

    const handleOptionsNext = async () => {
        if (!selectedFile) return;
        
        setIsLoading(true);
        setError(null);
        try {
            const previewData = await PreviewFile(selectedFile, importOptions);
            setPreview(previewData);
            setCurrentStep('data-preview');
        } catch (err) {
            setError(`Failed to preview file: ${err}`);
        } finally {
            setIsLoading(false);
        }
    };

    const handleImport = async () => {
        if (!selectedFile) return;
        
        setCurrentStep('importing');
        setError(null);
        setImportProgress(0);
        
        try {
            // Simulate progress updates
            const progressInterval = setInterval(() => {
                setImportProgress(prev => Math.min(prev + 10, 90));
            }, 200);
            
            const data = await ImportFile(selectedFile, importOptions);
            
            clearInterval(progressInterval);
            setImportProgress(100);
            
            // Small delay to show 100% progress
            setTimeout(() => {
                onImportComplete(data);
                onClose();
            }, 500);
        } catch (err) {
            setError(`Failed to import file: ${err}`);
            setCurrentStep('data-preview');
        }
    };

    const handleBack = () => {
        switch (currentStep) {
            case 'format-options':
                setCurrentStep('file-selection');
                break;
            case 'data-preview':
                setCurrentStep('format-options');
                break;
            case 'importing':
                // Can't go back while importing
                break;
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-50 overflow-y-auto">
            <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
                {/* Background overlay */}
                <div 
                    className="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75 dark:bg-gray-900 dark:bg-opacity-75"
                    onClick={currentStep !== 'importing' ? onClose : undefined}
                />

                {/* Modal panel */}
                <div className="inline-block w-full max-w-4xl my-8 text-left align-middle transition-all transform bg-white dark:bg-gray-800 shadow-xl rounded-lg">
                    {/* Header */}
                    <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
                        <div className="flex items-center justify-between">
                            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                                Import Data Wizard
                            </h2>
                            {currentStep !== 'importing' && (
                                <button
                                    onClick={onClose}
                                    className="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300"
                                >
                                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                </button>
                            )}
                        </div>
                        
                        {/* Progress indicator */}
                        <div className="mt-4">
                            <div className="flex items-center justify-between">
                                <div className={`flex items-center ${currentStep === 'file-selection' ? 'text-blue-600 dark:text-blue-400' : 'text-gray-600 dark:text-gray-400'}`}>
                                    <div className={`w-8 h-8 rounded-full flex items-center justify-center border-2 ${
                                        currentStep === 'file-selection' 
                                            ? 'border-blue-600 bg-blue-600 text-white dark:border-blue-400 dark:bg-blue-400' 
                                            : 'border-gray-300 dark:border-gray-600'
                                    }`}>
                                        1
                                    </div>
                                    <span className="ml-2 text-sm font-medium">Select File</span>
                                </div>
                                
                                <div className="flex-1 h-0.5 mx-4 bg-gray-200 dark:bg-gray-700" />
                                
                                <div className={`flex items-center ${
                                    currentStep === 'format-options' ? 'text-blue-600 dark:text-blue-400' : 
                                    currentStep === 'data-preview' || currentStep === 'importing' ? 'text-gray-600 dark:text-gray-400' : 
                                    'text-gray-400 dark:text-gray-500'
                                }`}>
                                    <div className={`w-8 h-8 rounded-full flex items-center justify-center border-2 ${
                                        currentStep === 'format-options' 
                                            ? 'border-blue-600 bg-blue-600 text-white dark:border-blue-400 dark:bg-blue-400' 
                                            : currentStep === 'data-preview' || currentStep === 'importing'
                                            ? 'border-gray-300 dark:border-gray-600'
                                            : 'border-gray-300 dark:border-gray-600'
                                    }`}>
                                        2
                                    </div>
                                    <span className="ml-2 text-sm font-medium">Configure Options</span>
                                </div>
                                
                                <div className="flex-1 h-0.5 mx-4 bg-gray-200 dark:bg-gray-700" />
                                
                                <div className={`flex items-center ${
                                    currentStep === 'data-preview' || currentStep === 'importing' ? 'text-blue-600 dark:text-blue-400' : 
                                    'text-gray-400 dark:text-gray-500'
                                }`}>
                                    <div className={`w-8 h-8 rounded-full flex items-center justify-center border-2 ${
                                        currentStep === 'data-preview' || currentStep === 'importing'
                                            ? 'border-blue-600 bg-blue-600 text-white dark:border-blue-400 dark:bg-blue-400' 
                                            : 'border-gray-300 dark:border-gray-600'
                                    }`}>
                                        3
                                    </div>
                                    <span className="ml-2 text-sm font-medium">Preview & Import</span>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Content */}
                    <div className="px-6 py-4" style={{ minHeight: '400px', maxHeight: '60vh', overflowY: 'auto' }}>
                        {error && (
                            <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 rounded-lg">
                                {error}
                            </div>
                        )}

                        {currentStep === 'file-selection' && (
                            <FileSelector 
                                onFileSelect={handleFileSelect}
                                isLoading={isLoading}
                            />
                        )}

                        {currentStep === 'format-options' && fileInfo && (
                            <FormatOptions
                                fileInfo={fileInfo}
                                options={importOptions}
                                onChange={setImportOptions}
                            />
                        )}

                        {currentStep === 'data-preview' && preview && (
                            <DataPreview
                                preview={preview}
                                options={importOptions}
                                onChange={setImportOptions}
                            />
                        )}

                        {currentStep === 'importing' && (
                            <ImportProgress progress={importProgress} />
                        )}
                    </div>

                    {/* Footer */}
                    {currentStep !== 'importing' && (
                        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700">
                            <div className="flex justify-between">
                                <button
                                    onClick={handleBack}
                                    disabled={currentStep === 'file-selection' || isLoading}
                                    className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    Back
                                </button>
                                
                                <div className="flex gap-2">
                                    <button
                                        onClick={onClose}
                                        className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600"
                                    >
                                        Cancel
                                    </button>
                                    
                                    {currentStep === 'format-options' && (
                                        <button
                                            onClick={handleOptionsNext}
                                            disabled={isLoading}
                                            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                                        >
                                            {isLoading ? 'Loading...' : 'Next'}
                                        </button>
                                    )}
                                    
                                    {currentStep === 'data-preview' && (
                                        <button
                                            onClick={handleImport}
                                            disabled={isLoading}
                                            className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                                        >
                                            Import Data
                                        </button>
                                    )}
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};