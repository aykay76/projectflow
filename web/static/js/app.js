// Application state
let currentEditingTask = null;
let currentView = 'kanban';
let hierarchyData = [];

// DOM elements
const modal = document.getElementById('task-modal');
const modalTitle = document.getElementById('modal-title');
const taskForm = document.getElementById('task-form');
const newTaskBtn = document.getElementById('new-task-btn');
const cancelBtn = document.getElementById('cancel-btn');
const closeBtn = document.querySelector('.close');
const kanbanViewBtn = document.getElementById('kanban-view-btn');
const hierarchyViewBtn = document.getElementById('hierarchy-view-btn');
const taskBoard = document.querySelector('.task-board');
const hierarchyView = document.getElementById('hierarchy-view');

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    initializeEventListeners();
});

function initializeEventListeners() {
    // View switching
    kanbanViewBtn.addEventListener('click', () => switchToView('kanban'));
    hierarchyViewBtn.addEventListener('click', () => switchToView('hierarchy'));

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

    // Make task cards draggable (for future drag-and-drop functionality)
    const taskCards = document.querySelectorAll('.task-card');
    taskCards.forEach(card => {
        card.draggable = true;
        card.addEventListener('dragstart', handleDragStart);
    });

    // Make columns drop targets
    const columns = document.querySelectorAll('.column');
    columns.forEach(column => {
        column.addEventListener('dragover', handleDragOver);
        column.addEventListener('drop', handleDrop);
    });
}

function openTaskModal(task = null) {
    currentEditingTask = task;
    
    if (task) {
        modalTitle.textContent = 'Edit Task';
        populateForm(task);
    } else {
        modalTitle.textContent = 'New Task';
        taskForm.reset();
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
}

async function handleTaskSubmit(event) {
    event.preventDefault();
    
    const formData = new FormData(taskForm);
    const taskData = {
        title: formData.get('title'),
        description: formData.get('description'),
        type: formData.get('type'),
        priority: formData.get('priority'),
        status: formData.get('status')
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

        if (response.ok) {
            closeTaskModal();
            window.location.reload(); // Simple refresh for now
            showMessage('Task saved successfully!', 'success');
        } else {
            const error = await response.text();
            showMessage(`Error: ${error}`, 'error');
        }
    } catch (error) {
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

function showMessage(text, type) {
    // Remove existing messages
    const existingMessages = document.querySelectorAll('.message');
    existingMessages.forEach(msg => msg.remove());

    // Create new message
    const message = document.createElement('div');
    message.className = `message message-${type}`;
    message.textContent = text;

    // Insert at the top of the container
    const container = document.querySelector('.container');
    container.insertBefore(message, container.firstChild);

    // Auto-remove after 5 seconds
    setTimeout(() => {
        if (message.parentNode) {
            message.remove();
        }
    }, 5000);
}

// Drag and drop functionality (basic implementation)
function handleDragStart(event) {
    event.dataTransfer.setData('text/plain', event.target.getAttribute('data-id'));
    event.target.style.opacity = '0.5';
}

function handleDragOver(event) {
    event.preventDefault();
    event.currentTarget.style.backgroundColor = '#f0f8ff';
}

function handleDrop(event) {
    event.preventDefault();
    event.currentTarget.style.backgroundColor = '';
    
    const taskId = event.dataTransfer.getData('text/plain');
    const newStatus = event.currentTarget.getAttribute('data-status');
    
    if (taskId && newStatus) {
        updateTaskStatus(taskId, newStatus);
    }
    
    // Reset opacity
    const draggedElement = document.querySelector(`[data-id="${taskId}"]`);
    if (draggedElement) {
        draggedElement.style.opacity = '1';
    }
}

async function updateTaskStatus(taskId, newStatus) {
    try {
        // First get the current task data
        const getResponse = await fetch(`/api/tasks/${taskId}`);
        if (!getResponse.ok) return;
        
        const task = await getResponse.json();
        task.status = newStatus;
        
        // Update the task
        const updateResponse = await fetch(`/api/tasks/${taskId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(task)
        });

        if (updateResponse.ok) {
            // Move the task card to the appropriate column
            const taskCard = document.querySelector(`[data-id="${taskId}"]`);
            const targetColumn = document.querySelector(`[data-status="${newStatus}"] .task-list`);
            
            if (taskCard && targetColumn) {
                targetColumn.appendChild(taskCard);
                showMessage('Task status updated!', 'success');
            }
        } else {
            showMessage('Failed to update task status.', 'error');
        }
    } catch (error) {
        console.error('Error updating task status:', error);
        showMessage('Failed to update task status.', 'error');
    }
}

// View switching functions
function switchToView(viewType) {
    currentView = viewType;
    
    if (viewType === 'kanban') {
        taskBoard.style.display = 'grid';
        hierarchyView.style.display = 'none';
        kanbanViewBtn.classList.add('active');
        hierarchyViewBtn.classList.remove('active');
    } else if (viewType === 'hierarchy') {
        taskBoard.style.display = 'none';
        hierarchyView.style.display = 'block';
        kanbanViewBtn.classList.remove('active');
        hierarchyViewBtn.classList.add('active');
        loadHierarchyView();
    }
}

async function loadHierarchyView() {
    try {
        const response = await fetch('/api/hierarchy');
        if (response.ok) {
            hierarchyData = await response.json();
            renderHierarchyView();
        } else {
            showMessage('Failed to load hierarchy view.', 'error');
        }
    } catch (error) {
        console.error('Error loading hierarchy:', error);
        showMessage('Failed to load hierarchy view.', 'error');
    }
}

function renderHierarchyView() {
    const container = document.querySelector('.hierarchy-container');
    container.innerHTML = '';
    
    if (hierarchyData.length === 0) {
        container.innerHTML = '<p>No tasks found. Create your first task to get started!</p>';
        return;
    }
    
    hierarchyData.forEach(hierarchyTask => {
        const element = createHierarchyElement(hierarchyTask, 0);
        container.appendChild(element);
    });
}

function createHierarchyElement(hierarchyTask, level) {
    const task = hierarchyTask.task || hierarchyTask; // Handle both old and new format
    const childTasks = hierarchyTask.child_tasks || [];
    
    const item = document.createElement('div');
    item.className = `hierarchy-item ${task.type}`;
    item.style.marginLeft = `${level * 20}px`;
    
    const hasChildren = childTasks && childTasks.length > 0;
    const toggleSymbol = hasChildren ? '▼' : '•';
    
    item.innerHTML = `
        <div class="hierarchy-content">
            <div class="hierarchy-info">
                <button class="hierarchy-toggle" ${!hasChildren ? 'style="visibility: hidden;"' : ''}>
                    ${toggleSymbol}
                </button>
                <span class="hierarchy-badge ${task.type}">${task.type.toUpperCase()}</span>
                <h4 class="hierarchy-title">${task.title}</h4>
                <div class="hierarchy-meta">
                    <span class="hierarchy-badge status-${task.status}">${task.status.replace('_', ' ')}</span>
                    <span class="hierarchy-badge priority-${task.priority}">${task.priority}</span>
                    ${hasChildren ? `<span>${childTasks.length} child${childTasks.length !== 1 ? 'ren' : ''}</span>` : ''}
                </div>
            </div>
            <div class="hierarchy-actions">
                <button class="btn btn-sm btn-secondary edit-task" data-id="${task.id}">Edit</button>
                <button class="btn btn-sm btn-danger delete-task" data-id="${task.id}">Delete</button>
            </div>
        </div>
    `;
    
    if (hasChildren) {
        const childrenContainer = document.createElement('div');
        childrenContainer.className = 'hierarchy-children';
        childrenContainer.id = `children-${task.id}`;
        
        childTasks.forEach(childHierarchyTask => {
            const childElement = createHierarchyElement(childHierarchyTask, level + 1);
            childrenContainer.appendChild(childElement);
        });
        
        item.appendChild(childrenContainer);
        
        // Add toggle functionality
        const toggleBtn = item.querySelector('.hierarchy-toggle');
        toggleBtn.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            toggleHierarchyItem(task.id);
        });
    }
    
    return item;
}

function toggleHierarchyItem(taskId) {
    const childrenContainer = document.getElementById(`children-${taskId}`);
    const parentItem = childrenContainer.parentElement;
    const toggleBtn = parentItem.querySelector('.hierarchy-toggle');
    
    if (childrenContainer.style.display === 'none') {
        childrenContainer.style.display = 'block';
        toggleBtn.textContent = '▼';
    } else {
        childrenContainer.style.display = 'none';
        toggleBtn.textContent = '▶';
    }
}

// Utility functions
function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    });
}

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
