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
        // Toggle menu on button click
        this.menuBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            this.toggleMenu();
        });

        // Close menu when clicking outside
        document.addEventListener('click', (e) => {
            if (this.isMenuOpen && !this.menu.contains(e.target) && !this.menuBtn.contains(e.target)) {
                this.closeMenu();
            }
        });

        // Close menu on escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isMenuOpen) {
                this.closeMenu();
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
        
        // Hide after animation
        setTimeout(() => {
            if (!this.isMenuOpen) {
                this.menu.style.display = 'none';
            }
        }, 200);
    }
}

// Initialize header menu when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.headerMenuManager = new HeaderMenuManager();
});

export default HeaderMenuManager;
