(function(window){
  var CONFIG = {endpoint: 'https://updog.bartel.com/view'};

  function trackPageview(data){
    if(navigator.sendBeacon){
      try { navigator.sendBeacon(CONFIG.endpoint, JSON.stringify(data)); return; }
      catch(e){ /* fallback below */ }
    }
    var img = new Image();
    img.src = CONFIG.endpoint.replace(/\/view$/,'/view.gif') +
              '?domain=' + encodeURIComponent(data.domain) +
              '&path=' + encodeURIComponent(data.path) +
              '&ref=' + encodeURIComponent(data.ref);
  }

  // Process queued events
  (window._uaq || []).forEach(function(args){
    if(args[0]==='pageview') trackPageview(args[1]);
    else if(args[0]==='config') Object.assign(CONFIG, args[1]);
  });

  // Override push for future events
  window._uaq.push = function(args){
    if(args[0]==='pageview') trackPageview(args[1]);
    else if(args[0]==='config') Object.assign(CONFIG, args[1]);
  };

  // SPA
  (function(history){
    var push = history.pushState;
    history.pushState = function(){
      push.apply(history, arguments);
      window._uaq.push(['pageview', {
        domain: location.hostname,
        path: location.pathname,
        ref: document.referrer
      }]);
    };
  })(history);

  // Back/forward navigation
  window.addEventListener('popstate', function(){
    window._uaq.push(['pageview', {
      domain: location.hostname,
      path: location.pathname,
      ref: document.referrer
    }]);
  });

})(window);

