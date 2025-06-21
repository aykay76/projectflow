/**
 * Task Detail Manager - Handles task detail modal functionality
 */
import { formatDate, formatDateTime, escapeHtml } from './utils.js';

export class TaskDetailManager {
    constructor(apiClient, stateManager) {
        this.apiClient = apiClient;
        this.stateManager = stateManager;
        this.modal = null;
        this.currentTask = null;
        
        this.init();
    }

    /**
     * Initialize the task detail manager
     */
    init() {
        this.modal = document.getElementById('task-detail-modal');
        if (!this.modal) {
            console.error('Task detail modal not found');
            return;
        }

        this.setupEventListeners();
        this.setupKanbanCardClickHandlers();
    }

    /**
     * Setup event listeners for the modal
     */
    setupEventListeners() {
        // Close modal events
        const closeBtn = this.modal.querySelector('.close');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => this.hideModal());
        }

        // Close modal when clicking outside
        this.modal.addEventListener('click', (e) => {
            if (e.target === this.modal) {
                this.hideModal();
            }
        });

        // Handle edit and delete buttons
        const editBtn = document.getElementById('detail-edit-btn');
        const deleteBtn = document.getElementById('detail-delete-btn');

        if (editBtn) {
            editBtn.addEventListener('click', () => this.handleEdit());
        }

        if (deleteBtn) {
            deleteBtn.addEventListener('click', () => this.handleDelete());
        }

        // Listen for ESC key to close modal
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.modal.style.display !== 'none') {
                this.hideModal();
            }
        });
    }

    /**
     * Setup click handlers for kanban cards
     */
    setupKanbanCardClickHandlers() {
        // Add click handlers to existing cards
        this.attachCardClickHandlers();

        // Listen for DOM changes to handle dynamically added cards
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                if (mutation.type === 'childList') {
                    mutation.addedNodes.forEach((node) => {
                        if (node.nodeType === Node.ELEMENT_NODE) {
                            if (node.classList && node.classList.contains('task-card')) {
                                this.attachClickHandler(node);
                            } else {
                                const cards = node.querySelectorAll('.task-card');
                                cards.forEach(card => this.attachClickHandler(card));
                            }
                        }
                    });
                }
            });
        });

        // Observe the kanban board for changes
        const kanbanView = document.getElementById('kanban-view');
        if (kanbanView) {
            observer.observe(kanbanView, {
                childList: true,
                subtree: true
            });
        }
    }

    /**
     * Attach click handlers to all existing kanban cards
     */
    attachCardClickHandlers() {
        const cards = document.querySelectorAll('.task-card');
        cards.forEach(card => this.attachClickHandler(card));
    }

    /**
     * Attach click handler to a single card
     */
    attachClickHandler(card) {
        // Avoid duplicate handlers
        if (card.dataset.clickHandlerAttached) {
            return;
        }

        card.addEventListener('click', (e) => {
            // Don't show modal if clicking on action buttons
            if (e.target.classList.contains('edit-task') || 
                e.target.classList.contains('delete-task') ||
                e.target.closest('.task-actions')) {
                return;
            }

            const taskId = card.dataset.id;
            if (taskId) {
                this.showTaskDetails(taskId);
            }
        });

        card.dataset.clickHandlerAttached = 'true';
    }

    /**
     * Show task details in modal
     */
    async showTaskDetails(taskId) {
        try {
            // Show loading state
            this.showLoadingState();
            
            // Fetch task details
            const task = await this.apiClient.getTask(taskId);
            if (!task) {
                throw new Error('Task not found');
            }

            this.currentTask = task;
            this.populateModal(task);
            this.showModal();

        } catch (error) {
            console.error('Error loading task details:', error);
            this.hideModal();
            
            // Show error message if notification manager is available
            if (window.projectFlowApp?.notificationManager?.showMessage) {
                window.projectFlowApp.notificationManager.showMessage(
                    'Failed to load task details: ' + error.message, 
                    'error'
                );
            }
        }
    }

    /**
     * Show loading state in modal
     */
    showLoadingState() {
        this.modal.querySelector('#detail-modal-title').textContent = 'Loading...';
        this.modal.querySelector('#detail-task-title').textContent = 'Loading task details...';
        this.modal.querySelector('#detail-task-description').textContent = '';
        this.showModal();
    }

    /**
     * Populate modal with task data
     */
    populateModal(task) {
        // Modal title
        const modalTitle = this.modal.querySelector('#detail-modal-title');
        modalTitle.textContent = `${task.display_id || task.id}`;

        // Task title and description
        const titleEl = this.modal.querySelector('#detail-task-title');
        const descriptionEl = this.modal.querySelector('#detail-task-description');
        
        titleEl.textContent = task.title || 'No title';
        descriptionEl.textContent = task.description || 'No description provided';

        // Task metadata
        this.populateTaskMetadata(task);

        // Show/hide conditional sections
        this.updateConditionalSections(task);
    }

    /**
     * Populate task metadata fields
     */
    populateTaskMetadata(task) {
        // Type
        const typeEl = this.modal.querySelector('#detail-task-type');
        if (typeEl) {
            typeEl.textContent = task.type || 'task';
            typeEl.className = `task-type task-type-${task.type || 'task'}`;
        }

        // Status
        const statusEl = this.modal.querySelector('#detail-task-status');
        if (statusEl) {
            statusEl.textContent = this.formatStatus(task.status);
            statusEl.className = `task-status status-${task.status || 'todo'}`;
        }

        // Priority
        const priorityEl = this.modal.querySelector('#detail-task-priority');
        if (priorityEl) {
            priorityEl.textContent = task.priority || 'medium';
            priorityEl.className = `task-priority priority-${task.priority || 'medium'}`;
        }

        // Created date
        const createdAtEl = this.modal.querySelector('#detail-task-created-at');
        if (createdAtEl) {
            createdAtEl.textContent = formatDate(task.created_at);
        }

        // Due date
        const dueDateEl = this.modal.querySelector('#detail-task-due-date');
        if (dueDateEl) {
            dueDateEl.textContent = task.due_date ? formatDate(task.due_date) : '';
        }

        // Started at
        const startedAtEl = this.modal.querySelector('#detail-task-started-at');
        if (startedAtEl) {
            startedAtEl.textContent = task.started_at ? formatDateTime(task.started_at) : '';
        }

        // Children count
        const childrenEl = this.modal.querySelector('#detail-task-children');
        if (childrenEl) {
            const childCount = task.children ? task.children.length : 0;
            childrenEl.textContent = childCount > 0 ? `${childCount} subtask${childCount > 1 ? 's' : ''}` : '';
        }
    }

    /**
     * Update visibility of conditional sections
     */
    updateConditionalSections(task) {
        // Due date row
        const dueDateRow = this.modal.querySelector('#detail-due-date-row');
        if (dueDateRow) {
            dueDateRow.style.display = task.due_date ? 'flex' : 'none';
        }

        // Started at row
        const startedAtRow = this.modal.querySelector('#detail-started-at-row');
        if (startedAtRow) {
            startedAtRow.style.display = task.started_at ? 'flex' : 'none';
        }

        // Children row
        const childrenRow = this.modal.querySelector('#detail-children-row');
        if (childrenRow) {
            const hasChildren = task.children && task.children.length > 0;
            childrenRow.style.display = hasChildren ? 'flex' : 'none';
        }
    }

    /**
     * Format status for display
     */
    formatStatus(status) {
        const statusMap = {
            'todo': 'To Do',
            'in_progress': 'In Progress',
            'done': 'Done',
            'blocked': 'Blocked'
        };
        return statusMap[status] || status;
    }

    /**
     * Show the modal
     */
    showModal() {
        if (this.modal) {
            this.modal.style.display = 'block';
            document.body.style.overflow = 'hidden'; // Prevent background scrolling
        }
    }

    /**
     * Hide the modal
     */
    hideModal() {
        if (this.modal) {
            this.modal.style.display = 'none';
            document.body.style.overflow = ''; // Restore scrolling
            this.currentTask = null;
        }
    }

    /**
     * Handle edit button click
     */
    handleEdit() {
        if (this.currentTask) {
            this.hideModal();
            
            // Trigger edit task functionality
            if (window.projectFlowApp?.taskManager?.editTask) {
                window.projectFlowApp.taskManager.editTask(this.currentTask.id);
            } else {
                // Fallback: trigger existing edit button
                const editBtn = document.querySelector(`[data-id="${this.currentTask.id}"] .edit-task`);
                if (editBtn) {
                    editBtn.click();
                }
            }
        }
    }

    /**
     * Handle delete button click
     */
    async handleDelete() {
        if (this.currentTask) {
            const confirmDelete = confirm(`Are you sure you want to delete "${this.currentTask.title}"?`);
            if (confirmDelete) {
                try {
                    this.hideModal();
                    
                    // Use task manager if available
                    if (window.projectFlowApp?.taskManager?.deleteTask) {
                        await window.projectFlowApp.taskManager.deleteTask(this.currentTask.id);
                    } else {
                        // Fallback: trigger existing delete button
                        const deleteBtn = document.querySelector(`[data-id="${this.currentTask.id}"] .delete-task`);
                        if (deleteBtn) {
                            deleteBtn.click();
                        }
                    }
                } catch (error) {
                    console.error('Error deleting task:', error);
                    if (window.projectFlowApp?.notificationManager?.showMessage) {
                        window.projectFlowApp.notificationManager.showMessage(
                            'Failed to delete task: ' + error.message, 
                            'error'
                        );
                    }
                }
            }
        }
    }

    /**
     * Refresh card click handlers (call this when cards are updated)
     */
    refreshCardHandlers() {
        // Clear existing handlers
        const cards = document.querySelectorAll('.task-card');
        cards.forEach(card => {
            card.removeAttribute('data-click-handler-attached');
        });
        
        // Re-attach handlers
        this.attachCardClickHandlers();
    }
}
