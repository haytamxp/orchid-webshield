# Orchid WebShield — General Architecture

## 1. Purpose

Orchid WebShield is a modular web-security validation platform combining:

- authorized offensive security assessment
- defensive monitoring and detection
- application security hardening and remediation
- security-event normalization
- risk analysis
- findings management
- CI/CD security validation
- security reporting

This document defines the initial architectural baseline of the project.

---

## 2. High-Level Architecture

The platform is divided into five major areas:

1. Frontend
2. Control Plane
3. Security Workers
4. Target Application (lab web application under test)
5. Data and Deployment

### Logical flow

    Frontend
        |
        v
    Control Plane
        |
        v
    Authentication / Authorization
        |
        v
    Target and Scope Validation
        |
        v
    Policy Engine
        |
        v
    Assessment
        |
        v
    Job
        |
        +--------------------+
        |                    |
        v                    v
    Offensive Worker    Defensive Worker
        |                    |
        | attacks            | monitors events
        v                    |
    Target Application ------+
    (webapp)
        |                    |
        +---------+----------+
                  |
                  v
           Result Normalization
                  |
            +-----+------+
            |            |
            v            v
         Findings      Events
            |            |
            +-----+------+
                  |
                  v
             Risk Analysis
                  |
                  v
            Dashboard / Report

The Target Application is the system under test. It emits security events that feed the defensive worker. It is not part of the control plane.

---

# 3. Team Ownership

### Person 1 — Offensive Security Engineer

Primary ownership:

    workers/offensive/

Responsibilities:

- test scope definition
- offensive threat modeling
- attack surface discovery
- HTTP, API, authentication, authorization and input-validation testing
- offensive evidence collection
- creation of normalized findings
- retest after remediation

Deliverables:

- docs/security-scope.md
- docs/threat-model.md
- findings
- retest results

---

### Person 2 — Blue Team / Detection Engineer

Primary ownership:

    workers/defensive/
    backend/app/alerts/

Responsibilities:

- log source definition
- security-event detection rules
- event correlation into incidents
- incident lifecycle
- alert generation
- detection metrics
- defensive evidence collection

Deliverables:

- docs/log-sources.md
- detection rules
- incidents
- detection metrics

---

### Person 3 — Application Security Engineer

Primary ownership:

    webapp/

Responsibilities:

- development of the lab target application
- application security architecture
- authentication, sessions and server-side RBAC
- API security and API security matrix
- application security logging (common event format)
- remediation of findings
- automated security regression tests
- asset criticality input for the risk engine

Deliverables:

- webapp/
- docs/app-security-architecture.md
- docs/api-security-matrix.md
- docs/remediation-log.md
- security regression tests

---

### Person 4 — Security Platform / DevSecOps Engineer

Primary ownership:

    backend/ (all except backend/app/alerts/)
    frontend/
    migrations/
    deploy/
    .github/workflows/
    docs/

Responsibilities:

- system architecture
- database and migrations
- central REST API
- platform authentication and authorization
- target management
- assessment lifecycle
- job orchestration
- common finding model
- common event model
- policy and scope enforcement
- risk engine
- audit trail
- dashboard
- CI/CD security gates
- reporting
- integration between offensive, defensive and application components

---

# 4. Control Plane

The control plane is the central coordination layer of WebShield.

It is responsible for:

- authentication
- authorization
- target management
- scope validation
- assessment management
- job management
- worker orchestration
- result normalization
- finding management
- event management
- risk analysis
- audit logging

Workers must never be directly exposed to the frontend.

---

# 5. Assessment

An assessment represents one authorized security assessment against a registered target.

## Assessment lifecycle

    CREATED
       |
       v
    VALIDATING
       |
       v
    QUEUED
       |
       v
    RUNNING
       |
       v
    ANALYZING
       |
       v
    COMPLETED

Failure states:

    FAILED
    CANCELLED

---

# 6. Jobs

A job represents one executable operation belonging to an assessment.

A job contains at minimum:

- id
- assessment_id
- type
- status
- target
- created_at
- started_at
- completed_at
- retry_count

## Job lifecycle

    QUEUED
       |
       v
    RUNNING
       |
       +------> COMPLETED
       |
       +------> FAILED
       |
       +------> CANCELLED

---

# 7. Policy Engine

Every security-sensitive operation must pass policy validation before a worker receives a job.

Minimum controls:

- authenticated actor
- authorized actor
- registered target
- scope validation
- allowed operation
- rate/concurrency limits
- audit logging

## Security flow

    Request
       |
       v
    Authentication
       |
       v
    Authorization
       |
       v
    Target validation
       |
       v
    Scope validation
       |
       v
    Policy validation
       |
       v
    Job creation
       |
       v
    Worker

A worker must never receive an unvalidated target directly from the frontend.

---

# 8. Findings

All security findings must use a common structure.

Minimum fields:

- id
- assessment_id
- title
- severity
- confidence
- asset
- endpoint
- owasp_category
- description
- evidence
- impact
- remediation
- status

Both offensive and defensive components can contribute information to the finding model.

## Finding lifecycle

    OPEN
      |
      v
    REMEDIATED
      |
      v
    RETEST
      |
      +------> PASS ------> CLOSED
      |
      +------> FAIL ------> REOPENED

## Example

```json
{
  "id": "WS-001",
  "assessment_id": "ASM-001",
  "title": "Broken Access Control",
  "asset": "AST-004",
  "endpoint": "/api/orders/123",
  "severity": "HIGH",
  "confidence": "HIGH",
  "owasp_category": "A01",
  "description": "...",
  "evidence": "...",
  "impact": "...",
  "remediation": "...",
  "status": "OPEN"
}
```

---

# 9. Events

Security telemetry must be normalized into a common event model.

Examples:

- HTTP request
- WAF event
- authentication event
- scanner event
- detection event
- application security event

Raw worker output must be validated before entering the normalized event model.

## Common event format

Events are JSON, one event per line (JSON Lines).

| Field | Required | Description |
|---|---|---|
| timestamp | Yes | ISO 8601 UTC, e.g. 2026-10-12T10:30:00Z |
| source | Yes | e.g. webapp |
| event_type | Yes | snake_case, see list below |
| src_ip | Yes | client IP address |
| severity | Yes | LOW, MEDIUM, HIGH, CRITICAL (uppercase) |
| user | No | username, null if unknown |
| endpoint | No | request path |
| method | No | HTTP method |
| status_code | No | HTTP status code |
| user_agent | No | raw user agent |
| details | No | free-form object |

## Application event types

| event_type | Default severity |
|---|---|
| login_success | LOW |
| logout | LOW |
| file_upload | LOW |
| login_failure | MEDIUM |
| authorization_denied | MEDIUM |
| password_reset | MEDIUM |
| admin_action | MEDIUM |
| suspicious_request | HIGH |

Events must never contain passwords, tokens, session cookies or file contents.

## Example

```json
{"timestamp":"2026-10-12T10:30:00Z","source":"webapp","event_type":"authorization_denied","src_ip":"10.10.10.25","user":"user01","endpoint":"/api/orders/123","method":"GET","status_code":403,"severity":"MEDIUM","details":{"reason":"not_owner"}}
```

## Incidents

Detection rules produce alerts. Correlation of several related events produces an incident.

Incident lifecycle:

    OPEN
      |
      v
    INVESTIGATING
      |
      v
    CONTAINED
      |
      v
    REMEDIATED
      |
      v
    CLOSED

---

# 10. Risk Analysis

Risk analysis combines multiple security signals.

Potential inputs:

- severity
- confidence
- asset criticality (provided by Person 3)
- exploitability (provided by Person 1)
- defensive evidence and detection gap (provided by Person 2)
- recurrence
- exposure

Example:

    High-severity finding
            +
    High-confidence evidence
            +
    Critical asset
            +
    Repeated suspicious activity
            |
            v
       High/Critical Risk

The risk engine should remain independent from individual offensive or defensive implementations.

---

# 11. Audit

Security-sensitive actions must be auditable.

Examples:

- assessment created
- target registered
- scope modified
- assessment started
- policy denied execution
- job created
- job dispatched
- finding created
- finding modified
- assessment cancelled

Audit records should contain at least:

- actor
- action
- resource
- timestamp
- result
- relevant metadata

---

# 12. Trust Boundaries

## Boundary 1 — User to API

Authentication and authorization are required.

## Boundary 2 — API to Policy Engine

Requests must be validated.

## Boundary 3 — Policy Engine to Workers

Only authorized jobs can cross this boundary.

## Boundary 4 — Workers to Control Plane

Worker output must be considered untrusted input and validated before processing.

## Boundary 5 — Control Plane to Database

Database access must use:

- parameterized queries
- least-privilege credentials
- protected secrets
- audit logging where appropriate

## Boundary 6 — Offensive Worker to Target Application

Only registered targets within the approved scope can be tested.

## Boundary 7 — Target Application to Event Pipeline

Events must follow the common event format and must never contain secrets.

---

# 13. Repository Structure

```text
orchid-webshield/
│
├── docs/
│
├── backend/
│   ├── app/
│   │   ├── api/
│   │   ├── auth/
│   │   ├── targets/
│   │   ├── assessments/
│   │   ├── jobs/
│   │   ├── findings/
│   │   ├── events/
│   │   ├── alerts/
│   │   ├── policies/
│   │   ├── risk/
│   │   └── audit/
│   │
│   └── tests/
│
├── workers/
│   ├── offensive/
│   └── defensive/
│
├── webapp/                  # target lab application (Person 3)
│   └── tests/security/      # security regression tests
│
├── frontend/
│
├── migrations/
│
├── deploy/
│
└── .github/
    └── workflows/
```

## Branches

```text
main
feature/architecture-foundation
feature/red-team          (Person 1)
feature/blue-team         (Person 2)
feature/app-security      (Person 3)
feature/platform          (Person 4)
```

Each member modifies only the directories they own. Changes to shared files (docs/) go through a pull request.

---

# 14. Target Application

The target application is the system that is attacked, monitored and hardened. It lives in webapp/ and is separate from the control plane.

Functional scope:

- registration, login, logout
- user profile and user dashboard
- admin dashboard and role management
- REST API (orders and other resources)
- database
- file upload
- password reset

Security architecture of every request:

    Request
       |
       v
    Authentication
       |
       v
    Authorization (RBAC, ownership)
       |
       v
    Validation
       |
       v
    Business logic
       |
       v
    Database

Rules:

- all authorization checks are performed server-side
- database access uses parameterized queries
- secrets are never stored in code
- security logging is designed in from the start and follows the common event format
- the application exposes a documented API security matrix (authentication, role, validation, rate limit, logging per endpoint)
- the application can be reset to a known state (database reset and seed) so that tests can be repeated

---

# 15. Remediation and Retest

    Finding (Person 1)
         |
         v
    Registered in WebShield (Person 4)
         |
         v
    Root cause analysis (Person 3)
         |
         v
    Fix + regression test + documentation (Person 3)
         |
         v
    CI/CD security gate (Person 4)
         |
         v
    Retest (Person 1)
         |
         +------> PASS ------> CLOSED
         |
         +------> FAIL ------> REOPENED
         |
         v
    Detection validation (Person 2)

For each finding, Person 3 provides: root cause, fix, commit, regression test and affected component.

Regression tests must be executable by the CI pipeline, so that a fixed vulnerability cannot silently reappear.

---

# 16. CI/CD

The pipeline runs:

    Code
     |
     +--> SAST
     +--> SCA (dependencies)
     +--> Secret scan
     +--> DAST
     +--> Security regression tests (webapp/tests/security)
     |
     v
    Security Gate

Gate rules:

| Result | Action |
|---|---|
| CRITICAL vulnerability | Pipeline fails |
| HIGH vulnerability | Review required |
| MEDIUM vulnerability | Warning |

---

# 17. Responsibility Matrix

| Function | P1 | P2 | P3 | P4 |
|---|---|---|---|---|
| Threat modeling | **Lead** | Support | Support | Support |
| Reconnaissance | **Lead** | - | - | Support |
| Pentest | **Lead** | Support | Support | Support |
| Findings | **Lead** | - | Fix | Integration |
| Target application | - | - | **Lead** | Support |
| Authentication / RBAC (target app) | Support | - | **Lead** | Support |
| API security | Lead test | Detection | Lead fix | Integration |
| Logging | Defines needs | **Lead** | Implements | Integration |
| Detection | Scenarios | **Lead** | Support | Integration |
| Incidents | - | **Lead** | Support | Integration |
| Risk engine | Risk context | Detection gap | Asset criticality | **Lead** |
| Dashboard | - | Data | Data | **Lead** |
| CI/CD | Security tests | Detection validation | Regression tests | **Lead** |
| Retest | **Lead** | Validates detection | Provides fix | Tracks |
| Reporting | Technical input | SOC input | Remediation input | **Lead** |

---

# 18. Interfaces Between Team Members

| Interface | Provided by the first | Provided by the second |
|---|---|---|
| P1 and P3 | Endpoint, test, finding, evidence, impact, recommendation | Root cause, fix, commit, regression test, affected component. Then P1 retests |
| P1 and P2 | Offensive scenarios (e.g. repeated failed authentication) | Detection rule, then detected YES or NO |
| P1 and P4 | Assets, endpoints, findings, evidence, severity, retest results | API, database, finding management, dashboard, reports |
| P2 and P3 | Detection needs | Application logs in the common event format |
| P2 and P4 | Normalized events, rules, incidents, metrics | Exposure through API and dashboard |
| P3 and P4 | Application, API, security controls, logs, tests | CI/CD, security gates, deployment, monitoring integration |

Every member produces standardized outputs that can be integrated automatically (finding, event and incident formats in sections 8 and 9).

---

# 19. Definition of Done

A task is done only when all of the following are true:

- code works
- tests pass
- documentation is written
- integration with the other modules works
- security review is done
- changes are committed to Git
- the interface with the other modules is validated

---

# 20. Target Delivery Scenario

The final demonstration shows the complete loop:

    ATTACK -> DETECTION -> INCIDENT -> FINDING -> FIX -> CI/CD -> RETEST -> VALIDATION

Example: Person 1 discovers that User A can read User B's order (WS-007, HIGH, A01 Broken Access Control). Person 4 registers the finding. Person 3 identifies the missing server-side ownership check, fixes it and adds a regression test. Person 2 creates a detection for repeated unauthorized object access. The pipeline passes. Person 1 retests and receives 403 Forbidden. Person 2 confirms the request was logged and detected. The finding becomes REMEDIATED, and the report shows: vulnerability fixed, offensive retest PASS, detection validation PASS, regression test PASS.