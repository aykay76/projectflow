/**
 * Utility functions for common operations
 */

// String utilities
export function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Date formatting utilities
export function formatDate(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'short', 
        day: 'numeric' 
    });
}

export function formatDateTime(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
        month: 'short', 
        day: 'numeric', 
        hour: '2-digit',
        minute: '2-digit'
    });
}

// Debounce utility
export function debounce(func, wait) {
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

// DOM utilities
export function showMessage(message, type = 'info', duration = 5000) {
    // Create or get message container
    let messageContainer = document.getElementById('message-container');
    if (!messageContainer) {
        messageContainer = document.createElement('div');
        messageContainer.id = 'message-container';
        messageContainer.className = 'message-container';
        document.body.appendChild(messageContainer);
    }

    // Create message element
    const messageElement = document.createElement('div');
    messageElement.className = `message ${type}`;
    messageElement.innerHTML = `
        <span class="message-text">${escapeHtml(message)}</span>
        <button class="message-close" onclick="this.parentElement.remove()">×</button>
    `;

    // Add to container
    messageContainer.appendChild(messageElement);

    // Auto-remove after duration
    if (duration > 0) {
        setTimeout(() => {
            if (messageElement.parentNode) {
                messageElement.remove();
            }
        }, duration);
    }

    // Add fade-in animation
    setTimeout(() => messageElement.classList.add('show'), 10);
}

export function showLoadingOverlay(message = 'Loading...') {
    let overlay = document.getElementById('loading-overlay');
    if (!overlay) {
        overlay = document.createElement('div');
        overlay.id = 'loading-overlay';
        overlay.className = 'loading-overlay';
        overlay.innerHTML = `
            <div class="loading-content">
                <div class="loading-spinner"></div>
                <div class="loading-text">${escapeHtml(message)}</div>
            </div>
        `;
        document.body.appendChild(overlay);
    } else {
        overlay.querySelector('.loading-text').textContent = message;
    }
    overlay.style.display = 'flex';
}

export function hideLoadingOverlay() {
    const overlay = document.getElementById('loading-overlay');
    if (overlay) {
        overlay.style.display = 'none';
    }
}

// Task type icons
export function getTaskTypeIcon(type) {
    const icons = {
        'epic': '🎯',
        'story': '📖',
        'task': '📋',
        'subtask': '📌'
    };
    return icons[type] || '📋';
}

// Task sorting utilities
/**
 * Extract numeric suffix from task display ID
 * @param {string} displayId - Task display ID (e.g., "PF-123", "ABC-456")
 * @returns {number} - Numeric suffix or 0 if not found
 */
export function extractTaskNumber(displayId) {
    if (!displayId || typeof displayId !== 'string') {
        return 0;
    }
    
    // Match pattern: prefix-number (e.g., PF-123, ABC-456)
    const match = displayId.match(/^[A-Z]+-(\d+)$/);
    if (match && match[1]) {
        return parseInt(match[1], 10);
    }
    
    return 0;
}

/**
 * Compare two tasks by their display ID numeric suffix
 * @param {Object} taskA - First task object
 * @param {Object} taskB - Second task object
 * @returns {number} - Comparison result (-1, 0, 1)
 */
export function compareTasksByNumber(taskA, taskB) {
    const numberA = extractTaskNumber(taskA.display_id);
    const numberB = extractTaskNumber(taskB.display_id);
    
    // Sort in ascending order (lowest numbers first)
    return numberA - numberB;
}

/**
 * Sort an array of tasks by their display ID numeric suffix
 * @param {Array} tasks - Array of task objects
 * @returns {Array} - Sorted array of tasks
 */
export function sortTasksByNumber(tasks) {
    if (!Array.isArray(tasks)) {
        return [];
    }
    
    return [...tasks].sort(compareTasksByNumber);
}

// Button state management
export function setButtonLoading(button, loading) {
    if (!button) return;
    
    if (loading) {
        button.classList.add('loading');
        button.disabled = true;
    } else {
        button.classList.remove('loading');
        button.disabled = false;
    }
}
