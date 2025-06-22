/**
 * ProjectFlow Documentation Manager
 * Handles documentation viewing with Markdown rendering
 */

import MarkdownRenderer from './markdown-renderer.js';
import { showMessage } from './utils.js';

export class DocsManager {
    constructor(apiClient) {
        this.apiClient = apiClient;
        this.markdownRenderer = new MarkdownRenderer();
        this.currentDoc = null;
        this.docsList = [];
        
        this.initializeElements();
        this.bindEvents();
        this.loadDocsList();
    }

    initializeElements() {
        // Modal elements
        this.modal = document.getElementById('docs-modal');
        this.closeBtn = document.querySelector('.docs-modal-close');
        this.openBtn = document.getElementById('docs-btn');
        
        // Navigation elements
        this.navList = document.getElementById('docs-nav-list');
        this.searchInput = document.getElementById('docs-search');
        
        // Content elements
        this.contentArea = document.getElementById('docs-content-area');
        this.loadingDiv = document.getElementById('docs-loading');
        
        // Quick links
        this.quickLinks = document.querySelectorAll('[data-doc]');
    }

    bindEvents() {
        // Modal open/close
        if (this.openBtn) {
            this.openBtn.addEventListener('click', () => this.showModal());
        }
        
        if (this.closeBtn) {
            this.closeBtn.addEventListener('click', () => this.hideModal());
        }
        
        // Close modal on backdrop click
        if (this.modal) {
            this.modal.addEventListener('click', (e) => {
                if (e.target === this.modal) {
                    this.hideModal();
                }
            });
        }
        
        // Search functionality
        if (this.searchInput) {
            this.searchInput.addEventListener('input', () => this.handleSearch());
        }
        
        // Quick links
        this.quickLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const docName = link.getAttribute('data-doc');
                this.loadDocument(docName);
            });
        });
        
        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isModalOpen()) {
                this.hideModal();
            }
        });
    }

    showModal() {
        if (this.modal) {
            this.modal.style.display = 'block';
            document.body.classList.add('modal-open');
            
            // Focus search input
            if (this.searchInput) {
                setTimeout(() => this.searchInput.focus(), 100);
            }
        }
    }

    hideModal() {
        if (this.modal) {
            this.modal.style.display = 'none';
            document.body.classList.remove('modal-open');
        }
    }

    isModalOpen() {
        return this.modal && this.modal.style.display === 'block';
    }

    async loadDocsList() {
        try {
            this.showLoading();
            
            const response = await fetch('/api/docs/list');
            if (!response.ok) {
                throw new Error(`Failed to load docs list: ${response.status}`);
            }
            
            const data = await response.json();
            this.docsList = data.docs || [];
            
            this.renderNavigation();
            this.hideLoading();
            
        } catch (error) {
            console.error('Error loading docs list:', error);
            showMessage('Failed to load documentation list', 'error');
            this.hideLoading();
        }
    }

    renderNavigation() {
        if (!this.navList) return;
        
        // Clear existing content
        this.navList.innerHTML = '';
        
        // Group docs by category (based on filename patterns)
        const categories = this.categorizeDocuments(this.docsList);
        
        // Render categories
        Object.entries(categories).forEach(([category, docs]) => {
            const categoryElement = this.createCategoryElement(category, docs);
            this.navList.appendChild(categoryElement);
        });
    }

    categorizeDocuments(docs) {
        const categories = {
            'Getting Started': [],
            'User Guides': [],
            'Technical Documentation': [],
            'Configuration': [],
            'Other': []
        };
        
        docs.forEach(doc => {
            const name = doc.filename.toLowerCase();
            
            if (name.includes('user') || name.includes('guide')) {
                categories['User Guides'].push(doc);
            } else if (name.includes('config') || name.includes('setup')) {
                categories['Configuration'].push(doc);
            } else if (name.includes('mcp') || name.includes('technical') || name.includes('postgresql')) {
                categories['Technical Documentation'].push(doc);
            } else if (name.includes('readme') || name.includes('getting-started')) {
                categories['Getting Started'].push(doc);
            } else {
                categories['Other'].push(doc);
            }
        });
        
        // Remove empty categories
        Object.keys(categories).forEach(key => {
            if (categories[key].length === 0) {
                delete categories[key];
            }
        });
        
        return categories;
    }

    createCategoryElement(category, docs) {
        const categoryDiv = document.createElement('div');
        categoryDiv.className = 'docs-category';
        
        const categoryHeader = document.createElement('h4');
        categoryHeader.className = 'docs-category-header';
        categoryHeader.textContent = category;
        categoryDiv.appendChild(categoryHeader);
        
        const docsList = document.createElement('ul');
        docsList.className = 'docs-category-list';
        
        docs.forEach(doc => {
            const listItem = document.createElement('li');
            listItem.className = 'docs-nav-item';
            
            const link = document.createElement('a');
            link.href = '#';
            link.className = 'docs-nav-link';
            link.textContent = this.formatDocTitle(doc.filename);
            link.setAttribute('data-doc', doc.filename);
            
            // Add metadata if available
            if (doc.title && doc.title !== doc.filename) {
                link.textContent = doc.title;
            }
            
            link.addEventListener('click', (e) => {
                e.preventDefault();
                this.loadDocument(doc.filename);
            });
            
            listItem.appendChild(link);
            docsList.appendChild(listItem);
        });
        
        categoryDiv.appendChild(docsList);
        return categoryDiv;
    }

    formatDocTitle(filename) {
        // Remove .md extension and convert to title case
        return filename
            .replace(/\.md$/, '')
            .split('-')
            .map(word => word.charAt(0).toUpperCase() + word.slice(1))
            .join(' ');
    }

    async loadDocument(docName) {
        try {
            this.showLoading();
            
            const response = await fetch(`/api/docs/${encodeURIComponent(docName)}`);
            if (!response.ok) {
                throw new Error(`Failed to load document: ${response.status}`);
            }
            
            const data = await response.json();
            this.currentDoc = data;
            
            await this.renderDocument(data);
            this.hideLoading();
            
            // Update active nav item
            this.updateActiveNavItem(docName);
            
        } catch (error) {
            console.error('Error loading document:', error);
            showMessage(`Failed to load document: ${docName}`, 'error');
            this.hideLoading();
        }
    }

    async renderDocument(doc) {
        if (!this.contentArea) return;
        
        try {
            // Render markdown content
            const htmlContent = await this.markdownRenderer.render(doc.content);
            
            // Create document wrapper
            const docWrapper = document.createElement('div');
            docWrapper.className = 'docs-document';
            
            // Add document metadata if available
            if (doc.title || doc.filename) {
                const headerDiv = document.createElement('div');
                headerDiv.className = 'docs-document-header';
                
                const title = document.createElement('h1');
                title.textContent = doc.title || this.formatDocTitle(doc.filename);
                headerDiv.appendChild(title);
                
                if (doc.last_modified) {
                    const lastModified = document.createElement('p');
                    lastModified.className = 'docs-last-modified';
                    lastModified.textContent = `Last updated: ${new Date(doc.last_modified).toLocaleDateString()}`;
                    headerDiv.appendChild(lastModified);
                }
                
                docWrapper.appendChild(headerDiv);
            }
            
            // Add rendered content
            const contentDiv = document.createElement('div');
            contentDiv.className = 'markdown-content';
            contentDiv.innerHTML = htmlContent;
            
            // Process internal links
            this.processInternalLinks(contentDiv);
            
            docWrapper.appendChild(contentDiv);
            
            // Replace content area
            this.contentArea.innerHTML = '';
            this.contentArea.appendChild(docWrapper);
            
        } catch (error) {
            console.error('Error rendering document:', error);
            this.contentArea.innerHTML = `
                <div class="docs-error">
                    <h2>⚠️ Error Loading Document</h2>
                    <p>Failed to render the document content.</p>
                    <p class="error-details">${error.message}</p>
                </div>
            `;
        }
    }

    processInternalLinks(contentDiv) {
        // Find all links that might be internal
        const links = contentDiv.querySelectorAll('a[href]');
        
        links.forEach(link => {
            const href = link.getAttribute('href');
            
            // Check if it's an internal doc link
            if (href.startsWith('#') || href.endsWith('.md') || this.isInternalDocLink(href)) {
                link.addEventListener('click', (e) => {
                    e.preventDefault();
                    const docName = this.extractDocName(href);
                    if (docName) {
                        this.loadDocument(docName);
                    }
                });
                link.classList.add('internal-link');
            } else if (href.startsWith('http')) {
                // External links open in new tab
                link.setAttribute('target', '_blank');
                link.setAttribute('rel', 'noopener noreferrer');
                link.classList.add('external-link');
            }
        });
    }

    isInternalDocLink(href) {
        // Check if the link might be to another document
        return this.docsList.some(doc => 
            href.includes(doc.filename) || 
            href.includes(doc.filename.replace('.md', ''))
        );
    }

    extractDocName(href) {
        // Extract document name from various link formats
        if (href.endsWith('.md')) {
            return href.split('/').pop();
        }
        if (href.startsWith('#')) {
            return href.substring(1);
        }
        return href;
    }

    updateActiveNavItem(docName) {
        // Remove active class from all nav links
        const navLinks = this.navList.querySelectorAll('.docs-nav-link');
        navLinks.forEach(link => link.classList.remove('active'));
        
        // Add active class to current document link
        const activeLink = this.navList.querySelector(`[data-doc="${docName}"]`);
        if (activeLink) {
            activeLink.classList.add('active');
        }
    }

    handleSearch() {
        const query = this.searchInput.value.toLowerCase().trim();
        
        if (!query) {
            this.renderNavigation();
            return;
        }
        
        // Filter documents based on search query
        const filteredDocs = this.docsList.filter(doc => 
            doc.filename.toLowerCase().includes(query) ||
            (doc.title && doc.title.toLowerCase().includes(query)) ||
            (doc.content && doc.content.toLowerCase().includes(query))
        );
        
        // Render filtered results
        this.renderSearchResults(filteredDocs, query);
    }

    renderSearchResults(docs, query) {
        if (!this.navList) return;
        
        this.navList.innerHTML = '';
        
        if (docs.length === 0) {
            const noResults = document.createElement('div');
            noResults.className = 'docs-no-results';
            noResults.innerHTML = `
                <p>No documents found for "${query}"</p>
                <p>Try a different search term.</p>
            `;
            this.navList.appendChild(noResults);
            return;
        }
        
        const resultsHeader = document.createElement('h4');
        resultsHeader.className = 'docs-search-results-header';
        resultsHeader.textContent = `Search Results (${docs.length})`;
        this.navList.appendChild(resultsHeader);
        
        const resultsList = document.createElement('ul');
        resultsList.className = 'docs-search-results';
        
        docs.forEach(doc => {
            const listItem = document.createElement('li');
            listItem.className = 'docs-nav-item';
            
            const link = document.createElement('a');
            link.href = '#';
            link.className = 'docs-nav-link';
            link.textContent = doc.title || this.formatDocTitle(doc.filename);
            link.setAttribute('data-doc', doc.filename);
            
            link.addEventListener('click', (e) => {
                e.preventDefault();
                this.loadDocument(doc.filename);
            });
            
            listItem.appendChild(link);
            resultsList.appendChild(listItem);
        });
        
        this.navList.appendChild(resultsList);
    }

    showLoading() {
        if (this.loadingDiv) {
            this.loadingDiv.style.display = 'block';
        }
    }

    hideLoading() {
        if (this.loadingDiv) {
            this.loadingDiv.style.display = 'none';
        }
    }
}
