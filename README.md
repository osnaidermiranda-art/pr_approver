# PR Approved

A lightweight HTTP service written in Go that automates GitHub pull request actions — approve, merge, or both — for repositories inside the `G97-TECH-MKT` organization.

## How it works

The service exposes a single `POST /git-hub` endpoint. You send a PR URL and an action, and the service calls the GitHub API on your behalf using a personal access token.

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
{
  "message": "Pull request approved and merged",
  "status": "success"
}
```

**400 Bad Request** — invalid URL, repository outside `G97-TECH-MKT`, or unknown action.

**405 Method Not Allowed** — only `POST` is accepted.

**500 Internal Server Error** — GitHub API call failed.

### Swagger UI

Interactive API docs are available at the root path:

```
http://localhost:8080/
```

---

## Requirements

- Go 1.25+
- A GitHub personal access token with `repo` and `pull_request` scopes
- Docker (optional, for containerized runs)

---

## Running locally

### With Go

```bash
cp .env.example .env
# Fill in your GITHUB_TOKEN in .env

export $(cat .env | xargs)
go run main.go
```

### With Docker Compose

```bash
cp .env.example .env
# Fill in your GITHUB_TOKEN in .env

docker compose up --build
```

The service will be available at `http://localhost:8080`.

---

## Environment variables

| Variable       | Required | Description |
|----------------|----------|-------------|
| `GITHUB_TOKEN` | yes      | GitHub personal access token used to authenticate API calls |
| `VALID_REPOS`  | yes      | Comma-separated list of allowed GitHub organization names (e.g. `G97-TECH-MKT,another-org`) |

---

## Project structure

```
.
├── main.go          # Entry point — registers routes and starts the HTTP server
├── server/
│   └── server.go    # Request handler, GitHub API calls, URL parsing
├── docs/
│   └── docs.go      # Auto-generated Swagger spec (do not edit manually)
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
5. Add the environment variable `GITHUB_TOKEN` under **Environment Variables**.
6. Click **Create Web Service**.

Once deployed, your service will be live at:

```
https://<your-service-name>.onrender.com
```
