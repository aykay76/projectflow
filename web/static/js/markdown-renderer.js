/**
 * Markdown Renderer Component
 * 
 * Provides rich Markdown rendering for documentation and chat messages.
 * Supports tables, code blocks with syntax highlighting, internal linking, and more.
 */

class MarkdownRenderer {
    constructor() {
        this.initialize();
    }

    /**
     * Initialize the Markdown renderer with configuration
     */
    initialize() {
        // Configuration for Markdown parsing
        this.config = {
            breaks: true,           // Support line breaks
            linkify: true,          // Auto-convert URLs to links
            typographer: true,      // Enable smart quotes and other typography
            html: false,           // Disable HTML for security
            xhtmlOut: true,        // Use XHTML-style tags
        };

        // Internal link patterns for documentation cross-references
        this.internalLinkPattern = /\[([^\]]+)\]\(([^)]+\.md(?:#[^)]*)?)\)/g;
        
        // Code language aliases for syntax highlighting
        this.languageAliases = {
            'js': 'javascript',
            'ts': 'typescript',
            'py': 'python',
            'sh': 'bash',
            'shell': 'bash',
            'yml': 'yaml'
        };
    }

    /**
     * Render Markdown content to HTML
     * @param {string} markdown - The Markdown content to render
     * @param {Object} options - Rendering options
     * @returns {string} HTML string
     */
    render(markdown, options = {}) {
        if (!markdown || typeof markdown !== 'string') {
            return '';
        }

        try {
            // Process internal links first
            let processedMarkdown = this.processInternalLinks(markdown, options.baseUrl);
            
            // Convert Markdown to HTML using our custom parser
            let html = this.parseMarkdown(processedMarkdown);
            
            // Post-process the HTML
            html = this.postProcessHtml(html, options);
            
            return html;
        } catch (error) {
            console.error('Markdown rendering error:', error);
            return `<div class="markdown-error">Error rendering content: ${error.message}</div>`;
        }
    }

    /**
     * Parse Markdown to HTML using a custom lightweight parser
     * This provides the core functionality without external dependencies
     */
    parseMarkdown(markdown) {
        let html = markdown;

        // Escape HTML first (security)
        html = this.escapeHtml(html);

        // Headers (# ## ### etc.)
        html = html.replace(/^### (.*$)/gim, '<h3>$1</h3>');
        html = html.replace(/^## (.*$)/gim, '<h2>$1</h2>');
        html = html.replace(/^# (.*$)/gim, '<h1>$1</h1>');

        // Bold **text** or __text__
        html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
        html = html.replace(/__(.*?)__/g, '<strong>$1</strong>');

        // Italic *text* or _text_
        html = html.replace(/\*(.*?)\*/g, '<em>$1</em>');
        html = html.replace(/_(.*?)_/g, '<em>$1</em>');

        // Inline code `code`
        html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

        // Code blocks ```language\ncode\n```
        html = html.replace(/```(\w+)?\n([\s\S]*?)\n```/g, (match, lang, code) => {
            const language = this.languageAliases[lang] || lang || 'text';
            return `<pre><code class="language-${language}">${this.escapeHtml(code.trim())}</code></pre>`;
        });

        // Tables
        html = this.parseTable(html);

        // Lists (unordered and ordered)
        html = this.parseLists(html);

        // Links [text](url)
        html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (match, text, url) => {
            const isExternal = url.startsWith('http') || url.startsWith('//');
            const target = isExternal ? ' target="_blank" rel="noopener noreferrer"' : '';
            return `<a href="${url}"${target}>${text}</a>`;
        });

        // Blockquotes > text
        html = html.replace(/^> (.+$)/gim, '<blockquote>$1</blockquote>');

        // Horizontal rules ---
        html = html.replace(/^---$/gm, '<hr>');

        // Line breaks
        html = html.replace(/\n/g, '<br>');

        // Clean up multiple <br> tags
        html = html.replace(/(<br>\s*){3,}/g, '<br><br>');

        return html;
    }

    /**
     * Parse tables in Markdown format
     */
    parseTable(html) {
        const tableRegex = /^(\|.*\|)\n(\|.*\|)\n((?:\|.*\|\n?)*)/gm;
        
        return html.replace(tableRegex, (match, header, separator, rows) => {
            const headerCells = header.split('|').slice(1, -1).map(cell => 
                `<th>${cell.trim()}</th>`
            ).join('');
            
            const rowElements = rows.trim().split('\n').map(row => {
                const cells = row.split('|').slice(1, -1).map(cell => 
                    `<td>${cell.trim()}</td>`
                ).join('');
                return `<tr>${cells}</tr>`;
            }).join('');
            
            return `<table><thead><tr>${headerCells}</tr></thead><tbody>${rowElements}</tbody></table>`;
        });
    }

    /**
     * Parse lists (both ordered and unordered)
     */
    parseLists(html) {
        const lines = html.split('<br>');
        const result = [];
        let inList = false;
        let listType = null;
        let listItems = [];

        for (let line of lines) {
            const unorderedMatch = line.match(/^(\s*)[•\-\*\+] (.+)$/);
            const orderedMatch = line.match(/^(\s*)\d+\. (.+)$/);

            if (unorderedMatch || orderedMatch) {
                const isOrdered = !!orderedMatch;
                const content = (unorderedMatch || orderedMatch)[2];
                
                if (!inList) {
                    inList = true;
                    listType = isOrdered ? 'ol' : 'ul';
                    listItems = [];
                }
                
                if ((isOrdered && listType === 'ol') || (!isOrdered && listType === 'ul')) {
                    listItems.push(`<li>${content}</li>`);
                } else {
                    // List type changed, close previous and start new
                    result.push(`<${listType}>${listItems.join('')}</${listType}>`);
                    listType = isOrdered ? 'ol' : 'ul';
                    listItems = [`<li>${content}</li>`];
                }
            } else {
                if (inList) {
                    result.push(`<${listType}>${listItems.join('')}</${listType}>`);
                    inList = false;
                    listType = null;
                    listItems = [];
                }
                result.push(line);
            }
        }

        // Close any remaining list
        if (inList) {
            result.push(`<${listType}>${listItems.join('')}</${listType}>`);
        }

        return result.join('<br>');
    }

    /**
     * Process internal documentation links
     */
    processInternalLinks(markdown, baseUrl = '/docs') {
        return markdown.replace(this.internalLinkPattern, (match, text, url) => {
            // Convert .md links to documentation routes
            let docUrl = url.replace(/\.md(#.*)?$/, '$1');
            
            // Handle relative paths
            if (!docUrl.startsWith('/')) {
                docUrl = `${baseUrl}/${docUrl}`;
            }
            
            return `[${text}](${docUrl})`;
        });
    }

    /**
     * Post-process HTML for additional features
     */
    postProcessHtml(html, options = {}) {
        // Add table wrapper for responsive design
        html = html.replace(/<table>/g, '<div class="table-wrapper"><table class="markdown-table">');
        html = html.replace(/<\/table>/g, '</table></div>');

        // Add classes to elements for styling
        html = html.replace(/<blockquote>/g, '<blockquote class="markdown-blockquote">');
        html = html.replace(/<code(?![^>]*class)/g, '<code class="markdown-inline-code"');
        html = html.replace(/<pre>/g, '<pre class="markdown-code-block">');

        // Add emoji support (basic)
        html = this.processEmojis(html);

        return html;
    }

    /**
     * Basic emoji processing for common emojis
     */
    processEmojis(html) {
        const emojiMap = {
            ':check:': '✅',
            ':x:': '❌',
            ':warning:': '⚠️',
            ':info:': 'ℹ️',
            ':bulb:': '💡',
            ':rocket:': '🚀',
            ':book:': '📚',
            ':gear:': '⚙️',
            ':chat:': '💬',
            ':folder:': '📂',
            ':file:': '📄'
        };

        for (const [shortcode, emoji] of Object.entries(emojiMap)) {
            html = html.replace(new RegExp(shortcode, 'g'), emoji);
        }

        return html;
    }

    /**
     * Escape HTML characters for security
     */
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    /**
     * Render Markdown into a DOM element
     * @param {HTMLElement} element - Target element
     * @param {string} markdown - Markdown content
     * @param {Object} options - Rendering options
     */
    renderToElement(element, markdown, options = {}) {
        if (!element) {
            console.error('No target element provided for Markdown rendering');
            return;
        }

        const html = this.render(markdown, options);
        element.innerHTML = html;
        
        // Add event listeners for enhanced functionality
        this.addEventListeners(element, options);
        
        // Trigger syntax highlighting if available
        this.highlightCode(element);
    }

    /**
     * Add event listeners for interactive features
     */
    addEventListeners(element, options = {}) {
        // Handle internal documentation links
        const links = element.querySelectorAll('a[href^="/docs"]');
        links.forEach(link => {
            link.addEventListener('click', (e) => {
                if (options.onInternalLink) {
                    e.preventDefault();
                    options.onInternalLink(link.getAttribute('href'));
                }
            });
        });

        // Handle copy code functionality
        const codeBlocks = element.querySelectorAll('pre code');
        codeBlocks.forEach(block => {
            this.addCopyButton(block);
        });
    }

    /**
     * Add copy button to code blocks
     */
    addCopyButton(codeBlock) {
        const pre = codeBlock.parentElement;
        if (pre.querySelector('.copy-button')) return; // Already has copy button

        const copyButton = document.createElement('button');
        copyButton.className = 'copy-button';
        copyButton.title = 'Copy code';
        copyButton.innerHTML = '📋';
        
        copyButton.addEventListener('click', async () => {
            try {
                await navigator.clipboard.writeText(codeBlock.textContent);
                copyButton.innerHTML = '✅';
                setTimeout(() => {
                    copyButton.innerHTML = '📋';
                }, 2000);
            } catch (err) {
                console.error('Failed to copy code:', err);
                copyButton.innerHTML = '❌';
                setTimeout(() => {
                    copyButton.innerHTML = '📋';
                }, 2000);
            }
        });

        pre.style.position = 'relative';
        pre.appendChild(copyButton);
    }

    /**
     * Highlight code blocks using Prism.js if available
     */
    highlightCode(element) {
        if (typeof Prism !== 'undefined') {
            Prism.highlightAllUnder(element);
        }
    }

    /**
     * Create a new instance for static usage
     */
    static create() {
        return new MarkdownRenderer();
    }

    /**
     * Static method for quick rendering
     */
    static render(markdown, options = {}) {
        const renderer = new MarkdownRenderer();
        return renderer.render(markdown, options);
    }
}

// Export for module usage
export default MarkdownRenderer;

// Also make available globally for non-module usage
window.MarkdownRenderer = MarkdownRenderer;
