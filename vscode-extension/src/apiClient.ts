import axios, { AxiosInstance } from 'axios';
import * as vscode from 'vscode';

export interface Task {
	id: string;
	display_id?: string;
	project_id: string;
	title: string;
	description: string;
	status: 'todo' | 'in_progress' | 'done' | 'blocked';
	priority: 'low' | 'medium' | 'high' | 'critical';
	type: 'epic' | 'story' | 'task' | 'subtask';
	parent_id?: string;
	children: string[];
	created_at: string;
	updated_at: string;
	started_at?: string;
}

export interface HierarchyTask extends Task {
	child_tasks: HierarchyTask[];
}

export interface Project {
	id: string;
	name: string;
	description: string;
	display_prefix: string;
	settings: any;
	created_at: string;
	updated_at: string;
}

export interface CreateTaskRequest {
	title: string;
	description?: string;
	status?: string;
	priority?: string;
	type?: string;
	parent_id?: string;
	project_id?: string;
}

export interface UpdateTaskRequest {
	title?: string;
	description?: string;
	status?: string;
	priority?: string;
	type?: string;
	parent_id?: string;
}

export class ProjectFlowApiClient {
	private client: AxiosInstance;
	private baseUrl!: string;
	private project!: string;
	private apiKey?: string;

	constructor() {
		console.log('[ProjectFlowApiClient] Constructor called');
		this.updateConfiguration();
		console.log('[ProjectFlowApiClient] Initial configuration loaded:', {
			baseUrl: this.baseUrl,
			project: this.project,
			hasApiKey: !!this.apiKey
		});
		
		this.client = axios.create({
			baseURL: this.baseUrl,
			timeout: 10000,
			headers: {
				'Content-Type': 'application/json',
			},
		});
		console.log('[ProjectFlowApiClient] HTTP client created with baseURL:', this.baseUrl);

		// Add request interceptor for authentication
		this.client.interceptors.request.use((config) => {
			if (this.apiKey) {
				config.headers['Authorization'] = `Bearer ${this.apiKey}`;
			}
			return config;
		});

		// Add response interceptor for error handling
		this.client.interceptors.response.use(
			(response) => response,
			(error) => {
				console.error('[ProjectFlowApiClient] HTTP request failed:', error);
				return Promise.reject(error);
			}
		);
		
		console.log('[ProjectFlowApiClient] Constructor completed');
	}

	public updateConfiguration(): void {
		const config = vscode.workspace.getConfiguration('projectflow');
		const oldBaseUrl = this.baseUrl;
		const oldProject = this.project;
		const oldApiKey = this.apiKey;
		
		this.baseUrl = config.get<string>('serverUrl', 'http://localhost:16191');
		this.project = config.get<string>('project', 'PF');
		this.apiKey = config.get<string>('apiKey');

		console.log('[ProjectFlowApiClient] Configuration updated:', {
			serverUrl: { old: oldBaseUrl, new: this.baseUrl },
			project: { old: oldProject, new: this.project },
			hasApiKey: { old: !!oldApiKey, new: !!this.apiKey }
		});

		if (this.client) {
			this.client.defaults.baseURL = this.baseUrl;
			console.log('[ProjectFlowApiClient] HTTP client baseURL updated to:', this.baseUrl);
		}
	}

	async getProjects(): Promise<Project[]> {
		try {
			const response = await this.client.get<Project[]>('/api/projects');
			return response.data;
		} catch (error) {
			throw this.handleError(error, 'Failed to fetch projects');
		}
	}

	async getTasks(): Promise<Task[]> {
		try {
			// Include project filtering in the tasks request
			const response = await this.client.get<Task[]>(`/api/tasks?project_id=${this.project}`);
			return response.data;
		} catch (error) {
			throw this.handleError(error, 'Failed to fetch tasks');
		}
	}

	async getTaskHierarchy(): Promise<HierarchyTask[]> {
		try {
			console.log('[ProjectFlowApiClient] Requesting task hierarchy for project:', this.project);
			console.log('[ProjectFlowApiClient] Request URL:', `${this.baseUrl}/api/hierarchy?project_id=${this.project}`);
			
			// Include project filtering in the hierarchy request
			const response = await this.client.get<HierarchyTask[]>(`/api/hierarchy?project_id=${this.project}`);
			
			console.log('[ProjectFlowApiClient] Task hierarchy response status:', response.status);
			console.log('[ProjectFlowApiClient] Task hierarchy response data length:', response.data.length);
			console.log('[ProjectFlowApiClient] Task hierarchy response data:', JSON.stringify(response.data, null, 2));
			
			return response.data;
		} catch (error) {
			console.error('[ProjectFlowApiClient] Error fetching task hierarchy:', error);
			if (axios.isAxiosError(error)) {
				console.error('[ProjectFlowApiClient] Axios error details:', {
					status: error.response?.status,
					statusText: error.response?.statusText,
					data: error.response?.data,
					headers: error.response?.headers
				});
			}
			throw this.handleError(error, 'Failed to fetch task hierarchy');
		}
	}

	async getTask(id: string): Promise<Task> {
		try {
			const response = await this.client.get<Task>(`/api/tasks/${id}`);
			return response.data;
		} catch (error) {
			throw this.handleError(error, `Failed to fetch task ${id}`);
		}
	}

	async createTask(task: CreateTaskRequest): Promise<Task> {
		try {
			const taskData = {
				...task,
				project_id: task.project_id || this.project,
			};
			const response = await this.client.post<Task>('/api/tasks', taskData);
			return response.data;
		} catch (error) {
			throw this.handleError(error, 'Failed to create task');
		}
	}

	async updateTask(id: string, updates: UpdateTaskRequest): Promise<Task> {
		try {
			const response = await this.client.put<Task>(`/api/tasks/${id}`, updates);
			return response.data;
		} catch (error) {
			throw this.handleError(error, `Failed to update task ${id}`);
		}
	}

	async deleteTask(id: string): Promise<void> {
		try {
			await this.client.delete(`/api/tasks/${id}`);
		} catch (error) {
			throw this.handleError(error, `Failed to delete task ${id}`);
		}
	}

	async checkConnection(): Promise<boolean> {
		try {
			const response = await this.client.get('/health');
			return response.status === 200;
		} catch (error) {
			return false;
		}
	}

	async getConnectionStatus(): Promise<{ connected: boolean; error?: string }> {
		try {
			console.log('[ProjectFlowApiClient] Checking connection to:', this.baseUrl);
			const response = await this.client.get('/health');
			console.log('[ProjectFlowApiClient] Health check response status:', response.status);
			console.log('[ProjectFlowApiClient] Health check response data:', response.data);
			return { connected: response.status === 200 };
		} catch (error) {
			console.error('[ProjectFlowApiClient] Health check failed:', error);
			if (axios.isAxiosError(error)) {
				console.error('[ProjectFlowApiClient] Health check axios error details:', {
					status: error.response?.status,
					statusText: error.response?.statusText,
					data: error.response?.data,
					code: error.code,
					message: error.message
				});
			}
			return { 
				connected: false, 
				error: this.handleError(error, 'Connection failed').message 
			};
		}
	}

	getCurrentProject(): string {
		return this.project;
	}

	getServerUrl(): string {
		return this.baseUrl;
	}

	private handleError(error: any, message: string): Error {
		if (axios.isAxiosError(error)) {
			if (error.response) {
				return new Error(`${message}: ${error.response.status} ${error.response.statusText}`);
			} else if (error.request) {
				return new Error(`${message}: No response from server. Check if ProjectFlow server is running at ${this.baseUrl}`);
			}
		}
		return new Error(`${message}: ${error.message || error}`);
	}
}
