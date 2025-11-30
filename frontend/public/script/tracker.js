(function (window, document) {
    'use strict';

    var CONFIG = {
      endpoint: 'https://updog.bartel.com/view',
    };

    /**
     * Send pageview data to the server
     */
    function trackPageview() {
        var payload = {
            domain: window.location.hostname,
            path: window.location.pathname,
            ref: document.referrer
        };

        // Use sendBeacon if available for reliable delivery
        if (navigator.sendBeacon) {
            navigator.sendBeacon(CONFIG.endpoint, JSON.stringify(payload));
        } else {
            // Fallback for older browsers
            var xhr = new XMLHttpRequest();
            xhr.open('POST', CONFIG.endpoint, true);
            xhr.setRequestHeader('Content-Type', 'text/plain;charset=UTF-8');
            xhr.send(JSON.stringify(payload));
        }
    }

    // Track initial page load
    if (document.readyState === 'complete') {
        trackPageview();
    } else {
        window.addEventListener('load', trackPageview);
    }

    // Handle SPA navigation (History API)
    var history = window.history;
    var pushState = history.pushState;

    // Override pushState to track changes
    history.pushState = function () {
        pushState.apply(history, arguments);
        trackPageview();
    };

    // Listen for popstate events (back/forward button)
    window.addEventListener('popstate', function () {
        trackPageview();
    });

})(window, document);
