// Core Application Bootstrap
class ThreadlyApp {
    constructor() {
        this.init();
    }

    init() {
        // Initialize HTMX
        if (typeof htmx !== 'undefined') {
            htmx.config.globalViewTransitions = true;
            htmx.config.revalidateOnLoad = false;
        }

        // Initialize Alpine.js
        if (typeof Alpine !== 'undefined') {
            Alpine.start();
        }

        // Setup global event listeners
        this.setupEventListeners();
        
        // Initialize client-specific functionality
        this.initClientFeatures();
    }

    setupEventListeners() {
        // Handle form submissions with HTMX
        document.addEventListener('htmx:afterRequest', (event) => {
            this.handleHtmxResponse(event);
        });

        // Handle errors
        document.addEventListener('htmx:responseError', (event) => {
            this.handleError(event);
        });

        // Auto-scroll messages
        const messagesContainer = document.getElementById('messages-container');
        if (messagesContainer) {
            this.autoScrollMessages(messagesContainer);
        }
    }

    initClientFeatures() {
        // Initialize heartbeat for keeping session alive
        this.startHeartbeat();
        
        // Initialize real-time updates
        this.initRealTimeUpdates();
    }

    handleHtmxResponse(event) {
        const target = event.target;
        
        // Clear form inputs after successful submission
        if (event.detail.successful) {
            const form = target.closest('form');
            if (form) {
                const input = form.querySelector('input[type="text"], input[type="email"]');
                if (input) {
                    input.value = '';
                }
            }
        }
    }

    handleError(event) {
        console.error('HTMX Error:', event);
        
        // Show user-friendly error message
        const errorDiv = document.createElement('div');
        errorDiv.className = 'fixed top-4 right-4 bg-red-500 text-white px-4 py-2 rounded-lg shadow-lg z-50';
        errorDiv.textContent = 'An error occurred. Please try again.';
        document.body.appendChild(errorDiv);
        
        setTimeout(() => {
            errorDiv.remove();
        }, 3000);
    }

    autoScrollMessages(container) {
        const observer = new MutationObserver(() => {
            container.scrollTop = container.scrollHeight;
        });
        
        observer.observe(container, { childList: true });
        
        // Initial scroll
        container.scrollTop = container.scrollHeight;
    }

    startHeartbeat() {
        // Send heartbeat every 30 seconds to keep session alive
        setInterval(() => {
            fetch('/client/heartbeat', { method: 'POST' })
                .catch(error => console.log('Heartbeat failed:', error));
        }, 30000);
    }

    initRealTimeUpdates() {
        // Poll for new messages every 5 seconds
        setInterval(() => {
            this.checkForNewMessages();
        }, 5000);
    }

    checkForNewMessages() {
        // Implementation for checking new messages
        // This would depend on the current page context
        const currentPath = window.location.pathname;
        if (currentPath.includes('/client/businesses/') && currentPath.includes('/messages')) {
            // We're on a chat page, check for new messages
            const businessId = currentPath.split('/')[3];
            this.fetchNewMessages(businessId);
        }
    }

    fetchNewMessages(businessId) {
        fetch(`/client/businesses/${businessId}/messages`)
            .then(response => response.json())
            .then(data => {
                // Update messages if needed
                // This would be implemented based on the API response structure
            })
            .catch(error => console.log('Failed to fetch new messages:', error));
    }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new ThreadlyApp();
});

// Export for potential use in other modules
window.ThreadlyApp = ThreadlyApp;
