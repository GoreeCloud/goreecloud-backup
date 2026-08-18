# GoreeCloud Backup Security

## Security posture

GoreeCloud Backup is a recovery-critical maintained fork. Security work must preserve recoverability, repository compatibility, privacy, and controlled change. A build that starts successfully is not sufficient evidence that a security change is safe.

**Wardveil Security by GoreeCloud** is the platform-wide security and protection identity used when GoreeCloud Backup presents security posture, safeguards, alerts, or protection controls. Wardveil is an identity and presentation layer; it does not replace the underlying backup engine, repository security model, authentication/authorization implementation, policies, monitoring, or recovery evidence that produces the technical state.

The initial engine remains Kopia-derived. GoreeCloud-specific security changes should stay isolated from repository format, encryption, content addressing, deduplication, snapshot serialization, retention deletion, garbage collection, storage-write, and restore semantics unless a separately reviewed recovery-critical change is explicitly required.

## Reporting security issues

Do not publish reusable credentials, private repository information, private keys, recovery codes, tokens, backup passwords, or sensitive recovery data in a public issue, pull request, log, screenshot, or discussion.

Security reports should describe the affected component, observed behavior, practical impact, reproduction conditions using synthetic or disposable data where possible, and any known affected version or commit. Active secret values must be omitted or redacted.

## Dependency and vulnerability maintenance

GoreeCloud Backup uses layered security-maintenance controls:

- Go vulnerability analysis through `govulncheck`;
- module-integrity verification through `go mod verify`;
- deterministic toolchain/module/dependency-input evidence retained by the GoreeCloud Security workflow;
- inherited lint, test, compatibility, coverage, race, and platform workflows;
- GoreeCloud-owned source/security validation for newly introduced secret material and security-policy drift;
- package-lock auditing for release-blocking production dependency findings where the npm registry is available;
- GitHub Dependency Review when the repository Dependency Graph is enabled.

A scanner result is evidence, not the whole security decision. Findings must be reviewed for reachability, exploitability, affected deployment paths, fixed versions, rollback compatibility, and recovery impact.

## Secret handling

Reusable secrets must not be committed to this repository. This includes passwords, private keys, API tokens, OAuth client secrets, database passwords, webhook tokens, session-signing secrets, encryption keys, recovery codes, and backup passwords.

Use placeholders in examples and documentation. Sensitive runtime values belong in an approved secret store or a protected service-specific file outside ordinary source control. Secret-bearing files must use least-privilege ownership and permissions and should be mounted read-only where modification is unnecessary.

If a reusable secret is committed, treat it as exposed even if the repository is private or the commit is later deleted. Revoke or rotate it, update the consuming service, validate the replacement, review relevant history/logs, and record the incident without reproducing the active value.

## Logging and observability

Security and operational logging must follow `docs/goreecloud/OBSERVABILITY.md`.

In particular, routine logs must not contain passwords, backup passwords, API keys, access/refresh tokens, authorization headers, raw cookies, private keys, encryption keys, recovery codes, protected backup contents, or other reusable secret material.

Authentication logging must not echo submitted failed-login usernames merely for diagnostic convenience. Remote addresses, forwarded-for values, user agents, filesystem paths, filenames, repository locations, and other potentially sensitive identifiers must be minimized and included only when a documented operational need and applicable policy permit them.

Wardveil security events must preserve the identity of the underlying component/control producing the event. A Wardveil label must never turn a scanner pass, process-health pass, successful login, or successful snapshot into a generalized recoverability claim.

## Known release blockers

The current inherited HTTP authentication path in `internal/server/server.go` still records the remote address and submitted username on failed login and may log remote address plus username on successful login when request logging is enabled. That behavior does not satisfy the GoreeCloud observability/privacy contract and must be reconciled and regression-tested before Stable classification.

The same inherited authentication path also requires explicit release evidence for the short-term authentication cookie's `HttpOnly`, `Secure`, and intentional `SameSite` behavior. These attributes must be reviewed and tested rather than inferred from transport topology.

GitHub Dependency Review is also required when technically available. The workflow is intentionally retained while Dependency Graph remains disabled for this fork; the GoreeCloud Security workflow supplements but does not silently replace that repository-native control.

## Security acceptance

Before a GoreeCloud Backup release can be classified as stable, applicable security evidence must include:

- supported software/dependency baselines;
- passing release-relevant source, lint, test, compatibility, and vulnerability gates;
- no unresolved release-blocking vulnerability without a documented exception;
- no known reusable secret in GoreeCloud-owned changes;
- reviewed Electron/package dependencies and lockfile integrity;
- privacy-safe Glaze UI dependencies with no analytics, trackers, advertising technology, remote fonts, or unapproved remote UI runtime;
- validated authentication and authorization behavior for the target deployment;
- validated denied-access behavior where applicable;
- validated privacy-safe authentication/security logging;
- reviewed and tested authentication cookie attributes;
- a viable rollback or recovery path for material updates;
- target-environment backup and representative restore evidence.

Glaze UI conformance, Wardveil presentation, source validation, or a successful build must never be treated as proof of recoverability.

## Security exceptions

A security exception must identify the affected component, finding or deferred update, technical reason, exposure and exploitability assessment, privacy/security impact, compensating controls, backup/recovery state, planned resolution, review condition, and responsible administrator. Convenience alone is not an acceptable exception reason.
