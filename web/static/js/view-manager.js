/**
 * View management and navigation functionality
 */
import { stateManager } from './state-manager.js';
import { taskManager } from './task-manager.js';
import { showMessage } from './utils.js';

class ViewManager {
    constructor() {
        this.initializeEventListeners();
        this.initializeViewButtons();
    }

    initializeEventListeners() {
        // Listen for view changes
        stateManager.addEventListener('view-changed', (data) => {
            this.onViewChanged(data.view, data.previousView);
        });

        // Listen for project changes to refresh current view
        stateManager.addEventListener('project-changed', (data) => {
            if (data.newProject) {
                this.refreshCurrentView();
            }
        });

        // Listen for refresh requests
        stateManager.addEventListener('refresh-view', (data) => {
            this.loadView(data.view);
        });
    }

    initializeViewButtons() {
        // Add click handlers for view switching buttons
        const kanbanBtn = document.getElementById('kanban-view-btn');
        const hierarchyBtn = document.getElementById('hierarchy-view-btn');
        const timelineBtn = document.getElementById('timeline-view-btn');

        if (kanbanBtn) {
            kanbanBtn.addEventListener('click', () => this.switchToView('kanban'));
        }
        if (hierarchyBtn) {
            hierarchyBtn.addEventListener('click', () => this.switchToView('hierarchy'));
        }
        if (timelineBtn) {
            timelineBtn.addEventListener('click', () => this.switchToView('timeline'));
        }

        // Set initial view and ensure proper display state
        this.updateViewButtons();
        this.initializeViewDisplay();
    }

    initializeViewDisplay() {
        // Ensure only the current view is visible on initial load
        const currentView = stateManager.getCurrentView();
        const allViews = ['kanban', 'hierarchy', 'timeline'];
        
        allViews.forEach(view => {
            const viewContainer = document.getElementById(`${view}-view`);
            if (viewContainer) {
                if (view === currentView) {
                    viewContainer.style.display = 'block';
                } else {
                    viewContainer.style.display = 'none';
                }
            }
        });
    }

    switchToView(viewName) {
        if (!['kanban', 'hierarchy', 'timeline'].includes(viewName)) {
            console.error('Invalid view name:', viewName);
            return;
        }

        console.log(`Switching to ${viewName} view`);
        stateManager.setCurrentView(viewName);
    }

    onViewChanged(newView, previousView) {
        console.log(`View changed from ${previousView} to ${newView}`);
        
        this.updateViewButtons();
        this.showViewContent(newView);
        this.hideViewContent(previousView);
        this.loadView(newView);
    }

    updateViewButtons() {
        const currentView = stateManager.getCurrentView();
        
        // Update button states
        document.querySelectorAll('.view-btn').forEach(btn => {
            btn.classList.remove('active');
        });

        const activeBtn = document.getElementById(`${currentView}-view-btn`);
        if (activeBtn) {
            activeBtn.classList.add('active');
        }
    }

    showViewContent(viewName) {
        const viewContainer = document.getElementById(`${viewName}-view`);
        if (viewContainer) {
            viewContainer.style.display = 'block';
        }
    }

    hideViewContent(viewName) {
        if (!viewName) return;
        
        const viewContainer = document.getElementById(`${viewName}-view`);
        if (viewContainer) {
            viewContainer.style.display = 'none';
        }
    }

    async loadView(viewName) {
        const currentProject = stateManager.getCurrentProject();
        
        if (!currentProject) {
            console.warn(`No project selected for ${viewName} view`);
            this.showNoProjectMessage(viewName);
            return;
        }

        console.log(`Loading ${viewName} view for project:`, currentProject.name);

        try {
            const tasks = await taskManager.loadTasks();
            
            switch (viewName) {
                case 'kanban':
                    this.loadKanbanView(tasks);
                    break;
                case 'hierarchy':
                    this.loadHierarchyView(tasks);
                    break;
                case 'timeline':
                    this.loadTimelineView(tasks);
                    break;
                default:
                    console.error('Unknown view:', viewName);
            }
        } catch (error) {
            console.error(`Error loading ${viewName} view:`, error);
            showMessage(`Failed to load ${viewName} view`, 'error');
        }
    }

    loadKanbanView(tasks) {
        console.log('Loading kanban view with', tasks.length, 'tasks');
        
        // Define status columns
        const statusColumns = ['todo', 'in_progress', 'done', 'blocked'];
        
        statusColumns.forEach(status => {
            const taskList = document.getElementById(`${status.replace('_', '-')}-tasks`);
            if (!taskList) {
                console.warn(`Task list not found for status: ${status}`);
                return;
            }
            
            // Filter tasks by status
            const statusTasks = tasks.filter(task => task.status === status);
            console.log(`Found ${statusTasks.length} tasks with status: ${status}`);
            
            // Clear existing tasks
            taskList.innerHTML = '';
            
            // Add tasks to column
            statusTasks.forEach(task => {
                const taskCard = taskManager.createTaskCard(task);
                taskList.appendChild(taskCard);
            });
        });
        
        // Re-initialize drag and drop functionality
        this.initializeDragAndDrop();
    }

    loadHierarchyView(tasks) {
        console.log('Loading hierarchy view with', tasks.length, 'tasks');
        const hierarchyContainer = document.querySelector('.hierarchy-container');
        
        if (!hierarchyContainer) {
            console.error('Hierarchy container not found');
            return;
        }
        
        if (tasks.length === 0) {
            hierarchyContainer.innerHTML = '<div class="no-tasks-message">No tasks found for this project</div>';
            return;
        }
        
        taskManager.buildHierarchyTree(tasks, hierarchyContainer);
    }

    loadTimelineView(tasks) {
        console.log('Loading timeline view with', tasks.length, 'tasks');
        const timelineContainer = document.querySelector('.timeline-container');
        
        if (!timelineContainer) {
            console.error('Timeline container not found');
            return;
        }
        
        if (tasks.length === 0) {
            timelineContainer.innerHTML = '<div class="no-tasks-message">No tasks found for this project</div>';
            return;
        }
        
        taskManager.buildTimelineView(tasks, timelineContainer);
    }

    showNoProjectMessage(viewName) {
        const containers = {
            kanban: document.querySelector('.kanban-board'),
            hierarchy: document.querySelector('.hierarchy-container'),
            timeline: document.querySelector('.timeline-container')
        };
        
        const container = containers[viewName];
        if (container) {
            container.innerHTML = '<div class="error-message">Please select a project first</div>';
        }
    }

    refreshCurrentView() {
        const currentView = stateManager.getCurrentView();
        this.loadView(currentView);
    }

    initializeDragAndDrop() {
        // Re-initialize drag and drop functionality for task cards
        const taskCards = document.querySelectorAll('.task-card');
        taskCards.forEach(card => {
            card.addEventListener('dragstart', this.handleDragStart.bind(this));
            card.addEventListener('dragend', this.handleDragEnd.bind(this));
        });
        
        // Re-initialize drop zones
        const taskLists = document.querySelectorAll('.task-list');
        taskLists.forEach(list => {
            list.addEventListener('dragover', this.handleDragOver.bind(this));
            list.addEventListener('drop', this.handleDrop.bind(this));
        });
    }

    handleDragStart(event) {
        this.draggedTask = event.target;
        event.target.style.opacity = '0.5';
        event.target.style.cursor = 'grabbing';
        
        // Store task data for the drag operation
        event.dataTransfer.setData('text/plain', event.target.dataset.id);
        event.dataTransfer.effectAllowed = 'move';
        
        // Add visual feedback
        event.target.classList.add('dragging');
        
        // Add drag-over styles to all columns
        document.querySelectorAll('.column').forEach(column => {
            column.classList.add('drag-active');
        });
    }

    handleDragOver(event) {
        event.preventDefault();
        event.dataTransfer.dropEffect = 'move';
        
        const column = event.target.closest('.column');
        if (column && !column.classList.contains('drag-over')) {
            // Remove drag-over from other columns
            document.querySelectorAll('.column').forEach(col => {
                col.classList.remove('drag-over');
            });
            // Add to current column
            column.classList.add('drag-over');
        }
    }

    async handleDrop(event) {
        event.preventDefault();
        
        const taskId = event.dataTransfer.getData('text/plain');
        const targetColumn = event.target.closest('.column');
        
        if (targetColumn && this.draggedTask) {
            const newStatus = targetColumn.dataset.status;
            const currentStatus = this.draggedTask.closest('.column').dataset.status;
            
            // Only move if dropping in a different column
            if (newStatus !== currentStatus) {
                // Move the task visually first for immediate feedback
                const targetTaskList = targetColumn.querySelector('.task-list');
                targetTaskList.appendChild(this.draggedTask);
                
                // Update the task status via API
                try {
                    await taskManager.moveTask(taskId, newStatus);
                } catch (error) {
                    // If API call fails, the task manager will refresh the view
                    console.error('Failed to move task:', error);
                }
            }
        }
        
        // Clean up drag state
        this.cleanupDragState();
    }

    handleDragEnd(event) {
        this.cleanupDragState();
    }

    cleanupDragState() {
        if (this.draggedTask) {
            this.draggedTask.style.opacity = '';
            this.draggedTask.style.cursor = 'grab';
            this.draggedTask.classList.remove('dragging');
            this.draggedTask = null;
        }
        
        // Remove visual feedback from all columns
        document.querySelectorAll('.column').forEach(column => {
            column.classList.remove('drag-active', 'drag-over');
        });
    }

    // Keyboard shortcuts for view switching
    initializeKeyboardShortcuts() {
        document.addEventListener('keydown', (event) => {
            // Only process shortcuts when not in input fields
            if (event.target.tagName === 'INPUT' || event.target.tagName === 'TEXTAREA') {
                return;
            }
            
            // Number keys for view switching
            if (event.key >= '1' && event.key <= '3') {
                event.preventDefault();
                const views = ['kanban', 'hierarchy', 'timeline'];
                this.switchToView(views[parseInt(event.key) - 1]);
            }
        });
    }

    initializeViewDisplay() {
        // Ensure only the current view is visible on initial load
        const currentView = stateManager.getCurrentView();
        const allViews = ['kanban', 'hierarchy', 'timeline'];
        
        allViews.forEach(view => {
            const viewContainer = document.getElementById(`${view}-view`);
            if (viewContainer) {
                if (view === currentView) {
                    viewContainer.style.display = 'block';
                } else {
                    viewContainer.style.display = 'none';
                }
            }
        });
    }
}

// Export both the class and singleton instance
export { ViewManager };
export const viewManager = new ViewManager();
