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

package main

import (
	"fmt"

	"github.com/bitjungle/gopca/pkg/transform"
	"github.com/bitjungle/gopca/pkg/types"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// TransformationType represents the type of transformation
type TransformationType string

const (
	TransformLog         TransformationType = "log"
	TransformSqrt        TransformationType = "sqrt"
	TransformSquare      TransformationType = "square"
	TransformStandardize TransformationType = "standardize"
	TransformMinMax      TransformationType = "minmax"
	TransformBin         TransformationType = "bin"
	TransformOneHot      TransformationType = "onehot"
)

// TransformOptions represents options for data transformation
type TransformOptions struct {
	Type     TransformationType `json:"type"`
	Columns  []string           `json:"columns"`
	BinCount int                `json:"binCount,omitempty"` // For binning
	MinValue float64            `json:"minValue,omitempty"` // For min-max scaling
	MaxValue float64            `json:"maxValue,omitempty"` // For min-max scaling

	// RemoveOriginal drops the source column after one-hot encoding it.
	// Absent from the JSON payload means false, i.e. the column is kept --
	// see transform.Options for why the flag is spelled this way round.
	RemoveOriginal bool `json:"removeOriginal,omitempty"` // For one-hot encoding
}

// TransformationResult represents the result of a transformation
type TransformationResult struct {
	Success            bool      `json:"success"`
	TransformedColumns []string  `json:"transformedColumns"`
	NewColumns         []string  `json:"newColumns,omitempty"`
	Messages           []string  `json:"messages"`
	Data               *FileData `json:"data"`
}

// ApplyTransformation applies a transformation to the data with undo support
func (a *App) ApplyTransformation(data *FileData, options TransformOptions) (*TransformationResult, error) {
	// Use the command pattern for undo support
	cmd := NewTransformCommand(a, data, options)
	if err := a.history.Execute(cmd, data); err != nil {
		return nil, fmt.Errorf("apply transformation: %w", err)
	}
	a.currentData = data // Store the current data
	a.markDirty()
	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "undo-redo-state-changed", a.GetUndoRedoState())
	}

	// Return the transformation result
	return cmd.result, nil
}

// applyTransformationInternal delegates to pkg/transform and maps the result
// back into the FileData-based TransformationResult expected by the Wails layer.
func (a *App) applyTransformationInternal(data *FileData, options TransformOptions) (*TransformationResult, error) {
	if data == nil || len(data.Data) == 0 {
		return nil, fmt.Errorf("no data to transform")
	}

	in := transform.Input{
		Data:               data.Data,
		Headers:            data.Headers,
		ColumnTypes:        data.ColumnTypes,
		CategoricalColumns: data.CategoricalColumns,
		Rows:               data.Rows,
		Columns:            data.Columns,
	}
	opts := transform.Options{
		Type:     transform.Type(options.Type),
		Columns:  options.Columns,
		BinCount: options.BinCount,
		MinValue: options.MinValue,
		MaxValue: options.MaxValue,

		RemoveOriginal: options.RemoveOriginal,
	}

	res, err := transform.Apply(in, opts)
	if err != nil {
		return nil, err
	}

	newData := &FileData{
		Headers:              res.Headers,
		Data:                 res.Data,
		Rows:                 data.Rows,
		Columns:              res.Columns,
		CategoricalColumns:   res.CategoricalColumns,
		NumericTargetColumns: make(map[string][]types.JSONFloat64),
		ColumnTypes:          res.ColumnTypes,
	}
	if data.RowNames != nil {
		newData.RowNames = make([]string, len(data.RowNames))
		copy(newData.RowNames, data.RowNames)
	}
	// Carry over any numeric target columns untouched.
	for k, v := range data.NumericTargetColumns {
		newData.NumericTargetColumns[k] = v
	}

	return &TransformationResult{
		Success:            true,
		TransformedColumns: res.TransformedColumns,
		NewColumns:         res.NewColumns,
		Messages:           res.Messages,
		Data:               newData,
	}, nil
}

// GetTransformableColumns returns columns eligible for the given transformation type.
func (a *App) GetTransformableColumns(data *FileData, transformType TransformationType) []string {
	in := transform.Input{
		Data:        data.Data,
		Headers:     data.Headers,
		ColumnTypes: data.ColumnTypes,
		Rows:        data.Rows,
		Columns:     data.Columns,
	}
	return transform.GetTransformableColumns(in, transform.Type(transformType))
}
