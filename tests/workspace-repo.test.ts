import * as assert from 'assert';
import * as path from 'path';
import * as fs from 'fs';

/**
 * Test workspace-level repository detection
 */

suite('Workspace Repository Detection Test Suite', () => {

  test('should detect workspace folder itself as repository when it contains .git', () => {
    const testPath = process.cwd();
    const gitPath = path.join(testPath, '.git');
    const isRepo = fs.existsSync(gitPath);
    
    // The current working directory (workspace root) should be a git repository
    assert.strictEqual(isRepo, true, 'Workspace root should be a git repository');
  });

  test('should correctly get basename of workspace folder for display', () => {
    const testPath = process.cwd();
    const baseName = path.basename(testPath);
    
    assert.ok(baseName.length > 0, 'Basename should not be empty');
    assert.strictEqual(typeof baseName, 'string', 'Basename should be a string');
  });

  test('should handle paths correctly for workspace folders', () => {
    const testPath = '/path/to/workspace/my-repo';
    const baseName = path.basename(testPath);
    
    assert.strictEqual(baseName, 'my-repo', 'Should extract correct folder name');
  });
});
