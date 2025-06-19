/**
 * DragDropManager - Handles all drag and drop functionality for task cards
 * 
 * Features:
 * - Task card dragging between columns
 * - Visual feedback during drag operations
 * - Status updates via API
 * - Cleanup and state management
 */

export class DragDropManager {
    constructor(apiClient, stateManager) {
        this.apiClient = apiClient;
        this.stateManager = stateManager;
        this.draggedTask = null;
        
        this.init();
    }

    /**
     * Initialize drag and drop event listeners
     */
    init() {
        // Global dragend event listener for cleanup
        document.addEventListener('dragend', (event) => {
            if (event.target.classList.contains('task-card')) {
                this.cleanupDragState();
            }
        });

        // Initialize existing cards
        this.initializeDraggableCards();
    }

    /**
     * Initialize draggable attributes for all task cards
     */
    initializeDraggableCards() {
        const taskCards = document.querySelectorAll('.task-card');
        taskCards.forEach(card => {
            this.setupTaskCardDragEvents(card);
        });
    }

    /**
     * Setup drag events for a single task card
     * @param {HTMLElement} card - The task card element
     */
    setupTaskCardDragEvents(card) {
        card.setAttribute('draggable', 'true');
        card.style.cursor = 'grab';
        
        // Remove existing listeners to prevent duplicates
        card.removeEventListener('dragstart', this.handleDragStart.bind(this));
        
        // Add drag event listeners
        card.addEventListener('dragstart', this.handleDragStart.bind(this));
    }

    /**
     * Setup drag events for column elements
     * @param {HTMLElement} list - The column task list element
     */
    setupColumnDragEvents(list) {
        list.addEventListener('dragover', this.handleDragOver.bind(this));
        list.addEventListener('drop', this.handleDrop.bind(this));
    }

    /**
     * Handle drag start event
     * @param {DragEvent} event - The drag start event
     */
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

    /**
     * Handle drag over event
     * @param {DragEvent} event - The drag over event
     */
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

    /**
     * Handle drop event
     * @param {DragEvent} event - The drop event
     */
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
                await this.moveTask(taskId, newStatus);
            }
        }
        
        // Clean up drag state
        this.cleanupDragState();
    }

    /**
     * Move a task to a new status
     * @param {string} taskId - The task ID
     * @param {string} newStatus - The new status
     */
    async moveTask(taskId, newStatus) {
        try {
            this.stateManager.showLoadingOverlay('Moving task...');
            
            // First get the current task data
            const task = await this.apiClient.getTask(taskId);
            
            // Update the task with new status
            const updatedTask = await this.apiClient.updateTask(taskId, {
                ...task,
                status: newStatus
            });
            
            this.stateManager.hideLoadingOverlay();
            
            if (updatedTask) {
                // Reload the page for now - this could be optimized to update state directly
                window.location.reload();
                this.stateManager.showMessage('Task moved successfully! 📁', 'success');
            } else {
                this.stateManager.showMessage('Failed to move task', 'error');
            }
            
        } catch (error) {
            this.stateManager.hideLoadingOverlay();
            console.error('Error moving task:', error);
            this.stateManager.showMessage('Failed to move task: ' + error.message, 'error');
        }
    }

    /**
     * Clean up drag state and visual indicators
     */
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

    /**
     * Refresh draggable cards after new tasks are added
     */
    refreshDraggableCards() {
        this.initializeDraggableCards();
    }

    /**
     * Setup drag events for a newly created task card
     * @param {HTMLElement} card - The new task card element
     */
    setupNewTaskCard(card) {
        this.setupTaskCardDragEvents(card);
    }

    /**
     * Setup drag events for a newly created column
     * @param {HTMLElement} column - The column element
     */
    setupNewColumn(column) {
        const taskList = column.querySelector('.task-list');
        if (taskList) {
            this.setupColumnDragEvents(taskList);
        }
    }
}
