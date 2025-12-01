/**
 * Load and display real-time pageviews
 */
function loadRealtimePageviews() {
    const tbody = document.getElementById('realtime-tbody');
    const statusText = document.getElementById('status-text');
    const domainId = getSelectedDomainId();

    const queryParams = domainId ? `?domain_id=${domainId}` : '';

    if (!tbody) return; // Not on realtime page

    // Update status
    if (statusText) {
        statusText.textContent = 'Updating...';
    }

    // Fetch pageviews from API
    fetch('/api/v1/pageviews' + queryParams)
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
                const timestamp = new Date(pageview.timestamp);
                const timeStr = formatTimestamp(timestamp);

                // Build row HTML
                row.innerHTML = `
                    <td class="time-cell">${timeStr}</td>
                    <td class="path-cell">${escapeHtml(pageview.path || '/')}</td>
                    <td>${formatEntry(pageview.country)}</td>
                    <td>${formatEntry(pageview.region)}</td>
                    <td>${formatEntry(pageview.browser)}</td>
                    <td>${formatEntry(pageview.os)}</td>
                    <td>${formatEntry(pageview.device)}</td>
                    <td class="referrer-cell ${pageview.referrer ? '' : 'direct'}">${formatReferrer(pageview.referrer)}</td>
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

function getCookie(cname) {
  let name = cname + "=";
  let decodedCookie = decodeURIComponent(document.cookie);
  let ca = decodedCookie.split(';');
  for(let i = 0; i <ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) == ' ') {
      c = c.substring(1);
    }
    if (c.indexOf(name) == 0) {
      return c.substring(name.length, c.length);
    }
  }
  return "";
}

function setCookie(cname, cvalue, exdays) {
  const d = new Date();
  d.setTime(d.getTime() + (exdays*24*60*60*1000));
  let expires = "expires="+ d.toUTCString();
  document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/";
}

function setSelectedDomainId(domainId) {
    setCookie('selected_domain_id', domainId, 30);
    location.reload();
}

function getSelectedDomainId() {
    return getCookie('selected_domain_id');
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
function formatEntry(entry) {
  return entry || '—';
}


/**
 * Format referrer
 */
function formatReferrer(referrer) {
    if (!referrer) {
        return 'Direct';
    }
    return escapeHtml(referrer);
}

/**
 * Escape HTML to prevent XSS
 */
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Date Range Selector Logic
 */
document.addEventListener('DOMContentLoaded', function () {
    const dateRangeSelector = document.getElementById('dateRangeSelector');
    const dateRangeDropdown = document.getElementById('dateRangeDropdown');
    const dateRangeLabel = document.getElementById('dateRangeLabel');
    const presetDateRange = document.getElementById('presetDateRange');
    const customDateInput = document.getElementById('customDateInput');

    if (!dateRangeSelector || !dateRangeDropdown) return;

    // Toggle dropdown
    dateRangeSelector.addEventListener('click', function (e) {
        e.stopPropagation();
        dateRangeDropdown.classList.toggle('show');
    });

    // Close dropdown when clicking outside
    document.addEventListener('click', function (e) {
        if (!dateRangeSelector.contains(e.target) && !dateRangeDropdown.contains(e.target)) {
            dateRangeDropdown.classList.remove('show');
        }
    });

    // Handle preset selection
    if (presetDateRange) {
        presetDateRange.addEventListener('change', function () {
            const range = this.value;
            if (range) {
                handleDateRangeSelection(range);
            }
        });
    }

    // Initialize Flatpickr for custom range
    if (customDateInput) {
        flatpickr(customDateInput, {
            mode: "range",
            dateFormat: "Y-m-d",
            theme: "dark",
            onClose: function (selectedDates, dateStr, instance) {
                if (selectedDates.length === 2) {
                    const from = selectedDates[0].toISOString();
                    const to = selectedDates[1].toISOString();
                    reloadWithDateRange(from, to);
                }
            }
        });
    }

    // Check URL parameters to set active state and label
    const urlParams = new URLSearchParams(window.location.search);
    const fromParam = urlParams.get('from');
    const toParam = urlParams.get('to');

    if (fromParam && toParam) {
        // Try to match with presets
        const fromDate = new Date(fromParam);
        const toDate = new Date(toParam);

        dateRangeLabel.textContent = `${formatDate(fromDate)} - ${formatDate(toDate)}`;
    }
});

function handleDateRangeSelection(range) {
    const now = new Date();
    let from = new Date();
    let to = new Date();

    // Reset hours to start/end of day
    to.setHours(23, 59, 59, 999);
    from.setHours(0, 0, 0, 0);

    switch (range) {
        case 'today':
            // from and to are already today
            break;
        case 'yesterday':
            from.setDate(now.getDate() - 1);
            to.setDate(now.getDate() - 1);
            to.setHours(23, 59, 59, 999);
            break;
        case 'this-week':
            // Assuming week starts on Monday
            const day = now.getDay() || 7; // Get current day number, converting Sun (0) to 7
            if (day !== 1) from.setHours(-24 * (day - 1));
            break;
        case 'last-week':
            const lastWeekDay = now.getDay() || 7;
            from.setDate(now.getDate() - lastWeekDay - 6);
            to.setDate(now.getDate() - lastWeekDay);
            to.setHours(23, 59, 59, 999);
            break;
        case 'last-14-days':
            from.setDate(now.getDate() - 14);
            break;
        case 'last-28-days':
            from.setDate(now.getDate() - 28);
            break;
        case 'this-month':
            from.setDate(1);
            break;
        case 'last-month':
            from.setMonth(now.getMonth() - 1);
            from.setDate(1);
            to.setDate(0); // Last day of previous month
            to.setHours(23, 59, 59, 999);
            break;
        case 'this-year':
            from.setMonth(0, 1);
            break;
        case 'last-year':
            from.setFullYear(now.getFullYear() - 1);
            from.setMonth(0, 1);
            to.setFullYear(now.getFullYear() - 1);
            to.setMonth(11, 31);
            to.setHours(23, 59, 59, 999);
            break;
        case 'all-time':
            from = new Date('2000-01-01');
            break;
    }

    reloadWithDateRange(from.toISOString(), to.toISOString());
}

function reloadWithDateRange(from, to) {
    const url = new URL(window.location.href);
    url.searchParams.set('from', from);
    url.searchParams.set('to', to);
    window.location.href = url.toString();
}

function formatDate(date) {
    return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

