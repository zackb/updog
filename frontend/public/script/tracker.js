(function (window, document) {
    'use strict';

    var CONFIG = {
        endpoint: 'https://updog.bartel.com/view'
    };

    function trackViaBeacon(data) {
        if (navigator.sendBeacon) {
            try {
                navigator.sendBeacon(CONFIG.endpoint, JSON.stringify(data));
                return true;
            } catch (e) {
                return false;
            }
        }
        return false;
    }

    function trackViaPixel(data) {
        var url = CONFIG.endpoint.replace(/\/view$/, '/view.gif') +
            '?domain=' + encodeURIComponent(data.domain) +
            '&path=' + encodeURIComponent(data.path) +
            '&ref=' + encodeURIComponent(data.ref);

        var img = new Image();
        img.src = url;
    }

    function trackPageview() {
        var data = {
            domain: window.location.hostname,
            path: window.location.pathname,
            ref: document.referrer
        };

        // Try sendBeacon first; fallback to pixel
        if (!trackViaBeacon(data)) {
            trackViaPixel(data);
        }
    }

    // Track initial load
    if (document.readyState === 'complete') {
        trackPageview();
    } else {
        window.addEventListener('load', trackPageview);
    }

    // SPA support: override pushState
    var pushState = history.pushState;
    history.pushState = function () {
        pushState.apply(history, arguments);
        trackPageview();
    };

    // Back/forward buttons
    window.addEventListener('popstate', trackPageview);

})(window, document);

