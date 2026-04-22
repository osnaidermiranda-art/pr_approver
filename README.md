# PR Approved

A lightweight HTTP service written in Go that automates GitHub pull request actions — approve, merge, or both — supporting multiple GitHub accounts via per-organization tokens.

## How it works

The service exposes a `POST /git-hub` endpoint. You send a PR URL and an action, and the service calls the GitHub API on your behalf using the token configured for that organization.

```
POST /git-hub
```

### Request body

```json
{
  "url": "https://github.com/G97-TECH-MKT/my-repo/pull/42",
  "action": "both"
}
```

| Field    | Type   | Required | Description |
|----------|--------|----------|-------------|
| `url`    | string | yes      | Full GitHub pull request URL |
| `action` | string | no       | `approve`, `merge`, or `both` (default: `both`) |

### Actions

| Action    | Behavior |
|-----------|----------|
| `approve` | Approves the pull request |
| `merge`   | Squash merges the pull request |
| `both`    | Approves and then squash merges (default) |

### Responses

**200 OK**
```json
{ "message": "Pull request approved and merged", "status": "success" }
```

**400 Bad Request** — invalid URL, organization not in `VALID_OWNERS`, no token configured for that organization, or unknown action.

**405 Method Not Allowed** — only `POST` is accepted.

**500 Internal Server Error** — GitHub API call failed.

### Swagger UI

Interactive API docs available at:

```
http://localhost:8080/
```

---

## Requirements

- Go 1.25+
- A GitHub personal access token per organization with `repo` scope
- Docker (optional)

---

## Setup

```bash
cp .env.example .env
# Fill in your tokens in .env
```

---

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `GITHUB_TOKEN_<OWNER>` | yes (at least one) | Token for the GitHub org/user. Replace `<OWNER>` with the org name uppercased and `-` replaced by `_`. Example: `GITHUB_TOKEN_G97_TECH_MKT` |
| `VALID_OWNERS` | yes | Comma-separated list of allowed GitHub organization names (e.g. `G97-TECH-MKT,InnerPro-Sports`) |

### Token naming convention

| Organization | Environment variable |
|---|---|
| `G97-TECH-MKT` | `GITHUB_TOKEN_G97_TECH_MKT` |
| `InnerPro-Sports` | `GITHUB_TOKEN_INNERPRO_SPORTS` |

---

## Running locally

### With Make

```bash
make run
```

### With Docker Compose

```bash
docker compose up --build
```

The service will be available at `http://localhost:8080`.

---

## Development

```bash
make test      # run all tests
make test-v    # verbose output
make lint      # go vet
make build     # compile binary
make tidy      # go mod tidy
make clean     # remove binary
```

---

## Project structure

```
.
├── main.go                        # Entry point — wiring and server startup
├── internal/
│   ├── handler/
│   │   ├── github.go              # HTTP layer: parse requests, write responses
│   │   └── dto.go                 # Request/response types
│   ├── service/
│   │   ├── github.go              # Business logic: orchestrate approve/merge
│   │   └── errors.go              # Domain errors
│   └── ghclient/
│       └── client.go              # GitHub API adapter (wraps go-github)
├── docs/                          # Auto-generated Swagger spec (do not edit)
├── Makefile
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

---

## Deploying to Render

1. Push the repository to GitHub.
2. Go to [render.com](https://render.com) → **New** → **Web Service**.
3. Connect your GitHub repository.
4. Set **Port** to `8080` — Render will detect the `Dockerfile` automatically.
5. Add your `GITHUB_TOKEN_<OWNER>` and `VALID_OWNERS` environment variables.
6. Click **Create Web Service**.

Once deployed, your service will be live at:

```
https://<your-service-name>.onrender.com
```
