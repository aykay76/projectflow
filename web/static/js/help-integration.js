/**
 * Context-Sensitive Help Integration
 * Provides consistent help functionality throughout ProjectFlow
 */

export class HelpIntegration {
    constructor(docsManager) {
        this.docsManager = docsManager;
        this.isInitialized = false;
        this.helpElementsAdded = new Set();
        this.init();
    }

    init() {
        if (this.isInitialized) {
            console.log('HelpIntegration already initialized, skipping...');
            return;
        }
        
        this.addHelpStyles();
        this.setupGlobalHelp();
        this.isInitialized = true;
        console.log('HelpIntegration initialized');
    }

    /**
     * Add CSS for help icons and tooltips
     */
    addHelpStyles() {
        const style = document.createElement('style');
        style.textContent = `
            /* Help Icon Styles */
            .help-icon {
                display: inline-flex;
                align-items: center;
                justify-content: center;
                width: 18px;
                height: 18px;
                border-radius: 50%;
                background: var(--color-text-muted);
                color: var(--color-bg-primary);
                font-size: 12px;
                font-weight: 600;
                margin-left: 0.5rem;
                cursor: pointer;
                transition: all 0.2s ease;
                text-decoration: none;
                user-select: none;
            }

            .help-icon:hover {
                background: var(--color-primary);
                color: white;
                transform: scale(1.1);
            }

            .help-icon.inline {
                width: 16px;
                height: 16px;
                font-size: 11px;
                margin-left: 0.25rem;
                vertical-align: middle;
            }

            /* Help Link Styles */
            .help-link {
                color: var(--color-primary);
                text-decoration: none;
                font-size: 0.875rem;
                display: inline-flex;
                align-items: center;
                gap: 0.25rem;
                transition: color 0.2s ease;
            }

            .help-link:hover {
                color: var(--color-primary-dark);
                text-decoration: underline;
            }

            .help-link .icon {
                font-size: 0.75rem;
            }

            /* Help Tooltip Styles */
            .help-tooltip {
                position: relative;
                display: inline-block;
            }

            .help-tooltip .tooltip-content {
                visibility: hidden;
                width: 200px;
                background-color: var(--color-bg-tertiary);
                color: var(--color-text-primary);
                text-align: left;
                border-radius: var(--border-radius);
                padding: 0.75rem;
                position: absolute;
                z-index: 1000;
                bottom: 125%;
                left: 50%;
                margin-left: -100px;
                opacity: 0;
                transition: opacity 0.3s;
                border: 1px solid var(--color-border);
                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
                font-size: 0.875rem;
                line-height: 1.4;
            }

            .help-tooltip .tooltip-content::after {
                content: "";
                position: absolute;
                top: 100%;
                left: 50%;
                margin-left: -5px;
                border-width: 5px;
                border-style: solid;
                border-color: var(--color-border) transparent transparent transparent;
            }

            .help-tooltip:hover .tooltip-content {
                visibility: visible;
                opacity: 1;
            }

            /* Help Section Styles */
            .help-section {
                margin-top: 1rem;
                padding: 1rem;
                background: var(--color-bg-secondary);
                border-radius: var(--border-radius);
                border-left: 3px solid var(--color-primary);
            }

            .help-section h4 {
                margin: 0 0 0.5rem 0;
                color: var(--color-text-primary);
                font-size: 0.875rem;
                font-weight: 600;
            }

            .help-section p {
                margin: 0 0 0.5rem 0;
                font-size: 0.875rem;
                color: var(--color-text-secondary);
                line-height: 1.4;
            }

            .help-section .help-links {
                display: flex;
                gap: 1rem;
                margin-top: 0.75rem;
            }

            /* Error Help Styles */
            .error-help {
                margin-top: 0.75rem;
                padding-top: 0.75rem;
                border-top: 1px solid var(--color-border-light);
            }

            .error-help .help-text {
                color: var(--color-text-muted);
                font-size: 0.875rem;
                margin-bottom: 0.5rem;
            }
        `;
        document.head.appendChild(style);
    }

    /**
     * Setup global help functionality
     */
    setupGlobalHelp() {
        // Make help functions available globally
        window.ProjectFlowHelp = {
            openDoc: (docName) => this.openDoc(docName),
            addHelpIcon: (element, docName, tooltip) => this.addHelpIcon(element, docName, tooltip),
            addHelpLink: (element, docName, text) => this.addHelpLink(element, docName, text),
            addHelpSection: (container, title, description, links) => this.addHelpSection(container, title, description, links)
        };
    }

    /**
     * Open a specific documentation section
     */
    openDoc(docName) {
        if (this.docsManager) {
            this.docsManager.openDocument(docName);
        } else {
            console.warn('DocsManager not available');
        }
    }

    /**
     * Add a help icon next to an element
     */
    addHelpIcon(element, docName, tooltip = null) {
        if (!element) return;
        
        // Check if help icon already exists to prevent duplicates
        if (element.querySelector('.help-icon')) {
            return element.querySelector('.help-icon');
        }

        const helpIcon = document.createElement('span');
        helpIcon.className = 'help-icon';
        helpIcon.textContent = '?';
        helpIcon.title = tooltip || `Get help about ${docName}`;
        helpIcon.setAttribute('data-doc', docName);
        
        helpIcon.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            this.openDoc(docName);
        });

        // Add tooltip if provided
        if (tooltip) {
            helpIcon.classList.add('help-tooltip');
            const tooltipContent = document.createElement('span');
            tooltipContent.className = 'tooltip-content';
            tooltipContent.textContent = tooltip;
            helpIcon.appendChild(tooltipContent);
        }

        element.appendChild(helpIcon);
        return helpIcon;
    }

    /**
     * Add a help link
     */
    addHelpLink(container, docName, text = '📚 Get Help') {
        if (!container) return;

        const helpLink = document.createElement('a');
        helpLink.href = '#';
        helpLink.className = 'help-link';
        helpLink.innerHTML = `<span class="icon">📚</span> ${text}`;
        
        helpLink.addEventListener('click', (e) => {
            e.preventDefault();
            this.openDoc(docName);
        });

        container.appendChild(helpLink);
        return helpLink;
    }

    /**
     * Add a help section with links
     */
    addHelpSection(container, title, description, links = []) {
        if (!container) return;

        const helpSection = document.createElement('div');
        helpSection.className = 'help-section';

        const titleElement = document.createElement('h4');
        titleElement.textContent = title;
        helpSection.appendChild(titleElement);

        if (description) {
            const descElement = document.createElement('p');
            descElement.textContent = description;
            helpSection.appendChild(descElement);
        }

        if (links.length > 0) {
            const linksContainer = document.createElement('div');
            linksContainer.className = 'help-links';

            links.forEach(link => {
                const linkElement = document.createElement('a');
                linkElement.href = '#';
                linkElement.className = 'help-link';
                linkElement.textContent = link.text;
                
                linkElement.addEventListener('click', (e) => {
                    e.preventDefault();
                    this.openDoc(link.docName);
                });

                linksContainer.appendChild(linkElement);
            });

            helpSection.appendChild(linksContainer);
        }

        container.appendChild(helpSection);
        return helpSection;
    }

    /**
     * Add help to error messages
     */
    addErrorHelp(errorContainer, helpText, docName) {
        if (!errorContainer) return;

        const errorHelp = document.createElement('div');
        errorHelp.className = 'error-help';

        const helpTextElement = document.createElement('div');
        helpTextElement.className = 'help-text';
        helpTextElement.textContent = helpText;
        errorHelp.appendChild(helpTextElement);

        const helpLink = document.createElement('a');
        helpLink.href = '#';
        helpLink.className = 'help-link';
        helpLink.innerHTML = '📚 Get Help';
        
        helpLink.addEventListener('click', (e) => {
            e.preventDefault();
            this.openDoc(docName);
        });

        errorHelp.appendChild(helpLink);
        errorContainer.appendChild(errorHelp);
        return errorHelp;
    }

    /**
     * Clear all help elements to prevent duplicates
     */
    clearHelpElements() {
        // Remove all existing help icons
        const existingIcons = document.querySelectorAll('.help-icon');
        existingIcons.forEach(icon => icon.remove());
        
        // Remove help sections
        const helpSections = document.querySelectorAll('.chat-help-hints, .modal-help-section, .filter-help-section');
        helpSections.forEach(section => section.remove());
        
        // Clear tracking
        this.helpElementsAdded.clear();
        
        console.log('Cleared existing help elements');
    }

    /**
     * Add contextual help based on the current view
     */
    addContextualHelp() {
        if (this.helpElementsAdded.has('contextual')) {
            console.log('Contextual help already added, skipping...');
            return;
        }
        
        console.log('Adding contextual help...');
        
        // Chat interface help
        this.addChatHelp();
        
        // Task creation help
        this.addTaskCreationHelp();
        
        // Project management help
        this.addProjectManagementHelp();
        
        // Kanban board help
        this.addKanbanHelp();
        
        // Filter help
        this.addFilterHelp();
        
        this.helpElementsAdded.add('contextual');
    }

    /**
     * Add help to chat interface
     */
    addChatHelp() {
        if (this.helpElementsAdded.has('chat')) return;
        
        const chatHeader = document.querySelector('.chat-header h3');
        if (chatHeader) {
            this.addHelpIcon(chatHeader, 'chat-interface-guide', 'Learn about chat commands and features');
        }

        const chatInput = document.querySelector('.chat-input-container');
        if (chatInput && !chatInput.querySelector('.chat-help-hints')) {
            const helpSection = document.createElement('div');
            helpSection.className = 'chat-help-hints';
            helpSection.innerHTML = `
                <p style="font-size: 0.75rem; color: var(--color-text-muted); margin: 0.5rem 0 0 0;">
                    💡 Try: "Create a task" or "Show me all projects" 
                    <a href="#" class="help-link" style="margin-left: 0.5rem;">📚 Chat Guide</a>
                </p>
            `;
            
            const helpLink = helpSection.querySelector('.help-link');
            helpLink.addEventListener('click', (e) => {
                e.preventDefault();
                this.openDoc('chat-interface-guide');
            });
            
            chatInput.appendChild(helpSection);
        }
        
        this.helpElementsAdded.add('chat');
    }

    /**
     * Add help to task creation
     */
    addTaskCreationHelp() {
        if (this.helpElementsAdded.has('task-creation')) return;
        
        const newTaskBtn = document.querySelector('#new-task-btn');
        if (newTaskBtn) {
            this.addHelpIcon(newTaskBtn, 'user-guide', 'Learn about creating and managing tasks');
        }

        const taskForm = document.querySelector('#task-form');
        if (taskForm) {
            const typeField = taskForm.querySelector('[name="type"]');
            if (typeField && typeField.parentElement) {
                this.addHelpIcon(typeField.parentElement, 'user-guide', 'Learn about task types: Epic, Story, Task, Subtask');
            }

            const priorityField = taskForm.querySelector('[name="priority"]');
            if (priorityField && priorityField.parentElement) {
                this.addHelpIcon(priorityField.parentElement, 'user-guide', 'Understand priority levels and when to use them');
            }
        }
        
        this.helpElementsAdded.add('task-creation');
    }

    /**
     * Add help to project management
     */
    addProjectManagementHelp() {
        if (this.helpElementsAdded.has('project-management')) return;
        
        const projectSelector = document.querySelector('#project-selector-btn');
        if (projectSelector) {
            this.addHelpIcon(projectSelector, 'user-guide', 'Learn about project organization and management');
        }

        const createProjectBtn = document.querySelector('#create-project-btn');
        if (createProjectBtn) {
            this.addHelpIcon(createProjectBtn, 'user-guide', 'Best practices for creating and organizing projects');
        }
        
        this.helpElementsAdded.add('project-management');
    }

    /**
     * Add help to Kanban board
     */
    addKanbanHelp() {
        if (this.helpElementsAdded.has('kanban')) return;
        
        const kanbanView = document.querySelector('#kanban-view');
        if (kanbanView) {
            const columns = kanbanView.querySelectorAll('.column h3');
            columns.forEach(column => {
                if (column.textContent.includes('To Do')) {
                    this.addHelpIcon(column, 'user-guide', 'Tasks that are ready to be started');
                } else if (column.textContent.includes('In Progress')) {
                    this.addHelpIcon(column, 'user-guide', 'Tasks currently being worked on');
                } else if (column.textContent.includes('Done')) {
                    this.addHelpIcon(column, 'user-guide', 'Completed tasks ready for review');
                } else if (column.textContent.includes('Blocked')) {
                    this.addHelpIcon(column, 'user-guide', 'Tasks that cannot proceed due to dependencies');
                }
            });
        }
        
        this.helpElementsAdded.add('kanban');
    }

    /**
     * Add help to filters
     */
    addFilterHelp() {
        if (this.helpElementsAdded.has('filter')) return;
        
        const filterBtn = document.querySelector('#filter-toggle-btn');
        if (filterBtn) {
            this.addHelpIcon(filterBtn, 'user-guide', 'Learn about filtering and searching tasks');
        }

        const filterPanel = document.querySelector('#filter-panel');
        if (filterPanel) {
            const filterTitle = filterPanel.querySelector('h3');
            if (filterTitle) {
                this.addHelpIcon(filterTitle, 'user-guide', 'Advanced filtering and search capabilities');
            }
        }
        
        this.helpElementsAdded.add('filter');
    }
}

export default HelpIntegration;
