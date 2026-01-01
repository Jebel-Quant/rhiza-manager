# Tests

This directory contains unit tests for the Rhiza Manager VS Code extension.

## Test Structure

- `helper-functions.test.ts` - Tests for git command execution helpers
- `repo-detection.test.ts` - Tests for repository detection logic
- `git-status.test.ts` - Tests for git status parsing and reporting

## Running Tests

To run the tests:

```bash
npm test
```

## Test Coverage

The tests cover:
- Git command execution and error handling
- Repository detection (identifying .git directories)
- Git status parsing (branch, dirty state, ahead/behind counts)
- Error handling for invalid paths and commands

## Adding New Tests

When adding new functionality to the extension, please add corresponding tests to ensure reliability and maintainability.
