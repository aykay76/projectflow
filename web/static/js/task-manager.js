/**
 * Task management functionality
 */
import { apiClient } from './api-client.js';
import { stateManager } from './state-manager.js';
import { showMessage, escapeHtml, formatDate, formatDateTime, getTaskTypeIcon } from './utils.js';

class TaskManager {
    constructor() {
        this.initializeEventListeners();
    }

    initializeEventListeners() {
        // Listen for project changes to reload tasks
        stateManager.addEventListener('project-changed', (data) => {
            if (data.newProject) {
                this.refreshTasks();
            }
        });
    }

    async loadTasks(projectId = null) {
        try {        const currentProject = stateManager.getCurrentProject();
        const targetProjectId = projectId || (currentProject ? currentProject.display_prefix : null);
            
            if (!targetProjectId) {
                console.warn('No project selected for loading tasks');
                return [];
            }
            
            console.log(`Loading tasks for project: ${targetProjectId}`);
            const tasks = await apiClient.loadTasks(targetProjectId);
            console.log(`Loaded ${tasks.length} tasks for project ${currentProject?.name || 'Unknown'}`);
            
            return tasks;
        } catch (error) {
            console.error('Error loading tasks:', error);
            return [];
        }
    }

    async createTask(taskData) {
        const currentProject = stateManager.getCurrentProject();
        if (!currentProject) {
            showMessage('Please select a project before creating tasks', 'error');
            throw new Error('No project selected');
        }

        // Add project context to task
        const taskWithProject = {
            ...taskData,
            project_id: currentProject.display_prefix
        };

        try {
            const newTask = await apiClient.createTask(taskWithProject);
            showMessage('Task created successfully! ✨', 'success');
            
            // Refresh current view
            this.refreshTasks();
            return newTask;
        } catch (error) {
            console.error('Error creating task:', error);
            throw error;
        }
    }

    async updateTask(taskId, taskData) {
        try {
            const updatedTask = await apiClient.updateTask(taskId, taskData);
            showMessage('Task updated successfully! 🎉', 'success');
            
            // Refresh current view
            this.refreshTasks();
            return updatedTask;
        } catch (error) {
            console.error('Error updating task:', error);
            throw error;
        }
    }

    async deleteTask(taskId) {
        if (!confirm('Are you sure you want to delete this task?')) {
            return false;
        }

        try {
            await apiClient.deleteTask(taskId);
            
            // Remove the task card from the DOM immediately for better UX
            const taskCard = document.querySelector(`[data-id="${taskId}"]`);
            if (taskCard) {
                taskCard.remove();
            }
            
            showMessage('Task deleted successfully!', 'success');
            return true;
        } catch (error) {
            console.error('Error deleting task:', error);
            // Refresh to restore the UI if deletion failed
            this.refreshTasks();
            throw error;
        }
    }

    async moveTask(taskId, newStatus) {
        try {
            await apiClient.moveTask(taskId, newStatus);
            console.log(`Task ${taskId} moved to ${newStatus}`);
            // Note: Don't show success message for drag-and-drop operations
            // as they provide immediate visual feedback
        } catch (error) {
            console.error('Error moving task:', error);
            showMessage('Failed to move task', 'error');
            // Refresh to restore the correct state
            this.refreshTasks();
            throw error;
        }
    }

    async getTaskForEditing(taskId) {
        try {
            return await apiClient.getTask(taskId);
        } catch (error) {
            console.error('Error loading task for editing:', error);
            showMessage('Failed to load task for editing.', 'error');
            throw error;
        }
    }

    createTaskCard(task) {
        const card = document.createElement('div');
        card.className = 'task-card';
        card.dataset.id = task.id;
        card.draggable = true;
        
        card.innerHTML = `
            <div class="task-header">
                <div class="task-header-left">
                    <span class="task-type task-type-${escapeHtml(task.type)}">${escapeHtml(task.type)}</span>
                    ${task.display_id ? `<span class="task-project-id">${escapeHtml(task.display_id)}</span>` : ''}
                </div>
                <span class="task-priority priority-${escapeHtml(task.priority)}">${escapeHtml(task.priority)}</span>
            </div>
            <h4 class="task-title">${escapeHtml(task.title)}</h4>
            <p class="task-description">${escapeHtml(task.description || '')}</p>
            <div class="task-meta">
                <span class="task-date">${formatDate(task.created_at)}</span>
                ${task.started_at ? `<span class="task-started-at">Started: ${formatDateTime(task.started_at)}</span>` : ''}
                ${task.due_date ? `<span class="task-due-date">Due: ${formatDate(task.due_date)}</span>` : ''}
                ${task.children && task.children.length > 0 ? `<span class="task-children">${task.children.length} subtasks</span>` : ''}
            </div>
        `;
        
        return card;
    }

    buildHierarchyTree(tasks, container) {
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
        container.innerHTML = this.renderHierarchyNode(rootTasks, 0);
        console.log('Hierarchy rendered, attaching click handlers...');
        
        // Attach click event listeners to clickable tasks
        this.attachHierarchyClickHandlers(container, taskMap);
    }

    renderHierarchyNode(tasks, level) {
        if (!tasks || tasks.length === 0) return '';
        
        return tasks.map(task => {
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
                    ${hasChildren ? `<div class="hierarchy-children">${this.renderHierarchyNode(task.children, level + 1)}</div>` : ''}
                </div>
            `;
        }).join('');
    }

    attachHierarchyClickHandlers(container, taskMap) {
        // Add click handlers for hierarchy toggles
        container.addEventListener('click', (event) => {
            if (event.target.classList.contains('hierarchy-toggle')) {
                event.preventDefault();
                event.stopPropagation();
                
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

        // Add click handlers for task details
        container.querySelectorAll('.clickable-task').forEach(element => {
            element.addEventListener('click', (event) => {
                if (!event.target.classList.contains('hierarchy-toggle')) {
                    const taskId = element.dataset.taskId;
                    const task = Array.from(taskMap.values()).find(t => t.id === taskId);
                    if (task) {
                        this.showTaskDetail(task);
                    }
                }
            });
        });
    }

    buildTimelineView(tasks, container) {
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

    showTaskDetail(task) {
        // This will be handled by the UI manager
        stateManager.dispatchEvent('show-task-detail', { task });
    }

    // Helper method to refresh tasks in current view
    refreshTasks() {
        const currentView = stateManager.getCurrentView();
        stateManager.dispatchEvent('refresh-view', { view: currentView });
    }

    // Task form data processing
    processTaskFormData(formData) {
        const taskData = {
            title: formData.get('title'),
            description: formData.get('description'),
            type: formData.get('type'),
            priority: formData.get('priority'),
            status: formData.get('status'),
            due_date: formData.get('due_date') || null,
            started_at: formData.get('started_at') ? new Date(formData.get('started_at')).toISOString() : null
        };

        // Add parent_id if provided
        const parentId = formData.get('parent_id');
        if (parentId) {
            taskData.parent_id = parentId;
        }

        return taskData;
    }

    // Task form population
    populateTaskForm(task, form) {
        form.querySelector('#task-title').value = task.title || '';
        form.querySelector('#task-description').value = task.description || '';
        form.querySelector('#task-type').value = task.type || 'task';
        form.querySelector('#task-priority').value = task.priority || 'medium';
        form.querySelector('#task-status').value = task.status || 'todo';
        form.querySelector('#task-due-date').value = task.due_date ? task.due_date.split('T')[0] : '';
        
        // Handle start date - convert from RFC3339 to datetime-local format
        if (task.started_at) {
            const startDate = new Date(task.started_at);
            const year = startDate.getFullYear();
            const month = String(startDate.getMonth() + 1).padStart(2, '0');
            const day = String(startDate.getDate()).padStart(2, '0');
            const hours = String(startDate.getHours()).padStart(2, '0');
            const minutes = String(startDate.getMinutes()).padStart(2, '0');
            form.querySelector('#task-started-at').value = `${year}-${month}-${day}T${hours}:${minutes}`;
        } else {
            form.querySelector('#task-started-at').value = '';
        }
    }
}

// Export both the class and singleton instance
export { TaskManager };
export const taskManager = new TaskManager();
