/**
 * Centralized state management for the application
 */

class StateManager {
    constructor() {
        this.state = {
            // UI state
            currentView: 'kanban',
            currentTheme: localStorage.getItem('theme') || 'light',
            
            // Task state
            currentEditingTask: null,
            
            // Project state
            currentProject: null,
            availableProjects: [],
            isLoadingProjects: false,
            
            // Filter state
            filters: {
                search: '',
                status: '',
                priority: '',
                type: '',
                overdue: false
            }
        };
        
        this.listeners = {};
        this.loadSavedState();
    }

    // Event system for state changes
    addEventListener(event, callback) {
        if (!this.listeners[event]) {
            this.listeners[event] = [];
        }
        this.listeners[event].push(callback);
    }

    removeEventListener(event, callback) {
        if (!this.listeners[event]) return;
        const index = this.listeners[event].indexOf(callback);
        if (index > -1) {
            this.listeners[event].splice(index, 1);
        }
    }

    dispatchEvent(event, data) {
        if (!this.listeners[event]) return;
        this.listeners[event].forEach(callback => {
            try {
                callback(data);
            } catch (error) {
                console.error('Error in state listener:', error);
            }
        });
    }

    // State getters
    getCurrentView() {
        return this.state.currentView;
    }

    getCurrentTheme() {
        return this.state.currentTheme;
    }

    getCurrentEditingTask() {
        return this.state.currentEditingTask;
    }

    getCurrentProject() {
        return this.state.currentProject;
    }

    getAvailableProjects() {
        return this.state.availableProjects;
    }

    getFilters() {
        return { ...this.state.filters };
    }

    // State setters
    setCurrentView(view) {
        if (view !== this.state.currentView) {
            const previousView = this.state.currentView;
            this.state.currentView = view;
            localStorage.setItem('projectflow_current_view', view);
            this.dispatchEvent('view-changed', { view, previousView });
        }
    }

    setCurrentTheme(theme) {
        if (theme !== this.state.currentTheme) {
            this.state.currentTheme = theme;
            localStorage.setItem('theme', theme);
            document.documentElement.setAttribute('data-theme', theme);
            this.dispatchEvent('theme-changed', { theme });
        }
    }

    toggleTheme() {
        const newTheme = this.state.currentTheme === 'light' ? 'dark' : 'light';
        this.setCurrentTheme(newTheme);
    }

    setCurrentEditingTask(task) {
        this.state.currentEditingTask = task;
        this.dispatchEvent('editing-task-changed', { task });
    }

    setCurrentProject(project) {
        if (project?.id !== this.state.currentProject?.id) {
            const previousProject = this.state.currentProject;
            this.state.currentProject = project;
            
            if (project) {
                localStorage.setItem('projectflow_current_project', project.id);
            } else {
                localStorage.removeItem('projectflow_current_project');
            }
            
            this.dispatchEvent('project-changed', { 
                newProject: project, 
                previousProject 
            });
        }
    }

    setAvailableProjects(projects) {
        this.state.availableProjects = projects;
        this.dispatchEvent('projects-refreshed', { projects });
    }

    setProjectsLoading(loading) {
        this.state.isLoadingProjects = loading;
        this.dispatchEvent('projects-loading-changed', { loading });
    }

    updateFilters(filters) {
        const previousFilters = { ...this.state.filters };
        this.state.filters = { ...this.state.filters, ...filters };
        this.saveFilterState();
        this.dispatchEvent('filters-changed', { 
            filters: this.state.filters, 
            previousFilters 
        });
    }

    clearAllFilters() {
        const previousFilters = { ...this.state.filters };
        this.state.filters = {
            search: '',
            status: '',
            priority: '',
            type: '',
            overdue: false
        };
        this.saveFilterState();
        this.dispatchEvent('filters-cleared', { previousFilters });
    }

    // Persistence
    loadSavedState() {
        // Load view preference
        const savedView = localStorage.getItem('projectflow_current_view');
        if (savedView && ['kanban', 'hierarchy', 'timeline'].includes(savedView)) {
            this.state.currentView = savedView;
        }

        // Load filter state
        this.loadFilterState();

        // Load project preference
        const savedProjectId = localStorage.getItem('projectflow_current_project');
        if (savedProjectId) {
            // This will be resolved when projects are loaded
            this.savedProjectId = savedProjectId;
        }
    }

    saveFilterState() {
        localStorage.setItem('projectflow_filters', JSON.stringify(this.state.filters));
    }

    loadFilterState() {
        const saved = localStorage.getItem('projectflow_filters');
        if (saved) {
            try {
                this.state.filters = { ...this.state.filters, ...JSON.parse(saved) };
            } catch (error) {
                console.error('Error loading filter state:', error);
            }
        }
    }

    // Auto-save for forms
    saveFormDraft(formName, data) {
        localStorage.setItem(`${formName}_draft`, JSON.stringify(data));
    }

    loadFormDraft(formName) {
        const saved = localStorage.getItem(`${formName}_draft`);
        if (saved) {
            try {
                return JSON.parse(saved);
            } catch (error) {
                console.error('Error loading form draft:', error);
            }
        }
        return null;
    }

    clearFormDraft(formName) {
        localStorage.removeItem(`${formName}_draft`);
    }

    // Helper method to restore saved project after projects are loaded
    restoreSavedProject(availableProjects) {
        if (this.savedProjectId) {
            const savedProject = availableProjects.find(p => p.id === this.savedProjectId);
            if (savedProject) {
                this.setCurrentProject(savedProject);
            }
            delete this.savedProjectId;
        }
    }
}

// Export singleton instance
export const stateManager = new StateManager();
