package main

import (
	"fmt"
	"sort"
	"strings"
	
	"github.com/bitjungle/gopca/pkg/types"
)

// Command interface defines operations that can be undone/redone
type Command interface {
	Execute() error
	Undo() error
	GetDescription() string
}

// CommandHistory manages the command history for undo/redo
type CommandHistory struct {
	commands []Command
	current  int
	maxSize  int
}

// NewCommandHistory creates a new command history with a maximum size
func NewCommandHistory(maxSize int) *CommandHistory {
	if maxSize <= 0 {
		maxSize = 100 // Default max history size
	}
	return &CommandHistory{
		commands: make([]Command, 0),
		current:  -1,
		maxSize:  maxSize,
	}
}

// Execute adds a command to the history and executes it
func (h *CommandHistory) Execute(cmd Command) error {
	// Execute the command first
	if err := cmd.Execute(); err != nil {
		return err
	}

	// Remove any commands after the current position (for redo)
	if h.current < len(h.commands)-1 {
		h.commands = h.commands[:h.current+1]
	}

	// Add the new command
	h.commands = append(h.commands, cmd)
	h.current++

	// Limit history size
	if len(h.commands) > h.maxSize {
		h.commands = h.commands[1:]
		h.current--
	}

	return nil
}

// Undo undoes the last command
func (h *CommandHistory) Undo() error {
	if !h.CanUndo() {
		return fmt.Errorf("nothing to undo")
	}

	cmd := h.commands[h.current]
	if err := cmd.Undo(); err != nil {
		return fmt.Errorf("undo failed: %w", err)
	}

	h.current--
	return nil
}

// Redo redoes the next command
func (h *CommandHistory) Redo() error {
	if !h.CanRedo() {
		return fmt.Errorf("nothing to redo")
	}

	h.current++
	cmd := h.commands[h.current]
	if err := cmd.Execute(); err != nil {
		h.current-- // Revert on failure
		return fmt.Errorf("redo failed: %w", err)
	}

	return nil
}

// CanUndo returns true if there are commands to undo
func (h *CommandHistory) CanUndo() bool {
	return h.current >= 0
}

// CanRedo returns true if there are commands to redo
func (h *CommandHistory) CanRedo() bool {
	return h.current < len(h.commands)-1
}

// GetHistory returns the command history with current position
func (h *CommandHistory) GetHistory() ([]string, int) {
	descriptions := make([]string, len(h.commands))
	for i, cmd := range h.commands {
		descriptions[i] = cmd.GetDescription()
	}
	return descriptions, h.current
}

// Clear clears the command history
func (h *CommandHistory) Clear() {
	h.commands = make([]Command, 0)
	h.current = -1
}

// CellEditCommand represents an edit to a single cell
type CellEditCommand struct {
	app      *App
	data     *FileData
	row      int
	col      int
	oldValue string
	newValue string
}

// NewCellEditCommand creates a new cell edit command
func NewCellEditCommand(app *App, data *FileData, row, col int, oldValue, newValue string) *CellEditCommand {
	return &CellEditCommand{
		app:      app,
		data:     data,
		row:      row,
		col:      col,
		oldValue: oldValue,
		newValue: newValue,
	}
}

// Execute applies the cell edit
func (c *CellEditCommand) Execute() error {
	if c.row >= len(c.data.Data) || c.col >= len(c.data.Data[c.row]) {
		return fmt.Errorf("invalid cell position: row=%d, col=%d", c.row, c.col)
	}
	c.data.Data[c.row][c.col] = c.newValue
	return nil
}

// Undo reverts the cell edit
func (c *CellEditCommand) Undo() error {
	if c.row >= len(c.data.Data) || c.col >= len(c.data.Data[c.row]) {
		return fmt.Errorf("invalid cell position: row=%d, col=%d", c.row, c.col)
	}
	c.data.Data[c.row][c.col] = c.oldValue
	return nil
}

// GetDescription returns a description of the command
func (c *CellEditCommand) GetDescription() string {
	return fmt.Sprintf("Edit cell [%d,%d]: '%s' → '%s'", c.row+1, c.col+1, c.oldValue, c.newValue)
}

// HeaderEditCommand represents an edit to a column header
type HeaderEditCommand struct {
	app      *App
	data     *FileData
	col      int
	oldValue string
	newValue string
}

// NewHeaderEditCommand creates a new header edit command
func NewHeaderEditCommand(app *App, data *FileData, col int, oldValue, newValue string) *HeaderEditCommand {
	return &HeaderEditCommand{
		app:      app,
		data:     data,
		col:      col,
		oldValue: oldValue,
		newValue: newValue,
	}
}

// Execute applies the header edit
func (c *HeaderEditCommand) Execute() error {
	if c.col >= len(c.data.Headers) {
		return fmt.Errorf("invalid column index: %d", c.col)
	}
	c.data.Headers[c.col] = c.newValue
	
	// Update column types if needed
	if c.data.ColumnTypes != nil {
		if colType, exists := c.data.ColumnTypes[c.oldValue]; exists {
			delete(c.data.ColumnTypes, c.oldValue)
			c.data.ColumnTypes[c.newValue] = colType
		}
	}
	
	// Update categorical columns if needed
	if c.data.CategoricalColumns != nil {
		if values, exists := c.data.CategoricalColumns[c.oldValue]; exists {
			delete(c.data.CategoricalColumns, c.oldValue)
			c.data.CategoricalColumns[c.newValue] = values
		}
	}
	
	// Update numeric target columns if needed
	if c.data.NumericTargetColumns != nil {
		if values, exists := c.data.NumericTargetColumns[c.oldValue]; exists {
			delete(c.data.NumericTargetColumns, c.oldValue)
			c.data.NumericTargetColumns[c.newValue] = values
		}
	}
	
	return nil
}

// Undo reverts the header edit
func (c *HeaderEditCommand) Undo() error {
	if c.col >= len(c.data.Headers) {
		return fmt.Errorf("invalid column index: %d", c.col)
	}
	c.data.Headers[c.col] = c.oldValue
	
	// Revert column types if needed
	if c.data.ColumnTypes != nil {
		if colType, exists := c.data.ColumnTypes[c.newValue]; exists {
			delete(c.data.ColumnTypes, c.newValue)
			c.data.ColumnTypes[c.oldValue] = colType
		}
	}
	
	// Revert categorical columns if needed
	if c.data.CategoricalColumns != nil {
		if values, exists := c.data.CategoricalColumns[c.newValue]; exists {
			delete(c.data.CategoricalColumns, c.newValue)
			c.data.CategoricalColumns[c.oldValue] = values
		}
	}
	
	// Revert numeric target columns if needed
	if c.data.NumericTargetColumns != nil {
		if values, exists := c.data.NumericTargetColumns[c.newValue]; exists {
			delete(c.data.NumericTargetColumns, c.newValue)
			c.data.NumericTargetColumns[c.oldValue] = values
		}
	}
	
	return nil
}

// GetDescription returns a description of the command
func (c *HeaderEditCommand) GetDescription() string {
	return fmt.Sprintf("Edit header '%s' → '%s'", c.oldValue, c.newValue)
}

// FillMissingValuesCommand represents a missing value fill operation
type FillMissingValuesCommand struct {
	app         *App
	data        *FileData
	oldData     *FileData
	strategy    string
	column      string
	customValue string
}

// NewFillMissingValuesCommand creates a new fill missing values command
func NewFillMissingValuesCommand(app *App, data *FileData, strategy, column, customValue string) *FillMissingValuesCommand {
	// Create a deep copy of the current data for undo
	oldData := &FileData{
		Headers:              data.Headers,
		RowNames:             data.RowNames,
		Data:                 make([][]string, len(data.Data)),
		Rows:                 data.Rows,
		Columns:              data.Columns,
		CategoricalColumns:   data.CategoricalColumns,
		NumericTargetColumns: data.NumericTargetColumns,
		ColumnTypes:          data.ColumnTypes,
	}
	
	// Deep copy the data
	for i := range data.Data {
		oldData.Data[i] = make([]string, len(data.Data[i]))
		copy(oldData.Data[i], data.Data[i])
	}
	
	return &FillMissingValuesCommand{
		app:         app,
		data:        data,
		oldData:     oldData,
		strategy:    strategy,
		column:      column,
		customValue: customValue,
	}
}

// Execute applies the fill operation
func (c *FillMissingValuesCommand) Execute() error {
	request := FillMissingValuesRequest{
		Strategy: c.strategy,
		Column:   c.column,
		Value:    c.customValue,
	}
	
	newData, err := c.app.FillMissingValues(c.data, request)
	if err != nil {
		return err
	}
	
	// Update the data in place
	c.data.Data = newData.Data
	return nil
}

// Undo reverts the fill operation
func (c *FillMissingValuesCommand) Undo() error {
	// Restore the old data
	c.data.Data = make([][]string, len(c.oldData.Data))
	for i := range c.oldData.Data {
		c.data.Data[i] = make([]string, len(c.oldData.Data[i]))
		copy(c.data.Data[i], c.oldData.Data[i])
	}
	return nil
}

// GetDescription returns a description of the command
func (c *FillMissingValuesCommand) GetDescription() string {
	if c.column == "" {
		return fmt.Sprintf("Fill missing values (all columns) with %s", c.strategy)
	}
	return fmt.Sprintf("Fill missing values in '%s' with %s", c.column, c.strategy)
}

// BatchCommand represents multiple commands executed as one
type BatchCommand struct {
	commands    []Command
	description string
}

// NewBatchCommand creates a new batch command
func NewBatchCommand(description string, commands ...Command) *BatchCommand {
	return &BatchCommand{
		commands:    commands,
		description: description,
	}
}

// Execute executes all commands in the batch
func (c *BatchCommand) Execute() error {
	for i, cmd := range c.commands {
		if err := cmd.Execute(); err != nil {
			// Undo previously executed commands
			for j := i - 1; j >= 0; j-- {
				c.commands[j].Undo() // Ignore undo errors
			}
			return fmt.Errorf("batch command failed at step %d: %w", i+1, err)
		}
	}
	return nil
}

// Undo undoes all commands in reverse order
func (c *BatchCommand) Undo() error {
	for i := len(c.commands) - 1; i >= 0; i-- {
		if err := c.commands[i].Undo(); err != nil {
			return fmt.Errorf("batch undo failed at step %d: %w", len(c.commands)-i, err)
		}
	}
	return nil
}

// GetDescription returns a description of the batch command
func (c *BatchCommand) GetDescription() string {
	return c.description
}

// DeleteRowsCommand represents deletion of multiple rows
type DeleteRowsCommand struct {
	app        *App
	data       *FileData
	rowIndices []int
	oldRows    [][]string
	oldRowNames []string
}

// NewDeleteRowsCommand creates a new delete rows command
func NewDeleteRowsCommand(app *App, data *FileData, rowIndices []int) *DeleteRowsCommand {
	// Sort indices in descending order for easier deletion
	sortedIndices := make([]int, len(rowIndices))
	copy(sortedIndices, rowIndices)
	sort.Sort(sort.Reverse(sort.IntSlice(sortedIndices)))
	
	// Save the rows that will be deleted (in original order for restoration)
	// Create a map to preserve original order
	rowMap := make(map[int][]string)
	rowNameMap := make(map[int]string)
	
	for _, idx := range rowIndices {
		if idx < len(data.Data) {
			rowData := make([]string, len(data.Data[idx]))
			copy(rowData, data.Data[idx])
			rowMap[idx] = rowData
			
			if data.RowNames != nil && idx < len(data.RowNames) {
				rowNameMap[idx] = data.RowNames[idx]
			}
		}
	}
	
	// Store in sorted order for undo
	oldRows := make([][]string, 0, len(rowIndices))
	oldRowNames := make([]string, 0)
	sortedOriginal := make([]int, len(rowIndices))
	copy(sortedOriginal, rowIndices)
	sort.Ints(sortedOriginal)
	
	for _, idx := range sortedOriginal {
		if row, exists := rowMap[idx]; exists {
			oldRows = append(oldRows, row)
			if name, hasName := rowNameMap[idx]; hasName {
				oldRowNames = append(oldRowNames, name)
			}
		}
	}
	
	return &DeleteRowsCommand{
		app:        app,
		data:       data,
		rowIndices: sortedIndices,
		oldRows:    oldRows,
		oldRowNames: oldRowNames,
	}
}

// Execute deletes the rows
func (c *DeleteRowsCommand) Execute() error {
	// Delete rows in descending order to maintain indices
	for _, idx := range c.rowIndices {
		if idx >= 0 && idx < len(c.data.Data) {
			c.data.Data = append(c.data.Data[:idx], c.data.Data[idx+1:]...)
			
			// Update row names if present
			if c.data.RowNames != nil && idx < len(c.data.RowNames) {
				c.data.RowNames = append(c.data.RowNames[:idx], c.data.RowNames[idx+1:]...)
			}
		}
	}
	
	// Update row count
	c.data.Rows = len(c.data.Data)
	
	return nil
}

// Undo restores the deleted rows
func (c *DeleteRowsCommand) Undo() error {
	// Get sorted indices for restoration
	indices := make([]int, len(c.rowIndices))
	copy(indices, c.rowIndices)
	sort.Ints(indices)
	
	// Restore rows and names
	for i, idx := range indices {
		if idx <= len(c.data.Data) && i < len(c.oldRows) {
			// Insert row at original position
			c.data.Data = append(c.data.Data[:idx], append([][]string{c.oldRows[i]}, c.data.Data[idx:]...)...)
			
			// Restore row name if it existed
			if c.data.RowNames != nil && i < len(c.oldRowNames) {
				c.data.RowNames = append(c.data.RowNames[:idx], append([]string{c.oldRowNames[i]}, c.data.RowNames[idx:]...)...)
			}
		}
	}
	
	// Update row count
	c.data.Rows = len(c.data.Data)
	
	return nil
}

// GetDescription returns a description of the command
func (c *DeleteRowsCommand) GetDescription() string {
	if len(c.rowIndices) == 1 {
		return fmt.Sprintf("Delete row %d", c.rowIndices[0]+1)
	}
	return fmt.Sprintf("Delete %d rows", len(c.rowIndices))
}

// DeleteColumnsCommand represents deletion of multiple columns
type DeleteColumnsCommand struct {
	app         *App
	data        *FileData
	colIndices  []int
	oldHeaders  []string
	oldColumns  [][]string
	oldTypes    map[string]string
	oldCategorical map[string][]string
	oldNumericTarget map[string][]float64
}

// NewDeleteColumnsCommand creates a new delete columns command
func NewDeleteColumnsCommand(app *App, data *FileData, colIndices []int) *DeleteColumnsCommand {
	// Sort indices in descending order
	sortedIndices := make([]int, len(colIndices))
	copy(sortedIndices, colIndices)
	sort.Sort(sort.Reverse(sort.IntSlice(sortedIndices)))
	
	// Create maps to preserve column data by original index
	headerMap := make(map[int]string)
	columnMap := make(map[int][]string)
	oldTypes := make(map[string]string)
	oldCategorical := make(map[string][]string)
	oldNumericTarget := make(map[string][]float64)
	
	for _, idx := range colIndices {
		if idx < len(data.Headers) {
			header := data.Headers[idx]
			headerMap[idx] = header
			
			// Save column data
			colData := make([]string, len(data.Data))
			for j := range data.Data {
				if idx < len(data.Data[j]) {
					colData[j] = data.Data[j][idx]
				}
			}
			columnMap[idx] = colData
			
			// Save metadata
			if data.ColumnTypes != nil {
				if colType, exists := data.ColumnTypes[header]; exists {
					oldTypes[header] = colType
				}
			}
			if data.CategoricalColumns != nil {
				if values, exists := data.CategoricalColumns[header]; exists {
					oldCategorical[header] = values
				}
			}
			if data.NumericTargetColumns != nil {
				if values, exists := data.NumericTargetColumns[header]; exists {
					// Convert JSONFloat64 to float64
					floatValues := make([]float64, len(values))
					for j, v := range values {
						floatValues[j] = float64(v)
					}
					oldNumericTarget[header] = floatValues
				}
			}
		}
	}
	
	// Store in sorted order for undo
	sortedOriginal := make([]int, len(colIndices))
	copy(sortedOriginal, colIndices)
	sort.Ints(sortedOriginal)
	
	oldHeaders := make([]string, 0, len(sortedOriginal))
	oldColumns := make([][]string, 0, len(sortedOriginal))
	
	for _, idx := range sortedOriginal {
		if header, exists := headerMap[idx]; exists {
			oldHeaders = append(oldHeaders, header)
			oldColumns = append(oldColumns, columnMap[idx])
		}
	}
	
	return &DeleteColumnsCommand{
		app:         app,
		data:        data,
		colIndices:  sortedIndices,
		oldHeaders:  oldHeaders,
		oldColumns:  oldColumns,
		oldTypes:    oldTypes,
		oldCategorical: oldCategorical,
		oldNumericTarget: oldNumericTarget,
	}
}

// Execute deletes the columns
func (c *DeleteColumnsCommand) Execute() error {
	// Delete columns in descending order
	for _, idx := range c.colIndices {
		if idx >= 0 && idx < len(c.data.Headers) {
			header := c.data.Headers[idx]
			
			// Delete header
			c.data.Headers = append(c.data.Headers[:idx], c.data.Headers[idx+1:]...)
			
			// Delete column data from each row
			for i := range c.data.Data {
				if idx < len(c.data.Data[i]) {
					c.data.Data[i] = append(c.data.Data[i][:idx], c.data.Data[i][idx+1:]...)
				}
			}
			
			// Remove metadata
			if c.data.ColumnTypes != nil {
				delete(c.data.ColumnTypes, header)
			}
			if c.data.CategoricalColumns != nil {
				delete(c.data.CategoricalColumns, header)
			}
			if c.data.NumericTargetColumns != nil {
				delete(c.data.NumericTargetColumns, header)
			}
		}
	}
	
	// Update column count
	c.data.Columns = len(c.data.Headers)
	
	return nil
}

// Undo restores the deleted columns
func (c *DeleteColumnsCommand) Undo() error {
	// Restore columns in original order
	indices := make([]int, len(c.colIndices))
	copy(indices, c.colIndices)
	sort.Ints(indices)
	
	for i, idx := range indices {
		if idx <= len(c.data.Headers) {
			header := c.oldHeaders[i]
			
			// Insert header
			c.data.Headers = append(c.data.Headers[:idx], append([]string{header}, c.data.Headers[idx:]...)...)
			
			// Insert column data
			for j := range c.data.Data {
				value := ""
				if j < len(c.oldColumns[i]) {
					value = c.oldColumns[i][j]
				}
				c.data.Data[j] = append(c.data.Data[j][:idx], append([]string{value}, c.data.Data[j][idx:]...)...)
			}
			
			// Restore metadata
			if c.data.ColumnTypes == nil {
				c.data.ColumnTypes = make(map[string]string)
			}
			if colType, exists := c.oldTypes[header]; exists {
				c.data.ColumnTypes[header] = colType
			}
			
			if c.data.CategoricalColumns == nil {
				c.data.CategoricalColumns = make(map[string][]string)
			}
			if values, exists := c.oldCategorical[header]; exists {
				c.data.CategoricalColumns[header] = values
			}
			
			if c.data.NumericTargetColumns == nil {
				c.data.NumericTargetColumns = make(map[string][]types.JSONFloat64)
			}
			if values, exists := c.oldNumericTarget[header]; exists {
				// Convert float64 back to JSONFloat64
				jsonValues := make([]types.JSONFloat64, len(values))
				for j, v := range values {
					jsonValues[j] = types.JSONFloat64(v)
				}
				c.data.NumericTargetColumns[header] = jsonValues
			}
		}
	}
	
	// Update column count
	c.data.Columns = len(c.data.Headers)
	
	return nil
}

// GetDescription returns a description of the command
func (c *DeleteColumnsCommand) GetDescription() string {
	if len(c.colIndices) == 1 && len(c.oldHeaders) > 0 {
		return fmt.Sprintf("Delete column '%s'", c.oldHeaders[0])
	}
	return fmt.Sprintf("Delete %d columns", len(c.colIndices))
}

// InsertRowCommand represents insertion of a new row
type InsertRowCommand struct {
	app      *App
	data     *FileData
	index    int
	rowName  string
}

// NewInsertRowCommand creates a new insert row command
func NewInsertRowCommand(app *App, data *FileData, index int) *InsertRowCommand {
	rowName := ""
	if data.RowNames != nil {
		rowName = fmt.Sprintf("Row%d", len(data.Data)+1)
	}
	
	return &InsertRowCommand{
		app:     app,
		data:    data,
		index:   index,
		rowName: rowName,
	}
}

// Execute inserts a new row
func (c *InsertRowCommand) Execute() error {
	// Create empty row
	newRow := make([]string, len(c.data.Headers))
	
	// Insert at specified index
	if c.index >= len(c.data.Data) {
		c.data.Data = append(c.data.Data, newRow)
		if c.data.RowNames != nil {
			c.data.RowNames = append(c.data.RowNames, c.rowName)
		}
	} else {
		c.data.Data = append(c.data.Data[:c.index], append([][]string{newRow}, c.data.Data[c.index:]...)...)
		if c.data.RowNames != nil {
			c.data.RowNames = append(c.data.RowNames[:c.index], append([]string{c.rowName}, c.data.RowNames[c.index:]...)...)
		}
	}
	
	// Update row count
	c.data.Rows = len(c.data.Data)
	
	return nil
}

// Undo removes the inserted row
func (c *InsertRowCommand) Undo() error {
	if c.index < len(c.data.Data) {
		c.data.Data = append(c.data.Data[:c.index], c.data.Data[c.index+1:]...)
		if c.data.RowNames != nil && c.index < len(c.data.RowNames) {
			c.data.RowNames = append(c.data.RowNames[:c.index], c.data.RowNames[c.index+1:]...)
		}
	}
	
	// Update row count
	c.data.Rows = len(c.data.Data)
	
	return nil
}

// GetDescription returns a description of the command
func (c *InsertRowCommand) GetDescription() string {
	return fmt.Sprintf("Insert row at position %d", c.index+1)
}

// InsertColumnCommand represents insertion of a new column
type InsertColumnCommand struct {
	app    *App
	data   *FileData
	index  int
	name   string
}

// NewInsertColumnCommand creates a new insert column command
func NewInsertColumnCommand(app *App, data *FileData, index int, name string) *InsertColumnCommand {
	if name == "" {
		name = fmt.Sprintf("Column%d", len(data.Headers)+1)
	}
	
	return &InsertColumnCommand{
		app:   app,
		data:  data,
		index: index,
		name:  name,
	}
}

// Execute inserts a new column
func (c *InsertColumnCommand) Execute() error {
	// Insert header
	if c.index >= len(c.data.Headers) {
		c.data.Headers = append(c.data.Headers, c.name)
	} else {
		c.data.Headers = append(c.data.Headers[:c.index], append([]string{c.name}, c.data.Headers[c.index:]...)...)
	}
	
	// Insert empty values in each row
	for i := range c.data.Data {
		if c.index >= len(c.data.Data[i]) {
			c.data.Data[i] = append(c.data.Data[i], "")
		} else {
			c.data.Data[i] = append(c.data.Data[i][:c.index], append([]string{""}, c.data.Data[i][c.index:]...)...)
		}
	}
	
	// Add column type as text by default
	if c.data.ColumnTypes == nil {
		c.data.ColumnTypes = make(map[string]string)
	}
	c.data.ColumnTypes[c.name] = "text"
	
	// Update column count
	c.data.Columns = len(c.data.Headers)
	
	return nil
}

// Undo removes the inserted column
func (c *InsertColumnCommand) Undo() error {
	// Find column index by name (in case columns were reordered)
	idx := -1
	for i, header := range c.data.Headers {
		if header == c.name {
			idx = i
			break
		}
	}
	
	if idx >= 0 {
		// Remove header
		c.data.Headers = append(c.data.Headers[:idx], c.data.Headers[idx+1:]...)
		
		// Remove column data
		for i := range c.data.Data {
			if idx < len(c.data.Data[i]) {
				c.data.Data[i] = append(c.data.Data[i][:idx], c.data.Data[i][idx+1:]...)
			}
		}
		
		// Remove metadata
		if c.data.ColumnTypes != nil {
			delete(c.data.ColumnTypes, c.name)
		}
	}
	
	// Update column count
	c.data.Columns = len(c.data.Headers)
	
	return nil
}

// GetDescription returns a description of the command
func (c *InsertColumnCommand) GetDescription() string {
	return fmt.Sprintf("Insert column '%s'", c.name)
}

// ToggleTargetColumnCommand represents toggling the #target suffix on a column
type ToggleTargetColumnCommand struct {
	app        *App
	data       *FileData
	colIndex   int
	oldName    string
	newName    string
	wasTarget  bool
}

// NewToggleTargetColumnCommand creates a new toggle target column command
func NewToggleTargetColumnCommand(app *App, data *FileData, colIndex int) *ToggleTargetColumnCommand {
	if colIndex >= len(data.Headers) {
		return nil
	}
	
	oldName := data.Headers[colIndex]
	newName := oldName
	wasTarget := false
	
	// Check if column already has #target suffix
	lowerName := strings.ToLower(oldName)
	if strings.HasSuffix(lowerName, "#target") || strings.HasSuffix(lowerName, "# target") {
		// Remove #target suffix
		wasTarget = true
		if strings.HasSuffix(oldName, "#target") {
			newName = strings.TrimSuffix(oldName, "#target")
		} else if strings.HasSuffix(oldName, "# target") {
			newName = strings.TrimSuffix(oldName, "# target")
		} else if strings.HasSuffix(oldName, "#Target") {
			newName = strings.TrimSuffix(oldName, "#Target")
		} else if strings.HasSuffix(oldName, "# Target") {
			newName = strings.TrimSuffix(oldName, "# Target")
		}
		newName = strings.TrimSpace(newName)
	} else {
		// Add #target suffix
		newName = oldName + "#target"
	}
	
	return &ToggleTargetColumnCommand{
		app:       app,
		data:      data,
		colIndex:  colIndex,
		oldName:   oldName,
		newName:   newName,
		wasTarget: wasTarget,
	}
}

// Execute toggles the target column suffix
func (c *ToggleTargetColumnCommand) Execute() error {
	// Use the existing header edit logic
	headerCmd := NewHeaderEditCommand(c.app, c.data, c.colIndex, c.oldName, c.newName)
	return headerCmd.Execute()
}

// Undo reverts the toggle
func (c *ToggleTargetColumnCommand) Undo() error {
	headerCmd := NewHeaderEditCommand(c.app, c.data, c.colIndex, c.newName, c.oldName)
	return headerCmd.Execute()
}

// GetDescription returns a description of the command
func (c *ToggleTargetColumnCommand) GetDescription() string {
	if c.wasTarget {
		return fmt.Sprintf("Remove target flag from '%s'", c.oldName)
	}
	return fmt.Sprintf("Mark '%s' as target column", c.oldName)
}