import * as assert from 'assert';
import * as path from 'path';
import * as fs from 'fs';
import { exec } from 'child_process';

/**
 * Test helper functions for git operations
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

suite('Helper Functions Test Suite', () => {

  test('runGitCommand should execute git commands successfully', async () => {
    // Test with a simple git command that should work in any directory
    const result = await runGitCommand(process.cwd(), 'git --version');
    assert.ok(result.includes('git version'), 'Should return git version');
  });

  test('runGitCommand should reject on invalid command', async () => {
    try {
      await runGitCommand(process.cwd(), 'git invalid-command-xyz');
      assert.fail('Should have thrown an error');
    } catch (error) {
      assert.ok(error, 'Should throw an error for invalid command');
    }
  });

  test('runGitCommand should reject on non-existent directory', async () => {
    try {
      await runGitCommand('/non/existent/path', 'git status');
      assert.fail('Should have thrown an error');
    } catch (error) {
      assert.ok(error, 'Should throw an error for non-existent directory');
    }
  });
});
