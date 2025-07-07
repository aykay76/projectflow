import * as assert from 'assert';
import * as vscode from 'vscode';
import { ProjectFlowApiClient } from '../apiClient';

suite('ProjectFlow Extension Test Suite', () => {
	vscode.window.showInformationMessage('Start all tests.');

	test('API Client Configuration', () => {
		const client = new ProjectFlowApiClient();
		assert.ok(client, 'API client should be created');
		
		// Test configuration methods
		assert.strictEqual(typeof client.getCurrentProject(), 'string');
		assert.strictEqual(typeof client.getServerUrl(), 'string');
	});

	test('API Client Connection Test', async () => {
		const client = new ProjectFlowApiClient();
		
		// Test connection status method
		const status = await client.getConnectionStatus();
		assert.ok(typeof status.connected === 'boolean', 'Connection status should return boolean');
		
		if (!status.connected) {
			assert.ok(status.error, 'Should have error message when disconnected');
		}
	});

	test('Extension Activation', async () => {
		// Test that extension is activated
		const extension = vscode.extensions.getExtension('aykay76.projectflow');
		assert.ok(extension, 'Extension should be found');
		
		if (!extension.isActive) {
			await extension.activate();
		}
		
		assert.ok(extension.isActive, 'Extension should be activated');
	});

	test('Commands Registration', async () => {
		// Test that commands are registered
		const commands = await vscode.commands.getCommands();
		
		const expectedCommands = [
			'projectflow.refreshTasks',
			'projectflow.createTask',
			'projectflow.editTask',
			'projectflow.deleteTask',
			'projectflow.markTaskComplete',
			'projectflow.markTaskInProgress',
			'projectflow.openTask',
			'projectflow.checkConnection',
			'projectflow.switchProject'
		];

		expectedCommands.forEach(command => {
			assert.ok(commands.includes(command), `Command ${command} should be registered`);
		});
	});

	test('Configuration Access', () => {
		const config = vscode.workspace.getConfiguration('projectflow');
		
		// Test that configuration properties exist
		assert.ok(config.has('enabled'), 'Should have enabled setting');
		assert.ok(config.has('serverUrl'), 'Should have serverUrl setting');
		assert.ok(config.has('project'), 'Should have project setting');
		assert.ok(config.has('refreshInterval'), 'Should have refreshInterval setting');
	});
});
