# Updog

Updog is a privacy-focused, self-hosted web analytics platform. It provides real-time insights into your website's traffic without compromising user privacy.

## Features

- **Real-time Analytics**: View current visitors, pageviews, and active sessions in real-time.
- **Privacy Focused**: Designed to respect user privacy.
- **Detailed Metrics**: Track top pages, referrers, device types, browsers, and operating systems.
- **Geographic Data**: Visualize visitor locations with country and city-level granularity (powered by MaxMind).
- **Multi-Domain Support**: Manage and track multiple domains from a single dashboard.
- **Flexible Storage**: Supports both SQLite (default) and PostgreSQL databases.
- **Self-Hosted**: Full control over your data and infrastructure.

## Getting Started

### Prerequisites

- **Go**: Version 1.25 or higher.
- **Make**: For running build commands.
- **Docker** (Optional): For containerized deployment.

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/zackb/updog.git
    cd updog
    ```

2.  Download dependencies:
    ```bash
    go mod download
    ```

### Building

To build the project, run:

```bash
make build
```

This will create the `updog` binary in the current directory.

### Running

To run the application locally:

```bash
make run
```

For development mode (with verbose logging):

```bash
make run-dev
```

The server will start on port `8080` by default. Access the dashboard at `http://localhost:8080`.

## Configuration

Updog is configured via environment variables.

| Variable | Description | Default |
| :--- | :--- | :--- |
| `HTTP_PORT` | The port the HTTP server listens on. | `8080` |
| `DATABASE_URL` | Database connection string. Use `postgres://...` for PostgreSQL. If empty or `sqlite`, defaults to SQLite. | `file:db.db?cache=shared&_fk=1` |
| `MAXMIND_CITY_DB` | Path to the MaxMind GeoLite2 City database. | `data/maxmind/GeoLite2-City.mmdb` |
| `TLS_CERT_PATH` | Path to the TLS certificate file for HTTPS. | `""` |
| `TLS_KEY_PATH` | Path to the TLS key file for HTTPS. | `""` |
| `DEV` | Set to `true` or `1` to enable development mode. | `false` |

## Usage

To track a website, add the following snippet to the `<head>` of your HTML pages:

```html
<script async src="https://your-updog-instance.com/static/script/ua.js"></script>
<script>
  window._uaq = window._uaq || [];
  function ua(){_uaq.push(arguments);}
  ua('pageview', {domain: location.hostname, path: location.pathname, ref: document.referrer});
  // Replace with your hosted Updog endpoint
  ua('config', {endpoint: 'https://your-updog-instance.com'});
</script>
```

Replace `https://your-updog-instance.com` with the URL of your Updog installation.

## Development

### Running Tests

To run the test suite:

```bash
make test
```

### Docker

To build the Docker image locally:

```bash
make docker-local
```

To build multi-architecture images (requires Docker Buildx):

```bash
make docker-multiarch
```

## License

[MIT](LICENSE)
