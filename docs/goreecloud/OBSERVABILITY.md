# GoreeCloud Backup — Observability and Audit Contract

## Status

This document defines the production observability contract for GoreeCloud Backup while it remains a Kopia-derived maintained fork. It is a release-governance contract, not evidence that target-environment monitoring or restore validation has already occurred.

Security-facing observability is presented under **Wardveil Security by GoreeCloud**. Wardveil identifies and presents security posture; it does not replace the underlying backup engine, authentication system, repository, monitor, notification service, policy, or recovery evidence that produced the state.

## Goals

Operational and security telemetry must make failures diagnosable without turning logs into a second store of protected backup data. Logging must support troubleshooting, monitoring, audit review, security investigation, controlled update validation, and recovery analysis while following GoreeCloud privacy and sensitive-information requirements.

Logs and status records must never be treated as proof that protected data is recoverable. Representative restoration remains the authoritative recovery test.

## Structured logging direction

The production target is structured, machine-readable event logging for GoreeCloud-owned application, security, administrative, and integration events. The inherited Kopia logger may continue to emit legacy formatted lines during the maintained-fork transition; those inherited lines are not automatically classified as meeting the GoreeCloud structured-event contract.

New GoreeCloud-owned event producers should prefer stable event names and typed or sanitized fields rather than concatenating untrusted values into prose. A later direct logging implementation may use JSON or another documented structured encoding, but this contract does not require a new runtime dependency merely to satisfy a format preference.

A transition must not remove useful inherited diagnostic evidence before an equivalent or better privacy-safe event exists.

## Event categories

The maintained application should expose or preserve enough structured context for the following event families where the underlying component supports them:

- process startup, readiness, shutdown, crash, and restart;
- repository connection, disconnection, degraded access, synchronization, and maintenance;
- snapshot scheduling, start, completion, cancellation, warning, and failure;
- restore request, start, completion, cancellation, validation, and failure;
- policy and retention changes;
- authentication success, authentication failure, logout, and rejected authentication material;
- authorization denial and rejected privileged operations;
- CSRF and other request-integrity failures;
- security-control failures and Wardveil security findings;
- desktop-shell certificate, IPC, navigation, update, and renderer failures;
- dependency/update checks and update failures;
- notification and monitoring integration delivery/failure;
- API and integration errors;
- storage-capacity and resource-pressure warnings;
- repository integrity and verification results;
- administrative configuration changes that materially affect recovery or security.

## Structured context

Where practical, events should use stable machine-readable names and fields rather than relying only on prose. Useful non-secret context may include:

- event name;
- severity;
- component;
- operation type;
- server-generated correlation identifier;
- internal task identifier;
- result category;
- duration;
- retry count;
- target repository identifier that does not expose a credential or secret URL;
- policy identifier;
- application/build version;
- upstream baseline where relevant;
- Wardveil control/category for security events;
- error class and safe diagnostic summary.

Correlation identifiers must be generated or validated by the application. Caller-supplied request identifiers must not automatically become trusted audit identifiers.

## Privacy and sensitive-information rules

The following must not be written to routine application, security, integration, or audit logs:

- passwords, backup passwords, API keys, access tokens, refresh tokens, bearer tokens, session cookies, CSRF tokens, private keys, encryption keys, recovery codes, or signing material;
- `Authorization` or `Cookie` header values;
- repository connection strings containing reusable credentials;
- raw request bodies that may contain credentials or protected backup configuration;
- backup file contents or restored file contents;
- database exports or secret-file contents;
- caller-controlled secrets embedded in URLs or query strings.

File names, filesystem paths, repository locations, usernames, IP addresses, forwarded-address values, user agents, and source identifiers may themselves be sensitive. They must be omitted, normalized, pseudonymized, or included only when there is a documented operational need and the relevant privacy/security policy permits it.

Failed-login logging must not echo the submitted username merely to make the event more descriptive. A failed credential attempt is useful as an event even when untrusted credential text is excluded.

## Authentication and authorization logging

Authentication and authorization events are security events under Wardveil Security, but their technical authority remains the server authentication/authorization implementation.

The production contract is:

- log a stable event category for success/failure/denial;
- never log submitted passwords or reusable authentication material;
- do not echo a submitted failed-login username;
- do not log remote address, forwarded-for data, or user agent by default solely for authentication telemetry;
- successful authentication may reference a trusted internal user identifier when the implementation has one and policy permits it;
- authorization denial should identify the denied capability or operation without leaking protected object contents;
- CSRF failures should identify the failure category without recording the CSRF value.

### Source authentication-hardening checkpoint

The maintained HTTP authentication boundary in `internal/server/server.go` now emits stable structured authentication event names without interpolating submitted usernames or remote addresses. Missing and invalid credentials are represented as bounded reason categories rather than copied caller-controlled identity data.

The short-term authentication optimization cookie is now explicitly `HttpOnly`, `Secure`, and `SameSite=Strict`. JWT validation is constrained to the intended HS256 algorithm, GoreeCloud Backup's inherited Kopia audience and issuer contract, normal temporal validity, and the authenticated subject. Regression tests in `internal/server/server_goreecloud_security_test.go` verify cookie attributes, subject binding, wrong issuer/audience rejection, and invalid-credential denial.

These source changes close the previously recorded authentication logging/cookie source blocker. They do not by themselves prove target-environment TLS termination, proxy behavior, authorization policy, log retention, or complete runtime security acceptance; those remain production-acceptance requirements.

## Error handling and retries

Errors must be visible and actionable. Components should distinguish transient errors from permanent or configuration errors when doing so is reliable.

Retries are appropriate only when:

- the operation is safe to retry;
- retrying will not create destructive duplicate behavior;
- retry attempts are bounded or use an explicit backoff policy;
- persistent failure becomes visible rather than looping silently;
- cancellation or shutdown can interrupt the retry path when appropriate.

User-facing failures must map to Glaze UI warning, degraded, denied-access, or error states instead of leaving controls indefinitely busy or silently ignoring the result.

## Monitoring and notification integration

GoreeCloud Backup should expose stable health/status information that can be consumed by GoreeCloud Monitor and should emit actionable failure events through GoreeCloud Notify when those integrations are implemented and accepted.

Monitoring and notification integrations must remain least privilege. A monitoring consumer should not need repository decryption material or backup contents merely to determine whether the service is healthy or a schedule has failed.

Security-relevant statuses may be presented as Wardveil Security information. The UI must identify the originating subsystem and must not transform a scanner pass, process-health pass, or successful snapshot into a blanket `Protected by Wardveil` or recoverability claim.

## Retention and access

Log retention must be deliberate and proportional to operational/security value. Production retention, access permissions, export, and deletion behavior must be documented for the target environment before Stable approval.

Logs containing security or administrative events require access appropriate to their sensitivity. Logs must not be made public or sent to an external telemetry provider by default.

## Validation requirements

Before Stable classification, representative validation must cover:

1. authentication success and failure without sensitive credential leakage;
2. authorization denial;
3. CSRF/request-integrity failure;
4. repository unavailable/degraded state;
5. snapshot success, failure, cancellation, and missed schedule where supported;
6. restore success and failure, with no restored content placed in logs;
7. notification failure and retry behavior;
8. monitoring health/degraded reporting;
9. Electron certificate/navigation/IPC rejection paths;
10. update-check and update-failure paths;
11. log access/retention configuration;
12. verification that logs do not contain passwords, tokens, private keys, raw cookies, protected content, or other prohibited material;
13. verification that new GoreeCloud-owned events are structured consistently and do not regress to unreviewed free-form sensitive-value interpolation.

## Release boundary

Passing source tests, Wardveil security validation, vulnerability scanning, or log-contract review is not sufficient for Stable classification. Production observability must be validated together with authentication/authorization, monitoring, notifications, repository operation, backup scheduling, representative restoration, and recovery evidence in the intended target environment.
