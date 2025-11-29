/**
 * Load and display real-time pageviews
 */
function loadRealtimePageviews() {
    const tbody = document.getElementById('realtime-tbody');
    const statusText = document.getElementById('status-text');

    if (!tbody) return; // Not on realtime page

    // Update status
    if (statusText) {
        statusText.textContent = 'Updating...';
    }

    // Fetch pageviews from API
    fetch('/api/v1/pageviews')
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to fetch pageviews');
            }
            return response.json();
        })
        .then(data => {
            // Clear loading state
            tbody.innerHTML = '';

            // Check if we have data
            if (!data || data.length === 0) {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="7" class="empty-cell">
                            <i class="fa-solid fa-inbox"></i> No pageviews yet
                        </td>
                    </tr>
                `;
                if (statusText) {
                    statusText.textContent = 'No data';
                }
                return;
            }

            // Populate table with pageviews
            data.forEach(pageview => {
                const row = document.createElement('tr');

                // Format timestamp
                const timestamp = new Date(pageview.Timestamp);
                const timeStr = formatTimestamp(timestamp);

                // Build row HTML
                row.innerHTML = `
                    <td class="time-cell">${timeStr}</td>
                    <td class="path-cell">${escapeHtml(pageview.Path || '/')}</td>
                    <td>${formatCountry(pageview.Country)}</td>
                    <td>${formatBrowser(pageview.Browser)}</td>
                    <td>${formatOS(pageview.OS)}</td>
                    <td>${formatDeviceType(pageview.DeviceType)}</td>
                    <td class="referrer-cell ${pageview.Referrer ? '' : 'direct'}">${formatReferrer(pageview.Referrer)}</td>
                `;

                tbody.appendChild(row);
            });

            // Update status
            if (statusText) {
                statusText.textContent = `Live (${data.length} pageviews)`;
            }
        })
        .catch(error => {
            console.error('Error loading pageviews:', error);
            tbody.innerHTML = `
                <tr>
                    <td colspan="7" class="empty-cell">
                        <i class="fa-solid fa-triangle-exclamation"></i> Error loading pageviews
                    </td>
                </tr>
            `;
            if (statusText) {
                statusText.textContent = 'Error';
            }
        });
}

/**
 * Format timestamp to relative or absolute time
 */
function formatTimestamp(date) {
    const now = new Date();
    const diffMs = now - date;
    const diffSecs = Math.floor(diffMs / 1000);
    const diffMins = Math.floor(diffSecs / 60);
    const diffHours = Math.floor(diffMins / 60);

    if (diffSecs < 60) {
        return `${diffSecs}s ago`;
    } else if (diffMins < 60) {
        return `${diffMins}m ago`;
    } else if (diffHours < 24) {
        return `${diffHours}h ago`;
    } else {
        return date.toLocaleString();
    }
}

/**
 * Format country name
 */
function formatCountry(country) {
    if (!country || !country.Name) {
        return '—';
    }
    return escapeHtml(country.Name);
}

/**
 * Format browser name
 */
function formatBrowser(browser) {
    if (!browser || !browser.Name) {
        return '—';
    }
    return escapeHtml(browser.Name);
}

/**
 * Format operating system name
 */
function formatOS(os) {
    if (!os || !os.Name) {
        return '—';
    }
    return escapeHtml(os.Name);
}

/**
 * Format device type
 */
function formatDeviceType(deviceType) {
    if (!deviceType || !deviceType.Name) {
        return '—';
    }
    const name = deviceType.Name;
    return escapeHtml(name.charAt(0).toUpperCase() + name.slice(1));
}

/**
 * Format referrer
 */
function formatReferrer(referrer) {
    if (!referrer || !referrer.Host) {
        return 'Direct';
    }
    return escapeHtml(referrer.Host);
}

/**
 * Escape HTML to prevent XSS
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

