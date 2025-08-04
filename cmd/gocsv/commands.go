package main

import (
	"fmt"
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