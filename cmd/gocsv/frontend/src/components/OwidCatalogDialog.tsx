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

import React, { useState, useEffect, useRef } from 'react';
import { SearchOWIDDatasets, LoadOWIDDataset } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';

type FileData = main.FileData;
type OWIDDataset = main.OWIDDataset;

interface OwidCatalogDialogProps {
    isOpen: boolean;
    onClose: () => void;
    onDataLoaded: (data: FileData) => void;
}

export const OwidCatalogDialog: React.FC<OwidCatalogDialogProps> = ({
    isOpen,
    onClose,
    onDataLoaded,
}) => {
    const [query, setQuery] = useState('');
    const [datasets, setDatasets] = useState<OWIDDataset[]>([]);
    const [selected, setSelected] = useState<OWIDDataset | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [isSearching, setIsSearching] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const searchRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    // Load full catalog on open; re-search when query changes (debounced).
    useEffect(() => {
        if (!isOpen) return;

        if (searchRef.current) clearTimeout(searchRef.current);
        searchRef.current = setTimeout(async () => {
            setIsSearching(true);
            try {
                const results = await SearchOWIDDatasets(query);
                setDatasets(results ?? []);
                // Clear selection if it's no longer in results
                if (selected && results && !results.find((d: OWIDDataset) => d.path === selected.path)) {
                    setSelected(null);
                }
            } catch (e) {
                setError(`Search failed: ${e}`);
            } finally {
                setIsSearching(false);
            }
        }, 200);

        return () => {
            if (searchRef.current) clearTimeout(searchRef.current);
        };
    }, [query, isOpen]);

    const handleImport = async () => {
        if (!selected) return;
        setIsLoading(true);
        setError(null);
        try {
            const data = await LoadOWIDDataset(selected.path);
            onDataLoaded(data);
            onClose();
        } catch (e) {
            setError(`Failed to load dataset: ${e}`);
        } finally {
            setIsLoading(false);
        }
    };

    const handleClose = () => {
        if (!isLoading) {
            setQuery('');
            setSelected(null);
            setError(null);
            onClose();
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-50 overflow-y-auto">
            <div className="flex items-center justify-center min-h-screen px-4">
                {/* Backdrop */}
                <div
                    className="fixed inset-0 bg-gray-500 bg-opacity-75 dark:bg-gray-900 dark:bg-opacity-80 transition-opacity"
                    onClick={handleClose}
                />

                {/* Dialog panel */}
                <div className="relative inline-block bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-3xl p-0 overflow-hidden">
                    {/* Header */}
                    <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
                        <div>
                            <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                                Browse OWID Datasets
                            </h2>
                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                                Our World in Data — curated datasets for analysis
                            </p>
                        </div>
                        <button
                            onClick={handleClose}
                            disabled={isLoading}
                            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 disabled:opacity-50"
                            aria-label="Close"
                        >
                            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </button>
                    </div>

                    {/* Search */}
                    <div className="px-6 py-3 border-b border-gray-200 dark:border-gray-700">
                        <div className="relative">
                            <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                            </svg>
                            <input
                                type="text"
                                value={query}
                                onChange={e => setQuery(e.target.value)}
                                placeholder="Search by topic, dataset name…"
                                className="w-full pl-9 pr-4 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                                autoFocus
                            />
                        </div>
                    </div>

                    {/* Dataset list */}
                    <div className="overflow-y-auto" style={{ maxHeight: '340px' }}>
                        {isSearching ? (
                            <div className="flex items-center justify-center py-10 text-gray-400 text-sm">
                                Searching…
                            </div>
                        ) : datasets.length === 0 ? (
                            <div className="flex items-center justify-center py-10 text-gray-400 text-sm">
                                No datasets match your search.
                            </div>
                        ) : (
                            <table className="w-full text-sm">
                                <thead className="sticky top-0 bg-gray-50 dark:bg-gray-700/80 text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide">
                                    <tr>
                                        <th className="px-4 py-2 text-left font-medium">Dataset</th>
                                        <th className="px-4 py-2 text-left font-medium">Namespace</th>
                                        <th className="px-4 py-2 text-left font-medium">Version</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-gray-100 dark:divide-gray-700">
                                    {datasets.map(d => (
                                        <tr
                                            key={d.path}
                                            onClick={() => setSelected(d)}
                                            className={`cursor-pointer transition-colors ${
                                                selected?.path === d.path
                                                    ? 'bg-blue-50 dark:bg-blue-900/30'
                                                    : 'hover:bg-gray-50 dark:hover:bg-gray-700/50'
                                            }`}
                                        >
                                            <td className="px-4 py-3">
                                                <div className="font-medium text-gray-900 dark:text-gray-100">{d.title}</div>
                                                <div className="text-xs text-gray-500 dark:text-gray-400 mt-0.5 line-clamp-1">{d.description}</div>
                                            </td>
                                            <td className="px-4 py-3 text-gray-600 dark:text-gray-300 whitespace-nowrap">{d.namespace}</td>
                                            <td className="px-4 py-3 text-gray-600 dark:text-gray-300 whitespace-nowrap font-mono text-xs">{d.version}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        )}
                    </div>

                    {/* Selected dataset info */}
                    {selected && (
                        <div className="px-6 py-3 bg-blue-50 dark:bg-blue-900/20 border-t border-blue-100 dark:border-blue-800 text-sm">
                            <span className="font-medium text-blue-800 dark:text-blue-200">Selected:</span>{' '}
                            <span className="text-blue-700 dark:text-blue-300">{selected.title}</span>
                            <span className="text-blue-500 dark:text-blue-400 ml-2 text-xs">({selected.path})</span>
                        </div>
                    )}

                    {/* Error */}
                    {error && (
                        <div className="px-6 py-3 bg-red-50 dark:bg-red-900/20 border-t border-red-100 dark:border-red-800 text-sm text-red-700 dark:text-red-300">
                            {error}
                        </div>
                    )}

                    {/* Loading bar */}
                    {isLoading && (
                        <div className="px-6 py-2 border-t border-gray-100 dark:border-gray-700">
                            <div className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                                <svg className="animate-spin w-4 h-4 text-blue-500" fill="none" viewBox="0 0 24 24">
                                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
                                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z"/>
                                </svg>
                                Downloading dataset…
                            </div>
                        </div>
                    )}

                    {/* Footer */}
                    <div className="flex items-center justify-between px-6 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
                        <span className="text-xs text-gray-400 dark:text-gray-500">
                            Data from{' '}
                            <span className="font-medium">Our World in Data</span>
                            {' '}· CC BY 4.0
                        </span>
                        <div className="flex gap-3">
                            <button
                                onClick={handleClose}
                                disabled={isLoading}
                                className="px-4 py-2 text-sm text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleImport}
                                disabled={!selected || isLoading}
                                className="px-4 py-2 text-sm text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                            >
                                {isLoading ? 'Importing…' : 'Import Dataset'}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};
