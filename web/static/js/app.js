// Application state
let currentEditingTask = null;
let currentView = 'kanban';
let hierarchyData = [];
let currentTheme = localStorage.getItem('theme') || 'light';

// DOM elements
const modal = document.getElementById('task-modal');
const modalTitle = document.getElementById('modal-title');
const taskForm = document.getElementById('task-form');
const newTaskBtn = document.getElementById('new-task-btn');
const cancelBtn = document.getElementById('cancel-btn');
const closeBtn = document.querySelector('.close');
const kanbanViewBtn = document.getElementById('kanban-view-btn');
const hierarchyViewBtn = document.getElementById('hierarchy-view-btn');
const timelineViewBtn = document.getElementById('timeline-view-btn');
const taskBoard = document.querySelector('.task-board');
const hierarchyView = document.getElementById('hierarchy-view');
const timelineView = document.getElementById('timeline-view');

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    initializeTheme();
    initializeEventListeners();
    initializeTimelineControls();
    initializeKeyboardShortcuts();
    initializeFiltering();
    initializeContextMenu();
    initializeAutoSave();
    updateOverdueIndicators();
    updateTaskCounts();
});

// Theme Management
function initializeTheme() {
    const themeToggle = document.getElementById('theme-toggle');
    const themeIcon = document.getElementById('theme-icon');
    
    // Apply saved theme
    document.documentElement.setAttribute('data-theme', currentTheme);
    updateThemeIcon();
    
    // Theme toggle event listener
    if (themeToggle) {
        themeToggle.addEventListener('click', toggleTheme);
    }
}

function toggleTheme() {
    currentTheme = currentTheme === 'light' ? 'dark' : 'light';
    document.documentElement.setAttribute('data-theme', currentTheme);
    localStorage.setItem('theme', currentTheme);
    updateThemeIcon();
    
    // Add a subtle animation to indicate theme change
    document.body.style.transition = 'background-color 0.3s ease, color 0.3s ease';
    setTimeout(() => {
        document.body.style.transition = '';
    }, 300);
}

function updateThemeIcon() {
    const themeIcon = document.getElementById('theme-icon');
    if (themeIcon) {
        themeIcon.textContent = currentTheme === 'light' ? '🌙' : '☀️';
    }
}

// Keyboard Shortcuts
function initializeKeyboardShortcuts() {
    document.addEventListener('keydown', (event) => {
        // Only process shortcuts when not in input fields
        if (event.target.tagName === 'INPUT' || event.target.tagName === 'TEXTAREA') {
            return;
        }
        
        // Cmd/Ctrl + K: New Task
        if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
            event.preventDefault();
            openTaskModal();
        }
        
        // Cmd/Ctrl + D: Toggle Theme
        if ((event.metaKey || event.ctrlKey) && event.key === 'd') {
            event.preventDefault();
            toggleTheme();
        }
        
        // F: Toggle Filters
        if (event.key === 'f' || event.key === 'F') {
            event.preventDefault();
            toggleFilterPanel();
        }
        
        // Number keys for view switching
        if (event.key >= '1' && event.key <= '3') {
            event.preventDefault();
            const views = ['kanban', 'hierarchy', 'timeline'];
            switchToView(views[parseInt(event.key) - 1]);
        }
        
        // Escape key to close modal or filter panel
        if (event.key === 'Escape') {
            if (modal.style.display === 'block') {
                closeTaskModal();
            } else if (document.getElementById('filter-panel').style.display === 'block') {
                toggleFilterPanel();
            }
        }
    });
}

function initializeEventListeners() {
    // View switching
    kanbanViewBtn.addEventListener('click', () => switchToView('kanban'));
    hierarchyViewBtn.addEventListener('click', () => switchToView('hierarchy'));
    timelineViewBtn.addEventListener('click', () => switchToView('timeline'));

    // Modal controls
    newTaskBtn.addEventListener('click', () => openTaskModal());
    cancelBtn.addEventListener('click', () => closeTaskModal());
    closeBtn.addEventListener('click', () => closeTaskModal());
    
    // Close modal when clicking outside
    window.addEventListener('click', (event) => {
        if (event.target === modal) {
            closeTaskModal();
        }
    });

    // Form submission
    taskForm.addEventListener('submit', handleTaskSubmit);

    // Task card actions
    document.addEventListener('click', (event) => {
        if (event.target.classList.contains('edit-task')) {
            const taskId = event.target.getAttribute('data-id');
            editTask(taskId);
        } else if (event.target.classList.contains('delete-task')) {
            const taskId = event.target.getAttribute('data-id');
            deleteTask(taskId);
        }
    });

    // Make task cards draggable using event delegation
    document.addEventListener('dragstart', (event) => {
        if (event.target.classList.contains('task-card')) {
            handleDragStart(event);
        }
    });

    // Make columns drop targets using event delegation
    document.addEventListener('dragover', (event) => {
        if (event.target.closest('.column')) {
            handleDragOver(event);
        }
    });

    document.addEventListener('drop', (event) => {
        if (event.target.closest('.column')) {
            handleDrop(event);
        }
    });

    // Initialize draggable attribute for existing task cards
    initializeDraggableCards();
}

function openTaskModal(task = null) {
    currentEditingTask = task;
    
    if (task) {
        modalTitle.textContent = 'Edit Task';
        populateForm(task);
    } else {
        modalTitle.textContent = 'New Task';
        taskForm.reset();
        // Restore form draft for new tasks
        setTimeout(() => restoreFormDraft(), 100);
    }
    
    modal.style.display = 'block';
    document.getElementById('task-title').focus();
}

function closeTaskModal() {
    modal.style.display = 'none';
    currentEditingTask = null;
    taskForm.reset();
}

function populateForm(task) {
    document.getElementById('task-title').value = task.title || '';
    document.getElementById('task-description').value = task.description || '';
    document.getElementById('task-type').value = task.type || 'task';
    document.getElementById('task-priority').value = task.priority || 'medium';
    document.getElementById('task-status').value = task.status || 'todo';
    document.getElementById('task-due-date').value = task.due_date ? task.due_date.split('T')[0] : '';
    
    // Handle start date - convert from RFC3339 to datetime-local format
    if (task.started_at) {
        const startDate = new Date(task.started_at);
        const year = startDate.getFullYear();
        const month = String(startDate.getMonth() + 1).padStart(2, '0');
        const day = String(startDate.getDate()).padStart(2, '0');
        const hours = String(startDate.getHours()).padStart(2, '0');
        const minutes = String(startDate.getMinutes()).padStart(2, '0');
        document.getElementById('task-started-at').value = `${year}-${month}-${day}T${hours}:${minutes}`;
    } else {
        document.getElementById('task-started-at').value = '';
    }
}

async function handleTaskSubmit(event) {
    event.preventDefault();
    
    showLoadingOverlay(currentEditingTask ? 'Updating task...' : 'Creating task...');
    
    const formData = new FormData(taskForm);
    const taskData = {
        title: formData.get('title'),
        description: formData.get('description'),
        type: formData.get('type'),
        priority: formData.get('priority'),
        status: formData.get('status'),
        due_date: formData.get('due_date') || null,
        started_at: formData.get('started_at') ? new Date(formData.get('started_at')).toISOString() : null
    };

    try {
        let response;
        if (currentEditingTask) {
            // Update existing task
            response = await fetch(`/api/tasks/${currentEditingTask.id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(taskData)
            });
        } else {
            // Create new task
            response = await fetch('/api/tasks', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(taskData)
            });
        }

        hideLoadingOverlay();

        if (response.ok) {
            clearFormDraft(); // Clear saved draft
            closeTaskModal();
            window.location.reload();
            showMessage(
                currentEditingTask ? 'Task updated successfully! 🎉' : 'Task created successfully! ✨',
                'success'
            );
        } else {
            const error = await response.text();
            showMessage(`Error: ${error}`, 'error');
        }
    } catch (error) {
        hideLoadingOverlay();
        console.error('Error saving task:', error);
        showMessage('Failed to save task. Please try again.', 'error');
    }
}

async function editTask(taskId) {
    try {
        const response = await fetch(`/api/tasks/${taskId}`);
        if (response.ok) {
            const task = await response.json();
            openTaskModal(task);
        } else {
            showMessage('Failed to load task for editing.', 'error');
        }
    } catch (error) {
        console.error('Error loading task:', error);
        showMessage('Failed to load task for editing.', 'error');
    }
}

async function deleteTask(taskId) {
    if (!confirm('Are you sure you want to delete this task?')) {
        return;
    }

    try {
        const response = await fetch(`/api/tasks/${taskId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            // Remove the task card from the DOM
            const taskCard = document.querySelector(`[data-id="${taskId}"]`);
            if (taskCard) {
                taskCard.remove();
            }
            showMessage('Task deleted successfully!', 'success');
        } else {
            showMessage('Failed to delete task.', 'error');
        }
    } catch (error) {
        console.error('Error deleting task:', error);
        showMessage('Failed to delete task. Please try again.', 'error');
    }
}

// Keyboard Shortcuts Helper
function showKeyboardShortcuts() {
    const shortcutsModal = document.createElement('div');
    shortcutsModal.className = 'modal';
    shortcutsModal.style.display = 'block';
    
    shortcutsModal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h2>⌨️ Keyboard Shortcuts</h2>
                <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
            </div>
            <div style="padding: 0 24px 24px 24px;">
                <div style="display: grid; gap: 16px;">
                    <div>
                        <h3 style="margin-bottom: 8px; color: var(--text-primary);">Navigation</h3>
                        <div style="display: grid; grid-template-columns: auto 1fr; gap: 8px 16px; font-size: 14px;">
                            <code style="background: var(--bg-tertiary); padding: 2px 6px; border-radius: 4px;">1</code>
                            <span>Switch to Kanban View</span>
                            <code style="background: var(--bg-tertiary); padding: 2px 6px; border-radius: 4px;">2</code>
                            <span>Switch to Hierarchy View</span>
                            <code style="background: var(--bg-tertiary); padding: 2px 6px; border-radius: 4px;">3</code>
                            <span>Switch to Timeline View</span>
                        </div>
                    </div>
                    <div>
                        <h3 style="margin-bottom: 8px; color: var(--text-primary);">Actions</h3>
                        <div style="display: grid; grid-template-columns: auto 1fr; gap: 8px 16px; font-size: 14px;">
                            <code style="background: var(--bg-tertiary); padding: 2px 6px; border-radius: 4px;">⌘/Ctrl + K</code>
                            <span>Create New Task</span>
                            <code style="background: var(--bg-tertiary); padding: 2px 6px; border-radius: 4px;">⌘/Ctrl + D</code>
                            <span>Toggle Dark/Light Theme</span>
                            <code style="background: var(--bg-tertiary); padding: 2px 6px; border-radius: 4px;">Escape</code>
                            <span>Close Modal</span>
                            <code style="background: var(--bg-tertiary); padding: 2px 6px; border-radius: 4px;">?</code>
                            <span>Show Keyboard Shortcuts</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `;
    
    document.body.appendChild(shortcutsModal);
    
    // Close on outside click
    shortcutsModal.addEventListener('click', (e) => {
        if (e.target === shortcutsModal) {
            shortcutsModal.remove();
        }
    });
}

// Enhanced message system with toast notifications
function showMessage(text, type = 'info', duration = 4000) {
    // Create toast container if it doesn't exist
    let toastContainer = document.querySelector('.toast-container');
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.className = 'toast-container';
        document.body.appendChild(toastContainer);
    }
    
    // Create toast element
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    
    const icons = {
        success: '✅',
        error: '❌',
        warning: '⚠️',
        info: 'ℹ️'
    };
    
    toast.innerHTML = `
        <div class="toast-content">
            <span class="toast-icon">${icons[type] || icons.info}</span>
            <span class="toast-message">${text}</span>
            <button class="toast-close" onclick="this.parentElement.parentElement.remove()">×</button>
        </div>
    `;
    
    toastContainer.appendChild(toast);
    
    // Auto remove after duration
    if (duration > 0) {
        setTimeout(() => {
            if (toast.parentElement) {
                toast.style.animation = 'toastSlideOut 0.3s ease-in forwards';
                setTimeout(() => toast.remove(), 300);
            }
        }, duration);
    }
}

// Loading overlay functionality
function showLoadingOverlay(message = 'Loading...') {
    let overlay = document.querySelector('.loading-overlay');
    if (!overlay) {
        overlay = document.createElement('div');
        overlay.className = 'loading-overlay';
        overlay.innerHTML = `
            <div class="loading-spinner">
                <div class="spinner"></div>
                <div class="loading-text">${message}</div>
            </div>
        `;
        document.body.appendChild(overlay);
    }
    
    overlay.querySelector('.loading-text').textContent = message;
    overlay.classList.add('show');
}

function hideLoadingOverlay() {
    const overlay = document.querySelector('.loading-overlay');
    if (overlay) {
        overlay.classList.remove('show');
    }
}

// Context menu functionality
let contextMenu = null;
let contextMenuTarget = null;

function initializeContextMenu() {
    // Create context menu
    contextMenu = document.createElement('div');
    contextMenu.className = 'context-menu';
    contextMenu.innerHTML = `
        <button class="context-menu-item" data-action="edit">
            <span class="context-menu-icon">✏️</span>
            Edit Task
        </button>
        <button class="context-menu-item" data-action="duplicate">
            <span class="context-menu-icon">📋</span>
            Duplicate Task
        </button>
        <button class="context-menu-item" data-action="move">
            <span class="context-menu-icon">📁</span>
            Move to...
        </button>
        <hr style="margin: 8px 0; border: none; border-top: 1px solid var(--border-color);">
        <button class="context-menu-item danger" data-action="delete">
            <span class="context-menu-icon">🗑️</span>
            Delete Task
        </button>
    `;
    document.body.appendChild(contextMenu);
    
    // Context menu event listeners
    contextMenu.addEventListener('click', handleContextMenuAction);
    
    // Add right-click listeners to task cards
    document.addEventListener('contextmenu', (event) => {
        const taskCard = event.target.closest('.task-card');
        if (taskCard) {
            event.preventDefault();
            showContextMenu(event, taskCard);
        }
    });
    
    // Hide context menu when clicking elsewhere
    document.addEventListener('click', hideContextMenu);
}

function showContextMenu(event, taskCard) {
    contextMenuTarget = taskCard;
    contextMenu.style.left = `${event.pageX}px`;
    contextMenu.style.top = `${event.pageY}px`;
    contextMenu.classList.add('show');
    
    // Adjust position if menu goes off screen
    setTimeout(() => {
        const rect = contextMenu.getBoundingClientRect();
        if (rect.right > window.innerWidth) {
            contextMenu.style.left = `${event.pageX - rect.width}px`;
        }
        if (rect.bottom > window.innerHeight) {
            contextMenu.style.top = `${event.pageY - rect.height}px`;
        }
    }, 0);
}

function hideContextMenu() {
    if (contextMenu) {
        contextMenu.classList.remove('show');
        contextMenuTarget = null;
    }
}

function handleContextMenuAction(event) {
    const action = event.target.closest('.context-menu-item')?.dataset.action;
    const taskId = contextMenuTarget?.dataset.id;
    
    if (!action || !taskId) return;
    
    hideContextMenu();
    
    switch (action) {
        case 'edit':
            editTask(taskId);
            break;
        case 'duplicate':
            duplicateTask(taskId);
            break;
        case 'move':
            showMoveTaskDialog(taskId);
            break;
        case 'delete':
            if (confirm('Are you sure you want to delete this task?')) {
                deleteTask(taskId);
            }
            break;
    }
}

async function duplicateTask(taskId) {
    try {
        showLoadingOverlay('Duplicating task...');
        const response = await fetch(`/api/tasks/${taskId}`);
        if (response.ok) {
            const task = await response.json();
            
            // Create duplicate with modified title
            const duplicateData = {
                ...task,
                title: `${task.title} (Copy)`,
                status: 'todo',
                started_at: null,
                completed_at: null
            };
            delete duplicateData.id;
            delete duplicateData.created_at;
            delete duplicateData.updated_at;
            
            const createResponse = await fetch('/api/tasks', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(duplicateData)
            });
            
            hideLoadingOverlay();
            
            if (createResponse.ok) {
                window.location.reload();
                showMessage('Task duplicated successfully! 📋', 'success');
            } else {
                showMessage('Failed to duplicate task', 'error');
            }
        }
    } catch (error) {
        hideLoadingOverlay();
        showMessage('Error duplicating task', 'error');
    }
}

function showMoveTaskDialog(taskId) {
    const moveModal = document.createElement('div');
    moveModal.className = 'modal';
    moveModal.style.display = 'block';
    
    moveModal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h2>📁 Move Task</h2>
                <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
            </div>
            <div style="padding: 0 24px 24px 24px;">
                <p style="margin-bottom: 16px;">Select the new status for this task:</p>
                <div style="display: grid; gap: 12px;">
                    <button class="btn btn-outline" onclick="moveTask('${taskId}', 'todo'); this.closest('.modal').remove();">
                        📝 To Do
                    </button>
                    <button class="btn btn-outline" onclick="moveTask('${taskId}', 'in_progress'); this.closest('.modal').remove();">
                        🔄 In Progress
                    </button>
                    <button class="btn btn-outline" onclick="moveTask('${taskId}', 'review'); this.closest('.modal').remove();">
                        👀 Review
                    </button>
                    <button class="btn btn-outline" onclick="moveTask('${taskId}', 'done'); this.closest('.modal').remove();">
                        ✅ Done
                    </button>
                </div>
            </div>
        </div>
    `;
    
    document.body.appendChild(moveModal);
    
    // Close on outside click
    moveModal.addEventListener('click', (e) => {
        if (e.target === moveModal) {
            moveModal.remove();
        }
    });
}

async function moveTask(taskId, newStatus) {
    try {
        showLoadingOverlay('Moving task...');
        
        // First get the current task data
        const response = await fetch(`/api/tasks/${taskId}`);
        if (!response.ok) {
            throw new Error('Failed to fetch task');
        }
        
        const task = await response.json();
        
        // Update the task with new status
        const updateResponse = await fetch(`/api/tasks/${taskId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                ...task,
                status: newStatus
            })
        });
        
        hideLoadingOverlay();
        
        if (updateResponse.ok) {
            window.location.reload();
            showMessage('Task moved successfully! 📁', 'success');
        } else {
            showMessage('Failed to move task', 'error');
        }
    } catch (error) {
        hideLoadingOverlay();
        console.error('Error moving task:', error);
        showMessage('Error moving task', 'error');
    }
}

// Utility functions
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Auto-save form data
function initializeAutoSave() {
    const formInputs = taskForm.querySelectorAll('input, textarea, select');
    formInputs.forEach(input => {
        input.addEventListener('input', debounce(() => {
            const formData = new FormData(taskForm);
            const data = Object.fromEntries(formData.entries());
            localStorage.setItem('taskFormDraft', JSON.stringify(data));
        }, 500));
    });
}

function restoreFormDraft() {
    const draft = localStorage.getItem('taskFormDraft');
    if (draft && !currentEditingTask) {
        const data = JSON.parse(draft);
        Object.keys(data).forEach(key => {
            const input = taskForm.querySelector(`[name="${key}"]`);
            if (input) {
                input.value = data[key];
            }
        });
        showMessage('Form draft restored', 'info', 2000);
    }
}

function clearFormDraft() {
    localStorage.removeItem('taskFormDraft');
}
