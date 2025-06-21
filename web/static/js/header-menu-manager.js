/**
 * Header Menu Manager
 * Handles the popup menu functionality in the header
 */
class HeaderMenuManager {
    constructor() {
        this.menuBtn = null;
        this.menu = null;
        this.isMenuOpen = false;
        this.init();
    }

    init() {
        this.menuBtn = document.getElementById('header-menu-btn');
        this.menu = document.getElementById('header-menu');

        if (this.menuBtn && this.menu) {
            this.bindEvents();
        }
    }

    bindEvents() {
        // Toggle menu on button click (supports both mouse and touch)
        this.menuBtn.addEventListener('click', (e) => {
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
                // Don't close if clicking on project dropdown
                const projectDropdown = document.getElementById('project-dropdown');
                if (!projectDropdown || !projectDropdown.contains(e.target)) {
                    this.closeMenu();
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
        if (this.isMenuOpen) {
            this.closeMenu();
        } else {
            this.openMenu();
        }
    }

    openMenu() {
        this.menu.style.display = 'block';
        this.menu.style.visibility = 'visible';
        this.isMenuOpen = true;
        this.menuBtn.setAttribute('aria-expanded', 'true');
        
        // Add animation class
        setTimeout(() => {
            this.menu.classList.add('menu-open');
        }, 10);
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

// Initialize header menu when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.headerMenuManager = new HeaderMenuManager();
});

export default HeaderMenuManager;
