# Visual Guide Documentation

This folder contains SVG visual guides that illustrate how the Rhiza Manager extension works.

## Files

### extension-overview.svg
Shows the complete VS Code interface with the Rhiza Repositories view integrated into the Explorer sidebar. Demonstrates how the extension fits into the overall VS Code UI.

**Key elements:**
- Explorer sidebar with "RHIZA REPOSITORIES" section
- Multiple repositories listed with their status
- Editor area showing documentation

### tree-view.svg
Detailed view of the Rhiza Repositories tree showing multiple repositories with different statuses.

**Key elements:**
- Repository names
- Branch indicators
- Clean/dirty status
- Commits ahead/behind counters
- Status legend explaining the indicators

### commands-menu.svg
Shows the context menu with available commands when interacting with the view.

**Key elements:**
- Refresh command
- Pull All Repositories command
- Fetch All Repositories command
- Menu icon in the view toolbar

### pull-progress.svg
Illustrates the progress notification that appears when pulling all repositories.

**Key elements:**
- Progress bar
- Status messages for completed repositories
- Current operation indicator

## Technical Details

All images are created as SVG files for:
- **Scalability**: Vector graphics that look good at any size
- **Small file size**: Text-based format that compresses well
- **Editability**: Can be easily modified if design changes are needed
- **Accessibility**: Can be read by screen readers
- **Dark theme**: Matches VS Code's default dark theme (#1e1e1e background)

## Color Scheme

The visuals use VS Code's default dark theme colors:
- Background: `#1e1e1e`
- Sidebar: `#252526`
- Text (primary): `#cccccc`
- Text (secondary): `#858585`
- Accent (blue): `#007acc`, `#3794ff`
- Success (green): `#89d185`
- Warning (yellow): `#cca700`
