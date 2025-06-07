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
const timelineViewBtn = document.getElementById('timeline-view-btn');
const taskBoard = document.querySelector('.task-board');
const hierarchyView = document.getElementById('hierarchy-view');
const timelineView = document.getElementById('timeline-view');

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    initializeEventListeners();
    initializeTimelineControls();
    updateOverdueIndicators();
});

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
    document.getElementById('task-due-date').value = task.due_date ? task.due_date.split('T')[0] : '';
}

async function handleTaskSubmit(event) {
    event.preventDefault();
    
    const formData = new FormData(taskForm);
    const taskData = {
        title: formData.get('title'),
        description: formData.get('description'),
        type: formData.get('type'),
        priority: formData.get('priority'),
        status: formData.get('status'),
        due_date: formData.get('due_date') || null
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

// Update overdue indicators for Kanban view
function updateOverdueIndicators() {
    const taskCards = document.querySelectorAll('.task-card');
    const today = new Date();
    today.setHours(0, 0, 0, 0); // Set to start of day for comparison
    
    taskCards.forEach(card => {
        const dueDateElement = card.querySelector('.task-due-date');
        if (dueDateElement) {
            // Extract date from "Due: Month Day, Year" format
            const dueDateText = dueDateElement.textContent.replace('Due: ', '');
            const dueDate = new Date(dueDateText);
            dueDate.setHours(0, 0, 0, 0);
            
            // Get task status from the parent column
            const column = card.closest('.column');
            const status = column ? column.getAttribute('data-status') : '';
            
            // Only mark as overdue if not done and past due date
            if (status !== 'done' && dueDate < today) {
                card.classList.add('overdue');
            } else if (status !== 'done' && dueDate <= new Date(today.getTime() + (3 * 24 * 60 * 60 * 1000))) {
                // Due within 3 days
                card.classList.add('due-soon');
            }
        }
    });
}

// View switching functions
function switchToView(viewType) {
    currentView = viewType;
    
    // Hide all views
    taskBoard.style.display = 'none';
    hierarchyView.style.display = 'none';
    timelineView.style.display = 'none';
    
    // Remove active class from all buttons
    kanbanViewBtn.classList.remove('active');
    hierarchyViewBtn.classList.remove('active');
    timelineViewBtn.classList.remove('active');
    
    if (viewType === 'kanban') {
        taskBoard.style.display = 'grid';
        kanbanViewBtn.classList.add('active');
    } else if (viewType === 'hierarchy') {
        hierarchyView.style.display = 'block';
        hierarchyViewBtn.classList.add('active');
        loadHierarchyView();
    } else if (viewType === 'timeline') {
        timelineView.style.display = 'block';
        timelineViewBtn.classList.add('active');
        loadTimelineView();
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

// Timeline view functions
let timelineRange = 60; // Default 60 days

async function loadTimelineView() {
    try {
        console.log('Loading timeline view...');
        const response = await fetch('/api/tasks');
        if (response.ok) {
            const tasks = await response.json();
            console.log('Loaded tasks for timeline:', tasks.length);
            renderTimelineView(tasks);
        } else {
            console.error('Failed to load tasks, status:', response.status);
            showMessage('Failed to load timeline view.', 'error');
        }
    } catch (error) {
        console.error('Error loading timeline:', error);
        showMessage('Failed to load timeline view.', 'error');
    }
}

function renderTimelineView(tasks) {
    console.log('Rendering timeline view with tasks:', tasks);
    const container = document.querySelector('.timeline-container');
    
    if (!container) {
        console.error('Timeline container not found!');
        showMessage('Timeline container not found in DOM.', 'error');
        return;
    }
    
    container.innerHTML = '';
    
    if (tasks.length === 0) {
        container.innerHTML = '<p>No tasks found. Create your first task to get started!</p>';
        return;
    }
    
    // Get current date and range (normalize to start of day for consistent comparison)
    const today = new Date();
    today.setHours(0, 0, 0, 0); // Set to start of day
    const endDate = new Date(today.getTime() + (timelineRange * 24 * 60 * 60 * 1000));
    
    console.log('Timeline date range:', today, 'to', endDate);
    
    // Filter tasks with due dates within range and sort by due date
    const tasksWithDueDates = tasks
        .filter(task => {
            if (!task.due_date) return false;
            const dueDate = new Date(task.due_date);
            dueDate.setHours(0, 0, 0, 0); // Normalize to start of day for comparison
            return dueDate >= today && dueDate <= endDate;
        })
        .sort((a, b) => new Date(a.due_date) - new Date(b.due_date));
    
    console.log('Tasks with due dates in range:', tasksWithDueDates);
    
    if (tasksWithDueDates.length === 0) {
        container.innerHTML = `<p>No tasks with due dates found in the next ${timelineRange} days. Add due dates to tasks to see them in timeline view.</p>`;
        return;
    }
    
    // Create timeline scale
    const timelineScale = createTimelineScale(today, endDate);
    container.appendChild(timelineScale);
    
    // Create timeline tasks with proper lane assignment
    const timelineTasksContainer = document.createElement('div');
    timelineTasksContainer.className = 'timeline-tasks';
    
    // Assign lanes to prevent overlapping
    const lanes = assignTimelineLanes(tasksWithDueDates, today, endDate);
    
    tasksWithDueDates.forEach((task, index) => {
        const taskElement = createTimelineTaskElement(task, today, endDate, lanes[index]);
        timelineTasksContainer.appendChild(taskElement);
    });
    
    // Set the height of the timeline container based on the number of lanes used
    const maxLane = Math.max(...lanes);
    timelineTasksContainer.style.minHeight = `${(maxLane + 1) * 140}px`;
    
    container.appendChild(timelineTasksContainer);
}

function createTimelineScale(startDate, endDate) {
    const scale = document.createElement('div');
    scale.className = 'timeline-scale';
    
    const totalDays = Math.ceil((endDate - startDate) / (24 * 60 * 60 * 1000));
    const interval = Math.max(1, Math.floor(totalDays / 10)); // Show about 10 markers
    
    for (let i = 0; i <= totalDays; i += interval) {
        const date = new Date(startDate.getTime() + (i * 24 * 60 * 60 * 1000));
        const marker = document.createElement('div');
        marker.className = 'timeline-marker';
        marker.style.left = `${(i / totalDays) * 100}%`;
        marker.innerHTML = `
            <div class="timeline-date">${date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</div>
        `;
        scale.appendChild(marker);
    }
    
    return scale;
}

// Assign lanes to timeline tasks to prevent overlapping
function assignTimelineLanes(tasks, startDate, endDate) {
    const totalDays = Math.ceil((endDate - startDate) / (24 * 60 * 60 * 1000));
    const lanes = [];
    const laneOccupancy = []; // Track which positions are occupied in each lane
    
    tasks.forEach((task, index) => {
        const dueDate = new Date(task.due_date);
        const daysFromStart = Math.ceil((dueDate - startDate) / (24 * 60 * 60 * 1000));
        const position = (daysFromStart / totalDays) * 100;
        
        // Calculate the range this task will occupy (task width is 200px, timeline is typically 800-1000px)
        const taskWidth = 20; // Approximate percentage width of task card
        const startPos = Math.max(0, position - taskWidth/2);
        const endPos = Math.min(100, position + taskWidth/2);
        
        // Find the first available lane
        let assignedLane = 0;
        let laneFound = false;
        
        while (!laneFound) {
            // Initialize lane if it doesn't exist
            if (!laneOccupancy[assignedLane]) {
                laneOccupancy[assignedLane] = [];
            }
            
            // Check if this lane is available for this position range
            const isLaneAvailable = !laneOccupancy[assignedLane].some(occupied => 
                (startPos < occupied.end && endPos > occupied.start)
            );
            
            if (isLaneAvailable) {
                // Assign this lane and mark it as occupied
                laneOccupancy[assignedLane].push({ start: startPos, end: endPos });
                lanes[index] = assignedLane;
                laneFound = true;
            } else {
                // Try next lane
                assignedLane++;
            }
        }
    });
    
    return lanes;
}

function createTimelineTaskElement(task, startDate, endDate, lane = 0) {
    const taskElement = document.createElement('div');
    taskElement.className = `timeline-task ${task.status}`;
    
    const dueDate = new Date(task.due_date);
    const totalDays = Math.ceil((endDate - startDate) / (24 * 60 * 60 * 1000));
    const daysFromStart = Math.ceil((dueDate - startDate) / (24 * 60 * 60 * 1000));
    const position = (daysFromStart / totalDays) * 100;
    
    // Check if task is overdue
    const isOverdue = new Date() > dueDate && task.status !== 'done';
    if (isOverdue) {
        taskElement.classList.add('overdue');
    }
    
    // Position task horizontally and vertically
    taskElement.style.left = `${Math.max(0, Math.min(100, position))}%`;
    taskElement.style.top = `${lane * 140}px`; // 140px spacing between lanes
    
    // Calculate progress based on status
    let progress = 0;
    switch (task.status) {
        case 'todo': progress = 0; break;
        case 'in_progress': progress = 50; break;
        case 'done': progress = 100; break;
        case 'blocked': progress = 25; break;
    }
    
    taskElement.innerHTML = `
        <div class="timeline-task-content">
            <div class="timeline-task-header">
                <span class="task-type task-type-${task.type}">${task.type}</span>
                <span class="task-priority priority-${task.priority}">${task.priority}</span>
            </div>
            <h5 class="timeline-task-title">${task.title}</h5>
            <div class="timeline-task-progress">
                <div class="progress-bar" style="width: ${progress}%"></div>
            </div>
            <div class="timeline-task-meta">
                <span class="task-due-date">Due: ${dueDate.toLocaleDateString()}</span>
                ${isOverdue ? '<span class="overdue-indicator">OVERDUE</span>' : ''}
            </div>
        </div>
    `;
    
    return taskElement;
}

// Timeline control handlers
function initializeTimelineControls() {
    const rangeSelect = document.getElementById('timeline-range');
    const todayBtn = document.getElementById('timeline-today-btn');
    
    if (rangeSelect) {
        rangeSelect.addEventListener('change', (e) => {
            timelineRange = parseInt(e.target.value);
            if (currentView === 'timeline') {
                loadTimelineView();
            }
        });
    }
    
    if (todayBtn) {
        todayBtn.addEventListener('click', () => {
            // Scroll timeline to today (if implemented with horizontal scroll)
            const container = document.querySelector('.timeline-container');
            if (container) {
                container.scrollLeft = 0;
            }
        });
    }
}
