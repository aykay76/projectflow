// Application state
let currentEditingTask = null;
let currentView = 'kanban';
let hierarchyData = [];
let currentTheme = localStorage.getItem('theme') || 'light';

// Project management state
let currentProject = null;
let availableProjects = [];
let isLoadingProjects = false;

console.log('ProjectFlow app.js loaded - starting initialization');

// DOM elements
let modal, modalTitle, taskForm, newTaskBtn, cancelBtn, closeBtn;
let kanbanViewBtn, hierarchyViewBtn, timelineViewBtn;
let taskBoard, hierarchyView, timelineView;
let taskDetailModal, detailModalClose, detailEditBtn, detailDeleteBtn;

// Project management DOM elements
let projectSelector, projectDropdown, currentProjectDisplay;
let projectManagementBtn;

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    // Initialize DOM elements
    modal = document.getElementById('task-modal');
    modalTitle = document.getElementById('modal-title');
    taskForm = document.getElementById('task-form');
    newTaskBtn = document.getElementById('new-task-btn');
    cancelBtn = document.getElementById('cancel-btn');
    closeBtn = document.querySelector('.close');
    kanbanViewBtn = document.getElementById('kanban-view-btn');
    hierarchyViewBtn = document.getElementById('hierarchy-view-btn');
    timelineViewBtn = document.getElementById('timeline-view-btn');
    taskBoard = document.querySelector('.task-board');
    hierarchyView = document.getElementById('hierarchy-view');
    timelineView = document.getElementById('timeline-view');
    
    // Task Detail Modal elements
    taskDetailModal = document.getElementById('task-detail-modal');
    detailModalClose = document.getElementById('detail-modal-close');
    detailEditBtn = document.getElementById('detail-edit-btn');
    detailDeleteBtn = document.getElementById('detail-delete-btn');

    // Project management elements
    projectSelector = document.getElementById('project-selector-btn');
    projectDropdown = document.getElementById('project-dropdown');
    currentProjectDisplay = document.getElementById('current-project-display');
    projectManagementBtn = document.getElementById('manage-projects-btn');

    console.log('Task Detail Modal DOM elements found:', {
        taskDetailModal: !!taskDetailModal,
        detailModalClose: !!detailModalClose,
        detailEditBtn: !!detailEditBtn,
        detailDeleteBtn: !!detailDeleteBtn
    });
    initializeTheme();
    initializeEventListeners();
    initializeProjectManagement();
    initializeTaskDetailModal();
    initializeTimelineControls();
    initializeKeyboardShortcuts();
    initializeFiltering();
    initializeContextMenu();
    initializeAutoSave();
    initializeMobileEnhancements();
    enhanceSearch();
    initializePerformanceMonitoring();
    loadFilterState();
    updateOverdueIndicators();
    updateTaskCounts();
    
    // Load saved view preference
    const savedView = localStorage.getItem('projectflow_current_view');
    if (savedView && ['kanban', 'hierarchy', 'timeline'].includes(savedView)) {
        switchToView(savedView);
    }
});

// View Management
function switchToView(viewName) {
    if (!['kanban', 'hierarchy', 'timeline'].includes(viewName)) {
        console.error('Invalid view name:', viewName);
        return;
    }
    
    // Update current view state
    currentView = viewName;
    
    // Hide all views
    if (taskBoard) taskBoard.style.display = 'none';
    if (hierarchyView) hierarchyView.style.display = 'none';
    if (timelineView) timelineView.style.display = 'none';
    
    // Show selected view
    switch (viewName) {
        case 'kanban':
            if (taskBoard) taskBoard.style.display = 'block';
            break;
        case 'hierarchy':
            if (hierarchyView) {
                hierarchyView.style.display = 'block';
                loadHierarchyView();
            }
            break;
        case 'timeline':
            if (timelineView) {
                timelineView.style.display = 'block';
                loadTimelineView();
            }
            break;
    }
    
    // Update button states
    document.querySelectorAll('.view-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Add active class to current view button
    const activeBtn = document.getElementById(`${viewName}-view-btn`);
    if (activeBtn) {
        activeBtn.classList.add('active');
    }
    
    // Save view preference
    localStorage.setItem('projectflow_current_view', viewName);
    
    showMessage(`Switched to ${viewName.charAt(0).toUpperCase() + viewName.slice(1)} view! 👁️`, 'info', 2000);
}

function loadHierarchyView() {
    const hierarchyContainer = document.querySelector('.hierarchy-container');
    if (!hierarchyContainer) return;
    
    // Show loading message
    hierarchyContainer.innerHTML = '<div class="loading-message">Loading hierarchy view...</div>';
    
    // Fetch tasks and build hierarchy
    fetch('/api/tasks')
        .then(response => response.json())
        .then(tasks => {
            buildHierarchyTree(tasks, hierarchyContainer);
        })
        .catch(error => {
            console.error('Error loading hierarchy:', error);
            hierarchyContainer.innerHTML = '<div class="error-message">Failed to load hierarchy view</div>';
        });
}

function buildHierarchyTree(tasks, container) {
    console.log('Building hierarchy tree with', tasks.length, 'tasks');
    // Create a map of tasks by ID for quick lookup
    const taskMap = new Map();
    tasks.forEach(task => taskMap.set(task.id, { ...task, children: [] }));
    
    // Build parent-child relationships
    const rootTasks = [];
    tasks.forEach(task => {
        if (task.parent_id && taskMap.has(task.parent_id)) {
            taskMap.get(task.parent_id).children.push(taskMap.get(task.id));
        } else {
            rootTasks.push(taskMap.get(task.id));
        }
    });
    
    // Render the hierarchy
    container.innerHTML = renderHierarchyNode(rootTasks, 0);
    console.log('Hierarchy rendered, attaching click handlers...');
    
    // Attach click event listeners to clickable tasks
    attachHierarchyClickHandlers(container, taskMap);
}

function renderHierarchyNode(tasks, level) {
    if (!tasks || tasks.length === 0) return '';
    
    return tasks.map(task => {
        const indent = '  '.repeat(level);
        const hasChildren = task.children && task.children.length > 0;
        const typeIcon = getTaskTypeIcon(task.type);
        const priorityClass = `priority-${task.priority}`;
        
        return `
            <div class="hierarchy-item" data-id="${task.id}" style="margin-left: ${level * 15}px;">
                <div class="hierarchy-task clickable-task" data-task-id="${task.id}" tabindex="0" role="button" aria-label="View details for ${task.title}">
                    ${hasChildren ? '<span class="hierarchy-toggle">▶</span>' : '<span class="hierarchy-spacer"></span>'}
                    <span class="task-type ${priorityClass}">${typeIcon} ${task.type}</span>
                    <span class="task-title">${task.title}</span>
                    <span class="task-status status-${task.status}">${task.status}</span>
                </div>
                ${hasChildren ? `<div class="hierarchy-children">${renderHierarchyNode(task.children, level + 1)}</div>` : ''}
            </div>
        `;
    }).join('');
}

function getTaskTypeIcon(type) {
    const icons = {
        'epic': '🎯',
        'story': '📖',
        'task': '📋',
        'subtask': '📌'
    };
    return icons[type] || '📋';
}

function loadTimelineView() {
    const timelineContainer = document.querySelector('.timeline-container');
    if (!timelineContainer) return;
    
    // Show loading message
    timelineContainer.innerHTML = '<div class="loading-message">Loading timeline view...</div>';
    
    // Fetch tasks for timeline
    fetch('/api/tasks')
        .then(response => response.json())
        .then(tasks => {
            buildTimelineView(tasks, timelineContainer);
        })
        .catch(error => {
            console.error('Error loading timeline:', error);
            timelineContainer.innerHTML = '<div class="error-message">Failed to load timeline view</div>';
        });
}

function buildTimelineView(tasks, container) {
    // Filter tasks with dates
    const tasksWithDates = tasks.filter(task => task.due_date || task.started_at);
    
    if (tasksWithDates.length === 0) {
        container.innerHTML = '<div class="no-data-message">No tasks with dates found for timeline view</div>';
        return;
    }
    
    // Sort tasks by date
    tasksWithDates.sort((a, b) => {
        const dateA = new Date(a.due_date || a.started_at || a.created_at);
        const dateB = new Date(b.due_date || b.started_at || b.created_at);
        return dateA - dateB;
    });
    
    // Render timeline
    const timelineHTML = tasksWithDates.map(task => {
        const date = new Date(task.due_date || task.started_at || task.created_at);
        const dateStr = date.toLocaleDateString();
        const typeIcon = getTaskTypeIcon(task.type);
        const statusClass = `status-${task.status}`;
        const priorityClass = `priority-${task.priority}`;
        
        return `
            <div class="timeline-item ${statusClass}" data-id="${task.id}">
                <div class="timeline-date">${dateStr}</div>
                <div class="timeline-content">
                    <div class="timeline-task">
                        <span class="task-type ${priorityClass}">${typeIcon} ${task.type}</span>
                        <span class="task-title">${task.title}</span>
                        <span class="task-status">${task.status}</span>
                    </div>
                    ${task.description ? `<div class="timeline-description">${task.description}</div>` : ''}
                </div>
            </div>
        `;
    }).join('');
    
    container.innerHTML = `<div class="timeline-items">${timelineHTML}</div>`;
}

function initializeTimelineControls() {
    const timelineModeSelect = document.getElementById('timeline-mode');
    const timelineTodayBtn = document.getElementById('timeline-today-btn');
    const timelineRangeSelect = document.getElementById('timeline-range');
    
    if (timelineModeSelect) {
        timelineModeSelect.addEventListener('change', () => {
            if (currentView === 'timeline') {
                loadTimelineView();
            }
        });
    }
    
    if (timelineTodayBtn) {
        timelineTodayBtn.addEventListener('click', () => {
            // Scroll to today's date in timeline
            const today = new Date().toLocaleDateString();
            const todayElement = document.querySelector(`.timeline-date:contains("${today}")`);
            if (todayElement) {
                todayElement.scrollIntoView({ behavior: 'smooth' });
            }
        });
    }
    
    if (timelineRangeSelect) {
        timelineRangeSelect.addEventListener('change', () => {
            if (currentView === 'timeline') {
                loadTimelineView();
            }
        });
    }
}

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

// Project Management System
function initializeProjectManagement() {
    console.log('Initializing project management system');
    
    // Load saved project preference
    const savedProjectId = localStorage.getItem('projectflow_current_project');
    
    // Initialize project selector event listeners
    if (projectSelector) {
        projectSelector.addEventListener('click', toggleProjectDropdown);
    }
    
    // Close dropdown when clicking outside
    document.addEventListener('click', (event) => {
        if (!event.target.closest('.project-selector-wrapper')) {
            closeProjectDropdown();
        }
    });
    
    // Load projects and set current project
    loadAvailableProjects().then(() => {
        if (savedProjectId) {
            const savedProject = availableProjects.find(p => p.id === savedProjectId);
            if (savedProject) {
                setCurrentProject(savedProject);
            } else {
                // Fallback to first available project or create default
                handleInitialProjectSetup();
            }
        } else {
            handleInitialProjectSetup();
        }
    });
    
    // Initialize create project button
    const createProjectBtn = document.getElementById('create-project-btn');
    if (createProjectBtn) {
        createProjectBtn.addEventListener('click', () => {
            closeProjectDropdown();
            showCreateProjectModal();
        });
    }
    
    // Initialize manage projects button
    if (projectManagementBtn) {
        projectManagementBtn.addEventListener('click', () => {
            closeProjectDropdown();
            showProjectManagementModal();
        });
    }
}

async function loadAvailableProjects() {
    if (isLoadingProjects) return;
    
    isLoadingProjects = true;
    try {
        console.log('Loading available projects...');
        const response = await fetch('/api/projects');
        if (!response.ok) {
            throw new Error(`Failed to load projects: ${response.status}`);
        }
        
        availableProjects = await response.json();
        console.log('Loaded projects:', availableProjects);
        updateProjectDropdown();
        
    } catch (error) {
        console.error('Error loading projects:', error);
        showMessage('Failed to load projects', 'error');
        availableProjects = [];
    } finally {
        isLoadingProjects = false;
    }
}

function updateProjectDropdown() {
    const projectList = document.getElementById('project-list');
    if (!projectList) return;
    
    if (isLoadingProjects) {
        projectList.innerHTML = '<div class="project-loading">Loading projects...</div>';
        return;
    }
    
    if (availableProjects.length === 0) {
        projectList.innerHTML = `
            <div class="project-empty">
                <p>No projects found</p>
                <p>Create your first project to get started!</p>
            </div>
        `;
        return;
    }
    
    projectList.innerHTML = availableProjects.map(project => `
        <div class="project-item ${currentProject && currentProject.id === project.id ? 'selected' : ''}" 
             data-project-id="${project.id}">
            <div class="project-item-name">${escapeHtml(project.name)}</div>
            <div class="project-item-description">${escapeHtml(project.description || 'No description')}</div>
        </div>
    `).join('');
    
    // Add click handlers for project items
    projectList.querySelectorAll('.project-item').forEach(item => {
        item.addEventListener('click', () => {
            const projectId = item.dataset.projectId;
            const project = availableProjects.find(p => p.id === projectId);
            if (project) {
                setCurrentProject(project);
                closeProjectDropdown();
            }
        });
    });
}

function setCurrentProject(project) {
    console.log('Setting current project:', project);
    currentProject = project;
    
    // Update UI
    updateCurrentProjectDisplay();
    updateProjectSelectorButton();
    
    // Save to localStorage
    localStorage.setItem('projectflow_current_project', project.id);
    
    // Reload tasks for new project context
    refreshCurrentView();
    
    showMessage(`Switched to project: ${project.name}`, 'success', 3000);
}

function updateCurrentProjectDisplay() {
    if (currentProjectDisplay) {
        if (currentProject) {
            currentProjectDisplay.textContent = `Current: ${currentProject.name}`;
        } else {
            currentProjectDisplay.textContent = 'No Project Selected';
        }
    }
}

function updateProjectSelectorButton() {
    const selectorText = document.getElementById('project-selector-text');
    if (selectorText) {
        if (currentProject) {
            selectorText.textContent = currentProject.name;
        } else {
            selectorText.textContent = 'Select Project';
        }
    }
}

function toggleProjectDropdown() {
    if (!projectDropdown) return;
    
    const isOpen = projectDropdown.style.display !== 'none';
    if (isOpen) {
        closeProjectDropdown();
    } else {
        openProjectDropdown();
    }
}

function openProjectDropdown() {
    if (!projectDropdown || !projectSelector) return;
    
    projectDropdown.style.display = 'block';
    projectSelector.classList.add('open');
    updateProjectDropdown();
}

function closeProjectDropdown() {
    if (!projectDropdown || !projectSelector) return;
    
    projectDropdown.style.display = 'none';
    projectSelector.classList.remove('open');
}

async function handleInitialProjectSetup() {
    if (availableProjects.length > 0) {
        // Use first available project
        setCurrentProject(availableProjects[0]);
    } else {
        // Create a default project
        try {
            const defaultProject = await createProject({
                name: 'Default Project',
                description: 'Default project for task management',
                display_prefix: 'PF'
            });
            availableProjects.push(defaultProject);
            setCurrentProject(defaultProject);
            updateProjectDropdown();
        } catch (error) {
            console.error('Failed to create default project:', error);
            showMessage('Failed to create default project', 'error');
        }
    }
}

async function createProject(projectData) {
    try {
        const response = await fetch('/api/projects', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(projectData)
        });
        
        if (!response.ok) {
            throw new Error(`Failed to create project: ${response.status}`);
        }
        
        return await response.json();
    } catch (error) {
        console.error('Error creating project:', error);
        throw error;
    }
}

function refreshCurrentView() {
    // Refresh the current view with new project context
    switch (currentView) {
        case 'kanban':
            loadTasks();
            break;
        case 'hierarchy':
            loadHierarchyView();
            break;
        case 'timeline':
            loadTimelineView();
            break;
    }
}

function showCreateProjectModal() {
    // TODO: Implement create project modal
    console.log('Create project modal not yet implemented');
    showMessage('Create project feature coming soon!', 'info');
}

function showProjectManagementModal() {
    // TODO: Implement project management modal
    console.log('Project management modal not yet implemented');
    showMessage('Project management feature coming soon!', 'info');
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
    
    // Add event listeners for hierarchy toggle functionality
    document.addEventListener('click', (event) => {
        if (event.target.classList.contains('hierarchy-toggle')) {
            event.preventDefault();
            event.stopPropagation(); // Prevent event from bubbling to parent task element
            
            const hierarchyItem = event.target.closest('.hierarchy-item');
            const children = hierarchyItem.querySelector('.hierarchy-children');
            
            if (children) {
                if (children.style.display === 'none') {
                    children.style.display = 'block';
                    event.target.textContent = '▼';
                } else {
                    children.style.display = 'none';
                    event.target.textContent = '▶';
                }
            }
        }
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

// Task Detail Modal Functions
function initializeTaskDetailModal() {
    if (detailModalClose) {
        detailModalClose.addEventListener('click', closeTaskDetailModal);
    }
    
    if (taskDetailModal) {
        taskDetailModal.addEventListener('click', (e) => {
            if (e.target === taskDetailModal) {
                closeTaskDetailModal();
            }
        });
    }
    
    if (detailEditBtn) {
        detailEditBtn.addEventListener('click', () => {
            const taskId = detailEditBtn.dataset.taskId;
            if (taskId) {
                closeTaskDetailModal();
                editTask(taskId);
            }
        });
    }
    
    if (detailDeleteBtn) {
        detailDeleteBtn.addEventListener('click', () => {
            const taskId = detailDeleteBtn.dataset.taskId;
            if (taskId && confirm('Are you sure you want to delete this task?')) {
                deleteTask(taskId);
                closeTaskDetailModal();
            }
        });
    }
    
    // Handle Escape key
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && taskDetailModal && taskDetailModal.style.display === 'block') {
            closeTaskDetailModal();
        }
    });
}

function attachHierarchyClickHandlers(container, taskMap) {
    const clickableTasks = container.querySelectorAll('.clickable-task');
    console.log('Attaching click handlers to', clickableTasks.length, 'clickable tasks');
    
    clickableTasks.forEach(taskElement => {
        const taskId = taskElement.dataset.taskId;
        
        taskElement.addEventListener('click', (e) => {
            // Don't open modal if clicking on the hierarchy toggle
            if (e.target.classList.contains('hierarchy-toggle')) {
                return;
            }
            
            e.preventDefault();
            console.log('Task clicked:', taskId);
            showTaskDetailModal(taskId, taskMap);
        });
        
        taskElement.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                console.log('Task activated via keyboard:', taskId);
                showTaskDetailModal(taskId, taskMap);
            }
        });
    });
}

function showTaskDetailModal(taskId, taskMap) {
    console.log('Showing task detail modal for:', taskId);
    const task = taskMap.get(taskId);
    
    if (!task || !taskDetailModal) {
        console.error('Task not found or modal not available:', !!task, !!taskDetailModal);
        return;
    }
    
    // Populate modal with task data
    const titleElement = document.getElementById('detail-task-title');
    const descriptionElement = document.getElementById('detail-task-description');
    const typeElement = document.getElementById('detail-task-type');
    const statusElement = document.getElementById('detail-task-status');
    const priorityElement = document.getElementById('detail-task-priority');
    const dueDateElement = document.getElementById('detail-task-due-date');
    const startedAtElement = document.getElementById('detail-task-started-at');
    const createdAtElement = document.getElementById('detail-task-created-at');
    const childrenElement = document.getElementById('detail-task-children');
    
    // Set basic info
    if (titleElement) titleElement.textContent = task.title;
    if (descriptionElement) {
        descriptionElement.textContent = task.description || 'No description provided';
        descriptionElement.style.fontStyle = task.description ? 'normal' : 'italic';
        descriptionElement.style.color = task.description ? 'inherit' : 'var(--text-secondary)';
    }
    
    // Set type with icon
    if (typeElement) {
        const typeIcon = getTaskTypeIcon(task.type);
        typeElement.innerHTML = `${typeIcon} ${task.type}`;
        typeElement.className = `task-type task-type-${task.type}`;
    }
    
    // Set status
    if (statusElement) {
        statusElement.textContent = task.status.replace('_', ' ');
        statusElement.className = `task-status status-${task.status}`;
    }
    
    // Set priority
    if (priorityElement) {
        priorityElement.textContent = task.priority;
        priorityElement.className = `task-priority priority-${task.priority}`;
    }
    
    // Handle dates
    const dueDateRow = document.getElementById('detail-due-date-row');
    if (task.due_date && dueDateElement && dueDateRow) {
        dueDateElement.textContent = new Date(task.due_date).toLocaleDateString();
        dueDateRow.style.display = 'flex';
    } else if (dueDateRow) {
        dueDateRow.style.display = 'none';
    }
    
    const startedAtRow = document.getElementById('detail-started-at-row');
    if (task.started_at && startedAtElement && startedAtRow) {
        startedAtElement.textContent = new Date(task.started_at).toLocaleString();
        startedAtRow.style.display = 'flex';
    } else if (startedAtRow) {
        startedAtRow.style.display = 'none';
    }
    
    if (createdAtElement) {
        createdAtElement.textContent = new Date(task.created_at).toLocaleString();
    }
    
    // Handle children count
    const childrenRow = document.getElementById('detail-children-row');
    if (task.children && task.children.length > 0 && childrenElement && childrenRow) {
        childrenElement.textContent = `${task.children.length} subtask${task.children.length !== 1 ? 's' : ''}`;
        childrenRow.style.display = 'flex';
    } else if (childrenRow) {
        childrenRow.style.display = 'none';
    }
    
    // Set task ID on action buttons
    if (detailEditBtn) detailEditBtn.dataset.taskId = taskId;
    if (detailDeleteBtn) detailDeleteBtn.dataset.taskId = taskId;
    
    // Show modal
    taskDetailModal.style.display = 'block';
    document.body.style.overflow = 'hidden';
}

function closeTaskDetailModal() {
    if (taskDetailModal) {
        taskDetailModal.style.display = 'none';
        document.body.style.overflow = '';
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

// Drag and Drop functionality
let draggedTask = null;

function initializeDraggableCards() {
    // Set draggable attribute for all task cards
    const taskCards = document.querySelectorAll('.task-card');
    taskCards.forEach(card => {
        card.setAttribute('draggable', 'true');
        card.style.cursor = 'grab';
    });
}

function handleDragStart(event) {
    draggedTask = event.target;
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

function handleDragOver(event) {
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

function handleDrop(event) {
    event.preventDefault();
    
    const taskId = event.dataTransfer.getData('text/plain');
    const targetColumn = event.target.closest('.column');
    
    if (targetColumn && draggedTask) {
        const newStatus = targetColumn.dataset.status;
        const currentStatus = draggedTask.closest('.column').dataset.status;
        
        // Only move if dropping in a different column
        if (newStatus !== currentStatus) {
            // Move the task visually first for immediate feedback
            const targetTaskList = targetColumn.querySelector('.task-list');
            targetTaskList.appendChild(draggedTask);
            
            // Update the task status via API
            moveTask(taskId, newStatus);
        }
    }
    
    // Clean up drag state
    cleanupDragState();
}

function cleanupDragState() {
    if (draggedTask) {
        draggedTask.style.opacity = '';
        draggedTask.style.cursor = 'grab';
        draggedTask.classList.remove('dragging');
        draggedTask = null;
    }
    
    // Remove visual feedback from all columns
    document.querySelectorAll('.column').forEach(column => {
        column.classList.remove('drag-active', 'drag-over');
    });
}

// Handle drag end event to clean up if drop doesn't occur
document.addEventListener('dragend', (event) => {
    if (event.target.classList.contains('task-card')) {
        cleanupDragState();
    }
});

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

// Filter persistence
function saveFilterState() {
    const filterState = {
        search: document.getElementById('filter-search')?.value || '',
        status: document.getElementById('filter-status')?.value || '',
        priority: document.getElementById('filter-priority')?.value || '',
        type: document.getElementById('filter-type')?.value || '',
        overdue: document.getElementById('filter-overdue')?.checked || false
    };
    localStorage.setItem('projectflow_filters', JSON.stringify(filterState));
}

function loadFilterState() {
    const saved = localStorage.getItem('projectflow_filters');
    if (saved) {
        try {
            const filterState = JSON.parse(saved);
            Object.keys(filterState).forEach(key => {
                const element = document.getElementById(`filter-${key}`);
                if (element) {
                    if (element.type === 'checkbox') {
                        element.checked = filterState[key];
                    } else {
                        element.value = filterState[key];
                    }
                }
            });
            // Apply the loaded filters and update visual state
            setTimeout(() => {
                applyFilters();
                updateFilterVisualState();
            }, 100);
        } catch (error) {
            console.error('Error loading filter state:', error);
            // Ensure all tasks are visible if filter loading fails
            setTimeout(() => {
                applyFilters();
                updateFilterVisualState();
            }, 100);
        }
    } else {
        // No saved filters, ensure all tasks are visible
        setTimeout(() => {
            applyFilters();
            updateFilterVisualState();
        }, 100);
    }
}

// Performance monitoring
function initializePerformanceMonitoring() {
    let animationFrames = 0;
    let lastTime = performance.now();
    
    function measureFPS() {
        animationFrames++;
        const currentTime = performance.now();
        
        if (currentTime >= lastTime + 1000) {
            const fps = Math.round((animationFrames * 1000) / (currentTime - lastTime));
            
            // Only log if FPS is concerning
            if (fps < 45) {
                console.warn(`Low FPS detected: ${fps}fps`);
            }
            
            animationFrames = 0;
            lastTime = currentTime;
        }
        
        requestAnimationFrame(measureFPS);
    }
    
    requestAnimationFrame(measureFPS);
}

// Enhanced mobile interactions
function initializeMobileEnhancements() {
    // Add touch feedback for mobile
    document.addEventListener('touchstart', (e) => {
        if (e.target.classList.contains('task-card') || e.target.closest('.task-card')) {
            e.target.closest('.task-card')?.classList.add('touch-active');
        }
    });

    document.addEventListener('touchend', (e) => {
        document.querySelectorAll('.touch-active').forEach(card => {
            card.classList.remove('touch-active');
        });
    });

    // Enhanced swipe gestures for task cards
    let startX, startY;
    document.addEventListener('touchstart', (e) => {
        if (e.target.closest('.task-card')) {
            startX = e.touches[0].clientX;
            startY = e.touches[0].clientY;
        }
    });

    document.addEventListener('touchmove', (e) => {
        if (!startX || !startY) return;
        
        const card = e.target.closest('.task-card');
        if (!card) return;

        const currentX = e.touches[0].clientX;
        const currentY = e.touches[0].clientY;
        const diffX = currentX - startX;
        const diffY = currentY - startY;

        // If horizontal swipe is significant and vertical is minimal
        if (Math.abs(diffX) > 50 && Math.abs(diffY) < 30) {
            e.preventDefault();
            
            if (diffX > 0) {
                // Swipe right - show green background (mark as done)
                card.style.background = 'linear-gradient(90deg, var(--success-color) 0%, var(--bg-primary) 100%)';
                card.style.transform = `translateX(${Math.min(diffX / 4, 20)}px)`;
            } else {
                // Swipe left - show red background (delete)
                card.style.background = 'linear-gradient(90deg, var(--bg-primary) 0%, var(--danger-color) 100%)';
                card.style.transform = `translateX(${Math.max(diffX / 4, -20)}px)`;
            }
        }
    });

    document.addEventListener('touchend', (e) => {
        const card = e.target.closest('.task-card');
        if (card && startX) {
            const currentX = e.changedTouches[0].clientX;
            const diffX = currentX - startX;

            // Reset visual state
            card.style.background = '';
            card.style.transform = '';

            // Execute action if swipe was significant enough
            if (Math.abs(diffX) > 100) {
                const taskId = card.dataset.id;
                if (diffX > 0) {
                    // Right swipe - mark as done
                    moveTask(taskId, 'done');
                } else {
                    // Left swipe - delete (with confirmation)
                    if (confirm('Delete this task?')) {
                        deleteTask(taskId);
                    }
                }
            }
        }
        startX = null;
        startY = null;
    });
}

// Enhanced search with better performance
function enhanceSearch() {
    const searchInput = document.getElementById('filter-search');
    if (!searchInput) return;

    // Add search suggestions
    const suggestionsContainer = document.createElement('div');
    suggestionsContainer.className = 'search-suggestions';
    searchInput.parentNode.appendChild(suggestionsContainer);

    let searchCache = new Map();
    
    searchInput.addEventListener('input', debounce((e) => {
        const query = e.target.value.toLowerCase().trim();
        
        if (query.length === 0) {
            suggestionsContainer.style.display = 'none';
            applyFilters();
            return;
        }

        // Check cache first
        if (searchCache.has(query)) {
            showSearchSuggestions(searchCache.get(query), suggestionsContainer);
        } else {
            // Search through task titles and descriptions
            const tasks = document.querySelectorAll('.task-card');
            const matches = [];
            
            tasks.forEach(task => {
                const title = task.querySelector('.task-title')?.textContent.toLowerCase() || '';
                const desc = task.querySelector('.task-description')?.textContent.toLowerCase() || '';
                
                if (title.includes(query) || desc.includes(query)) {
                    matches.push({
                        id: task.dataset.id,
                        title: task.querySelector('.task-title')?.textContent || '',
                        type: 'task'
                    });
                }
            });

            searchCache.set(query, matches);
            showSearchSuggestions(matches, suggestionsContainer);
        }
        
        applyFilters();
    }, 300));

    // Hide suggestions when clicking outside
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.filter-group')) {
            suggestionsContainer.style.display = 'none';
        }
    });
}

function showSearchSuggestions(matches, container) {
    if (matches.length === 0) {
        container.style.display = 'none';
        return;
    }

    container.innerHTML = matches.slice(0, 5).map(match => `
        <div class="search-suggestion" data-task-id="${match.id}">
            <span class="suggestion-icon">📋</span>
            <span class="suggestion-text">${match.title}</span>
        </div>
    `).join('');

    container.style.display = 'block';

    // Add click handlers for suggestions
    container.querySelectorAll('.search-suggestion').forEach(item => {
        item.addEventListener('click', () => {
            const taskId = item.dataset.taskId;
            const taskCard = document.querySelector(`[data-id="${taskId}"]`);
            if (taskCard) {
                taskCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
                taskCard.classList.add('highlight');
                setTimeout(() => taskCard.classList.remove('highlight'), 2000);
            }
            container.style.display = 'none';
        });
    });
}

// Filtering functionality
function initializeFiltering() {
    const filterToggleBtn = document.getElementById('filter-toggle-btn');
    const filterPanel = document.getElementById('filter-panel');
    const clearFiltersBtn = document.getElementById('clear-filters-btn');
    const saveFilterBtn = document.getElementById('save-filter-btn');
    
    // Toggle filter panel
    if (filterToggleBtn) {
        filterToggleBtn.addEventListener('click', toggleFilterPanel);
    }
    
    // Clear all filters
    if (clearFiltersBtn) {
        clearFiltersBtn.addEventListener('click', clearAllFilters);
    }
    
    // Save current filter state
    if (saveFilterBtn) {
        saveFilterBtn.addEventListener('click', () => {
            saveFilterState();
            showMessage('Filter settings saved! 💾', 'success', 2000);
        });
    }
    
    // Add event listeners to filter inputs
    const filterInputs = [
        'filter-search',
        'filter-status', 
        'filter-priority',
        'filter-type',
        'filter-overdue'
    ];
    
    filterInputs.forEach(inputId => {
        const input = document.getElementById(inputId);
        if (input) {
            input.addEventListener('input', debounce(() => {
                applyFilters();
                saveFilterState();
                updateFilterVisualState();
            }, 300));
            input.addEventListener('change', () => {
                applyFilters();
                saveFilterState(); 
                updateFilterVisualState();
            });
        }
    });
}

function toggleFilterPanel() {
    const filterPanel = document.getElementById('filter-panel');
    const filterToggleBtn = document.getElementById('filter-toggle-btn');
    
    if (filterPanel.style.display === 'none' || filterPanel.style.display === '') {
        filterPanel.style.display = 'block';
        filterPanel.classList.add('show');
        filterToggleBtn.textContent = '🔍 Hide Filters';
        filterToggleBtn.classList.add('active');
    } else {
        filterPanel.style.display = 'none';
        filterPanel.classList.remove('show');
        filterToggleBtn.textContent = '🔍 Filters';
        filterToggleBtn.classList.remove('active');
    }
}

function applyFilters() {
    const searchTerm = document.getElementById('filter-search')?.value.toLowerCase() || '';
    const statusFilter = document.getElementById('filter-status')?.value || '';
    const priorityFilter = document.getElementById('filter-priority')?.value || '';
    const typeFilter = document.getElementById('filter-type')?.value || '';
    const overdueFilter = document.getElementById('filter-overdue')?.value || '';
    
    const taskCards = document.querySelectorAll('.task-card');
    let visibleCount = 0;
    
    // Update filter visual indicators
    updateFilterVisualState();
    
    // If no filters are active, show all tasks
    const hasActiveFilters = searchTerm || statusFilter || priorityFilter || typeFilter || overdueFilter;
    if (!hasActiveFilters) {
        taskCards.forEach(card => {
            card.style.display = 'block';
            visibleCount++;
        });
        updateFilterResultsCount(visibleCount, taskCards.length);
        hideActiveFilters();
        return;
    }
    
    // Show active filters
    showActiveFilters();
    
    taskCards.forEach(card => {
        let visible = true;
        
        // Search filter
        if (searchTerm) {
            const title = card.querySelector('.task-title')?.textContent.toLowerCase() || '';
            const description = card.querySelector('.task-description')?.textContent.toLowerCase() || '';
            if (!title.includes(searchTerm) && !description.includes(searchTerm)) {
                visible = false;
            }
        }
        
        // Status filter
        if (statusFilter) {
            const column = card.closest('.column');
            const columnStatus = column?.dataset.status || '';
            if (columnStatus !== statusFilter) {
                visible = false;
            }
        }
        
        // Priority filter
        if (priorityFilter) {
            const priorityElement = card.querySelector('.task-priority');
            const priority = priorityElement?.textContent.toLowerCase() || '';
            if (priority !== priorityFilter) {
                visible = false;
            }
        }
        
        // Type filter
        if (typeFilter) {
            const typeElement = card.querySelector('.task-type');
            const type = typeElement?.textContent.toLowerCase() || '';
            if (type !== typeFilter) {
                visible = false;
            }
        }
        
        // Overdue filter
        if (overdueFilter === 'overdue') {
            const dueDateElement = card.querySelector('.task-due-date');
            if (dueDateElement) {
                const dueDate = new Date(dueDateElement.textContent);
                const now = new Date();
                if (dueDate >= now) {
                    visible = false;
                }
            } else {
                visible = false;
            }
        } else if (overdueFilter === 'due-soon') {
            const dueDateElement = card.querySelector('.task-due-date');
            if (dueDateElement) {
                const dueDate = new Date(dueDateElement.textContent);
                const now = new Date();
                const threeDaysFromNow = new Date(now.getTime() + (3 * 24 * 60 * 60 * 1000));
                if (dueDate < now || dueDate > threeDaysFromNow) {
                    visible = false;
                }
            } else {
                visible = false;
            }
        }
        
        // Show/hide card
        if (visible) {
            card.style.display = 'block';
            visibleCount++;
        } else {
            card.style.display = 'none';
        }
    });
    
    // Update filter results count
    updateFilterResultsCount(visibleCount, taskCards.length);
}

function clearAllFilters() {
    document.getElementById('filter-search').value = '';
    document.getElementById('filter-status').value = '';
    document.getElementById('filter-priority').value = '';
    document.getElementById('filter-type').value = '';
    document.getElementById('filter-overdue').value = '';
    
    // Clear saved filters
    localStorage.removeItem('projectflow_filters');
    
    // Update visual state
    updateFilterVisualState();
    hideActiveFilters();
    
    // Show all tasks
    const taskCards = document.querySelectorAll('.task-card');
    taskCards.forEach(card => {
        card.style.display = 'block';
    });
    
    updateFilterResultsCount(taskCards.length, taskCards.length);
    showMessage('All filters cleared! 🧹', 'info', 2000);
}

// New functions for filter visual management
function updateFilterVisualState() {
    const filterToggleBtn = document.getElementById('filter-toggle-btn');
    const filterInputs = [
        'filter-search',
        'filter-status',
        'filter-priority', 
        'filter-type',
        'filter-overdue'
    ];
    
    let activeCount = 0;
    
    // Check each filter and update its visual state
    filterInputs.forEach(inputId => {
        const input = document.getElementById(inputId);
        if (input && input.value && input.value.trim() !== '') {
            input.classList.add('filter-active');
            activeCount++;
        } else if (input) {
            input.classList.remove('filter-active');
        }
    });
    
    // Update filter button state
    if (filterToggleBtn) {
        if (activeCount > 0) {
            filterToggleBtn.classList.add('filter-toggle-active');
            filterToggleBtn.setAttribute('data-active-count', activeCount);
            filterToggleBtn.title = `${activeCount} filter(s) active`;
        } else {
            filterToggleBtn.classList.remove('filter-toggle-active');
            filterToggleBtn.removeAttribute('data-active-count');
            filterToggleBtn.title = 'Toggle Filters (F)';
        }
    }
}

function showActiveFilters() {
    const container = document.getElementById('active-filters-container');
    const list = document.getElementById('active-filters-list');
    
    if (!container || !list) return;
    
    const activeFilters = getActiveFilters();
    
    if (activeFilters.length === 0) {
        hideActiveFilters();
        return;
    }
    
    list.innerHTML = activeFilters.map(filter => `
        <div class="filter-badge">
            <span>${filter.label}: ${filter.value}</span>
            <button class="filter-badge-remove" onclick="removeFilter('${filter.key}')" title="Remove filter">×</button>
        </div>
    `).join('');
    
    container.style.display = 'block';
}

function hideActiveFilters() {
    const container = document.getElementById('active-filters-container');
    if (container) {
        container.style.display = 'none';
    }
}

function getActiveFilters() {
    const filters = [];
    
    const searchValue = document.getElementById('filter-search')?.value;
    if (searchValue && searchValue.trim()) {
        filters.push({
            key: 'search',
            label: 'Search',
            value: `"${searchValue}"`
        });
    }
    
    const statusValue = document.getElementById('filter-status')?.value;
    if (statusValue) {
        const statusLabels = {
            'todo': 'To Do',
            'in_progress': 'In Progress',
            'done': 'Done',
            'blocked': 'Blocked'
        };
        filters.push({
            key: 'status',
            label: 'Status',
            value: statusLabels[statusValue] || statusValue
        });
    }
    
    const priorityValue = document.getElementById('filter-priority')?.value;
    if (priorityValue) {
        filters.push({
            key: 'priority',
            label: 'Priority',
            value: priorityValue.charAt(0).toUpperCase() + priorityValue.slice(1)
        });
    }
    
    const typeValue = document.getElementById('filter-type')?.value;
    if (typeValue) {
        filters.push({
            key: 'type',
            label: 'Type',
            value: typeValue.charAt(0).toUpperCase() + typeValue.slice(1)
        });
    }
    
    const overdueValue = document.getElementById('filter-overdue')?.value;
    if (overdueValue) {
        const overdueLabels = {
            'overdue': 'Overdue Tasks',
            'due-soon': 'Due Soon'
        };
        filters.push({
            key: 'overdue',
            label: 'Due Status',
            value: overdueLabels[overdueValue] || overdueValue
        });
    }
    
    return filters;
}

function removeFilter(filterKey) {
    const input = document.getElementById(`filter-${filterKey}`);
    if (input) {
        input.value = '';
        applyFilters();
        saveFilterState();
    }
}

function updateFilterResultsCount(visible, total) {
    const filterContent = document.querySelector('.filter-content h3');
    if (filterContent) {
        filterContent.textContent = `🔍 Advanced Filters (${visible}/${total} tasks)`;
    }
}

// Task count and overdue indicators
function updateTaskCounts() {
    const columns = document.querySelectorAll('.column');
    columns.forEach(column => {
        const tasks = column.querySelectorAll('.task-card:not([style*="display: none"])');
        const header = column.querySelector('h3');
        const status = column.dataset.status;
        
        let emoji = '';
        switch(status) {
            case 'todo': emoji = '📝'; break;
            case 'in_progress': emoji = '🔄'; break;
            case 'review': emoji = '👀'; break;
            case 'done': emoji = '✅'; break;
            default: emoji = '📋';
        }
        
        header.textContent = `${emoji} ${header.textContent.replace(/^[📝🔄👀✅📋]\s*/, '').replace(/\s*\(\d+\)$/, '')} (${tasks.length})`;
    });
}

function updateOverdueIndicators() {
    const taskCards = document.querySelectorAll('.task-card');
    const now = new Date();
    
    taskCards.forEach(card => {
        const dueDateElement = card.querySelector('.task-due-date');
        if (dueDateElement) {
            const dueDate = new Date(dueDateElement.textContent);
            
            // Remove existing indicators
            card.classList.remove('overdue', 'due-soon');
            
            if (dueDate < now) {
                card.classList.add('overdue');
            } else if (dueDate <= new Date(now.getTime() + (3 * 24 * 60 * 60 * 1000))) {
                card.classList.add('due-soon');
            }
        }
    });
}

// Utility Functions
function escapeHtml(text) {
    if (!text) return '';
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return text.replace(/[&<>"']/g, function(m) { return map[m]; });
}
