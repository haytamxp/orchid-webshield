# WebShield — Data Contract

**Version:** 1.0  
**Date:** 2026-10-14  
**Status:** Initial contract — implementation and integration status specified per section.

## 1. Purpose

This document defines the data exchanged between WebShield components, especially the defensive worker (P2), the WebShield Lab application (P3), and the backend/control plane (P4).

The objective is to ensure that all components use compatible event and alert formats.

## 2. Security Event Contract

### 2.1 Required fields

| Field | Type | Description |
|---|---|---|
| `schema_version` | integer | Contract version. Must be `1`. |
| `event_id` | string | Unique event identifier. |
| `timestamp` | string | Event timestamp in RFC3339 UTC format. |
| `source` | string | Component that produced the event, e.g. `webapp`. |
| `event_type` | string | Event category, e.g. `login_failure`. |
| `severity` | string | `info`, `low`, `medium`, `high`, or `critical`. |
| `src_ip` | string | Valid IPv4 or IPv6 address. |
| `request_id` | string | Identifier used to trace the request. |

### 2.2 Optional fields

| Field | Type | Description |
|---|---|---|
| `actor_id` | string | Identifier of the actor, if available. |
| `user` | string | Username or user reference, if available. |
| `target` | string | Target resource. |
| `endpoint` | string | Requested endpoint or route. |
| `method` | string | HTTP method. |
| `status_code` | integer | HTTP response status code. |
| `outcome` | string | Operation outcome, e.g. `success` or `failure`. |
| `reason_code` | string | Reason for the outcome, e.g. `invalid_credentials`. |
| `user_agent` | string | Client user-agent, if available. |
| `details` | object | Additional structured metadata. Must not contain secrets. |

Optional fields may be omitted when unavailable.

### 2.3 Example

```json
{
  "schema_version": 1,
  "event_id": "evt-20261014-000001",
  "timestamp": "2026-10-14T12:00:00Z",
  "source": "webapp",
  "event_type": "login_failure",
  "severity": "low",
  "src_ip": "192.168.1.10",
  "request_id": "req-20261014-000001",
  "outcome": "failure",
  "reason_code": "invalid_credentials",
  "endpoint": "/api/v1/auth/login",
  "method": "POST",
  "status_code": 401
}
```

### 2.4 Initial event types

| Event type | Suggested event severity | Meaning |
|---|---|---|
| `login_success` | `info` | Successful login. |
| `login_failure` | `low` | Failed login attempt. |
| `authorization_denied` | `medium` | Access denied by authorization controls. |

These are initial examples, not an exhaustive list. The `event_type` field remains extensible to future sources such as Suricata and Zeek.

**Important:** Event severity describes an individual event. Alert severity describes the risk determined by a detection rule; the two values do not have to match.

## 3. Event Transport — JSON Lines

The initial MVP uses JSON Lines (`.jsonl`): one JSON event per line.

The defensive worker reads events, decodes each JSON object, validates it against the event contract, and keeps valid events for downstream processing.

Invalid JSON lines or invalid events are reported as errors and skipped. Processing continues with subsequent lines when possible.

The exact input source in the integrated environment, including the Docker volume or alternative transport, remains to be agreed by the team.

## 4. Alert Contract — Defensive Worker to Backend

### 4.1 Proposed fields

| Field | Type | Description |
|---|---|---|
| `schema_version` | integer | Alert contract version. |
| `alert_id` | string | Unique alert identifier. |
| `created_at` | string | Alert creation timestamp in RFC3339 UTC format. |
| `source` | string | Alert producer, e.g. `defensive-worker`. |
| `rule_id` | string | Identifier of the rule that triggered the alert. |
| `title` | string | Short alert title. |
| `severity` | string | `low`, `medium`, `high`, or `critical`. |
| `confidence` | number | Optional confidence score, if defined by the team. |
| `event_ids` | array of strings | IDs of events supporting the alert. |
| `evidence` | object or array | Relevant evidence supporting the detection. Must not contain secrets. |

This is the proposed alert contract. Exact validation rules and the representation of `evidence` must be agreed before integration.

### 4.2 Proposed backend endpoint

`POST /api/v1/security/alerts`

The endpoint, authentication method, response format, and integration behavior must be confirmed with P4 before implementation.

The backend should validate incoming alerts, authenticate the producer, limit request size, persist accepted alerts, and handle duplicate `alert_id` values safely.

## 5. Security Requirements

- Never hardcode credentials, API tokens, or other secrets.
- Do not include passwords, session tokens, or other secrets in event details or alert evidence.
- Validate incoming data before processing it.
- Use UTC timestamps.
- Preserve event IDs for traceability and correlation.
- Restrict access to the alert ingestion endpoint.
- Define retry, timeout, and duplicate-handling behavior before production integration.

## 6. Implementation Status

| Component | Status |
|---|---|
| `SecurityEvent` Go structure | Implemented locally. |
| Event validation | Implemented and tested locally. |
| JSON Lines reader | Implemented and tested locally. |
| Detection rules and correlation | Not yet implemented in the current MVP. |
| Alert generation and serialization | Contract proposed; implementation pending. |
| Backend alert endpoint | Must be confirmed with P4. |
| Docker input source and transport | Must be agreed and integrated. |

## 7. Open Integration Decisions

The team must confirm:

1. The source and location of JSONL events in the integrated environment.
2. Whether events are exchanged through a shared Docker volume or another mechanism.
3. The final alert schema, including the type of `evidence` and the optional `confidence` field.
4. The backend endpoint and worker authentication method.
5. The backend response format and HTTP status codes.
6. Timeout, retry, and duplicate-alert handling.
7. Whether event and alert contract changes require a schema version update.

## 8. End-to-End Flow

The intended defensive pipeline is:

**Event input → Ingestion → Validation and normalization → Detection → Correlation → Alert generation → Backend ingestion**

The current implementation covers the initial event structure, validation, and JSON Lines reading. The remaining stages will be developed and integrated according to the team's agreed plan.