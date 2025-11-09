# GoCSV MSIX Assets

This directory contains visual assets required for the GoCSV Desktop application when bundled in the MSIX package.

## Required Assets

### Square Tile Assets

1. **GoCSV_Square150x150Logo.png**
   - Dimensions: 150x150 pixels
   - Format: PNG with transparency or solid background
   - Purpose: Start Menu medium tile and Microsoft Store listing
   - Design: Should match GoCSV branding (CSV/spreadsheet theme)

2. **GoCSV_Square44x44Logo.png**
   - Dimensions: 44x44 pixels
   - Format: PNG with transparency or solid background
   - Purpose: Start Menu app list and taskbar icon
   - Design: Simplified version of the main GoCSV logo

## Design Guidelines

### Color Scheme
- Should complement existing GoCSV icon colors
- Consider using green/blue tones to differentiate from GoPCA (which uses blue)
- Maintain visual consistency with existing `cmd/gocsv/build/icons/` assets

### Microsoft Store Requirements
- Follow [Microsoft Store icon guidelines](https://learn.microsoft.com/en-us/windows/apps/design/style/iconography/app-icon-design)
- Ensure sufficient contrast for both light and dark backgrounds
- Avoid overly complex details in the 44x44 version
- Test visibility at various DPI scales (100%, 125%, 150%, 200%)

### Asset Generation

**Recommended approach**:
1. Start with existing `cmd/gocsv/build/icons/icon-256.png` or `icon-128.png`
2. Resize to 150x150 and 44x44 using high-quality image scaling
3. Optionally add padding/margins if the icon doesn't work well as a perfect square
4. Save as PNG with transparency

**Tools**:
- ImageMagick: `magick icon-256.png -resize 150x150 GoCSV_Square150x150Logo.png`
- Photoshop/GIMP: Manual resizing with high-quality interpolation
- Online tools: Use PNG-compatible resizing services

## Integration with MSIX Package

These assets are referenced in `cmd/gopca-desktop/AppxManifest.template.xml`:

```xml
<Application Id="GoCSV"
             Executable="GoCSV.exe"
             EntryPoint="Windows.FullTrustApplication">
  <uap:VisualElements
    DisplayName="GoCSV"
    Description="CSV data preparation tool for GoPCA Suite"
    Square150x150Logo="Assets\GoCSV_Square150x150Logo.png"
    Square44x44Logo="Assets\GoCSV_Square44x44Logo.png"
    ... />
</Application>
```

The build script (`scripts/windows/build-msix.sh`) copies these assets to the MSIX package during the build process.

## Status

✅ **All required MSIX assets have been created** (generated from `docs/images/GoCSV-icon-1024-black.png`)

### Completed
- [x] GoCSV_Square44x44Logo.png (44×44, 3.3KB)
- [x] GoCSV_Square71x71Logo.png (71×71, 5.5KB)
- [x] GoCSV_Square150x150Logo.png (150×150, 14KB)
- [x] GoCSV_Square300x300Logo.png (300×300, 42KB)
- [x] GoCSV_StoreLogo.png (50×50, 3.7KB)

### TODO
- [ ] Test assets at different DPI scales (100%, 125%, 150%, 200%)
- [ ] Verify visual consistency with GoPCA assets
- [ ] Validate against Microsoft Store asset checker
- [ ] Test in actual MSIX package installation

## References

- Microsoft Store icon design: https://learn.microsoft.com/en-us/windows/apps/design/style/iconography/app-icon-design
- MSIX visual assets: https://learn.microsoft.com/en-us/windows/apps/design/style/iconography/app-icon-construction
- Existing GoCSV icons: `cmd/gocsv/build/icons/`
