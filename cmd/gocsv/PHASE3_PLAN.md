# GoCSV Phase 3: Enhanced Features - Implementation Plan

## Overview
Phase 3 focuses on advanced features that enhance the user experience and provide deeper insights into data quality before PCA analysis. This phase builds on the solid foundation of Phase 1 (core editor) and Phase 2 (format support).

## Timeline: 2 weeks

## Progress Status
- ✅ **Missing Value Handling Tools** - COMPLETED
- ✅ **Data Quality Report Feature** - COMPLETED
- ⏳ **Direct GoPCA Integration** - IN PROGRESS
- ⏳ **Undo/Redo Functionality** - PENDING
- ⏳ **Import Wizard UI** - PENDING
- ⏳ **Advanced Data Transformations** - PENDING

## Task Breakdown

### 1. Missing Value Handling Tools (3 days) ✅ COMPLETED

#### 1.1 Visual Missing Value Indicators
- **Backend (app.go)**:
  ```go
  type MissingValueStats struct {
      TotalCells      int                    `json:"totalCells"`
      MissingCells    int                    `json:"missingCells"`
      MissingPercent  float64                `json:"missingPercent"`
      ColumnStats     map[string]ColumnStats `json:"columnStats"`
      RowStats        map[int]RowStats       `json:"rowStats"`
  }
  
  func (a *App) AnalyzeMissingValues(data *FileData) *MissingValueStats
  ```

- **Frontend Components**:
  - `MissingValueHeatmap.tsx`: Visual heatmap showing missing value patterns
  - `MissingValueSummary.tsx`: Statistics panel with column/row summaries
  - Update `CSVGrid.tsx` to highlight cells with missing values

#### 1.2 Missing Value Operations
- **Fill strategies**:
  - Fill with mean/median/mode (numeric columns)
  - Fill with most frequent value (categorical columns)
  - Forward/backward fill
  - Interpolation for time series
  - Custom value fill

- **UI Implementation**:
  - Right-click context menu on columns
  - "Missing Values" toolbar with quick actions
  - Batch operations dialog

#### 1.3 Missing Value Report
- Generate detailed report showing:
  - Missing value patterns (random, systematic)
  - Correlation between missing values in different columns
  - Recommendations for handling strategies

### 2. Data Quality Report Feature (3 days) ✅ COMPLETED

#### 2.1 Data Quality Metrics
- **Backend Analysis**:
  ```go
  type DataQualityReport struct {
      DataProfile      DataProfile           `json:"dataProfile"`
      ColumnAnalysis   []ColumnAnalysis      `json:"columnAnalysis"`
      QualityScore     float64               `json:"qualityScore"`
      Issues           []QualityIssue        `json:"issues"`
      Recommendations  []Recommendation      `json:"recommendations"`
  }
  
  type ColumnAnalysis struct {
      Name         string              `json:"name"`
      Type         string              `json:"type"`
      Stats        ColumnStatistics    `json:"stats"`
      Distribution DistributionInfo    `json:"distribution"`
      Outliers     []OutlierInfo       `json:"outliers"`
  }
  ```

#### 2.2 Quality Checks
- **Automated checks**:
  - Duplicate row detection
  - Outlier detection (IQR, Z-score)
  - Data type consistency
  - Value range validation
  - Correlation analysis for multicollinearity
  - Normality tests for numeric columns

#### 2.3 Report UI
- **Components**:
  - `DataQualityDashboard.tsx`: Main report view
  - `DistributionChart.tsx`: Histograms and box plots
  - `CorrelationMatrix.tsx`: Visual correlation heatmap
  - `QualityScoreCard.tsx`: Overall data quality score

- **Export options**:
  - PDF report generation
  - HTML report with interactive charts
  - JSON for programmatic access

### 3. Direct GoPCA Integration (2 days) ⏳ IN PROGRESS

#### 3.1 Inter-Process Communication
- **Approach 1: File-based handoff**:
  ```go
  func (a *App) OpenInGoPCA(data *FileData) error {
      // Save to temp file
      tempFile := SaveToTemp(data)
      
      // Launch GoPCA with file path
      cmd := exec.Command("gopca-desktop", "--open", tempFile)
      return cmd.Start()
  }
  ```

- **Approach 2: URL scheme** (if GoPCA supports):
  ```go
  func (a *App) OpenInGoPCAURL(data *FileData) error {
      // Encode data in URL
      url := fmt.Sprintf("gopca://open?data=%s", encodeData(data))
      return browser.OpenURL(url)
  }
  ```

#### 3.2 Shared Validation
- Extract validation logic into shared package:
  ```go
  // pkg/validation/validator.go
  type Validator struct {
      Rules []ValidationRule
  }
  
  func (v *Validator) Validate(data *FileData) []ValidationError
  ```

#### 3.3 UI Integration
- "Open in GoPCA" button with status indicators
- Check if GoPCA is installed
- Show validation status before handoff

### 4. Undo/Redo Functionality (2 days)

#### 4.1 Command Pattern Implementation
- **Backend**:
  ```go
  type Command interface {
      Execute() error
      Undo() error
      GetDescription() string
  }
  
  type CommandHistory struct {
      commands []Command
      current  int
  }
  
  // Example command
  type CellEditCommand struct {
      row, col int
      oldValue, newValue string
      data *FileData
  }
  ```

#### 4.2 Supported Operations
- Cell edits
- Row/column operations (add, delete, move)
- Fill operations
- Data transformations
- Import operations

#### 4.3 UI Implementation
- Keyboard shortcuts (Cmd+Z, Cmd+Shift+Z)
- Edit menu with undo/redo
- History panel showing recent operations

### 5. Import Wizard UI (2 days)

#### 5.1 Multi-Step Wizard
- **Step 1: File Selection**
  - Drag & drop area
  - Recent files list
  - Format auto-detection

- **Step 2: Format-Specific Options**
  - Excel: Sheet selection, range selection
  - JSON: Path to data array, field mapping
  - CSV: Delimiter detection, encoding selection

- **Step 3: Data Preview**
  - First 100 rows preview
  - Column type detection results
  - Encoding issues detection

- **Step 4: Import Options**
  - Column selection
  - Type overrides
  - Header row selection
  - Row name column selection

#### 5.2 Components
- `ImportWizard.tsx`: Main wizard container
- `FileSelector.tsx`: File selection step
- `FormatOptions.tsx`: Format-specific options
- `DataPreview.tsx`: Preview with type detection
- `ImportProgress.tsx`: Progress bar for large files

### 6. Advanced Data Transformations (1 day)

#### 6.1 Column Operations
- **Type conversions**:
  - String to numeric with error handling
  - Numeric to categorical (binning)
  - Date/time parsing

- **Transformations**:
  - Log/sqrt/power transformations
  - Standardization (z-score)
  - Min-max scaling
  - One-hot encoding for categorical

#### 6.2 Row Operations
- Sort by multiple columns
- Filter by complex conditions
- Remove duplicates with key selection
- Sampling (random, stratified)

## Technical Considerations

### Performance
- Virtualized grid for large datasets (already implemented)
- Web Workers for heavy computations
- Streaming processing for large file imports
- Lazy loading for data quality analysis

### State Management
- Consider Redux or Zustand for complex state
- Separate stores for:
  - Data state
  - UI state
  - Command history
  - Analysis results

### Testing Strategy
- Unit tests for all data transformations
- Integration tests for import/export
- E2E tests for wizard flows
- Performance benchmarks for large datasets

### Documentation
- User guide for each feature
- Video tutorials for complex operations
- API documentation for shared packages
- Migration guide from other tools

## Dependencies
- **Charting**: Recharts (already used) or D3.js for advanced visualizations
- **PDF Generation**: jsPDF or puppeteer
- **Statistics**: Simple Statistics or custom implementations
- **File Processing**: Maintain current approach with backend processing

## Success Metrics
- Missing value handling reduces data prep time by 40%
- Quality reports catch 95% of common data issues
- Direct GoPCA integration used in 80% of workflows
- Undo/redo improves user confidence (measure via analytics)
- Import wizard handles 99% of files without errors

## Risk Mitigation
- **Complexity creep**: Keep features focused on PCA preparation
- **Performance**: Profile and optimize before adding features
- **Browser limits**: Plan for large file handling strategies
- **Integration issues**: Implement fallback for GoPCA handoff

## Future Considerations (Post-Phase 3)
- Plugin system for custom transformations
- Collaboration features (share sessions)
- Cloud storage integration
- Automated data cleaning pipelines
- Machine learning for type detection