# Orchid WebShield — General Architecture

## 1. Purpose

Orchid WebShield is a modular web-security assessment platform combining:

- authorized offensive security assessment
- defensive monitoring
- security-event normalization
- risk analysis
- findings management
- security reporting

This document defines the initial architectural baseline of the project.

---

## 2. High-Level Architecture

The platform is divided into four major areas:

1. Frontend
2. Control Plane
3. Security Workers
4. Data and Deployment

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

---

## 3. Team Ownership

### Person 1 — Offensive Security

Primary ownership:

    workers/offensive/

Responsibilities:

- discovery
- HTTP security testing
- API security testing
- authentication testing
- authorization testing
- input-validation testing
- offensive evidence collection

---

### Person 2 — Defensive Security

Primary ownership:

    workers/defensive/

Additional ownership:

    backend/app/alerts/

Responsibilities:

- log ingestion
- security-event detection
- detection rules
- event correlation
- alert generation
- defensive evidence collection

---

### Person 3 — Platform

Primary ownership:

    frontend/
    backend/app/api/
    backend/app/auth/
    backend/app/targets/
    migrations/

Responsibilities:

- frontend
- REST API
- authentication
- authorization implementation
- target management
- database
- migrations

---

### Person 4 — Architecture, Orchestration and Security Integration

Primary ownership:

    docs/

    backend/app/assessments/
    backend/app/jobs/
    backend/app/findings/
    backend/app/events/
    backend/app/policies/
    backend/app/risk/
    backend/app/audit/

    .github/workflows/
    deploy/

Responsibilities:

- system architecture
- assessment lifecycle
- job orchestration
- common finding model
- common event model
- policy/scope enforcement
- risk integration
- audit trail
- CI security
- integration between offensive and defensive components

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
- description
- evidence
- remediation
- status

Both offensive and defensive components can contribute information to the finding model.

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

---

# 10. Risk Analysis

Risk analysis combines multiple security signals.

Potential inputs:

- severity
- confidence
- asset criticality
- exploitability
- defensive evidence
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
├── frontend/
│
├── migrations/
│
├── deploy/
│
└── .github/
    └── workflows/
