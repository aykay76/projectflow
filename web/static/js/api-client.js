/**
 * Centralized API client for all backend communication
 */
import { showMessage, showLoadingOverlay, hideLoadingOverlay } from './utils.js';

class ApiClient {
    constructor() {
        this.baseUrl = '/api';
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseUrl}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`HTTP ${response.status}: ${errorText}`);
            }

            return await response.json();
        } catch (error) {
            console.error(`API request failed: ${endpoint}`, error);
            throw error;
        }
    }

    // Task operations
    async loadTasks(projectId = null) {
        try {
            const endpoint = projectId ? `/tasks?project_id=${encodeURIComponent(projectId)}` : '/tasks';
            return await this.request(endpoint);
        } catch (error) {
            showMessage('Failed to load tasks', 'error');
            throw error;
        }
    }

    async createTask(taskData) {
        try {
            return await this.request('/tasks', {
                method: 'POST',
                body: JSON.stringify(taskData)
            });
        } catch (error) {
            showMessage('Failed to create task', 'error');
            throw error;
        }
    }

    async updateTask(taskId, taskData) {
        try {
            return await this.request(`/tasks/${taskId}`, {
                method: 'PUT',
                body: JSON.stringify(taskData)
            });
        } catch (error) {
            showMessage('Failed to update task', 'error');
            throw error;
        }
    }

    async deleteTask(taskId) {
        try {
            await this.request(`/tasks/${taskId}`, {
                method: 'DELETE'
            });
        } catch (error) {
            showMessage('Failed to delete task', 'error');
            throw error;
        }
    }

    async getTask(taskId) {
        try {
            return await this.request(`/tasks/${taskId}`);
        } catch (error) {
            showMessage('Failed to load task', 'error');
            throw error;
        }
    }

    async moveTask(taskId, newStatus) {
        try {
            return await this.updateTask(taskId, { status: newStatus });
        } catch (error) {
            showMessage('Failed to move task', 'error');
            throw error;
        }
    }

    // Project operations
    async loadProjects() {
        try {
            return await this.request('/projects');
        } catch (error) {
            showMessage('Failed to load projects', 'error');
            throw error;
        }
    }

    async createProject(projectData) {
        try {
            return await this.request('/projects', {
                method: 'POST',
                body: JSON.stringify(projectData)
            });
        } catch (error) {
            showMessage('Failed to create project', 'error');
            throw error;
        }
    }

    async updateProject(projectId, projectData) {
        try {
            return await this.request(`/projects/${projectId}`, {
                method: 'PUT',
                body: JSON.stringify(projectData)
            });
        } catch (error) {
            showMessage('Failed to update project', 'error');
            throw error;
        }
    }

    async deleteProject(projectId) {
        try {
            await this.request(`/projects/${projectId}`, {
                method: 'DELETE'
            });
        } catch (error) {
            showMessage('Failed to delete project', 'error');
            throw error;
        }
    }

    async getProject(projectId) {
        try {
            return await this.request(`/projects/${projectId}`);
        } catch (error) {
            showMessage('Failed to load project', 'error');
            throw error;
        }
    }
}

// Export both the class and singleton instance
export { ApiClient };
export const apiClient = new ApiClient();
