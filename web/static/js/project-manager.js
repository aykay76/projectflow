/**
 * Project management functionality
 */
import { apiClient } from './api-client.js';
import { stateManager } from './state-manager.js';
import { showMessage } from './utils.js';

class ProjectManager {
    constructor() {
        console.log('ProjectManager constructor called');
        this.initializeEventListeners();
        this.setupDOMEventListeners();
    }

    initializeEventListeners() {
        // Listen for state changes
        stateManager.addEventListener('project-changed', (data) => {
            this.onProjectChanged(data.newProject, data.previousProject);
        });

        stateManager.addEventListener('projects-refreshed', (data) => {
            this.onProjectsRefreshed(data.projects);
        });
    }

    setupDOMEventListeners() {
        // Wait for DOM to be ready
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => {
                this.attachProjectDropdownListeners();
            });
        } else {
            this.attachProjectDropdownListeners();
        }
    }

    attachProjectDropdownListeners() {
        const projectSelectorBtn = document.getElementById('project-selector-btn');
        if (projectSelectorBtn) {
            projectSelectorBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                this.toggleProjectDropdown();
            });
            console.log('Project selector button event listener attached');
        } else {
            console.warn('Project selector button not found in DOM');
        }

        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            const projectDropdown = document.getElementById('project-dropdown');
            const projectSelector = document.getElementById('project-selector-btn');
            
            if (projectDropdown && projectSelector) {
                const isClickInsideDropdown = projectDropdown.contains(e.target);
                const isClickOnSelector = projectSelector.contains(e.target);
                
                if (!isClickInsideDropdown && !isClickOnSelector) {
                    this.closeProjectDropdown();
                }
            }
        });

        // Set up create project button
        const createProjectBtn = document.getElementById('create-project-btn');
        if (createProjectBtn) {
            createProjectBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                this.showCreateProjectDialog();
            });
        }

        // Set up manage projects button
        const manageProjectsBtn = document.getElementById('manage-projects-btn');
        if (manageProjectsBtn) {
            manageProjectsBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                this.showManageProjectsDialog();
            });
        }
    }

    async loadAvailableProjects() {
        console.log('loadAvailableProjects called, loading state:', stateManager.state.isLoadingProjects);
        if (stateManager.state.isLoadingProjects) return;
        
        stateManager.setProjectsLoading(true);
        
        try {
            console.log('Loading available projects...');
            const projects = await apiClient.loadProjects();
            console.log('Loaded projects:', projects);
            
            stateManager.setAvailableProjects(projects);
            this.updateProjectDropdown();
            
            // Restore saved project if available
            stateManager.restoreSavedProject(projects);
            
            return projects;
        } catch (error) {
            console.error('Error loading projects:', error);
            stateManager.setAvailableProjects([]);
        } finally {
            stateManager.setProjectsLoading(false);
        }
    }

    async createProject(projectData) {
        try {
            const newProject = await apiClient.createProject(projectData);
            
            // Refresh projects list
            await this.loadAvailableProjects();
            
            // Set as current project
            stateManager.setCurrentProject(newProject);
            
            showMessage(`Project "${newProject.name}" created successfully!`, 'success');
            return newProject;
        } catch (error) {
            console.error('Error creating project:', error);
            throw error;
        }
    }

    async updateProject(projectId, projectData) {
        try {
            const updatedProject = await apiClient.updateProject(projectId, projectData);
            
            // Update local cache
            const projects = stateManager.getAvailableProjects();
            const index = projects.findIndex(p => p.id === projectId);
            if (index > -1) {
                projects[index] = updatedProject;
                stateManager.setAvailableProjects([...projects]);
            }
            
            // Update current project if it's the one being updated
            const currentProject = stateManager.getCurrentProject();
            if (currentProject && currentProject.id === projectId) {
                stateManager.setCurrentProject(updatedProject);
            }
            
            this.updateProjectDropdown();
            showMessage(`Project "${updatedProject.name}" updated successfully!`, 'success');
            return updatedProject;
        } catch (error) {
            console.error('Error updating project:', error);
            throw error;
        }
    }

    async deleteProject(projectId) {
        try {
            await apiClient.deleteProject(projectId);
            
            // Refresh projects
            await this.loadAvailableProjects();
            
            // If we deleted the current project, switch to another one
            const currentProject = stateManager.getCurrentProject();
            if (currentProject?.id === projectId) {
                const projects = stateManager.getAvailableProjects();
                if (projects.length > 0) {
                    stateManager.setCurrentProject(projects[0]);
                } else {
                    stateManager.setCurrentProject(null);
                }
            }
            
            showMessage('Project deleted successfully', 'success');
        } catch (error) {
            console.error('Error deleting project:', error);
            throw error;
        }
    }

    async switchToProject(projectId) {
        const projects = stateManager.getAvailableProjects();
        const project = projects.find(p => p.id === projectId);
        if (project) {
            stateManager.setCurrentProject(project);
            showMessage(`Switched to project: ${project.name}`, 'success');
        }
    }

    async handleInitialProjectSetup() {
        const projects = stateManager.getAvailableProjects();
        
        if (projects.length > 0) {
            // Use first available project
            stateManager.setCurrentProject(projects[0]);
        } else {
            // Create a default project
            try {
                const defaultProject = await this.createProject({
                    name: 'Default Project',
                    description: 'Default project for task management',
                    display_prefix: 'PF'
                });
                showMessage('Created default project for you!', 'success');
            } catch (error) {
                console.error('Failed to create default project:', error);
                showMessage('Failed to create default project', 'error');
            }
        }
    }

    updateProjectDropdown() {
        console.log('updateProjectDropdown called');
        const projectList = document.getElementById('project-list');
        if (!projectList) {
            console.warn('project-list element not found');
            return;
        }
        
        const projects = stateManager.getAvailableProjects();
        const currentProject = stateManager.getCurrentProject();
        const isLoading = stateManager.state.isLoadingProjects;
        
        console.log('Updating dropdown with projects:', projects);
        console.log('Current project:', currentProject);
        console.log('Is loading:', isLoading);
        
        if (isLoading) {
            projectList.innerHTML = '<div class="project-loading">Loading projects...</div>';
            return;
        }
        
        if (projects.length === 0) {
            projectList.innerHTML = `
                <div class="project-empty">
                    <p>No projects found</p>
                    <p>Create your first project to get started!</p>
                </div>
            `;
            return;
        }
        
        projectList.innerHTML = projects.map(project => `
            <div class="project-item ${currentProject && currentProject.id === project.id ? 'selected' : ''}" 
                 data-project-id="${project.id}">
                <div class="project-item-name">${this.escapeHtml(project.name)}</div>
                <div class="project-item-description">${this.escapeHtml(project.description || 'No description')}</div>
            </div>
        `).join('');
        
        // Add click handlers for project items
        projectList.querySelectorAll('.project-item').forEach(item => {
            item.addEventListener('click', () => {
                const projectId = item.dataset.projectId;
                this.switchToProject(projectId);
                this.closeProjectDropdown();
            });
        });
    }

    updateCurrentProjectDisplay() {
        const currentProjectDisplay = document.getElementById('current-project-display');
        const currentProject = stateManager.getCurrentProject();
        
        if (currentProjectDisplay) {
            if (currentProject) {
                currentProjectDisplay.textContent = `Current: ${currentProject.name}`;
            } else {
                currentProjectDisplay.textContent = 'No Project Selected';
            }
        }
    }

    updateProjectSelectorButton() {
        const selectorText = document.getElementById('project-selector-text');
        const currentProject = stateManager.getCurrentProject();
        
        if (selectorText) {
            if (currentProject) {
                selectorText.textContent = currentProject.name;
            } else {
                selectorText.textContent = 'Select Project';
            }
        }
    }

    toggleProjectDropdown() {
        const projectDropdown = document.getElementById('project-dropdown');
        const projectSelector = document.getElementById('project-selector-btn');
        
        if (!projectDropdown || !projectSelector) return;
        
        const isOpen = projectDropdown.style.display !== 'none';
        if (isOpen) {
            this.closeProjectDropdown();
        } else {
            this.openProjectDropdown();
        }
    }

    openProjectDropdown() {
        const projectDropdown = document.getElementById('project-dropdown');
        const projectSelector = document.getElementById('project-selector-btn');
        
        if (!projectDropdown || !projectSelector) return;
        
        projectDropdown.style.display = 'block';
        projectSelector.classList.add('open');
        this.updateProjectDropdown();
    }

    closeProjectDropdown() {
        const projectDropdown = document.getElementById('project-dropdown');
        const projectSelector = document.getElementById('project-selector-btn');
        
        if (!projectDropdown || !projectSelector) return;
        
        projectDropdown.style.display = 'none';
        projectSelector.classList.remove('open');
    }

    // Dialog methods
    showCreateProjectDialog() {
        // TODO: Implement create project dialog
        console.log('Create project dialog - to be implemented');
        showMessage('Create project dialog - coming soon!', 'info');
        this.closeProjectDropdown();
    }

    showManageProjectsDialog() {
        // TODO: Implement manage projects dialog
        console.log('Manage projects dialog - to be implemented');
        showMessage('Manage projects dialog - coming soon!', 'info');
        this.closeProjectDropdown();
    }

    // Event handlers
    onProjectChanged(newProject, previousProject) {
        console.log('Project changed:', newProject?.name || 'None');
        this.updateCurrentProjectDisplay();
        this.updateProjectSelectorButton();
        
        // Notify other parts of the app that project has changed
        // This will trigger task reloading in other modules
    }

    onProjectsRefreshed(projects) {
        console.log('Projects refreshed, count:', projects.length);
        this.updateProjectDropdown();
    }

    // Utility methods
    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Validation methods
    async validateOnDemand() {
        const currentProject = stateManager.getCurrentProject();
        if (!currentProject) {
            await this.handleInitialProjectSetup();
            return false;
        }
        
        return await this.validateCurrentProject();
    }

    async validateCurrentProject() {
        const currentProject = stateManager.getCurrentProject();
        if (!currentProject) {
            console.warn('No current project to validate');
            return false;
        }
        
        try {
            const response = await fetch(`/api/projects/${currentProject.id}`);
            if (!response.ok) {
                // Only handle actual 404s (project deleted), not other HTTP errors
                if (response.status === 404) {
                    console.warn('Current project no longer exists, switching to default');
                    await this.handleInitialProjectSetup();
                    return false;
                } else {
                    // For other errors (500, network issues, etc.), just log and continue
                    console.warn(`Project validation failed with status ${response.status}, but continuing with current project`);
                    return true;
                }
            }
            return true;
        } catch (error) {
            // Network errors, etc. - don't switch projects, just log
            console.warn('Error validating current project (network/connectivity issue):', error.message);
            return true; // Assume project is still valid, network might be temporarily down
        }
    }

    // Utility methods for compatibility
    getCurrentProject() {
        return stateManager.getCurrentProject();
    }

    getAvailableProjects() {
        return stateManager.getAvailableProjects();
    }
}

// Export both the class and singleton instance
export { ProjectManager };
export const projectManager = new ProjectManager();
