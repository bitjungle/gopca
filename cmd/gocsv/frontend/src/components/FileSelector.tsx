import React, { useState, useCallback } from 'react';
import { SelectFileForImport } from '../../wailsjs/go/main/App';

interface FileSelectorProps {
    onFileSelect: (filePath: string) => void;
    isLoading: boolean;
}

export const FileSelector: React.FC<FileSelectorProps> = ({ onFileSelect, isLoading }) => {
    const [isDragging, setIsDragging] = useState(false);
    const [recentFiles] = useState<string[]>([
        // TODO: Implement recent files storage
    ]);

    const handleBrowse = async () => {
        try {
            const filePath = await SelectFileForImport();
            if (filePath) {
                onFileSelect(filePath);
            }
        } catch (err) {
            console.error('Error selecting file:', err);
        }
    };

    const handleDragOver = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
    }, []);

    const handleDragEnter = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(true);
    }, []);

    const handleDragLeave = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(false);
    }, []);

    const handleDrop = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragging(false);
        
        // Note: Due to Wails limitations, we can't directly handle dropped files
        // Show a message to use the browse button instead
        alert('Please use the "Browse Files" button to select files. Drag and drop is not supported in the file dialog.');
    }, []);

    return (
        <div className="space-y-6">
            {/* Drag and drop area */}
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
                <svg className={`mx-auto h-12 w-12 ${isDragging ? 'text-blue-500' : 'text-gray-400'} transition-colors`} stroke="currentColor" fill="none" viewBox="0 0 48 48" aria-hidden="true">
                    <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
                <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                    {isDragging ? (
                        <span className="font-medium text-blue-600 dark:text-blue-400">
                            Drop files here
                        </span>
                    ) : (
                        <span>
                            Drag and drop files here, or{' '}
                            <button
                                onClick={handleBrowse}
                                disabled={isLoading}
                                className="font-medium text-blue-600 dark:text-blue-400 hover:text-blue-500 dark:hover:text-blue-300"
                            >
                                browse
                            </button>
                        </span>
                    )}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                    Supports CSV, TSV, Excel (.xlsx, .xls), and JSON files
                </p>
            </div>

            {/* Browse button */}
            <div className="text-center">
                <button
                    onClick={handleBrowse}
                    disabled={isLoading}
                    className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                    {isLoading ? 'Loading...' : 'Browse Files'}
                </button>
            </div>

            {/* Recent files */}
            {recentFiles.length > 0 && (
                <div>
                    <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                        Recent Files
                    </h3>
                    <div className="space-y-1">
                        {recentFiles.map((file, index) => (
                            <button
                                key={index}
                                onClick={() => onFileSelect(file)}
                                disabled={isLoading}
                                className="w-full text-left px-3 py-2 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors disabled:opacity-50"
                            >
                                <div className="flex items-center gap-2">
                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                    </svg>
                                    <span className="truncate">{file}</span>
                                </div>
                            </button>
                        ))}
                    </div>
                </div>
            )}

            {/* File format info */}
            <div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
                <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Supported File Formats
                </h3>
                <ul className="space-y-1 text-xs text-gray-600 dark:text-gray-400">
                    <li className="flex items-center gap-2">
                        <span className="font-mono bg-gray-200 dark:bg-gray-600 px-1 rounded">.csv</span>
                        Comma-separated values
                    </li>
                    <li className="flex items-center gap-2">
                        <span className="font-mono bg-gray-200 dark:bg-gray-600 px-1 rounded">.tsv</span>
                        Tab-separated values
                    </li>
                    <li className="flex items-center gap-2">
                        <span className="font-mono bg-gray-200 dark:bg-gray-600 px-1 rounded">.xlsx/.xls</span>
                        Microsoft Excel
                    </li>
                    <li className="flex items-center gap-2">
                        <span className="font-mono bg-gray-200 dark:bg-gray-600 px-1 rounded">.json</span>
                        JSON arrays (coming soon)
                    </li>
                </ul>
            </div>
        </div>
    );
};