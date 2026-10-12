# WebShield Lab Application (webapp)

Target web application of the Orchid WebShield project.
It is the system that is attacked (offensive worker), monitored (defensive worker)
and hardened (application security). It is **not** part of the control plane.

Owner: Person 3 (Application Security Engineer).

> This application is for use in an authorized lab environment only.

## Status

| Module | Feature | State |
|---|---|---|
| M1 | Project skeleton, configuration, SQLite database, `/health` | Done |
| M2 | Authentication and sessions | Planned |
| M3 | RBAC and authorization | Planned |
| M4 | Business features (profile, orders, upload, password reset) | Planned |
| M5 | REST API and API security matrix | Planned |
| M6 | Security logging (common event format) | Planned |

## Requirements

- Go 1.26 or later

## Run

From the `webapp/` directory:

```powershell
Copy-Item .env.example .env
go run ./cmd/server
```

On Linux or macOS, use `cp .env.example .env` instead of `Copy-Item`.

Expected output:

```
database ready: 4 tables
webapp listening on :8081
```

The server must be started from the `webapp/` directory, because `.env` is read
from the current directory.

## Check that it works

```powershell
Invoke-RestMethod http://localhost:8081/health
```

With curl:

```
curl http://localhost:8081/health
```

Expected response:

```json
{"status":"ok"}
```

## Configuration

Settings are read from environment variables, or from the `.env` file.
A variable already defined in the environment takes priority over `.env`.

| Variable | Default | Description |
|---|---|---|
| `WEBAPP_PORT` | `8081` | Listening port |
| `WEBAPP_DATABASE_PATH` | `database/webshield.db` | SQLite database file |
| `WEBAPP_UPLOAD_DIR` | `uploads` | Directory for uploaded files |
| `WEBAPP_LOG_FILE` | `logs/security_events.jsonl` | Security event log (JSON Lines) |
| `WEBAPP_SESSION_SECRET` | `change-me` | Secret used for sessions |
| `WEBAPP_COOKIE_SECURE` | `false` | Set to `true` when served over HTTPS |

`.env` contains local settings and must never be committed.
Use `.env.example` as the template.

## Database

SQLite, created automatically at startup from `database/schema.sql`.

| Table | Content |
|---|---|
| `users` | Accounts, password hashes and roles (`USER` or `ADMIN`) |
| `orders` | Orders, each owned by one user |
| `uploads` | Metadata of uploaded files |
| `password_resets` | Password reset tokens (hashed) |

To reset the database, stop the server and delete `database/webshield.db`.
It is recreated at the next start.

## Structure

```text
webapp/
├── cmd/server/          # program entry point
├── internal/
│   ├── config/          # configuration loading
│   ├── db/              # database connection and schema
│   ├── models/          # data access
│   ├── handlers/        # HTTP routes
│   ├── middleware/      # authentication and authorization
│   └── seclog/          # security event logging
├── database/            # schema.sql
├── templates/           # HTML pages
├── static/              # CSS
├── tests/security/      # security regression tests
└── uploads/             # uploaded files (content not committed)
```

## Endpoints

| Endpoint | Method | Auth | Description |
|---|---|---|---|
| `/health` | GET | No | Liveness check |

The full list of endpoints will be documented in `docs/api-security-matrix.md`.