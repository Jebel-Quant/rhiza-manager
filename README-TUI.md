# Rhiza Manager TUI

A Terminal User Interface (TUI) version of Rhiza Manager built with Go. Manage multiple Git repositories from a beautiful terminal interface.

## Features

- 📊 **Repository Status View** - See all your repositories with their current status:
  - Current branch
  - Clean/Dirty status
  - Commits ahead/behind remote
  - Template sync status (commits behind rhiza template)
- 🔄 **Refresh Status** - Update the status of all repositories
- ⬇️ **Pull Repositories** - Pull changes from remote for selected repositories
- 🔄 **Fetch Repositories** - Fetch updates from remote for selected repositories
- 🔀 **Sync Templates** - Materialize/sync rhiza templates for selected repositories (requires `rhiza` CLI)
- ✅ **Multi-select** - Select multiple repositories to perform bulk operations

## Installation

### Prerequisites

- Go 1.21 or higher
- Git installed and available in your PATH

### Build

```bash
go mod download
go build -o rhiza-tui ./cmd/rhiza-tui
```

Or install directly:

```bash
go install ./cmd/rhiza-tui
```

## Configuration

Create a `config.json` file in the same directory as the executable (or specify a path as the first argument):

```json
{
  "repositories": [
    {
      "name": "my-project",
      "path": "/Users/username/projects/my-project"
    },
    {
      "name": "another-repo",
      "path": "/Users/username/projects/another-repo"
    }
  ]
}
```

- `name`: Display name for the repository (can be anything)
- `path`: Absolute or relative path to the Git repository

You can copy `config.json.example` to `config.json` and modify it:

```bash
cp config.json.example config.json
# Edit config.json with your repositories
```

## Usage

Run the TUI:

```bash
./rhiza-tui
```

Or with a custom config path:

```bash
./rhiza-tui /path/to/config.json
```

### Controls

- **↑/↓** or **k/j**: Navigate up/down through repositories
- **Space**: Toggle selection of the current repository
- **a**: Select all repositories
- **d**: Deselect all repositories
- **r**: Refresh status of all repositories
- **p**: Pull selected repositories
- **f**: Fetch selected repositories
- **s**: Sync/Materialize rhiza templates for selected repositories
- **q** or **Ctrl+C**: Quit

### Workflow

1. **Select repositories**: Use arrow keys to navigate and spacebar to select repositories
2. **Pull or Fetch**: Press `p` to pull or `f` to fetch selected repositories
3. **Sync Templates**: Press `s` to sync/materialize rhiza templates (requires `rhiza` CLI installed)
4. **Refresh**: Press `r` to refresh the status of all repositories

## Status Indicators

- **Branch name** (e.g., `main`, `develop`) - The current Git branch
- **clean** - No uncommitted changes in the working directory
- **dirty** - There are uncommitted changes
- **↑N** - Number of commits your local branch is ahead of the remote
- **↓N** - Number of commits your local branch is behind the remote
- **template: ↓N** - Number of commits behind the rhiza template (if repository uses rhiza)
- **template: up-to-date** - Repository is synced with the rhiza template

## Example

```
 Rhiza Manager 

> ✓ my-project     main · clean · ↑0 ↓2 · template: ↓5
  ✓ another-repo   develop · dirty · ↑1 ↓0 · template: up-to-date
    third-repo     feature-branch · clean · ↑0 ↓0

Status refreshed

↑/↓: navigate  space: select  r: refresh  p: pull  f: fetch  s: sync  a: select all  d: deselect all  q: quit
```

## Requirements

- Git must be installed and available in your PATH
- Each repository path in the config must be a valid Git repository (contain a `.git` directory)
- Remote tracking information requires an upstream branch to be set
- For template syncing: `rhiza` CLI must be installed (`pip install rhiza`) or `uvx` available (will use `uvx rhiza`)

## Differences from VS Code Extension

- **Configuration-driven**: Repositories are defined in a JSON config file instead of auto-detected from workspace folders
- **Multi-select**: Select specific repositories to pull/fetch instead of always operating on all repositories
- **Terminal-based**: Runs in any terminal, no IDE required

## Troubleshooting

### "Failed to read config file"

Make sure `config.json` exists in the current directory or provide the path as an argument.

### "Error: repository not found"

Verify that the paths in your `config.json` are correct and point to valid Git repositories.

### Upstream branch not set

If you see `↑0 ↓0` for all repositories, you may need to set upstream branches:

```bash
git branch --set-upstream-to=origin/main main
```

## License

MIT

