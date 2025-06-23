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
        this.llmInfo = null; // Store LLM provider information
        this.llmHealthy = false; // Track LLM health status
        this.chatMode = 'translated'; // 'translated' (default) or 'direct'
        
        // Initialize chat interface
        this.init();
        
        // Load conversation from localStorage
        this.loadConversationFromStorage();
        
        // Load chat mode preference
        this.loadChatModePreference();
        
        // Check LLM status
        this.checkLLMStatus();
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
        this.chatPanel.innerHTML = 
            '<div class="chat-header">' +
                '<h3>💬 Chat Assistant</h3>' +
                '<div class="chat-header-actions">' +
                    '<button class="chat-mode-btn" id="chat-mode-btn" title="Chat Mode" aria-label="Chat Mode">' +
                        '<span id="chat-mode-indicator">🔄</span>' +
                    '</button>' +
                    '<button class="chat-status-btn" id="chat-status-btn" title="LLM Status" aria-label="LLM Status">' +
                        '<span class="status-indicator" id="llm-status-indicator">⚪</span>' +
                    '</button>' +
                    '<button class="chat-close-btn" title="Close Chat" aria-label="Close Chat">×</button>' +
                '</div>' +
            '</div>' +
            '<div class="chat-mode-panel" id="chat-mode-panel" style="display: none;">' +
                '<div class="mode-content">' +
                    '<h4>💬 Chat Mode</h4>' +
                    '<div class="mode-options">' +
                        '<label class="mode-option">' +
                            '<input type="radio" name="chat-mode" value="translated" id="mode-translated" checked>' +
                            '<div class="mode-info">' +
                                '<strong>🔄 Smart Assistant</strong>' +
                                '<p>Translates your requests into ProjectFlow actions (recommended)</p>' +
                            '</div>' +
                        '</label>' +
                        '<label class="mode-option">' +
                            '<input type="radio" name="chat-mode" value="direct" id="mode-direct">' +
                            '<div class="mode-info">' +
                                '<strong>🤖 Direct LLM</strong>' +
                                '<p>Chat directly with the language model</p>' +
                            '</div>' +
                        '</label>' +
                    '</div>' +
                    '<div class="mode-actions">' +
                        '<button class="btn-apply-mode" id="apply-mode-btn">Apply</button>' +
                    '</div>' +
                '</div>' +
            '</div>' +
            '<div class="chat-status-panel" id="chat-status-panel" style="display: none;">' +
                '<div class="status-content">' +
                    '<h4>🤖 LLM Status</h4>' +
                    '<div id="llm-status-details">Checking...</div>' +
                    '<div class="status-actions">' +
                        '<button class="btn-refresh-status" id="refresh-status-btn">🔄 Refresh</button>' +
                    '</div>' +
                '</div>' +
            '</div>' +
            '<div class="chat-messages" id="chat-messages"></div>' +
            '<div class="chat-input-container">' +
                '<div class="chat-input-wrapper">' +
                    '<textarea ' +
                        'class="chat-input" ' +
                        'id="chat-input" ' +
                        'placeholder="Ask me to create tasks, list projects, or anything else..."' +
                        'rows="1"' +
                        'maxlength="1000"' +
                    '></textarea>' +
                    '<button class="chat-send-btn" id="chat-send-btn" title="Send Message" aria-label="Send Message">' +
                        '➤' +
                    '</button>' +
                '</div>' +
            '</div>';

        // Add to document
        document.body.appendChild(this.toggleBtn);
        document.body.appendChild(this.chatPanel);

        // Get references to elements
        this.messagesContainer = document.getElementById('chat-messages');
        this.inputElement = document.getElementById('chat-input');
        this.sendBtn = document.getElementById('chat-send-btn');
        this.closeBtn = this.chatPanel.querySelector('.chat-close-btn');
        this.statusBtn = document.getElementById('chat-status-btn');
        this.statusPanel = document.getElementById('chat-status-panel');
        this.statusIndicator = document.getElementById('llm-status-indicator');
        this.statusDetails = document.getElementById('llm-status-details');
        this.refreshStatusBtn = document.getElementById('refresh-status-btn');
        this.modeBtn = document.getElementById('chat-mode-btn');
        this.modePanel = document.getElementById('chat-mode-panel');
        this.modeIndicator = document.getElementById('chat-mode-indicator');
        this.modeTranslated = document.getElementById('mode-translated');
        this.modeDirect = document.getElementById('mode-direct');
        this.applyModeBtn = document.getElementById('apply-mode-btn');
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
        
        // Mode panel toggle and actions
        this.modeBtn.addEventListener('click', () => this.toggleModePanel());
        this.applyModeBtn.addEventListener('click', () => this.applyChatMode());
        this.modeTranslated.addEventListener('change', () => this.updateModePreview());
        this.modeDirect.addEventListener('change', () => this.updateModePreview());
        
        // Status panel toggle
        this.statusBtn.addEventListener('click', () => this.toggleStatusPanel());
        this.refreshStatusBtn.addEventListener('click', () => this.checkLLMStatus());
        
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
            content: 'Hi! I\'m your ProjectFlow assistant. I can help you manage tasks and projects using natural language.\n\n' +
                'Try asking me things like:\n' +
                '• "Create a high priority task to fix the login bug"\n' +
                '• "List all tasks in the PF project"\n' +
                '• "Show me task PF-123"\n' +
                '• "Mark task PF-123 as done"\n\n' +
                'What would you like me to help you with?',
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
            let response;
            
            if (this.chatMode === 'direct') {
                // Direct LLM mode
                if (!this.llmHealthy) {
                    throw new Error('Direct LLM mode requires a healthy LLM connection. Please check the LLM status or switch to Smart Assistant mode.');
                }
                
                const llmResponse = await this.sendDirectLLMMessage(message);
                
                // Create response object similar to translated mode
                response = {
                    response: llmResponse,
                    conversation_id: this.currentConversationId || 'direct-' + Date.now(),
                    actions_taken: [],
                    task_ids: [],
                    project_ids: [],
                    confidence: 0.9,
                    intent: 'direct_llm',
                    requires_confirmation: false
                };
            } else {
                // Translated mode (default)
                response = await this.apiClient.sendChatMessage(message, this.currentConversationId);
            }
            
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
                    requires_confirmation: response.requires_confirmation || false,
                    chat_mode: this.chatMode
                }
            };

            this.hideTypingIndicator();
            this.addMessageToUI(assistantMessage);
            this.messages.push(assistantMessage);
            this.saveConversationToStorage();

            // Show notification for successful actions (only in translated mode)
            if (this.chatMode === 'translated' && response.actions_taken && response.actions_taken.length > 0) {
                const actionText = response.actions_taken.join(', ');
                if (this.notificationManager) {
                    this.notificationManager.showSuccess('Action completed: ' + actionText);
                }
                
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
            
            // Check if this is an LLM-related error and provide specific guidance
            let errorContent = 'Sorry, I encountered an error: ' + (error.message || 'Please try again.');
            let helpContent = '';
            
            // Add mode-specific and LLM-specific error guidance
            if (this.chatMode === 'direct') {
                helpContent += '\n\n🤖 **Direct LLM Mode:**\n';
                helpContent += '• This mode requires a healthy LLM connection\n';
                helpContent += '• Try switching to Smart Assistant mode for basic functionality\n';
                
                if (this.llmInfo) {
                    if (this.llmInfo.provider === 'ollama') {
                        helpContent += '• Make sure Ollama is running with `ollama serve`\n';
                        helpContent += '• Verify the model "' + (this.llmInfo.model || 'llama3.2') + '" is installed\n';
                    } else if (this.llmInfo.provider === 'groq') {
                        helpContent += '• Check your Groq API key configuration\n';
                    } else if (this.llmInfo.provider === 'openai') {
                        helpContent += '• Check your OpenAI API key configuration\n';
                    }
                }
            } else {
                // Translated mode error guidance
                if (error.message && error.message.includes('LLM')) {
                    helpContent += '\n\n🔧 **LLM Troubleshooting:**\n';
                    
                    if (this.llmInfo) {
                        if (this.llmInfo.provider === 'ollama') {
                            helpContent += '• Make sure Ollama is running with `ollama serve`\n';
                            helpContent += '• Verify the model "' + (this.llmInfo.model || 'llama3.2') + '" is installed with `ollama pull ' + (this.llmInfo.model || 'llama3.2') + '`\n';
                            helpContent += '• Check Ollama logs for detailed error information\n';
                        } else if (this.llmInfo.provider === 'groq') {
                            helpContent += '• Check your Groq API key configuration\n';
                            helpContent += '• Verify internet connectivity\n';
                        } else if (this.llmInfo.provider === 'openai') {
                            helpContent += '• Check your OpenAI API key configuration\n';
                            helpContent += '• Verify internet connectivity\n';
                        }
                    } else {
                        helpContent += '• LLM service may be disabled or misconfigured\n';
                        helpContent += '• Check server logs for more details\n';
                    }
                    
                    helpContent += '\n💬 **Meanwhile:** You can still use basic commands like "create task: fix bug"';
                }
            }
            
            helpContent += '\n\n💡 Need help? Check out the [Chat Interface Guide](chat-interface-guide) to learn about available commands and troubleshooting tips.';
            
            const errorMessage = {
                id: 'error-' + Date.now(),
                role: 'assistant',
                content: errorContent + helpContent,
                timestamp: new Date(),
                metadata: {
                    intent: 'error',
                    confidence: 0,
                    hasHelpLink: true,
                    llmProvider: (this.llmInfo && this.llmInfo.provider) || 'unknown',
                    chat_mode: this.chatMode
                }
            };

            this.addMessageToUI(errorMessage);
            this.messages.push(errorMessage);
            if (this.notificationManager) {
                this.notificationManager.showError('Failed to send message. Please try again.');
            }
        } finally {
            this.setLoading(false);
        }
    }

    /**
     * Check LLM status and capabilities
     */
    async checkLLMStatus() {
        try {
            const response = await fetch('/api/llm/info');
            if (response.ok) {
                this.llmInfo = await response.json();
                this.llmHealthy = this.llmInfo.enabled && this.llmInfo.status === 'healthy';
                console.log('LLM Status:', this.llmInfo);
                
                // Update UI based on LLM status
                this.updateLLMStatusInUI();
            } else {
                console.warn('Failed to get LLM status:', response.status);
                this.llmHealthy = false;
            }
        } catch (error) {
            console.error('Error checking LLM status:', error);
            this.llmHealthy = false;
        }
    }

    /**
     * Toggle status panel visibility
     */
    toggleStatusPanel() {
        const isVisible = this.statusPanel.style.display !== 'none';
        this.statusPanel.style.display = isVisible ? 'none' : 'block';
        
        if (!isVisible) {
            // Refresh status when opening panel
            this.checkLLMStatus();
        }
    }

    /**
     * Update UI elements based on LLM status
     */
    updateLLMStatusInUI() {
        if (!this.llmInfo) return;
        
        // Update status indicator
        if (this.statusIndicator) {
            if (this.llmInfo.enabled && this.llmHealthy) {
                this.statusIndicator.textContent = '🟢';
                this.statusIndicator.title = this.llmInfo.provider + ' - Healthy';
            } else if (this.llmInfo.enabled && !this.llmHealthy) {
                this.statusIndicator.textContent = '🟡';
                this.statusIndicator.title = this.llmInfo.provider + ' - Unhealthy';
            } else {
                this.statusIndicator.textContent = '⚪';
                this.statusIndicator.title = 'LLM Disabled';
            }
        }
        
        // Update status details
        if (this.statusDetails) {
            let statusHTML = '';
            
            if (this.llmInfo.enabled) {
                statusHTML += '<div class="status-item"><strong>Provider:</strong> ' + (this.llmInfo.provider || 'Unknown') + '</div>';
                
                if (this.llmInfo.model) {
                    statusHTML += '<div class="status-item"><strong>Model:</strong> ' + this.llmInfo.model + '</div>';
                }
                
                statusHTML += '<div class="status-item"><strong>Status:</strong> <span class="status-' + this.llmInfo.status + '">' + this.llmInfo.status + '</span></div>';
                
                if (this.llmInfo.metadata && this.llmInfo.metadata.health_error) {
                    statusHTML += '<div class="status-item error"><strong>Error:</strong> ' + this.llmInfo.metadata.health_error + '</div>';
                }
                
                statusHTML += '<div class="status-item"><strong>Last Check:</strong> ' + new Date(this.llmInfo.timestamp).toLocaleTimeString() + '</div>';
            } else {
                statusHTML += '<div class="status-item"><strong>Status:</strong> <span class="status-disabled">Disabled</span></div>';
                statusHTML += '<div class="status-item">LLM functionality is currently disabled in the configuration.</div>';
            }
            
            this.statusDetails.innerHTML = statusHTML;
        }
        
        // Update chat input placeholder based on LLM status
        if (this.inputElement) {
            this.updateInputPlaceholder();
        }
    }

    /**
     * Send direct message to LLM (bypassing translation)
     */
    async sendDirectLLMMessage(message) {
        try {
            const llmRequest = {
                messages: [
                    { role: 'user', content: message }
                ],
                max_tokens: 1000,
                temperature: 0.7
            };

            const response = await fetch('/api/llm/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(llmRequest)
            });

            if (!response.ok) {
                throw new Error('LLM request failed: ' + response.status + ' ' + response.statusText);
            }

            const data = await response.json();
            return data.response.choices[0] && data.response.choices[0].message && data.response.choices[0].message.content || 'No response from LLM';
        } catch (error) {
            console.error('Direct LLM message failed:', error);
            throw error;
        }
    }

    /**
     * Toggle mode panel visibility
     */
    toggleModePanel() {
        const isVisible = this.modePanel.style.display !== 'none';
        
        // Close other panels first
        this.statusPanel.style.display = 'none';
        
        this.modePanel.style.display = isVisible ? 'none' : 'block';
        
        if (!isVisible) {
            // Set current mode in the UI
            if (this.chatMode === 'translated') {
                this.modeTranslated.checked = true;
            } else {
                this.modeDirect.checked = true;
            }
            this.updateModePreview();
        }
    }

    /**
     * Update mode preview text
     */
    updateModePreview() {
        // Update mode indicator
        const selectedMode = this.modeTranslated.checked ? 'translated' : 'direct';
        
        if (selectedMode === 'translated') {
            this.modeIndicator.textContent = '🔄';
            this.modeBtn.title = 'Smart Assistant Mode';
        } else {
            this.modeIndicator.textContent = '🤖';
            this.modeBtn.title = 'Direct LLM Mode';
        }
    }

    /**
     * Apply selected chat mode
     */
    applyChatMode() {
        const selectedMode = this.modeTranslated.checked ? 'translated' : 'direct';
        
        if (selectedMode !== this.chatMode) {
            this.chatMode = selectedMode;
            
            // Update UI
            this.updateModePreview();
            this.updateChatModeInUI();
            
            // Save preference
            localStorage.setItem('projectflow_chat_mode', this.chatMode);
            
            // Notify user
            const modeNames = {
                'translated': 'Smart Assistant',
                'direct': 'Direct LLM'
            };
            
            this.addSystemMessage('Switched to ' + modeNames[this.chatMode] + ' mode.');
            
            console.log('Chat mode changed to:', this.chatMode);
        }
        
        // Close panel
        this.modePanel.style.display = 'none';
    }

    /**
     * Update UI based on current chat mode
     */
    updateChatModeInUI() {
        // Update header indicator
        if (this.chatMode === 'translated') {
            this.modeIndicator.textContent = '🔄';
            this.modeBtn.title = 'Smart Assistant Mode - Translates requests to ProjectFlow actions';
        } else {
            this.modeIndicator.textContent = '🤖';
            this.modeBtn.title = 'Direct LLM Mode - Chat directly with language model';
        }
        
        // Update input placeholder
        this.updateInputPlaceholder();
    }

    /**
     * Update input placeholder based on mode and LLM status
     */
    updateInputPlaceholder() {
        if (!this.inputElement) return;
        
        if (this.chatMode === 'direct') {
            if (this.llmHealthy) {
                this.inputElement.placeholder = 'Chat directly with ' + (this.llmInfo && this.llmInfo.provider || 'LLM') + '...';
            } else {
                this.inputElement.placeholder = 'Direct LLM mode requires a healthy LLM connection...';
            }
        } else {
            // Translated mode
            if (this.llmInfo && this.llmInfo.enabled && this.llmHealthy) {
                this.inputElement.placeholder = 'Ask me to create tasks, list projects, or anything else... (' + this.llmInfo.provider + ')';
            } else if (this.llmInfo && this.llmInfo.enabled && !this.llmHealthy) {
                this.inputElement.placeholder = 'LLM (' + (this.llmInfo.provider || 'unknown') + ') is currently unavailable. Basic commands still work...';
            } else {
                this.inputElement.placeholder = 'Ask me to create tasks, list projects, etc...';
            }
        }
    }

    /**
     * Add system message to conversation
     */
    addSystemMessage(content) {
        const systemMessage = {
            id: 'system-' + Date.now(),
            role: 'system',
            content: content,
            timestamp: new Date(),
            metadata: {
                intent: 'system',
                confidence: 1.0
            }
        };

        this.addMessageToUI(systemMessage);
        this.messages.push(systemMessage);
        this.saveConversationToStorage();
    }

    /**
     * Save conversation to localStorage
     */
    saveConversationToStorage() {
        try {
            const conversationData = {
                messages: this.messages,
                currentConversationId: this.currentConversationId,
                timestamp: Date.now()
            };
            localStorage.setItem('projectflow_chat_conversation', JSON.stringify(conversationData));
        } catch (error) {
            console.warn('Failed to save conversation to storage:', error);
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
                
                // Check if data is not too old (24 hours)
                const maxAge = 24 * 60 * 60 * 1000; // 24 hours in milliseconds
                if (Date.now() - conversationData.timestamp < maxAge) {
                    this.messages = conversationData.messages || [];
                    this.currentConversationId = conversationData.currentConversationId;
                    
                    // Restore messages to UI if chat is open
                    if (this.isOpen && this.chatMessages) {
                        this.messages.forEach(message => {
                            this.addMessageToUI(message);
                        });
                    }
                } else {
                    // Clear old conversation
                    localStorage.removeItem('projectflow_chat_conversation');
                }
            }
        } catch (error) {
            console.warn('Failed to load conversation from storage:', error);
            // Clear corrupted data
            localStorage.removeItem('projectflow_chat_conversation');
        }
    }

    /**
     * Load chat mode preference from localStorage
     */
    loadChatModePreference() {
        try {
            const savedMode = localStorage.getItem('projectflow_chat_mode');
            if (savedMode && ['translated', 'direct'].includes(savedMode)) {
                this.chatMode = savedMode;
            }
        } catch (error) {
            console.warn('Failed to load chat mode preference:', error);
        }
        
        // Update UI to reflect loaded mode
        this.updateChatModeInUI();
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
