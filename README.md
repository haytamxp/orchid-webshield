# Orchid WebShield

### Modular Web Application Security Assessment & Defense Platform

[![WebShield P4 CI](https://github.com/haytamxp/orchid-webshield/actions/workflows/backend-ci.yml/badge.svg?branch=main)](https://github.com/haytamxp/orchid-webshield/actions/workflows/backend-ci.yml)

**Orchid WebShield** is a cybersecurity engineering project designed to bring authorized offensive security testing, defensive monitoring, application security validation, and DevSecOps practices into one modular platform.

The project aims to help security engineers assess web application weaknesses, collect and normalize security findings, correlate defensive events, evaluate risk, and validate security improvements through automated CI/CD checks.

> **Project status:** Active development. The repository currently contains foundational Go services, a lab target web application, architectural documentation, and an operational backend CI pipeline. Several platform components remain under development.

---

## Table of Contents

- [Project Vision](#project-vision)
- [Core Capabilities](#core-capabilities)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Repository Structure](#repository-structure)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Security and DevSecOps](#security-and-devsecops)
- [Development Roadmap](#development-roadmap)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [Responsible Use](#responsible-use)
- [License](#license)

---

## Project Vision

WebShield is being developed around a central principle:

**Security testing should lead to measurable detection, actionable remediation, and verifiable security improvements.**

Rather than treating offensive testing and defensive monitoring as unrelated activities, WebShield is designed to connect their results through a common platform.

The intended workflow is:

1. Register a target application and define the authorized assessment scope.
2. Validate the requested operation against platform policies.
3. Execute approved security assessment tasks.
4. Collect and normalize findings and defensive events.
5. Correlate relevant events and assess the potential risk.
6. Present results for investigation and remediation.
7. Validate code changes through automated quality and security checks.
8. Repeat the assessment to verify improvements.

This workflow describes the platform's architectural direction; not every stage is implemented yet.

## Core Capabilities

### Offensive Security

The planned offensive worker is organized around modular security testing components.

Areas of development include:

- HTTP and API security assessment.
- Authentication and authorization testing.
- Input-validation assessment.
- Attack-surface discovery.
- Evidence collection and finding normalization.
- Retesting after remediation.

### Defensive Security

The defensive worker is intended to process security events and identify suspicious activity.

Planned capabilities include:

- Security-event ingestion and normalization.
- Detection rules.
- Event correlation.
- Alert generation.
- Incident-oriented evidence collection.
- Detection metrics and investigation support.

### Central Control Plane

The Go backend provides the foundation for a central coordination layer. Its intended responsibilities include:

- Platform API and health monitoring.
- Authentication and authorization.
- Target and assessment management.
- Scope and policy validation.
- Job orchestration.
- Findings and event management.
- Risk analysis, auditing, and reporting.

### Security-Focused Development

WebShield also uses automated checks to identify defects and security issues during development rather than waiting until deployment.

The current CI pipeline covers backend quality checks, static security analysis, Go vulnerability analysis, and repository secret scanning.

---

## Architecture

The architectural design separates the application under test from the platform responsible for coordinating its assessment.

```text
                    Security Engineer
                           |
                           v
                    Frontend / UI
                           |
                           v
                  Go Backend / API
                    Control Plane
                           |
             +-------------+-------------+
             |             |             |
             v             v             v
        Scope and       Assessment     Policy and
        Target          Management     Validation
        Validation
             |
             v
      Authorized Security Jobs
             |
        +----+----------------+
        |                     |
        v                     v
 Offensive Worker      Defensive Worker
        |                     ^
        v                     |
        +----> Lab Web App ---+
                    |
                    v
            Security Events
                    |
                    v
          Findings and Correlation
                    |
                    v
              Risk Analysis
                    |
                    v
             Reports / Dashboard

       CI/CD validates code changes separately
```

The diagram represents the intended architecture, not a claim that all integrations are operational.

### Architectural Components

| Component | Responsibility |
|---|---|
| `backend/` | Central Go HTTP API and control-plane foundation |
| `webapp/` | Lab target application used for application-security testing |
| `workers/offensive/` | Planned modular offensive security tests |
| `workers/defensive/` | Planned security-event processing and detection |
| `frontend/` | Frontend structure and integration foundation |
| `devsecops/` | CI documentation, local validation, and security-gate guidance |
| `.github/workflows/` | GitHub Actions automation |
| `docs/` | Architecture, design decisions, and security policies |

For further detail, see [`docs/GENERAL_ARCHITECTURE.md`](docs/GENERAL_ARCHITECTURE.md).

---

## Technology Stack

| Technology | Purpose |
|---|---|
| Go | Backend and lab application services |
| SQLite | Local database for the lab web application |
| Git and GitHub | Source control and collaborative development |
| GitHub Actions | Continuous Integration |
| Gosec | Static security analysis for Go code |
| govulncheck | Go vulnerability analysis |
| Gitleaks | Repository secret detection |
| PowerShell | Local validation and Windows development workflows |

**Toolchain:** The backend CI workflow currently pins Go 1.26.9.

The stack may evolve as additional modules are implemented.

---

## Repository Structure

```text
orchid-webshield/
├── .github/
│   └── workflows/
│       └── backend-ci.yml
├── backend/
│   ├── cmd/server/
│   ├── internal/
│   │   ├── api/
│   │   ├── config/
│   │   ├── handlers/
│   │   └── middleware/
│   ├── app/
│   └── go.mod
├── webapp/
│   ├── cmd/server/
│   ├── internal/
│   │   ├── config/
│   │   ├── db/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── models/
│   │   └── validation/
│   ├── database/
│   ├── tests/
│   ├── .env.example
│   └── go.mod
├── workers/
│   ├── offensive/
│   │   ├── api/
│   │   ├── authentication/
│   │   ├── authorization/
│   │   ├── discovery/
│   │   ├── http/
│   │   └── input-validation/
│   └── defensive/
│       ├── alerts/
│       ├── common/
│       ├── correlation/
│       ├── detection/
│       ├── ingestion/
│       └── rules/
├── frontend/
├── devsecops/
│   ├── ci/
│   ├── sast/
│   ├── sca/
│   └── secrets/
├── docs/
│   ├── GENERAL_ARCHITECTURE.md
│   └── security/
│       └── CI_SECURITY_GATES.md
└── README.md
```

Some listed directories are scaffolding for future implementation.

---

## Getting Started

### Prerequisites

Install:

- [Go](https://go.dev/dl/) — Go 1.26 or a compatible version supported by the modules.
- [Git](https://git-scm.com/downloads).
- PowerShell on Windows, or a compatible terminal on Linux/macOS.

### 1. Clone the Repository

```powershell
git clone https://github.com/haytamxp/orchid-webshield.git
Set-Location orchid-webshield
```

### 2. Run the Backend

Open a terminal in the repository root and run:

```powershell
Set-Location backend

go test -count=1 -mod=readonly ./...
go run ./cmd/server
```

By default, the backend listens on `127.0.0.1:8080`.

In a second terminal, check its health endpoint:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/api/v1/health
```

The health endpoint verifies that the backend HTTP service is responding. It does not verify every subsystem of the future platform.

### 3. Run the Lab Web Application

Open another terminal:

```powershell
Set-Location webapp

Copy-Item .env.example .env
go test -count=1 -mod=readonly ./...
go run ./cmd/server
```

The lab web application defaults to port `8081`.

Check its health endpoint:

```powershell
Invoke-RestMethod http://localhost:8081/health
```

Expected response:

```json
{
  "status": "ok"
}
```

Start each service in its own terminal.

For additional configuration and database details, see [`webapp/README.md`](webapp/README.md).

---

## Configuration

### Backend

The backend reads environment variables and validates its HTTP server configuration at startup.

| Variable | Default | Purpose |
|---|---|---|
| `HTTP_ADDR` | `127.0.0.1:8080` | HTTP listen address |
| `HTTP_READ_TIMEOUT` | `10s` | Request read timeout |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | Request-header timeout |
| `HTTP_WRITE_TIMEOUT` | `15s` | Response write timeout |
| `HTTP_IDLE_TIMEOUT` | `60s` | Keep-alive idle timeout |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown timeout |
| `HTTP_MAX_HEADER_BYTES` | `1048576` | Maximum HTTP header size |
| `MAX_REQUEST_BODY_BYTES` | `1048576` | Maximum request body size |
| `LOG_LEVEL` | `info` | Logging verbosity |

Environment variable names and defaults correspond to the current backend configuration.

### Lab Web Application

The `webapp/.env.example` file documents the lab application's configuration.

| Variable | Default | Purpose |
|---|---|---|
| `WEBAPP_PORT` | `8081` | Listening port |
| `WEBAPP_DATABASE_PATH` | `database/webshield.db` | SQLite database path |
| `WEBAPP_UPLOAD_DIR` | `uploads` | Upload directory |
| `WEBAPP_LOG_FILE` | `logs/security_events.jsonl` | Security event log path |
| `WEBAPP_SESSION_SECRET` | `change-me` | Session signing secret |
| `WEBAPP_COOKIE_SECURE` | `false` | Secure-cookie setting |

The lab application's default session secret is for development only. Configure an appropriate secret for local testing, never commit real secrets, and do not expose the application publicly with development settings.

---

## Security and DevSecOps

WebShield maintains a dedicated CI workflow:

[` .github/workflows/backend-ci.yml `](.github/workflows/backend-ci.yml)

The workflow currently includes these checks:

| Check | Purpose |
|---|---|
| Go toolchain verification | Ensure the expected version is used |
| `gofmt` | Validate Go formatting |
| `go vet` | Detect suspicious code patterns |
| `go test` | Execute Go tests |
| Race detector | Detect data races in tests |
| `go build` | Verify backend packages compile |
| Gosec | Analyze Go source for potential security issues |
| govulncheck | Analyze Go code and dependencies for known vulnerabilities |
| Gitleaks | Detect potential secrets in repository history and content |

### Security Gate Principles

- A failed check must be investigated.
- A scan that cannot complete is not a clean scan.
- Findings should be assessed and remediated according to their actual risk.
- Secrets must not be committed to source control.
- CI success does not prove that the application is completely secure.
- Required status checks should be enforced through repository rules before relying on CI to protect `main`.

**Current scope limitation:** The existing CI validates the backend module and scans the repository for secrets. Dedicated CI for the independent `webapp/` module, frontend, and security workers remains future work.

Local backend validation instructions are available in [`devsecops/ci/README.md`](devsecops/ci/README.md).

---

## Development Roadmap

The roadmap is iterative and subject to implementation and testing.

- [x] Establish the repository structure and architecture documentation.
- [x] Create the Go backend foundation and health endpoint.
- [x] Establish the lab web application's Go module and local database foundation.
- [x] Implement the initial backend CI workflow and security gates.
- [ ] Expand automated tests and security regression coverage.
- [ ] Build out the target application's security-testing scenarios.
- [ ] Implement modular offensive assessment workers.
- [ ] Implement defensive event ingestion, detection, correlation, and alerts.
- [ ] Define and test stable interfaces between workers and the backend.
- [ ] Implement finding normalization, risk analysis, and reporting.
- [ ] Add CI for the separate modules and frontend when ready.
- [ ] Define and test deployment configuration and operational monitoring.
- [ ] Enforce required CI status checks on `main`.

---

## Documentation

| Document | Description |
|---|---|
| [General Architecture](docs/GENERAL_ARCHITECTURE.md) | System components, responsibilities, and intended data flow |
| [CI Security Gates](docs/security/CI_SECURITY_GATES.md) | Security checks, failure classifications, and merge policy |
| [CI Operations](devsecops/ci/README.md) | Local validation and CI troubleshooting |
| [SAST Guidance](devsecops/sast/README.md) | Static analysis scope and usage |
| [Dependency Scanning](devsecops/sca/README.md) | Dependency vulnerability analysis |
| [Secret Scanning](devsecops/secrets/README.md) | Secret detection principles |
| [Lab Application README](webapp/README.md) | Target application setup and configuration |

---

## Contributing

Contributions should preserve clear component ownership and minimize unnecessary coupling.

1. Create a focused branch for the change.
2. Keep changes within the relevant component whenever possible.
3. Add tests or update existing tests when behavior changes.
4. Run the applicable local quality and security checks.
5. Submit a pull request and review all CI results.
6. Document architectural changes and integration contracts.

Never include real credentials, access tokens, session secrets, sensitive logs, or private assessment data in a commit.

---

## Responsible Use

WebShield is intended for **authorized security assessments and controlled laboratory environments**.

Only run offensive assessment tasks against systems that you own or have explicit permission to test. Define the permitted scope in advance, avoid disrupting production systems, and handle collected evidence responsibly.

The project is under active development and must not be considered production-ready solely because individual CI checks pass.

---

## License

No `LICENSE` file is currently defined in this repository. Until a license is added, do not assume that the source code is available for unrestricted reuse or redistribution.
