// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

import React from 'react';
import { ErrorBoundary, ErrorAlert } from '@gopca/ui-components';
import { DataTable, SelectionTable, MatrixIllustration, HelpWrapper } from '../index';
import { useFileDataContext } from '../../contexts/FileDataContext';
import { usePCAContext } from '../../contexts/PCAContext';
import { useGoCSVContext } from '../../contexts/GoCSVContext';
import { logger } from '../../utils/logger';

/**
 * Step 1: Load Data — file picker, GoCSV button, matrix illustration, sample
 * datasets — plus the loaded data table and file error alert.
 *
 * All state comes from FileDataContext, PCAContext, and GoCSVContext.
 * No props needed.
 */
export function DataLoadSection() {
    const {
        fileData, fileError, datasetId, loading: fileLoading, clearFileError,
    } = useFileDataContext();
    const {
        loading, excludedRows,
        handleLoadDataset, handleNativeFileSelectWithReset,
        handleRowSelectionChange, handleColumnSelectionChange,
    } = usePCAContext();
    const { goCSVStatus, isCheckingGoCSV, handleGoCSVAction } = useGoCSVContext();

    return (
        <>
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
                                {fileLoading ? 'Loading...' : 'Choose File'}
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
        </>
    );
}
