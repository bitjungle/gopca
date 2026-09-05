// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

import React, { useState, useEffect } from 'react';
import { ExecuteFilterRows, PreviewFilter } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';

type FileData = main.FileData;

interface FilterRowsDialogProps {
    isOpen: boolean;
    onClose: () => void;
    fileData: FileData;
    onFilterComplete: (data: FileData) => void;
}

// Mirrors the FilterOperator constants in cmd/gocsv/filter.go.
const operators = [
    { value: 'equals', label: 'is', needsValue: true },
    { value: 'not_equals', label: 'is not', needsValue: true },
    { value: 'contains', label: 'contains', needsValue: true },
    { value: 'not_contains', label: 'does not contain', needsValue: true },
    { value: 'greater', label: 'is greater than', needsValue: true },
    { value: 'greater_equal', label: 'is at least', needsValue: true },
    { value: 'less', label: 'is less than', needsValue: true },
    { value: 'less_equal', label: 'is at most', needsValue: true },
    { value: 'is_empty', label: 'is empty', needsValue: false },
    { value: 'is_not_empty', label: 'is not empty', needsValue: false }
];

export const FilterRowsDialog: React.FC<FilterRowsDialogProps> = ({
    isOpen,
    onClose,
    fileData,
    onFilterComplete
}) => {
    const [column, setColumn] = useState('');
    const [operator, setOperator] = useState('equals');
    const [value, setValue] = useState('');
    const [mode, setMode] = useState<'keep' | 'remove'>('keep');
    const [preview, setPreview] = useState<main.FilterPreview | null>(null);
    const [isApplying, setIsApplying] = useState(false);

    const headers: string[] = fileData?.headers || [];
    const needsValue = operators.find((o) => o.value === operator)?.needsValue ?? true;

    // Start on the first column each time the dialog opens, and clear anything
    // left from last time. The dialog is not unmounted when closed, so without
    // this the previous filter would still be sitting there.
    useEffect(() => {
        if (isOpen) {
            setColumn(headers[0] || '');
            setOperator('equals');
            setValue('');
            setMode('keep');
            setPreview(null);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isOpen]);

    // Ask the backend what the filter would do, on every change.
    //
    // The count is what makes the operation safe to use: a filter that would
    // empty the table, or one that matches nothing, is worth seeing before it
    // is applied rather than after.
    useEffect(() => {
        if (!isOpen || !fileData || !column) {
            return;
        }
        let cancelled = false;
        const run = async () => {
            try {
                const result = await PreviewFilter(fileData, {
                    column,
                    operator,
                    value: needsValue ? value : '',
                    mode
                } as main.FilterCondition);
                if (!cancelled) {
                    setPreview(result);
                }
            } catch (err) {
                console.error('Error previewing filter:', err);
            }
        };
        void run();
        return () => {
            cancelled = true;
        };
    }, [isOpen, fileData, column, operator, value, mode, needsValue]);

    if (!isOpen) {
        return null;
    }

    const canApply = preview !== null && !preview.error && !isApplying;

    const handleApply = async () => {
        setIsApplying(true);
        try {
            const updated = await ExecuteFilterRows(fileData, {
                column,
                operator,
                value: needsValue ? value : '',
                mode
            } as main.FilterCondition);
            onFilterComplete(updated);
            onClose();
        } catch (err) {
            console.error('Error filtering rows:', err);
        } finally {
            setIsApplying(false);
        }
    };

    const selectClass =
        'w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg ' +
        'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100';

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-lg mx-4">
                <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
                    <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Filter Rows</h2>
                </div>

                <div className="px-6 py-4 space-y-4">
                    <div className="grid grid-cols-3 gap-3">
                        <div className="col-span-1">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                Column
                            </label>
                            <select className={selectClass} value={column} onChange={(e) => setColumn(e.target.value)}>
                                {headers.map((h) => (
                                    <option key={h} value={h}>{h}</option>
                                ))}
                            </select>
                        </div>
                        <div className="col-span-1">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                Condition
                            </label>
                            <select className={selectClass} value={operator} onChange={(e) => setOperator(e.target.value)}>
                                {operators.map((o) => (
                                    <option key={o.value} value={o.value}>{o.label}</option>
                                ))}
                            </select>
                        </div>
                        <div className="col-span-1">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                Value
                            </label>
                            <input
                                type="text"
                                className={`${selectClass} ${needsValue ? '' : 'opacity-40'}`}
                                value={needsValue ? value : ''}
                                disabled={!needsValue}
                                onChange={(e) => setValue(e.target.value)}
                            />
                        </div>
                    </div>

                    <fieldset>
                        <legend className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            Matching rows
                        </legend>
                        <div className="flex gap-4">
                            <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 cursor-pointer">
                                <input
                                    type="radio"
                                    name="filter-mode"
                                    checked={mode === 'keep'}
                                    onChange={() => setMode('keep')}
                                />
                                Keep them
                            </label>
                            <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 cursor-pointer">
                                <input
                                    type="radio"
                                    name="filter-mode"
                                    checked={mode === 'remove'}
                                    onChange={() => setMode('remove')}
                                />
                                Remove them
                            </label>
                        </div>
                    </fieldset>

                    {/* What the filter would do, before it does it. */}
                    <div className="text-sm rounded-lg p-3 bg-gray-50 dark:bg-gray-700/50">
                        {preview?.error ? (
                            <span className="text-red-600 dark:text-red-400">{preview.error}</span>
                        ) : preview ? (
                            <>
                                <span className="text-gray-700 dark:text-gray-300">
                                    {preview.matched} of {preview.total} rows match.{' '}
                                    <strong>{preview.remaining}</strong> would remain.
                                </span>
                                {preview.remaining === 0 && (
                                    <span className="block mt-1 text-amber-600 dark:text-amber-400">
                                        This would leave no data. Undo is available, but check the
                                        condition first.
                                    </span>
                                )}
                                {preview.remaining === preview.total && (
                                    <span className="block mt-1 text-gray-500 dark:text-gray-400">
                                        This would change nothing.
                                    </span>
                                )}
                            </>
                        ) : (
                            <span className="text-gray-500 dark:text-gray-400">…</span>
                        )}
                    </div>

                    <p className="text-xs text-gray-500 dark:text-gray-400">
                        Empty cells match only “is empty”. They are never picked up by a negative
                        condition, so a filter about one column will not remove rows for having no
                        value in it.
                    </p>
                </div>

                <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleApply}
                        disabled={!canApply}
                        className="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-40 disabled:cursor-not-allowed"
                    >
                        {isApplying ? 'Filtering…' : 'Apply Filter'}
                    </button>
                </div>
            </div>
        </div>
    );
};
