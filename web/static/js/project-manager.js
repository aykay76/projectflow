/**
 * Project management functionality
 */
import { apiClient } from './api-client.js';
import { stateManager } from './state-manager.js';
import { showMessage } from './utils.js';

class ProjectManager {
    constructor() {
        this.isToggling = false; // Add flag to prevent rapid toggling
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
        console.log('ProjectManager: setupDOMEventListeners called, document.readyState:', document.readyState);
        // Wait for DOM to be ready
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => {
                console.log('ProjectManager: DOMContentLoaded event fired');
                // Add additional delay to ensure header menu is initialized
                setTimeout(() => {
                    this.attachProjectDropdownListeners();
                }, 100);
            });
        } else {
            console.log('ProjectManager: DOM already ready, attaching listeners');
            // Add delay to ensure other components are initialized
            setTimeout(() => {
                this.attachProjectDropdownListeners();
            }, 100);
        }
    }

    attachProjectDropdownListeners() {
        console.log('attachProjectDropdownListeners called');
        const projectSelectorBtn = document.getElementById('project-selector-btn');
        console.log('Project selector button element:', projectSelectorBtn);
        
        if (projectSelectorBtn) {
            // Remove any existing listeners to prevent duplicates
            const newBtn = projectSelectorBtn.cloneNode(true);
            projectSelectorBtn.parentNode.replaceChild(newBtn, projectSelectorBtn);
            
            // Add the event listener to the new element
            newBtn.addEventListener('click', (e) => {
                console.log('Project selector button clicked!');
                e.preventDefault();
                e.stopPropagation();
                e.stopImmediatePropagation();
                
                // Prevent rapid double-clicks
                if (this.isToggling) {
                    console.log('Already toggling, ignoring click');
                    return;
                }
                
                this.isToggling = true;
                setTimeout(() => {
                    this.isToggling = false;
                }, 300);
                
                this.toggleProjectDropdown();
            });
            console.log('Project selector button event listener attached');
        } else {
            console.error('Project selector button not found in DOM');
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
            
            // If no current project is set after restoration, set the first available project
            const currentProject = stateManager.getCurrentProject();
            if (!currentProject && projects.length > 0) {
                console.log('No current project set, using first available project');
                stateManager.setCurrentProject(projects[0]);
            }
            
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
            console.error('project-list element not found');
            return;
        }
        
        const projects = stateManager.getAvailableProjects();
        const currentProject = stateManager.getCurrentProject();
        const isLoading = stateManager.state.isLoadingProjects;
        
        console.log('Available projects:', projects);
        console.log('Current project:', currentProject);
        console.log('Is loading:', isLoading);
        
        if (isLoading) {
            projectList.innerHTML = '<div class="project-loading">Loading projects...</div>';
            return;
        }
        
        if (projects.length === 0) {
            console.log('No projects available');
            projectList.innerHTML = `
                <div class="project-empty">
                    <p>No projects found</p>
                    <p>Create your first project to get started!</p>
                </div>
            `;
            return;
        }
        
        console.log('Rendering project list with', projects.length, 'projects');
        projectList.innerHTML = projects.map(project => `
            <div class="project-item ${currentProject && currentProject.id === project.id ? 'selected' : ''}" 
                 data-project-id="${project.id}">
                <div class="project-item-name">${this.escapeHtml(project.name)}</div>
                <div class="project-item-description">${this.escapeHtml(project.description || 'No description')}</div>
            </div>
        `).join('');
        
        console.log('Project list HTML updated');
        
        // Add click handlers for project items
        const projectItems = projectList.querySelectorAll('.project-item');
        console.log('Adding click handlers to', projectItems.length, 'project items');
        projectItems.forEach(item => {
            item.addEventListener('click', () => {
                const projectId = item.dataset.projectId;
                console.log('Project item clicked, projectId:', projectId);
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
        console.log('toggleProjectDropdown called');
        const projectDropdown = document.getElementById('project-dropdown');
        const projectSelector = document.getElementById('project-selector-btn');
        
        console.log('Project dropdown element:', projectDropdown);
        console.log('Project selector element:', projectSelector);
        
        if (!projectDropdown || !projectSelector) {
            console.error('Required elements not found for project dropdown');
            return;
        }
        
        // Check current state more reliably
        const isCurrentlyVisible = projectDropdown.style.display === 'block';
        console.log('Dropdown is currently visible:', isCurrentlyVisible);
        
        if (isCurrentlyVisible) {
            console.log('Closing dropdown');
            this.closeProjectDropdown();
        } else {
            console.log('Opening dropdown');
            this.openProjectDropdown();
        }
    }

    openProjectDropdown() {
        console.log('openProjectDropdown called');
        const projectDropdown = document.getElementById('project-dropdown');
        const projectSelector = document.getElementById('project-selector-btn');
        
        if (!projectDropdown || !projectSelector) {
            console.error('Cannot open dropdown - elements not found');
            return;
        }
        
        console.log('Setting dropdown display to block');
        projectDropdown.style.display = 'block';
        projectSelector.classList.add('open');
        this.updateProjectDropdown();
        console.log('Dropdown opened successfully');
    }

    closeProjectDropdown() {
        console.log('closeProjectDropdown called');
        const projectDropdown = document.getElementById('project-dropdown');
        const projectSelector = document.getElementById('project-selector-btn');
        
        if (!projectDropdown || !projectSelector) return;
        
        projectDropdown.style.display = 'none';
        projectSelector.classList.remove('open');
        console.log('Dropdown closed');
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
