/**
 * ChatManager - Handles the chat interface and communication with the backend
 */

import { ApiClient } from './api-client.js';
import { showMessage } from './utils.js';

export class ChatManager {
    constructor(apiClient, notificationManager) {
        this.apiClient = apiClient || new ApiClient();
        this.notificationManager = notificationManager;
        this.isOpen = false;
        this.isLoading = false;
        this.currentConversationId = null;
        this.messages = [];
        
        // Initialize chat interface
        this.init();
        
        // Load conversation from localStorage
        this.loadConversationFromStorage();
    }

    /**
     * Initialize the chat interface
     */
    init() {
        this.createChatElements();
        this.bindEvents();
        console.log('ChatManager initialized');
    }

    /**
     * Create chat UI elements
     */
    createChatElements() {
        // Create chat toggle button
        this.toggleBtn = document.createElement('button');
        this.toggleBtn.className = 'chat-toggle-btn';
        this.toggleBtn.innerHTML = '💬';
        this.toggleBtn.title = 'Open Chat Assistant';
        this.toggleBtn.setAttribute('aria-label', 'Open Chat Assistant');
        
        // Create chat panel
        this.chatPanel = document.createElement('div');
        this.chatPanel.className = 'chat-panel';
        this.chatPanel.innerHTML = `
            <div class="chat-header">
                <h3>💬 Chat Assistant</h3>
                <div class="chat-header-actions">
                    <button class="chat-close-btn" title="Close Chat" aria-label="Close Chat">×</button>
                </div>
            </div>
            <div class="chat-messages" id="chat-messages"></div>
            <div class="chat-input-container">
                <div class="chat-input-wrapper">
                    <textarea 
                        class="chat-input" 
                        id="chat-input" 
                        placeholder="Ask me to create tasks, list projects, or anything else..."
                        rows="1"
                        maxlength="1000"
                    ></textarea>
                    <button class="chat-send-btn" id="chat-send-btn" title="Send Message" aria-label="Send Message">
                        ➤
                    </button>
                </div>
            </div>
        `;

        // Add to document
        document.body.appendChild(this.toggleBtn);
        document.body.appendChild(this.chatPanel);

        // Get references to elements
        this.messagesContainer = document.getElementById('chat-messages');
        this.inputElement = document.getElementById('chat-input');
        this.sendBtn = document.getElementById('chat-send-btn');
        this.closeBtn = this.chatPanel.querySelector('.chat-close-btn');
    }

    /**
     * Bind event listeners
     */
    bindEvents() {
        // Toggle chat panel
        this.toggleBtn.addEventListener('click', () => this.toggleChat());
        this.closeBtn.addEventListener('click', () => this.closeChat());

        // Also bind to header chat toggle button if it exists
        const headerChatBtn = document.getElementById('chat-toggle-btn');
        if (headerChatBtn) {
            headerChatBtn.addEventListener('click', () => this.toggleChat());
        }

        // Send message
        this.sendBtn.addEventListener('click', () => this.sendMessage());
        
        // Enter key to send (Shift+Enter for new line)
        this.inputElement.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        // Auto-resize textarea
        this.inputElement.addEventListener('input', () => this.autoResizeInput());

        // Close on escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isOpen) {
                this.closeChat();
            }
        });

        // Close when clicking outside (mobile)
        this.chatPanel.addEventListener('click', (e) => {
            e.stopPropagation();
        });
        
        document.addEventListener('click', (e) => {
            const headerChatBtn = document.getElementById('chat-toggle-btn');
            const isToggleBtn = this.toggleBtn.contains(e.target) || (headerChatBtn && headerChatBtn.contains(e.target));
            
            if (this.isOpen && !this.chatPanel.contains(e.target) && !isToggleBtn) {
                // Only auto-close on mobile
                if (window.innerWidth <= 768) {
                    this.closeChat();
                }
            }
        });
    }

    /**
     * Toggle chat panel open/closed
     */
    toggleChat() {
        if (this.isOpen) {
            this.closeChat();
        } else {
            this.openChat();
        }
    }

    /**
     * Open chat panel
     */
    openChat() {
        this.isOpen = true;
        this.chatPanel.classList.add('open');
        this.toggleBtn.classList.add('active');
        this.toggleBtn.innerHTML = '×';
        this.toggleBtn.title = 'Close Chat Assistant';
        
        // Update header button if it exists
        const headerChatBtn = document.getElementById('chat-toggle-btn');
        if (headerChatBtn) {
            headerChatBtn.classList.add('chat-active');
        }
        
        // Focus input
        setTimeout(() => {
            this.inputElement.focus();
        }, 100);

        // Show welcome message if no conversation exists
        if (this.messages.length === 0) {
            this.showWelcomeMessage();
        }
    }

    /**
     * Close chat panel
     */
    closeChat() {
        this.isOpen = false;
        this.chatPanel.classList.remove('open');
        this.toggleBtn.classList.remove('active');
        this.toggleBtn.innerHTML = '💬';
        this.toggleBtn.title = 'Open Chat Assistant';
        
        // Update header button if it exists
        const headerChatBtn = document.getElementById('chat-toggle-btn');
        if (headerChatBtn) {
            headerChatBtn.classList.remove('chat-active');
        }
    }

    /**
     * Show welcome message
     */
    showWelcomeMessage() {
        const welcomeMessage = {
            id: 'welcome-' + Date.now(),
            role: 'assistant',
            content: `Hi! I'm your ProjectFlow assistant. I can help you manage tasks and projects using natural language. 

Try asking me things like:
• "Create a high priority task to fix the login bug"
• "List all tasks in the PF project"
• "Show me task PF-123"
• "Mark task PF-123 as done"

What would you like me to help you with?`,
            timestamp: new Date(),
            metadata: {
                intent: 'welcome',
                confidence: 1.0
            }
        };

        this.addMessageToUI(welcomeMessage);
        this.messages.push(welcomeMessage);
    }

    /**
     * Send a message
     */
    async sendMessage() {
        const message = this.inputElement.value.trim();
        if (!message || this.isLoading) {
            return;
        }

        // Clear input
        this.inputElement.value = '';
        this.autoResizeInput();

        // Add user message to UI
        const userMessage = {
            id: 'user-' + Date.now(),
            role: 'user',
            content: message,
            timestamp: new Date()
        };

        this.addMessageToUI(userMessage);
        this.messages.push(userMessage);
        this.saveConversationToStorage();

        // Show typing indicator
        this.showTypingIndicator();
        this.setLoading(true);

        try {
            // Send to backend
            const response = await this.apiClient.sendChatMessage(message, this.currentConversationId);
            
            // Update conversation ID
            this.currentConversationId = response.conversation_id;

            // Add assistant response to UI
            const assistantMessage = {
                id: 'assistant-' + Date.now(),
                role: 'assistant',
                content: response.response,
                timestamp: new Date(),
                metadata: {
                    actions_taken: response.actions_taken || [],
                    task_ids: response.task_ids || [],
                    project_ids: response.project_ids || [],
                    confidence: response.confidence || 0,
                    intent: response.intent || 'unknown',
                    requires_confirmation: response.requires_confirmation || false
                }
            };

            this.hideTypingIndicator();
            this.addMessageToUI(assistantMessage);
            this.messages.push(assistantMessage);
            this.saveConversationToStorage();

            // Show notification for successful actions
            if (response.actions_taken && response.actions_taken.length > 0) {
                const actionText = response.actions_taken.join(', ');
                this.notificationManager?.showSuccess(`Action completed: ${actionText}`);
                
                // Trigger refresh if tasks/projects were modified
                if (response.actions_taken.some(action => 
                    ['create_task', 'update_task', 'delete_task', 'create_project', 'update_project', 'delete_project'].includes(action)
                )) {
                    this.triggerUIRefresh();
                }
            }

        } catch (error) {
            console.error('Failed to send chat message:', error);
            this.hideTypingIndicator();
            
            const errorMessage = {
                id: 'error-' + Date.now(),
                role: 'assistant',
                content: `Sorry, I encountered an error: ${error.message || 'Please try again.'}\n\n💡 Need help? Check out the [Chat Interface Guide](chat-interface-guide) to learn about available commands and troubleshooting tips.`,
                timestamp: new Date(),
                metadata: {
                    intent: 'error',
                    confidence: 0,
                    hasHelpLink: true
                }
            };

            this.addMessageToUI(errorMessage);
            this.messages.push(errorMessage);
            this.notificationManager?.showError('Failed to send message. Please try again.');
        } finally {
            this.setLoading(false);
        }
    }

    /**
     * Add message to UI
     */
    addMessageToUI(message) {
        const messageElement = document.createElement('div');
        messageElement.className = `chat-message ${message.role}`;
        messageElement.innerHTML = `
            <div class="chat-message-content">
                ${this.formatMessageContent(message.content)}
                <div class="chat-message-meta">
                    ${this.formatTimestamp(message.timestamp)}
                    ${message.metadata?.confidence ? ` • ${Math.round(message.metadata.confidence * 100)}% confidence` : ''}
                </div>
            </div>
        `;

        this.messagesContainer.appendChild(messageElement);
        this.scrollToBottom();
    }

    /**
     * Format message content (basic markdown support)
     */
    formatMessageContent(content) {
        return content
            .replace(/\n/g, '<br>')
            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
            .replace(/\*(.*?)\*/g, '<em>$1</em>')
            .replace(/`(.*?)`/g, '<code>$1</code>')
            .replace(/•/g, '•'); // Preserve bullet points
    }

    /**
     * Format timestamp
     */
    formatTimestamp(timestamp) {
        const now = new Date();
        const messageTime = new Date(timestamp);
        const diffMs = now - messageTime;
        const diffMins = Math.floor(diffMs / 60000);
        
        if (diffMins < 1) {
            return 'Just now';
        } else if (diffMins < 60) {
            return `${diffMins}m ago`;
        } else if (diffMins < 1440) {
            return `${Math.floor(diffMins / 60)}h ago`;
        } else {
            return messageTime.toLocaleDateString();
        }
    }

    /**
     * Show typing indicator
     */
    showTypingIndicator() {
        if (this.typingIndicator) {
            return;
        }

        this.typingIndicator = document.createElement('div');
        this.typingIndicator.className = 'chat-typing';
        this.typingIndicator.innerHTML = `
            <div class="chat-typing-dots">
                <div class="chat-typing-dot"></div>
                <div class="chat-typing-dot"></div>
                <div class="chat-typing-dot"></div>
            </div>
            <span>Assistant is thinking...</span>
        `;

        this.messagesContainer.appendChild(this.typingIndicator);
        this.scrollToBottom();
    }

    /**
     * Hide typing indicator
     */
    hideTypingIndicator() {
        if (this.typingIndicator) {
            this.typingIndicator.remove();
            this.typingIndicator = null;
        }
    }

    /**
     * Set loading state
     */
    setLoading(loading) {
        this.isLoading = loading;
        this.sendBtn.disabled = loading;
        this.inputElement.disabled = loading;
        
        if (loading) {
            this.sendBtn.innerHTML = '⏳';
        } else {
            this.sendBtn.innerHTML = '➤';
        }
    }

    /**
     * Auto-resize input textarea
     */
    autoResizeInput() {
        this.inputElement.style.height = 'auto';
        const newHeight = Math.min(this.inputElement.scrollHeight, 120);
        this.inputElement.style.height = newHeight + 'px';
    }

    /**
     * Scroll to bottom of messages
     */
    scrollToBottom() {
        setTimeout(() => {
            this.messagesContainer.scrollTop = this.messagesContainer.scrollHeight;
        }, 100);
    }

    /**
     * Save conversation to localStorage
     */
    saveConversationToStorage() {
        try {
            const conversationData = {
                id: this.currentConversationId,
                messages: this.messages,
                lastUpdated: new Date().toISOString()
            };
            localStorage.setItem('projectflow_chat_conversation', JSON.stringify(conversationData));
        } catch (error) {
            console.warn('Failed to save conversation to localStorage:', error);
        }
    }

    /**
     * Load conversation from localStorage
     */
    loadConversationFromStorage() {
        try {
            const savedData = localStorage.getItem('projectflow_chat_conversation');
            if (savedData) {
                const conversationData = JSON.parse(savedData);
                
                // Only load if conversation is less than 24 hours old
                const lastUpdated = new Date(conversationData.lastUpdated);
                const now = new Date();
                const hoursDiff = (now - lastUpdated) / (1000 * 60 * 60);
                
                if (hoursDiff < 24) {
                    this.currentConversationId = conversationData.id;
                    this.messages = conversationData.messages || [];
                    
                    // Restore messages to UI
                    this.messages.forEach(message => {
                        this.addMessageToUI(message);
                    });
                } else {
                    // Clear old conversation
                    localStorage.removeItem('projectflow_chat_conversation');
                }
            }
        } catch (error) {
            console.warn('Failed to load conversation from localStorage:', error);
            localStorage.removeItem('projectflow_chat_conversation');
        }
    }

    /**
     * Clear conversation
     */
    clearConversation() {
        this.messages = [];
        this.currentConversationId = null;
        this.messagesContainer.innerHTML = '';
        localStorage.removeItem('projectflow_chat_conversation');
        this.showWelcomeMessage();
    }

    /**
     * Trigger UI refresh for other components
     */
    triggerUIRefresh() {
        // Dispatch custom event for other managers to listen to
        window.dispatchEvent(new CustomEvent('chat:dataChanged', {
            detail: {
                timestamp: new Date(),
                source: 'chat'
            }
        }));
    }

    /**
     * Get conversation history
     */
    async getConversationHistory() {
        if (!this.currentConversationId) {
            return [];
        }

        try {
            return await this.apiClient.getChatHistory(this.currentConversationId);
        } catch (error) {
            console.error('Failed to get conversation history:', error);
            return [];
        }
    }

    /**
     * Destroy the chat manager
     */
    destroy() {
        if (this.toggleBtn) {
            this.toggleBtn.remove();
        }
        if (this.chatPanel) {
            this.chatPanel.remove();
        }
        this.saveConversationToStorage();
    }
}
