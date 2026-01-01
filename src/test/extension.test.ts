import * as assert from 'assert';
import * as vscode from 'vscode';
import * as myExtension from '../extension';

suite('Extension Test Suite', () => {
	vscode.window.showInformationMessage('Start all tests.');

	test('Extension exports activate function', () => {
		assert.ok(myExtension.activate, 'activate function should be exported');
		assert.strictEqual(typeof myExtension.activate, 'function', 'activate should be a function');
	});

	test('Extension exports deactivate function', () => {
		assert.ok(myExtension.deactivate, 'deactivate function should be exported');
		assert.strictEqual(typeof myExtension.deactivate, 'function', 'deactivate should be a function');
	});

	test('Extension exports runGitCommand function', () => {
		assert.ok(myExtension.runGitCommand, 'runGitCommand function should be exported');
		assert.strictEqual(typeof myExtension.runGitCommand, 'function', 'runGitCommand should be a function');
	});

	test('Extension exports getRepoStatus function', () => {
		assert.ok(myExtension.getRepoStatus, 'getRepoStatus function should be exported');
		assert.strictEqual(typeof myExtension.getRepoStatus, 'function', 'getRepoStatus should be a function');
	});

	test('runGitCommand should execute git commands successfully', async () => {
		// Test with a simple git command that should work in any directory
		const result = await myExtension.runGitCommand(process.cwd(), 'git --version');
		assert.ok(result.includes('git version'), 'Should return git version');
	});

	test('runGitCommand should reject on invalid command', async () => {
		try {
			await myExtension.runGitCommand(process.cwd(), 'git invalid-command-xyz');
			assert.fail('Should have thrown an error');
		} catch (error) {
			assert.ok(error, 'Should throw an error for invalid command');
		}
	});

	test('getRepoStatus should return status object with required fields', async () => {
		const status = await myExtension.getRepoStatus(process.cwd());
		
		assert.ok(status, 'Status should be defined');
		assert.ok(typeof status.branch === 'string', 'Branch should be a string');
		assert.ok(typeof status.dirty === 'boolean', 'Dirty should be a boolean');
		assert.ok(typeof status.ahead === 'number', 'Ahead should be a number');
		assert.ok(typeof status.behind === 'number', 'Behind should be a number');
	});

	test('getRepoStatus should detect current branch', async () => {
		const status = await myExtension.getRepoStatus(process.cwd());
		
		assert.notStrictEqual(status.branch, '', 'Branch should not be empty');
		assert.notStrictEqual(status.branch, 'unknown', 'Branch should be detected');
	});

	test('getRepoStatus should handle invalid repository path', async () => {
		const status = await myExtension.getRepoStatus('/non/existent/path');
		
		assert.strictEqual(status.branch, 'unknown', 'Should return unknown for invalid path');
		assert.strictEqual(status.dirty, false, 'Should return false for dirty status');
		assert.strictEqual(status.ahead, 0, 'Should return 0 for ahead count');
		assert.strictEqual(status.behind, 0, 'Should return 0 for behind count');
	});

	test('ahead and behind counts should be non-negative', async () => {
		const status = await myExtension.getRepoStatus(process.cwd());
		
		assert.ok(status.ahead >= 0, 'Ahead count should be non-negative');
		assert.ok(status.behind >= 0, 'Behind count should be non-negative');
	});
});
