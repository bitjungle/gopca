# Screenshot Update Checklist for GoPCA Documentation

This checklist guides the process of updating all screenshots to showcase the new Plotly-based visualization system.

## Prerequisites
- [ ] GoPCA Desktop application built and running
- [ ] GoCSV application built and running  
- [ ] Example datasets available (Iris, Wine, NIR, Swiss Roll)
- [ ] Screenshot tool ready (macOS: Cmd+Shift+4, Windows: Snipping Tool, Linux: gnome-screenshot)

## Screenshots to Capture

### 1. Main Application Screenshots

#### gopca-overview.jpg
- [ ] Launch GoPCA Desktop
- [ ] Load Iris or Wine dataset
- [ ] Ensure Scores Plot is visible with groups colored
- [ ] Show confidence ellipses enabled
- [ ] Capture full application window
- [ ] Light theme recommended
- [ ] Resolution: 1920x1080 or higher

#### gopca-scoreplot-example.png
- [ ] Load Iris dataset
- [ ] Select Scores Plot visualization
- [ ] Enable confidence ellipses (95% level)
- [ ] Enable group coloring by species
- [ ] Zoom to show clear separation between groups
- [ ] Capture just the plot area with Plotly toolbar visible
- [ ] High resolution PNG format

#### GoPCA-plots.jpg (Collage)
Create a collage showing all 7 main visualizations:
- [ ] Scores Plot with groups and ellipses
- [ ] Biplot with loading vectors
- [ ] Loadings Plot (bar chart)
- [ ] Scree Plot with cumulative variance
- [ ] Circle of Correlations
- [ ] Diagnostic Plot (T² vs Q)
- [ ] Eigencorrelation Plot heatmap

For each plot in the collage:
- Use consistent dataset (Wine recommended for variety)
- Show Plotly toolbar
- Use light theme for consistency
- Arrange in a 3x3 grid (leave center or corner empty for title)

### 2. GoCSV Screenshots

#### gocsv-overview.jpg
- [ ] Launch GoCSV
- [ ] Load a dataset with mixed types (numeric and categorical columns)
- [ ] Show the spreadsheet view with data
- [ ] Ensure column headers are visible
- [ ] Show type indicators if available
- [ ] Capture full application window

#### gocsv-qr-example.png
- [ ] Load dataset with some missing values or quality issues
- [ ] Show Quality Report panel or validation results
- [ ] Highlight any warnings or issues detected
- [ ] Show recommendations for data cleaning
- [ ] Capture relevant portion of the window

## Technical Requirements

### Image Specifications
- **Format**: PNG for individual plots, JPG for overview/collage images
- **Resolution**: Minimum 1920x1080, prefer 2x retina quality
- **File size**: Optimize to keep under 1MB per image
- **Naming**: Keep existing filenames for easy replacement

### Visual Guidelines
- **Window size**: Consistent across screenshots
- **Theme**: Use light theme unless showing theme options
- **Data**: Use recognizable example datasets
- **Features**: Show key Plotly features:
  - Toolbar with zoom, pan, export buttons
  - Interactive tooltips (if possible to capture)
  - Professional appearance

### Plotly-Specific Elements to Highlight
- [ ] Modebar (toolbar) visible in all plot screenshots
- [ ] Export button prominently shown
- [ ] Clean, modern appearance
- [ ] GoPCA watermark visible (if enabled)
- [ ] Fullscreen toggle button visible

## Post-Processing

### Image Optimization
```bash
# Optimize PNG files (requires optipng)
optipng -o2 docs/images/*.png

# Optimize JPG files (requires jpegoptim)
jpegoptim --size=1000k docs/images/*.jpg

# Alternative: Use ImageOptim app on macOS
```

### Verification
- [ ] All images display correctly in README.md
- [ ] File sizes are reasonable (< 1MB each)
- [ ] Images are sharp and readable
- [ ] Plotly interface elements are clearly visible
- [ ] No debug panels or development artifacts visible

## README Updates

After capturing new screenshots:
1. [ ] Replace old images in `docs/images/`
2. [ ] Verify image links in README.md still work
3. [ ] Update any captions if visualization names changed
4. [ ] Add note about Plotly interactivity if not present
5. [ ] Commit changes with descriptive message

## Additional Considerations

### Optional Enhancements
- [ ] Add animated GIF showing plot interactivity
- [ ] Include dark theme example
- [ ] Show export dialog or exported PNG
- [ ] Capture 3D plot rotation (for 3D Scores/Biplot)

### Future Screenshots
Consider adding:
- Workflow diagram showing data flow
- Before/after comparison (Recharts vs Plotly)
- Export options demonstration
- Palette customization examples

## Completion
- [ ] All required screenshots captured
- [ ] Images optimized for repository
- [ ] README.md updated if needed
- [ ] Changes committed and pushed
- [ ] Pull request created referencing issue #280