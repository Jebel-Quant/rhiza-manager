import * as assert from 'assert';
import { exec } from 'child_process';

/**
 * Test git status parsing functionality
 */

function runGitCommand(repoPath: string, command: string): Promise<string> {
  return new Promise((resolve, reject) => {
    exec(command, { cwd: repoPath }, (err, stdout, stderr) => {
      if (err) {
        reject(stderr || err.message);
      } else {
        resolve(stdout.trim());
      }
    });
  });
}

async function getRepoStatus(repoPath: string) {
  try {
    const branch = (await runGitCommand(repoPath, "git branch --show-current")) || "detached";
    const dirtyOutput = await runGitCommand(repoPath, "git status --porcelain");
    const dirty = dirtyOutput.length > 0;
    let ahead = 0, behind = 0;

    try {
      const revList = await runGitCommand(
        repoPath,
        "git rev-list --left-right --count HEAD...@{upstream}"
      );
      [ahead, behind] = revList.split("\t").map(Number);
    } catch { /* upstream may not exist */ }

    return { branch, dirty, ahead, behind };
  } catch {
    return { branch: "unknown", dirty: false, ahead: 0, behind: 0 };
  }
}

suite('Git Status Test Suite', () => {

  test('getRepoStatus should return status object with required fields', async () => {
    const status = await getRepoStatus(process.cwd());
    
    assert.ok(status, 'Status should be defined');
    assert.ok(typeof status.branch === 'string', 'Branch should be a string');
    assert.ok(typeof status.dirty === 'boolean', 'Dirty should be a boolean');
    assert.ok(typeof status.ahead === 'number', 'Ahead should be a number');
    assert.ok(typeof status.behind === 'number', 'Behind should be a number');
  });

  test('getRepoStatus should detect current branch', async () => {
    const status = await getRepoStatus(process.cwd());
    
    assert.notStrictEqual(status.branch, '', 'Branch should not be empty');
    assert.notStrictEqual(status.branch, 'unknown', 'Branch should be detected');
  });

  test('getRepoStatus should handle invalid repository path', async () => {
    const status = await getRepoStatus('/non/existent/path');
    
    assert.strictEqual(status.branch, 'unknown', 'Should return unknown for invalid path');
    assert.strictEqual(status.dirty, false, 'Should return false for dirty status');
    assert.strictEqual(status.ahead, 0, 'Should return 0 for ahead count');
    assert.strictEqual(status.behind, 0, 'Should return 0 for behind count');
  });

  test('ahead and behind counts should be non-negative', async () => {
    const status = await getRepoStatus(process.cwd());
    
    assert.ok(status.ahead >= 0, 'Ahead count should be non-negative');
    assert.ok(status.behind >= 0, 'Behind count should be non-negative');
  });
});
