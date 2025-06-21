/**
 * Header Menu Manager
 * Handles the popup menu functionality in the header
 */
class HeaderMenuManager {
    constructor() {
        this.menuBtn = null;
        this.menu = null;
        this.isMenuOpen = false;
        this.isInitialized = false;
        this.retryCount = 0;
        this.maxRetries = 5;
        
        // Defer initialization to ensure DOM is ready
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => this.init());
        } else {
            this.init();
        }
    }

    init() {
        console.log('HeaderMenuManager: init called, isInitialized:', this.isInitialized);
        if (this.isInitialized) {
            return;
        }
        
        // Try to find elements
        this.menuBtn = document.getElementById('header-menu-btn');
        this.menu = document.getElementById('header-menu');
        
        console.log('HeaderMenuManager: menuBtn element:', this.menuBtn);
        console.log('HeaderMenuManager: menu element:', this.menu);

        if (this.menuBtn && this.menu) {
            this.bindEvents();
            this.isInitialized = true;
            console.log('HeaderMenuManager: Successfully initialized');
        } else {
            // If elements not found and haven't exceeded max retries, try again
            if (this.retryCount < this.maxRetries) {
                this.retryCount++;
                console.log(`HeaderMenuManager: Elements not found, retrying... (${this.retryCount}/${this.maxRetries})`);
                setTimeout(() => this.init(), 100);
            } else {
                console.error('HeaderMenuManager: Failed to initialize after maximum retries');
            }
        }
    }

    bindEvents() {
        console.log('HeaderMenuManager: Binding events');
        // Toggle menu on button click (supports both mouse and touch)
        this.menuBtn.addEventListener('click', (e) => {
            console.log('Header menu button clicked');
            e.stopPropagation();
            this.toggleMenu();
        });

        // Add touch support for mobile devices
        this.menuBtn.addEventListener('touchstart', (e) => {
            e.preventDefault(); // Prevent double tap zoom
        });

        // Close menu when clicking outside
        document.addEventListener('click', (e) => {
            if (this.isMenuOpen && !this.menu.contains(e.target) && !this.menuBtn.contains(e.target)) {
                // Don't close if clicking on project dropdown or project selector
                const projectDropdown = document.getElementById('project-dropdown');
                const projectSelector = document.getElementById('project-selector-btn');
                const isProjectDropdownClick = projectDropdown && projectDropdown.contains(e.target);
                const isProjectSelectorClick = projectSelector && projectSelector.contains(e.target);
                
                if (!isProjectDropdownClick && !isProjectSelectorClick) {
                    console.log('HeaderMenuManager: Closing menu due to outside click');
                    this.closeMenu();
                } else {
                    console.log('HeaderMenuManager: Ignoring click on project selector/dropdown');
                }
            }
        });

        // Touch events for mobile
        document.addEventListener('touchstart', (e) => {
            if (this.isMenuOpen && !this.menu.contains(e.target) && !this.menuBtn.contains(e.target)) {
                const projectDropdown = document.getElementById('project-dropdown');
                if (!projectDropdown || !projectDropdown.contains(e.target)) {
                    this.closeMenu();
                }
            }
        });

        // Close menu on escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isMenuOpen) {
                this.closeMenu();
                this.menuBtn.focus(); // Return focus to menu button
            }
            
            // Handle arrow key navigation within menu
            if (this.isMenuOpen && (e.key === 'ArrowDown' || e.key === 'ArrowUp')) {
                this.handleKeyboardNavigation(e);
            }
        });

        // Prevent menu from closing when clicking inside it
        this.menu.addEventListener('click', (e) => {
            e.stopPropagation();
        });
    }

    toggleMenu() {
        console.log('HeaderMenuManager: toggleMenu called, isMenuOpen:', this.isMenuOpen);
        if (this.isMenuOpen) {
            this.closeMenu();
        } else {
            this.openMenu();
        }
    }

    openMenu() {
        console.log('HeaderMenuManager: Opening menu');
        this.menu.style.display = 'block';
        this.menu.style.visibility = 'visible';
        this.isMenuOpen = true;
        this.menuBtn.setAttribute('aria-expanded', 'true');
        
        // Add animation class
        setTimeout(() => {
            this.menu.classList.add('menu-open');
        }, 10);
        console.log('HeaderMenuManager: Menu opened');
    }

    closeMenu() {
        this.menu.classList.remove('menu-open');
        this.isMenuOpen = false;
        this.menuBtn.setAttribute('aria-expanded', 'false');
        
        // Close any open submenus like project dropdown
        const projectDropdown = document.getElementById('project-dropdown');
        if (projectDropdown) {
            projectDropdown.style.display = 'none';
        }
        
        // Hide after animation
        setTimeout(() => {
            if (!this.isMenuOpen) {
                this.menu.style.display = 'none';
                this.menu.style.visibility = 'hidden';
            }
        }, 200);
    }

    /**
     * Handle keyboard navigation within the menu
     */
    handleKeyboardNavigation(e) {
        e.preventDefault();
        const focusableElements = this.menu.querySelectorAll('button, [tabindex="0"]');
        const currentIndex = Array.from(focusableElements).indexOf(document.activeElement);
        
        let nextIndex;
        if (e.key === 'ArrowDown') {
            nextIndex = currentIndex < focusableElements.length - 1 ? currentIndex + 1 : 0;
        } else if (e.key === 'ArrowUp') {
            nextIndex = currentIndex > 0 ? currentIndex - 1 : focusableElements.length - 1;
        }
        
        if (nextIndex !== undefined && focusableElements[nextIndex]) {
            focusableElements[nextIndex].focus();
        }
    }

    /**
     * Get the current menu state
     */
    isOpen() {
        return this.isMenuOpen;
    }

    /**
     * Force close the menu (useful for other components)
     */
    forceClose() {
        if (this.isMenuOpen) {
            this.closeMenu();
        }
    }

    /**
     * Get reference to menu element
     */
    getMenuElement() {
        return this.menu;
    }
}

export default HeaderMenuManager;
