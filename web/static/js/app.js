/**
 * ProjectFlow - Main Application Entry Point
 * Orchestrates all modules and manages the overall application lifecycle
 */

import { showMessage, showLoadingOverlay, hideLoadingOverlay } from './utils.js';
import { ApiClient } from './api-client.js';
import { stateManager } from './state-manager.js';
import { ProjectManager } from './project-manager.js';
import { TaskManager } from './task-manager.js';
import { TaskDetailManager } from './task-detail-manager.js';
import { DragDropManager } from './drag-drop-manager.js';
import { UIManager } from './ui-manager.js';
import { FilterManager } from './filter-manager.js';
import { NotificationManager } from './notification-manager.js';
import { ViewManager } from './view-manager.js';
import HeaderMenuManager from './header-menu-manager.js';

/**
 * Main Application Class
 */
class ProjectFlowApp {
    constructor() {
        console.log('ProjectFlow app initializing...');
        
        // Initialize core modules
        this.apiClient = new ApiClient();
        this.stateManager = stateManager; // Use singleton
        this.notificationManager = new NotificationManager();
        
        // Initialize feature modules
        this.projectManager = new ProjectManager(this.apiClient, this.stateManager);
        this.taskManager = new TaskManager(this.apiClient, this.stateManager, this.notificationManager);
        this.taskDetailManager = new TaskDetailManager(this.apiClient, this.stateManager);
        this.dragDropManager = new DragDropManager(this.apiClient, this.stateManager);
        this.uiManager = new UIManager(this.stateManager, this.taskManager);
        this.filterManager = new FilterManager(this.stateManager);
        this.viewManager = new ViewManager();
        this.headerMenuManager = new HeaderMenuManager();
        
        // Store header menu manager globally for debugging
        window.headerMenuManager = this.headerMenuManager;
        
        console.log('HeaderMenuManager created:', this.headerMenuManager);
        
        // Store references globally for compatibility
        window.projectFlowApp = this;
        this.setupGlobalReferences();
        
        this.init();
    }

    /**
     * Setup global references for backward compatibility
     */
    setupGlobalReferences() {
        // Make key functions available globally for event handlers and inline scripts
        window.createTask = this.taskManager.createTask?.bind(this.taskManager);
        window.updateTask = this.taskManager.updateTask?.bind(this.taskManager);
        window.deleteTask = this.taskManager.deleteTask?.bind(this.taskManager);
        window.moveTask = this.taskManager.moveTask?.bind(this.taskManager);
        
        window.createProject = this.projectManager.createProject?.bind(this.projectManager);
        window.switchProject = this.projectManager.switchProject?.bind(this.projectManager);
        
        window.showMessage = showMessage;
        window.showLoadingOverlay = showLoadingOverlay;
        window.hideLoadingOverlay = hideLoadingOverlay;
        
        window.switchView = this.viewManager.switchView?.bind(this.viewManager);
    }

    /**
     * Initialize the application
     */
    async init() {
        try {
            console.log('Initializing ProjectFlow application...');
            
            // Wait for DOM to be ready
            if (document.readyState === 'loading') {
                await new Promise(resolve => {
                    document.addEventListener('DOMContentLoaded', resolve);
                });
            }

            // Initialize modules in order
            await this.initializeModules();
            
            // Setup event listeners
            this.setupEventListeners();
            
            // Load initial data
            await this.loadInitialData();
            
            console.log('ProjectFlow application initialized successfully');
            
        } catch (error) {
            console.error('Failed to initialize ProjectFlow application:', error);
            showMessage('Failed to initialize application: ' + error.message, 'error');
        }
    }

    /**
     * Initialize all modules
     */
    async initializeModules() {
        // Core initialization - only call if methods exist
        if (this.stateManager.init) await this.stateManager.init();
        if (this.apiClient.init) await this.apiClient.init();
        
        // Feature initialization - only call if methods exist
        if (this.projectManager.init) await this.projectManager.init();
        if (this.taskManager.init) await this.taskManager.init();
        if (this.viewManager.init) await this.viewManager.init();
        
        // UI enhancements - only call if methods exist
        if (this.dragDropManager.init) this.dragDropManager.init();
        if (this.uiManager.init) this.uiManager.init();
        if (this.filterManager.init) this.filterManager.init();
        if (this.notificationManager.init) this.notificationManager.init();
        // HeaderMenuManager initializes itself in constructor
    }

    /**
     * Setup global event listeners
     */
    setupEventListeners() {
        // Project events - only setup if methods exist
        if (this.projectManager.on) {
            this.projectManager.on('project-changed', (project) => {
                if (this.viewManager.onProjectChanged) this.viewManager.onProjectChanged(project);
                if (this.taskManager.onProjectChanged) this.taskManager.onProjectChanged(project);
            });
        }

        // Task events - only setup if methods exist
        if (this.taskManager.on) {
            this.taskManager.on('task-created', (task) => {
                if (this.viewManager.onTaskCreated) this.viewManager.onTaskCreated(task);
                if (this.dragDropManager.refreshDraggableCards) this.dragDropManager.refreshDraggableCards();
            });

            this.taskManager.on('task-updated', (task) => {
                if (this.viewManager.onTaskUpdated) this.viewManager.onTaskUpdated(task);
            });

            this.taskManager.on('task-deleted', (taskId) => {
                if (this.viewManager.onTaskDeleted) this.viewManager.onTaskDeleted(taskId);
            });
        }

        // View change events - only setup if methods exist
        if (this.viewManager.on) {
            this.viewManager.on('view-changed', (viewName) => {
                this.stateManager.setCurrentView(viewName);
                if (this.dragDropManager.refreshDraggableCards) this.dragDropManager.refreshDraggableCards();
            });
        }
        
        // Window events
        window.addEventListener('resize', () => {
            if (this.viewManager.handleResize) this.viewManager.handleResize();
        });

        // Handle browser back/forward
        window.addEventListener('popstate', (event) => {
            if (event.state) {
                this.handlePopState(event.state);
            }
        });
    }

    /**
     * Load initial application data
     */
    async loadInitialData() {
        try {
            // Load projects first - only if method exists
            if (this.projectManager.loadAvailableProjects) {
                await this.projectManager.loadAvailableProjects();
            }
            
            // Load tasks for current project - only if method exists
            const currentProject = this.projectManager.getCurrentProject ? this.projectManager.getCurrentProject() : null;
            if (currentProject && this.taskManager.loadTasks) {
                await this.taskManager.loadTasks(currentProject.id);
            }
            
        } catch (error) {
            console.error('Failed to load initial data:', error);
            showMessage('Failed to load data: ' + error.message, 'error');
        }
    }

    /**
     * Handle browser popstate for navigation
     */
    handlePopState(state) {
        if (state.view && this.viewManager.switchView) {
            this.viewManager.switchView(state.view, false);
        }
        if (state.project && this.projectManager.switchProject) {
            this.projectManager.switchProject(state.project, false);
        }
    }

    /**
     * Get application state for debugging
     */
    getState() {
        return {
            currentProject: this.projectManager.getCurrentProject ? this.projectManager.getCurrentProject() : null,
            currentView: this.stateManager.getCurrentView(),
            tasksLoaded: this.taskManager.getTaskCount ? this.taskManager.getTaskCount() : 0,
            theme: this.uiManager.getCurrentTheme ? this.uiManager.getCurrentTheme() : 'light'
        };
    }

    /**
     * Clean up resources
     */
    destroy() {
        // Remove event listeners and clean up - only if methods exist
        if (this.projectManager.destroy) this.projectManager.destroy();
        if (this.taskManager.destroy) this.taskManager.destroy();
        if (this.viewManager.destroy) this.viewManager.destroy();
        if (this.dragDropManager.destroy) this.dragDropManager.destroy();
        if (this.filterManager.destroy) this.filterManager.destroy();
        
        // Clear global references
        delete window.projectFlowApp;
    }
}

// Initialize the application when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        new ProjectFlowApp();
    });
} else {
    new ProjectFlowApp();
}

// Export for module usage
export { ProjectFlowApp };
