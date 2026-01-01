import * as assert from 'assert';
import * as path from 'path';
import * as fs from 'fs';

/**
 * Test repository detection logic
 */

suite('Repository Detection Test Suite', () => {

  test('should identify directory with .git as a repository', () => {
    const testPath = process.cwd();
    const gitPath = path.join(testPath, '.git');
    const isRepo = fs.existsSync(gitPath);
    
    // The current working directory should be a git repository
    assert.strictEqual(isRepo, true, 'Current directory should be a git repository');
  });

  test('should correctly check if path is a directory', () => {
    const testPath = process.cwd();
    const isDirectory = fs.statSync(testPath).isDirectory();
    
    assert.strictEqual(isDirectory, true, 'Current path should be a directory');
  });

  test('should handle non-existent paths gracefully', () => {
    const nonExistentPath = path.join(process.cwd(), 'non-existent-directory-xyz');
    const exists = fs.existsSync(nonExistentPath);
    
    assert.strictEqual(exists, false, 'Non-existent path should return false');
  });
});
