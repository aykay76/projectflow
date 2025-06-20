/**
 * UI Manager - Handles UI interactions, themes, keyboard shortcuts, and mobile enhancements
 */
class UIManager {
    constructor() {
        this.currentTheme = localStorage.getItem('theme') || 'light';
        this.keyboardShortcuts = new Map();
        this.isMobile = this.detectMobile();
        
        // Apply theme immediately to avoid flash
        document.documentElement.setAttribute('data-theme', this.currentTheme);
        
        // Don't initialize theme here - wait for DOM to be ready
        this.initializeKeyboardShortcuts();
        this.initializeMobileEnhancements();
    }

    /**
     * Initialize the UI Manager - called after DOM is ready
     */
    init() {
        this.initializeTheme();
    }

    /**
     * Detect if running on mobile device
     */
    detectMobile() {
        return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) ||
               window.innerWidth <= 768;
    }

    /**
     * Initialize theme system
     */
    initializeTheme() {
        // Apply theme and update UI elements
        this.applyTheme(this.currentTheme);
        
        // Theme toggle button
        const themeToggle = document.getElementById('theme-toggle');
        if (themeToggle) {
            themeToggle.addEventListener('click', () => this.toggleTheme());
            console.log('Theme toggle event listener attached');
        } else {
            console.warn('Theme toggle button not found');
        }

        // Listen for system theme changes
        if (window.matchMedia) {
            const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
            mediaQuery.addEventListener('change', (e) => {
                if (!localStorage.getItem('theme')) {
                    this.applyTheme(e.matches ? 'dark' : 'light');
                }
            });
        }
    }

    /**
     * Apply theme to document
     */
    applyTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        this.currentTheme = theme;
        
        // Update theme toggle button icon
        const themeIcon = document.getElementById('theme-icon');
        const themeToggle = document.getElementById('theme-toggle');
        
        if (themeIcon) {
            themeIcon.textContent = theme === 'dark' ? '☀️' : '🌙';
        }
        
        if (themeToggle) {
            themeToggle.title = `Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`;
        }
    }

    /**
     * Toggle between light and dark themes
     */
    toggleTheme() {
        const newTheme = this.currentTheme === 'dark' ? 'light' : 'dark';
        this.applyTheme(newTheme);
        localStorage.setItem('theme', newTheme);
        
        // Show feedback - try multiple notification methods
        if (window.projectFlowApp?.notificationManager?.showMessage) {
            window.projectFlowApp.notificationManager.showMessage(
                `Switched to ${newTheme} theme! ${newTheme === 'dark' ? '🌙' : '☀️'}`, 
                'info', 
                2000
            );
        } else if (window.messageManager?.showMessage) {
            window.messageManager.showMessage(
                `Switched to ${newTheme} theme! ${newTheme === 'dark' ? '🌙' : '☀️'}`, 
                'info', 
                2000
            );
        } else if (window.showMessage) {
            window.showMessage(
                `Switched to ${newTheme} theme! ${newTheme === 'dark' ? '🌙' : '☀️'}`, 
                'info'
            );
        }
    }

    /**
     * Initialize keyboard shortcuts
     */
    initializeKeyboardShortcuts() {
        // Register default shortcuts
        this.registerShortcut('KeyN', () => this.handleNewTask(), { ctrlKey: true });
        this.registerShortcut('KeyS', (e) => this.handleSave(e), { ctrlKey: true });
        this.registerShortcut('KeyF', (e) => this.handleSearch(e), { ctrlKey: true });
        this.registerShortcut('Escape', () => this.handleEscape());
        this.registerShortcut('Digit1', () => this.switchView('kanban'), { ctrlKey: true });
        this.registerShortcut('Digit2', () => this.switchView('hierarchy'), { ctrlKey: true });
        this.registerShortcut('Digit3', () => this.switchView('timeline'), { ctrlKey: true });

        // Global keyboard event listener
        document.addEventListener('keydown', (e) => this.handleKeyDown(e));
    }

    /**
     * Register a keyboard shortcut
     */
    registerShortcut(key, handler, modifiers = {}) {
        const shortcutKey = this.createShortcutKey(key, modifiers);
        this.keyboardShortcuts.set(shortcutKey, handler);
    }

    /**
     * Create a unique key for the shortcut
     */
    createShortcutKey(key, modifiers) {
        const parts = [];
        if (modifiers.ctrlKey) parts.push('ctrl');
        if (modifiers.altKey) parts.push('alt');
        if (modifiers.shiftKey) parts.push('shift');
        if (modifiers.metaKey) parts.push('meta');
        parts.push(key.toLowerCase());
        return parts.join('+');
    }

    /**
     * Handle keyboard events
     */
    handleKeyDown(e) {
        // Don't trigger shortcuts when typing in inputs
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.isContentEditable) {
            return;
        }

        const shortcutKey = this.createShortcutKey(e.code, {
            ctrlKey: e.ctrlKey,
            altKey: e.altKey,
            shiftKey: e.shiftKey,
            metaKey: e.metaKey
        });

        const handler = this.keyboardShortcuts.get(shortcutKey);
        if (handler) {
            e.preventDefault();
            handler(e);
        }
    }

    /**
     * Keyboard shortcut handlers
     */
    handleNewTask() {
        const newTaskBtn = document.getElementById('new-task-btn');
        if (newTaskBtn) {
            newTaskBtn.click();
        }
    }

    handleSave(e) {
        e.preventDefault();
        const modal = document.getElementById('task-modal');
        if (modal && modal.style.display !== 'none') {
            const saveBtn = document.querySelector('#task-form button[type="submit"]');
            if (saveBtn) {
                saveBtn.click();
            }
        }
    }

    handleSearch(e) {
        e.preventDefault();
        const searchInput = document.getElementById('search-input');
        if (searchInput) {
            searchInput.focus();
            searchInput.select();
        }
    }

    handleEscape() {
        // Close any open modals
        const modals = document.querySelectorAll('.modal');
        modals.forEach(modal => {
            if (modal.style.display !== 'none') {
                const closeBtn = modal.querySelector('.close');
                if (closeBtn) {
                    closeBtn.click();
                }
            }
        });

        // Clear search
        const searchInput = document.getElementById('search-input');
        if (searchInput && searchInput.value) {
            searchInput.value = '';
            searchInput.dispatchEvent(new Event('input'));
        }
    }

    switchView(viewName) {
        if (window.viewManager) {
            window.viewManager.switchToView(viewName);
        }
    }

    /**
     * Initialize mobile enhancements
     */
    initializeMobileEnhancements() {
        if (!this.isMobile) return;

        // Add mobile class to body
        document.body.classList.add('mobile');

        // Touch gestures for mobile
        this.initializeTouchGestures();
        
        // Mobile-specific UI adjustments
        this.adjustMobileUI();
        
        // Handle orientation changes
        window.addEventListener('orientationchange', () => {
            setTimeout(() => this.adjustMobileUI(), 100);
        });
    }

    /**
     * Initialize touch gestures
     */
    initializeTouchGestures() {
        let touchStartX = 0;
        let touchStartY = 0;
        let touchEndX = 0;
        let touchEndY = 0;

        document.addEventListener('touchstart', (e) => {
            touchStartX = e.changedTouches[0].screenX;
            touchStartY = e.changedTouches[0].screenY;
        });

        document.addEventListener('touchend', (e) => {
            touchEndX = e.changedTouches[0].screenX;
            touchEndY = e.changedTouches[0].screenY;
            this.handleSwipeGesture();
        });

        const handleSwipeGesture = () => {
            const deltaX = touchEndX - touchStartX;
            const deltaY = touchEndY - touchStartY;
            const minSwipeDistance = 50;

            // Horizontal swipes for view switching
            if (Math.abs(deltaX) > Math.abs(deltaY) && Math.abs(deltaX) > minSwipeDistance) {
                if (deltaX > 0) {
                    // Swipe right - previous view
                    this.switchToPreviousView();
                } else {
                    // Swipe left - next view
                    this.switchToNextView();
                }
            }
        };
    }

    /**
     * Switch to previous view (mobile gesture)
     */
    switchToPreviousView() {
        const views = ['kanban', 'hierarchy', 'timeline'];
        const currentIndex = views.indexOf(window.viewManager?.currentView || 'kanban');
        const previousIndex = currentIndex > 0 ? currentIndex - 1 : views.length - 1;
        this.switchView(views[previousIndex]);
    }

    /**
     * Switch to next view (mobile gesture)
     */
    switchToNextView() {
        const views = ['kanban', 'hierarchy', 'timeline'];
        const currentIndex = views.indexOf(window.viewManager?.currentView || 'kanban');
        const nextIndex = currentIndex < views.length - 1 ? currentIndex + 1 : 0;
        this.switchView(views[nextIndex]);
    }

    /**
     * Adjust UI for mobile devices
     */
    adjustMobileUI() {
        // Adjust modal sizes for mobile
        const modals = document.querySelectorAll('.modal');
        modals.forEach(modal => {
            if (this.isMobile) {
                modal.style.width = '95%';
                modal.style.height = '90%';
                modal.style.top = '5%';
                modal.style.left = '2.5%';
            }
        });

        // Adjust dropdown positions
        const dropdowns = document.querySelectorAll('.dropdown-content');
        dropdowns.forEach(dropdown => {
            if (this.isMobile) {
                dropdown.style.position = 'fixed';
                dropdown.style.width = '90%';
                dropdown.style.left = '5%';
                dropdown.style.maxHeight = '60vh';
                dropdown.style.overflow = 'auto';
            }
        });
    }

    /**
     * Show loading spinner
     */
    showLoading(element, message = 'Loading...') {
        if (!element) return;
        
        element.classList.add('loading');
        const originalContent = element.innerHTML;
        element.innerHTML = `
            <div class="loading-spinner">
                <div class="spinner"></div>
                <span>${message}</span>
            </div>
        `;
        
        return () => {
            element.classList.remove('loading');
            element.innerHTML = originalContent;
        };
    }

    /**
     * Animate element
     */
    animateElement(element, animation, duration = 300) {
        if (!element) return Promise.resolve();
        
        return new Promise((resolve) => {
            element.style.animation = `${animation} ${duration}ms ease-in-out`;
            setTimeout(() => {
                element.style.animation = '';
                resolve();
            }, duration);
        });
    }

    /**
     * Smooth scroll to element
     */
    scrollToElement(element, offset = 0) {
        if (!element) return;
        
        const elementPosition = element.getBoundingClientRect().top + window.pageYOffset;
        const offsetPosition = elementPosition - offset;
        
        window.scrollTo({
            top: offsetPosition,
            behavior: 'smooth'
        });
    }

    /**
     * Get current theme
     */
    getCurrentTheme() {
        return this.currentTheme;
    }

    /**
     * Check if mobile
     */
    getIsMobile() {
        return this.isMobile;
    }
}

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = UIManager;
}
