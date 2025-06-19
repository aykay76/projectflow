/**
 * Filter Manager - Handles search, filtering, and task organization
 */
class FilterManager {
    constructor() {
        this.filters = {
            search: '',
            status: 'all',
            priority: 'all',
            type: 'all',
            assignee: 'all',
            dueDate: 'all'
        };
        
        this.searchHistory = JSON.parse(localStorage.getItem('projectflow_search_history') || '[]');
        this.maxSearchHistory = 10;
        
        this.initializeFiltering();
        this.initializeSearch();
    }

    /**
     * Initialize filtering system
     */
    initializeFiltering() {
        // Status filter
        const statusFilter = document.getElementById('status-filter');
        if (statusFilter) {
            statusFilter.addEventListener('change', (e) => {
                this.setFilter('status', e.target.value);
                this.applyFilters();
            });
        }

        // Priority filter
        const priorityFilter = document.getElementById('priority-filter');
        if (priorityFilter) {
            priorityFilter.addEventListener('change', (e) => {
                this.setFilter('priority', e.target.value);
                this.applyFilters();
            });
        }

        // Type filter
        const typeFilter = document.getElementById('type-filter');
        if (typeFilter) {
            typeFilter.addEventListener('change', (e) => {
                this.setFilter('type', e.target.value);
                this.applyFilters();
            });
        }

        // Due date filter
        const dueDateFilter = document.getElementById('due-date-filter');
        if (dueDateFilter) {
            dueDateFilter.addEventListener('change', (e) => {
                this.setFilter('dueDate', e.target.value);
                this.applyFilters();
            });
        }

        // Clear filters button
        const clearFiltersBtn = document.getElementById('clear-filters-btn');
        if (clearFiltersBtn) {
            clearFiltersBtn.addEventListener('click', () => this.clearAllFilters());
        }

        // Load saved filter state
        this.loadFilterState();
    }

    /**
     * Initialize search functionality
     */
    initializeSearch() {
        const searchInput = document.getElementById('search-input');
        const searchBtn = document.getElementById('search-btn');
        const clearSearchBtn = document.getElementById('clear-search-btn');

        if (searchInput) {
            // Real-time search with debouncing
            let searchTimeout;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(searchTimeout);
                searchTimeout = setTimeout(() => {
                    this.setFilter('search', e.target.value);
                    this.applyFilters();
                }, 300);
            });

            // Search on Enter key
            searchInput.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.preventDefault();
                    this.performSearch(e.target.value);
                }
            });

            // Show search suggestions
            searchInput.addEventListener('focus', () => this.showSearchSuggestions());
            searchInput.addEventListener('blur', () => {
                // Delay hiding to allow clicking on suggestions
                setTimeout(() => this.hideSearchSuggestions(), 200);
            });
        }

        if (searchBtn) {
            searchBtn.addEventListener('click', () => {
                const query = searchInput ? searchInput.value : '';
                this.performSearch(query);
            });
        }

        if (clearSearchBtn) {
            clearSearchBtn.addEventListener('click', () => this.clearSearch());
        }

        // Initialize advanced search modal
        this.initializeAdvancedSearch();
    }

    /**
     * Initialize advanced search modal
     */
    initializeAdvancedSearch() {
        const advancedSearchBtn = document.getElementById('advanced-search-btn');
        const advancedSearchModal = document.getElementById('advanced-search-modal');
        const advancedSearchForm = document.getElementById('advanced-search-form');
        const closeAdvancedSearch = document.getElementById('close-advanced-search');

        if (advancedSearchBtn && advancedSearchModal) {
            advancedSearchBtn.addEventListener('click', () => {
                advancedSearchModal.style.display = 'block';
            });
        }

        if (closeAdvancedSearch && advancedSearchModal) {
            closeAdvancedSearch.addEventListener('click', () => {
                advancedSearchModal.style.display = 'none';
            });
        }

        if (advancedSearchForm) {
            advancedSearchForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.performAdvancedSearch(new FormData(e.target));
                if (advancedSearchModal) {
                    advancedSearchModal.style.display = 'none';
                }
            });
        }
    }

    /**
     * Set a filter value
     */
    setFilter(filterType, value) {
        this.filters[filterType] = value;
        this.saveFilterState();
    }

    /**
     * Get current filter value
     */
    getFilter(filterType) {
        return this.filters[filterType];
    }

    /**
     * Apply all current filters to tasks
     */
    applyFilters() {
        const tasks = window.taskManager?.getAllTasks() || [];
        const filteredTasks = this.filterTasks(tasks);
        
        // Update task display
        this.updateTaskDisplay(filteredTasks);
        
        // Update filter indicators
        this.updateFilterIndicators();
        
        // Update task counts
        this.updateTaskCounts(filteredTasks);
        
        console.log(`Applied filters: ${Object.entries(this.filters).filter(([k, v]) => v !== 'all' && v !== '').map(([k, v]) => `${k}:${v}`).join(', ')}`);
    }

    /**
     * Filter tasks based on current filters
     */
    filterTasks(tasks) {
        return tasks.filter(task => {
            // Search filter
            if (this.filters.search && !this.matchesSearch(task, this.filters.search)) {
                return false;
            }

            // Status filter
            if (this.filters.status !== 'all' && task.status !== this.filters.status) {
                return false;
            }

            // Priority filter
            if (this.filters.priority !== 'all' && task.priority !== this.filters.priority) {
                return false;
            }

            // Type filter
            if (this.filters.type !== 'all' && task.type !== this.filters.type) {
                return false;
            }

            // Due date filter
            if (this.filters.dueDate !== 'all' && !this.matchesDueDateFilter(task, this.filters.dueDate)) {
                return false;
            }

            return true;
        });
    }

    /**
     * Check if task matches search query
     */
    matchesSearch(task, query) {
        if (!query) return true;
        
        const searchTerms = query.toLowerCase().split(' ');
        const searchableText = [
            task.title,
            task.description,
            task.status,
            task.priority,
            task.type,
            task.display_id
        ].join(' ').toLowerCase();

        return searchTerms.every(term => searchableText.includes(term));
    }

    /**
     * Check if task matches due date filter
     */
    matchesDueDateFilter(task, filter) {
        if (!task.due_date) return filter === 'no-date';
        
        const dueDate = new Date(task.due_date);
        const today = new Date();
        const daysDiff = Math.ceil((dueDate - today) / (1000 * 60 * 60 * 24));

        switch (filter) {
            case 'overdue':
                return daysDiff < 0;
            case 'today':
                return daysDiff === 0;
            case 'tomorrow':
                return daysDiff === 1;
            case 'this-week':
                return daysDiff >= 0 && daysDiff <= 7;
            case 'next-week':
                return daysDiff > 7 && daysDiff <= 14;
            case 'this-month':
                return daysDiff >= 0 && daysDiff <= 30;
            case 'no-date':
                return false; // Task has a date, so doesn't match 'no-date'
            default:
                return true;
        }
    }

    /**
     * Update task display based on filtered tasks
     */
    updateTaskDisplay(filteredTasks) {
        const currentView = window.viewManager?.currentView || 'kanban';
        
        switch (currentView) {
            case 'kanban':
                this.updateKanbanDisplay(filteredTasks);
                break;
            case 'hierarchy':
                this.updateHierarchyDisplay(filteredTasks);
                break;
            case 'timeline':
                this.updateTimelineDisplay(filteredTasks);
                break;
        }
    }

    /**
     * Update kanban board display
     */
    updateKanbanDisplay(filteredTasks) {
        const taskCards = document.querySelectorAll('.task-card');
        const filteredTaskIds = new Set(filteredTasks.map(t => t.id));
        
        taskCards.forEach(card => {
            const taskId = card.dataset.id;
            if (filteredTaskIds.has(taskId)) {
                card.style.display = 'block';
            } else {
                card.style.display = 'none';
            }
        });
    }

    /**
     * Update hierarchy view display
     */
    updateHierarchyDisplay(filteredTasks) {
        if (window.viewManager?.buildHierarchyTree) {
            const hierarchyContainer = document.querySelector('.hierarchy-container');
            if (hierarchyContainer) {
                window.viewManager.buildHierarchyTree(filteredTasks, hierarchyContainer);
            }
        }
    }

    /**
     * Update timeline view display
     */
    updateTimelineDisplay(filteredTasks) {
        if (window.viewManager?.buildTimelineView) {
            const timelineContainer = document.querySelector('.timeline-container');
            if (timelineContainer) {
                window.viewManager.buildTimelineView(filteredTasks, timelineContainer);
            }
        }
    }

    /**
     * Update filter indicators
     */
    updateFilterIndicators() {
        const activeFilters = Object.entries(this.filters)
            .filter(([key, value]) => value !== 'all' && value !== '');
        
        const filterIndicator = document.getElementById('filter-indicator');
        if (filterIndicator) {
            if (activeFilters.length > 0) {
                filterIndicator.style.display = 'block';
                filterIndicator.textContent = `${activeFilters.length} filter${activeFilters.length > 1 ? 's' : ''} active`;
            } else {
                filterIndicator.style.display = 'none';
            }
        }

        // Update clear filters button
        const clearFiltersBtn = document.getElementById('clear-filters-btn');
        if (clearFiltersBtn) {
            clearFiltersBtn.style.display = activeFilters.length > 0 ? 'block' : 'none';
        }
    }

    /**
     * Update task counts
     */
    updateTaskCounts(filteredTasks) {
        const statusCounts = {
            total: filteredTasks.length,
            todo: filteredTasks.filter(t => t.status === 'todo').length,
            'in_progress': filteredTasks.filter(t => t.status === 'in_progress').length,
            done: filteredTasks.filter(t => t.status === 'done').length,
            blocked: filteredTasks.filter(t => t.status === 'blocked').length
        };

        // Update status counts in UI
        Object.entries(statusCounts).forEach(([status, count]) => {
            const countElement = document.getElementById(`${status}-count`);
            if (countElement) {
                countElement.textContent = count;
            }
        });

        // Update column headers with counts
        const columns = document.querySelectorAll('.task-column');
        columns.forEach(column => {
            const status = column.dataset.status;
            if (status && statusCounts[status] !== undefined) {
                const header = column.querySelector('h3');
                if (header) {
                    const baseText = header.textContent.split(' (')[0];
                    header.textContent = `${baseText} (${statusCounts[status]})`;
                }
            }
        });
    }

    /**
     * Perform search
     */
    performSearch(query) {
        this.setFilter('search', query);
        this.addToSearchHistory(query);
        this.applyFilters();
        
        if (query) {
            window.messageManager?.showMessage(`Searching for: "${query}"`, 'info', 2000);
        }
    }

    /**
     * Perform advanced search
     */
    performAdvancedSearch(formData) {
        const searchQuery = formData.get('search-query') || '';
        const status = formData.get('search-status') || 'all';
        const priority = formData.get('search-priority') || 'all';
        const type = formData.get('search-type') || 'all';
        const dueDate = formData.get('search-due-date') || 'all';

        // Set all filters at once
        this.filters = {
            search: searchQuery,
            status: status,
            priority: priority,
            type: type,
            assignee: 'all',
            dueDate: dueDate
        };

        // Update UI filters
        this.updateFilterUI();
        
        // Apply filters
        this.applyFilters();
        
        // Add to search history if there's a query
        if (searchQuery) {
            this.addToSearchHistory(searchQuery);
        }
        
        window.messageManager?.showMessage('Advanced search applied!', 'info', 2000);
    }

    /**
     * Update filter UI elements
     */
    updateFilterUI() {
        Object.entries(this.filters).forEach(([filterType, value]) => {
            const element = document.getElementById(`${filterType}-filter`);
            if (element) {
                element.value = value;
            }
        });

        const searchInput = document.getElementById('search-input');
        if (searchInput) {
            searchInput.value = this.filters.search;
        }
    }

    /**
     * Clear search
     */
    clearSearch() {
        this.setFilter('search', '');
        const searchInput = document.getElementById('search-input');
        if (searchInput) {
            searchInput.value = '';
        }
        this.applyFilters();
    }

    /**
     * Clear all filters
     */
    clearAllFilters() {
        this.filters = {
            search: '',
            status: 'all',
            priority: 'all',
            type: 'all',
            assignee: 'all',
            dueDate: 'all'
        };
        
        this.updateFilterUI();
        this.applyFilters();
        
        window.messageManager?.showMessage('All filters cleared!', 'info', 2000);
    }

    /**
     * Add query to search history
     */
    addToSearchHistory(query) {
        if (!query || this.searchHistory.includes(query)) return;
        
        this.searchHistory.unshift(query);
        if (this.searchHistory.length > this.maxSearchHistory) {
            this.searchHistory = this.searchHistory.slice(0, this.maxSearchHistory);
        }
        
        localStorage.setItem('projectflow_search_history', JSON.stringify(this.searchHistory));
    }

    /**
     * Show search suggestions
     */
    showSearchSuggestions() {
        if (this.searchHistory.length === 0) return;
        
        let suggestionsContainer = document.getElementById('search-suggestions');
        if (!suggestionsContainer) {
            suggestionsContainer = document.createElement('div');
            suggestionsContainer.id = 'search-suggestions';
            suggestionsContainer.className = 'search-suggestions';
            
            const searchInput = document.getElementById('search-input');
            if (searchInput && searchInput.parentNode) {
                searchInput.parentNode.appendChild(suggestionsContainer);
            }
        }
        
        suggestionsContainer.innerHTML = this.searchHistory
            .map(query => `<div class="search-suggestion" data-query="${query}">${query}</div>`)
            .join('');
        
        suggestionsContainer.style.display = 'block';
        
        // Add click handlers
        suggestionsContainer.querySelectorAll('.search-suggestion').forEach(suggestion => {
            suggestion.addEventListener('click', () => {
                const query = suggestion.dataset.query;
                const searchInput = document.getElementById('search-input');
                if (searchInput) {
                    searchInput.value = query;
                }
                this.performSearch(query);
                this.hideSearchSuggestions();
            });
        });
    }

    /**
     * Hide search suggestions
     */
    hideSearchSuggestions() {
        const suggestionsContainer = document.getElementById('search-suggestions');
        if (suggestionsContainer) {
            suggestionsContainer.style.display = 'none';
        }
    }

    /**
     * Save filter state to localStorage
     */
    saveFilterState() {
        localStorage.setItem('projectflow_filters', JSON.stringify(this.filters));
    }

    /**
     * Load filter state from localStorage
     */
    loadFilterState() {
        const savedFilters = localStorage.getItem('projectflow_filters');
        if (savedFilters) {
            try {
                this.filters = { ...this.filters, ...JSON.parse(savedFilters) };
                this.updateFilterUI();
            } catch (error) {
                console.warn('Failed to load filter state:', error);
            }
        }
    }

    /**
     * Get current filters
     */
    getCurrentFilters() {
        return { ...this.filters };
    }

    /**
     * Get filtered tasks count
     */
    getFilteredTasksCount() {
        const tasks = window.taskManager?.getAllTasks() || [];
        return this.filterTasks(tasks).length;
    }
}

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = FilterManager;
}
