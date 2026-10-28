# 🔗 LinkForge

**A self-hosted URL shortener built with Go, PostgreSQL, Docker, and Nginx.**

LinkForge turns long URLs into short, shareable links through a responsive web interface and REST API. It features persistent PostgreSQL storage, HTTP redirects, rate limiting, and a containerized deployment workflow powered by Docker Compose and GitHub Actions.

The project includes automated testing, vulnerability checks, Docker image scanning, and publishing container images to GitHub Container Registry (GHCR).

## ✨ Features

- **URL shortening:** Convert long HTTP and HTTPS URLs into short links.
- **Web interface:** Responsive frontend built with HTML, CSS, and JavaScript.
- **REST API:** Create short links programmatically.
- **Persistent storage:** Store short-code mappings in PostgreSQL.
- **HTTP redirects:** Redirect visitors to their original URLs.
- **Rate limiting:** Limit requests to the URL-shortening endpoint.
- **Reverse proxy:** Route HTTP traffic through Nginx.
- **Containerized deployment:** Run the application and database with Docker Compose.
- **Continuous integration:** Automate Go tests, builds, and security checks with GitHub Actions.
- **Container registry:** Build and publish Docker images to GitHub Container Registry (GHCR).

## 🛠️ Tech Stack

| Technology | Purpose |
|---|---|
| Go | Backend API and URL redirection |
| PostgreSQL | Persistent link storage |
| HTML, CSS, JavaScript | Web interface |
| Docker | Application containerization |
| Docker Compose | Multi-container orchestration |
| Nginx | Reverse proxy |
| GitHub Actions | CI/CD automation |
| GHCR | Container image storage |
| Trivy | Vulnerability and configuration scanning |
| govulncheck | Go dependency vulnerability checks |

## 🏗️ Architecture

```text
                 User / Browser
                       |
                       v
                 Nginx :80
                       |
                       v
                 Go API :8080
                   /       \
                  v         v
           Web Interface  PostgreSQL
           HTML/CSS/JS    Link Storage
```

Nginx acts as the HTTP entry point and forwards requests to the Go API. The API serves the application routes, validates submitted URLs, and stores short-code mappings in PostgreSQL. When a visitor opens a short link, the API looks up its destination and returns an HTTP redirect.

The services communicate over Docker Compose's internal network. PostgreSQL and the API do not need to be exposed directly to the host for Nginx to communicate with them.

## 📁 Project Structure

```text
linkforge/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   ├── links.go
│   │   ├── links_test.go
│   │   └── ratelimit.go
│   └── store/
│       └── postgres.go
├── db/
│   └── schema.sql
├── nginx/
│   └── default.conf
├── web/
│   ├── index.html
│   ├── style.css
│   └── app.js
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── publish.yml
├── compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

*The workflow filenames above assume the CI and image-publishing workflows are named `ci.yml` and `publish.yml` in your repository.*

## 🚀 Getting Started

### Prerequisites

Install the following tools:

- [Git](https://git-scm.com/)
- [Docker Engine](https://docs.docker.com/engine/install/)
- [Docker Compose](https://docs.docker.com/compose/)

Go is also useful if you want to run tests or build the application directly outside Docker.

### 1. Clone the repository

```bash
git clone https://github.com/utkarshjain1338-hub/linkforge.git
cd linkforge
```

### 2. Start the application locally

For a source-based development deployment:

```bash
docker compose -f compose.yml up -d --build
```

This builds the Go application image from the local source and starts the services defined in `compose.yml`.

Check service status:

```bash
docker compose ps
```

View logs:

```bash
docker compose logs -f
```

### 3. Open the website

Visit:

**http://localhost/**

Paste a valid URL, click **Shorten URL**, and copy the generated short link.

The default Compose configuration binds published ports to `127.0.0.1`, keeping access local to the machine running Docker.

### 4. Use the published GHCR image

LinkForge's application image is published to GitHub Container Registry:

```text
ghcr.io/utkarshjain1338-hub/linkforge:latest
```

To pull the image:

```bash
docker pull ghcr.io/utkarshjain1338-hub/linkforge:latest
```

To deploy using the published image, configure the `api` service in `compose.yml` with the image reference instead of `build: .`:

```yaml
api:
  image: ghcr.io/utkarshjain1338-hub/linkforge:latest
```

Keep the rest of the API service configuration, environment variables, dependencies, and networking settings consistent with your project. Ensure the local Compose file, Nginx configuration, and database schema file are present.

Then start or update the stack:

```bash
docker compose pull api
docker compose up -d
```

> **Note:** The `build: .` approach is intended for local source-based development. The GHCR image approach runs the published application image. Choose the approach that matches your deployment.

### 5. Check application health

If the API exposes a `/health` endpoint, test it with:

```bash
curl -i http://localhost/health
```

A successful response indicates that the health endpoint is responding. The actual health response depends on the implementation.

### 6. Create a short link using the API

```bash
curl -i -X POST http://localhost/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'
```

A successful request returns `201 Created` with a JSON response similar to:

```json
{
  "code": "a1b2c3d4",
  "short_url": "http://localhost/a1b2c3d4",
  "target_url": "https://example.com"
}
```

The values above are illustrative; the application generates the actual short code.

### 7. Follow a short link

Use the `short_url` returned by the API:

```bash
curl -i http://localhost/a1b2c3d4
```

An existing short code should return an HTTP redirect to its original destination.

### 8. Stop the application

```bash
docker compose -f compose.yml down
```

This stops and removes the Compose containers and network while preserving data stored in named volumes.

**Warning:** `docker compose down -v` also removes named volumes and can delete your stored PostgreSQL data.

## 📡 API Reference

### `GET /health`

Checks the application's health endpoint, if enabled by the application.

### `POST /shorten`

Creates a short link.

**Request body**

```json
{
  "url": "https://example.com"
}
```

**Successful response:** `201 Created`

Returns the generated short code, short URL, and original destination.

Invalid request bodies and unsupported URL schemes should be rejected by the handler.

### `GET /{code}`

Looks up a short code and redirects the client to its original URL.

- `302 Found` — the short link exists and redirects.
- `404 Not Found` — the short code does not exist.

## ⚙️ Configuration

The application uses the `DATABASE_URL` environment variable to connect to PostgreSQL.

Within Docker Compose, the database hostname should be the service name, typically `db`, rather than `localhost`. Docker Compose provides internal DNS so services on the same Compose network can resolve one another by service name.

Example connection string for a local development environment:

```text
postgres://linkforge:devpassword@db:5432/linkforge?sslmode=disable
```

Use credentials appropriate to your environment. Do not reuse development credentials for a public deployment.

### Local-only access

The default port bindings should use loopback addresses:

```yaml
ports:
  - "127.0.0.1:80:80"
```

The API can use Docker's internal networking without publishing port `8080` on the host if Nginx is its entry point. PostgreSQL can likewise remain internal if the host does not need direct database access.

With loopback bindings, the website is intended to be accessed from the same machine running Docker, rather than directly from other devices on the LAN.

## 🔄 CI/CD Pipeline

GitHub Actions automates checks and container image publishing.

### Continuous integration

The CI workflow can run on pushes to `main` and pull requests targeting `main`. Its checks include:

- Go unit tests using `go test ./...`
- Go application build checks
- Go dependency vulnerability checks with `govulncheck`
- Docker image build verification
- Filesystem vulnerability, secret, and misconfiguration scanning with Trivy
- Docker image vulnerability scanning with Trivy

### Container image publishing

The GHCR publishing workflow builds the Docker image and publishes tags to:

```text
ghcr.io/utkarshjain1338-hub/linkforge:latest
```

It can also publish a commit-specific tag, allowing a particular build to be identified.

The publishing workflow uses GitHub Actions' `GITHUB_TOKEN` with package write permissions. The package's visibility and access settings determine whether authentication is required when pulling the image.

### Deployment workflow

The deployment process is:

1. Push application changes to GitHub.
2. GitHub Actions runs the configured CI checks.
3. The publishing workflow builds and pushes the application image to GHCR.
4. On the machine running LinkForge, pull the updated image and recreate the API service.

For a manual update:

```bash
docker compose pull api
docker compose up -d api
```

These commands update the API service according to the image reference in Compose. They do not automatically configure a remote deployment or guarantee that the CI checks must pass before publishing; that behavior depends on the workflow configuration.

## 🧪 Testing

Run the Go test suite:

```bash
go test ./...
```

Build the application:

```bash
go build ./...
```

Build the Docker image:

```bash
docker build -t linkforge:test .
```

Run these commands from the repository root in an environment with Go and Docker installed.

## 🔐 Security Notes

- Only HTTP and HTTPS destination URLs should be accepted.
- Request bodies should be size-limited by the API handler.
- The shortening endpoint uses an in-memory rate limiter.
- Database credentials should be managed securely and should not be hardcoded for production deployments.
- Public deployments should use HTTPS and appropriate reverse-proxy configuration.
- The current rate limiter is process-local; multi-instance deployments need a shared rate-limiting strategy.
- Short links can redirect visitors to external destinations. Only shorten links you trust.
- Keep PostgreSQL and internal API ports private unless direct access is explicitly required.
- Review Trivy and `govulncheck` findings and update vulnerable dependencies where appropriate.

## 🎯 What I Learned

Building LinkForge provided hands-on practice with:

- Designing and testing a REST API in Go.
- Integrating Go with PostgreSQL.
- Managing multi-container applications with Docker Compose.
- Serving application traffic through Nginx.
- Implementing HTTP redirects and request rate limiting.
- Building and publishing Docker images with GitHub Actions.
- Using GHCR as a container image registry.
- Automating testing, builds, and security checks in CI.
- Debugging Docker networking, port conflicts, and container startup failures.

---

**LinkForge** — a hands-on project for learning Go development, containers, and practical DevOps workflows.
