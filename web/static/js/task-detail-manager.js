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
        this.setupTaskActionHandlers();
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
            this.openEditModal(this.currentTask.id);
        }
    }

    /**
     * Handle delete button click
     */
    async handleDelete() {
        if (this.currentTask) {
            this.hideModal();
            await this.handleDeleteTask(this.currentTask.id);
        }
    }

    /**
     * Setup event delegation for task action buttons (edit/delete)
     */
    setupTaskActionHandlers() {
        // Use event delegation to handle dynamically added buttons
        document.addEventListener('click', (e) => {
            // Handle edit button clicks
            if (e.target.classList.contains('edit-task')) {
                e.preventDefault();
                e.stopPropagation();
                const taskId = e.target.dataset.id;
                if (taskId) {
                    this.openEditModal(taskId);
                }
            }
            
            // Handle delete button clicks
            if (e.target.classList.contains('delete-task')) {
                e.preventDefault();
                e.stopPropagation();
                const taskId = e.target.dataset.id;
                if (taskId) {
                    this.handleDeleteTask(taskId);
                }
            }
        });
    }

    /**
     * Open edit modal for a task
     */
    async openEditModal(taskId) {
        try {
            const task = await this.apiClient.getTask(taskId);
            if (!task) {
                throw new Error('Task not found');
            }

            this.populateEditModal(task);
            this.showEditModal();

        } catch (error) {
            console.error('Error loading task for editing:', error);
            if (window.projectFlowApp?.notificationManager?.showMessage) {
                window.projectFlowApp.notificationManager.showMessage(
                    'Failed to load task for editing: ' + error.message, 
                    'error'
                );
            }
        }
    }

    /**
     * Populate the edit modal with task data
     */
    populateEditModal(task) {
        const modal = document.getElementById('task-modal');
        if (!modal) return;

        // Set modal title
        const modalTitle = modal.querySelector('#modal-title');
        if (modalTitle) {
            modalTitle.textContent = `Edit Task - ${task.display_id || task.id}`;
        }

        // Populate form fields
        const form = modal.querySelector('#task-form');
        if (form) {
            form.dataset.taskId = task.id; // Store task ID for update

            const titleInput = form.querySelector('#task-title');
            const descriptionInput = form.querySelector('#task-description');
            const typeSelect = form.querySelector('#task-type');
            const prioritySelect = form.querySelector('#task-priority');
            const statusSelect = form.querySelector('#task-status');
            const dueDateInput = form.querySelector('#task-due-date');
            const startedAtInput = form.querySelector('#task-started-at');

            if (titleInput) titleInput.value = task.title || '';
            if (descriptionInput) descriptionInput.value = task.description || '';
            if (typeSelect) typeSelect.value = task.type || 'task';
            if (prioritySelect) prioritySelect.value = task.priority || 'medium';
            if (statusSelect) statusSelect.value = task.status || 'todo';
            
            if (dueDateInput && task.due_date) {
                // Convert date to YYYY-MM-DD format
                const dueDate = new Date(task.due_date);
                dueDateInput.value = dueDate.toISOString().split('T')[0];
            }

            if (startedAtInput && task.started_at) {
                // Convert to datetime-local format
                const startedAt = new Date(task.started_at);
                startedAtInput.value = startedAt.toISOString().slice(0, 16);
            }
        }

        // Setup form submission for updates
        this.setupEditFormSubmission();
    }

    /**
     * Setup form submission for task updates
     */
    setupEditFormSubmission() {
        const form = document.getElementById('task-form');
        if (!form) return;

        // Remove existing listeners to avoid duplicates
        const newForm = form.cloneNode(true);
        form.parentNode.replaceChild(newForm, form);

        newForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const taskId = newForm.dataset.taskId;
            if (!taskId) return;

            const formData = new FormData(newForm);
            const taskData = {
                title: formData.get('title'),
                description: formData.get('description'),
                type: formData.get('type'),
                priority: formData.get('priority'),
                status: formData.get('status'),
                due_date: formData.get('due_date') || null,
                started_at: formData.get('started_at') || null
            };

            try {
                await this.apiClient.updateTask(taskId, taskData);
                this.hideEditModal();
                
                // Refresh the view
                if (window.projectFlowApp?.taskManager?.refreshTasks) {
                    window.projectFlowApp.taskManager.refreshTasks();
                }

                if (window.projectFlowApp?.notificationManager?.showMessage) {
                    window.projectFlowApp.notificationManager.showMessage(
                        'Task updated successfully! 🎉', 
                        'success'
                    );
                }
            } catch (error) {
                console.error('Error updating task:', error);
                if (window.projectFlowApp?.notificationManager?.showMessage) {
                    window.projectFlowApp.notificationManager.showMessage(
                        'Failed to update task: ' + error.message, 
                        'error'
                    );
                }
            }
        });

        // Setup cancel button
        const cancelBtn = newForm.querySelector('#cancel-btn');
        if (cancelBtn) {
            cancelBtn.addEventListener('click', () => {
                this.hideEditModal();
            });
        }

        // Setup close button
        const modal = document.getElementById('task-modal');
        const closeBtn = modal?.querySelector('.close');
        if (closeBtn) {
            closeBtn.onclick = () => this.hideEditModal();
        }
    }

    /**
     * Show the edit modal
     */
    showEditModal() {
        const modal = document.getElementById('task-modal');
        if (modal) {
            modal.style.display = 'block';
            document.body.style.overflow = 'hidden';
        }
    }

    /**
     * Hide the edit modal
     */
    hideEditModal() {
        const modal = document.getElementById('task-modal');
        if (modal) {
            modal.style.display = 'none';
            document.body.style.overflow = '';
            
            // Clear form data
            const form = modal.querySelector('#task-form');
            if (form) {
                form.reset();
                form.removeAttribute('data-task-id');
            }
        }
    }

    /**
     * Handle delete task
     */
    async handleDeleteTask(taskId) {
        try {
            const task = await this.apiClient.getTask(taskId);
            if (!task) {
                throw new Error('Task not found');
            }

            const confirmDelete = confirm(`Are you sure you want to delete "${task.title}"?`);
            if (confirmDelete) {
                await this.apiClient.deleteTask(taskId);
                
                // Remove the task card from the DOM immediately for better UX
                const taskCard = document.querySelector(`[data-id="${taskId}"]`);
                if (taskCard) {
                    taskCard.remove();
                }
                
                if (window.projectFlowApp?.notificationManager?.showMessage) {
                    window.projectFlowApp.notificationManager.showMessage(
                        'Task deleted successfully!', 
                        'success'
                    );
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
