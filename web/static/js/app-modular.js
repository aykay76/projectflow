/**
 * ProjectFlow - Main Application Entry Point
 * Orchestrates all modules and manages the overall application lifecycle
 */

import { showMessage, showLoadingOverlay, hideLoadingOverlay } from './utils.js';
import { apiClient } from './api-client.js';
import { stateManager } from './state-manager.js';
import { projectManager } from './project-manager.js';
import { taskManager } from './task-manager.js';
import { DragDropManager } from './drag-drop-manager.js';

/**
 * Main Application Class
 */
class ProjectFlowApp {
    constructor() {
        console.log('ProjectFlow app initializing...');
        
        // Use imported modules
        this.apiClient = apiClient;
        this.stateManager = stateManager;
        this.projectManager = projectManager;
        this.taskManager = taskManager;
        
        // Initialize drag and drop with dependencies
        this.dragDropManager = new DragDropManager(this.apiClient, this.stateManager);
        
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
        window.showMessage = showMessage;
        window.showLoadingOverlay = showLoadingOverlay;
        window.hideLoadingOverlay = hideLoadingOverlay;
        
        // Add convenient access to managers
        window.apiClient = this.apiClient;
        window.stateManager = this.stateManager;
        window.projectManager = this.projectManager;
        window.taskManager = this.taskManager;
        window.dragDropManager = this.dragDropManager;
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
        // Initialize drag and drop
        if (this.dragDropManager.init) this.dragDropManager.init();
        
        console.log('All modules initialized');
    }

    /**
     * Setup global event listeners
     */
    setupEventListeners() {
        // Window events
        window.addEventListener('resize', () => {
            console.log('Window resized');
        });

        console.log('Event listeners setup complete');
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
