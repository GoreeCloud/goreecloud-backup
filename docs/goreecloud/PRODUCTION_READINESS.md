# GoreeCloud Backup — Production Readiness Evidence

## Current classification

**Active development — not Stable and not approved for production replacement of an existing Kopia deployment.**

This record separates source-level evidence from target-environment acceptance. It is intentionally fail-closed: an unknown, unavailable, skipped, stale, cancelled by a newer head, or unverified required gate is not equivalent to a pass.

## Governing principle

GoreeCloud Backup is recovery-critical software. A successful build, a green security scan, a completed snapshot, or a polished Glaze UI does not prove that protected information can be recovered.

Stable classification requires both:

1. source/release evidence showing that the candidate is maintainable, secure, testable, compatible, and intentionally packaged; and
2. target-environment evidence showing that the actual deployment can back up, retain, monitor, notify, and restore representative protected information safely.

Wardveil Security by GoreeCloud presents security/protection posture. It does not replace the technical evidence below and must not turn an incomplete gate set into a generalized protection claim.

## Evidence freshness and exact-candidate rule

Release evidence belongs to the exact candidate commit that produced it. A passing workflow on an older commit is useful historical evidence but does not automatically validate a newer source head.

When a newer commit supersedes a running workflow, cancellation is not a source failure, but the cancelled run is also not a passing result for the newer candidate. Documentation-only changes still require the source/governance gates whose path filters cover those files.

The release record must distinguish:

- **passed** — the required gate evaluated the exact candidate and succeeded;
- **failed** — the exact-candidate gate evaluated and reported a blocking result;
- **blocked/unavailable** — infrastructure, repository configuration, permissions, or another external condition prevented evaluation;
- **pending** — evaluation has started or is queued but has not completed;
- **not applicable** — the requirement is genuinely outside the candidate's supported scope and the reason is documented.

`Not applicable` must not be used as a shortcut for difficult, unavailable, or inconvenient validation.

## Source and repository gates

The following source gates are required for a release candidate:

- exact upstream baseline and maintained-fork provenance recorded;
- Apache-2.0 and applicable third-party attribution preserved;
- controlled branch/PR review path rather than unreviewed production mutation;
- Go module verification;
- Go vulnerability analysis;
- production Electron dependency audit;
- deterministic dependency-input evidence;
- inherited unit/integration tests relevant to the candidate;
- lint and formatting gates;
- race detector where supported by the inherited project;
- compatibility tests against established Kopia behavior;
- build/package validation on supported platforms;
- HTML UI tests;
- GoreeCloud Glaze UI source-conformance validation;
- GoreeCloud/Wardveil security-source validation;
- Dependency Review when GitHub Dependency Graph is available;
- privacy-safe authentication logging and explicit authentication-cookie attribute tests;
- no unresolved release-blocking vulnerability unless an approved, documented exception exists;
- no known reusable secret in GoreeCloud-owned source changes;
- no unreviewed repository-format, encryption, key-derivation, deduplication, content-addressing, retention-deletion, garbage-collection, or restore-semantics change.

## User-experience gates

Stable visual/interaction acceptance requires representative runtime review of:

- Glaze UI light and dark appearance;
- Compact, Medium, Expanded, and Wide adaptive ranges;
- keyboard navigation and focus visibility;
- reduced motion;
- reduced transparency and no-backdrop-filter fallbacks;
- increased contrast and forced colors;
- loading and long-running operation states;
- success states;
- empty states;
- warning and degraded states;
- denied-access states;
- validation errors and server errors;
- destructive confirmation and recovery-sensitive actions;
- dialogs, menus, notifications, navigation, forms, tables, technical output, and settings;
- Wardveil security status/alert presentation without overstating the underlying evidence;
- unique GoreeCloud Backup application artwork across favicon, desktop/launcher, packaging, and other supported product surfaces;
- absence of material inherited upstream presentation that conflicts with the approved GoreeCloud product identity unless explicitly documented as a compatibility exception.

## Authentication, authorization, and privacy gates

Before Stable approval, the target candidate must prove:

- intended authentication works;
- invalid credentials are denied;
- least-privilege authorization works for each supported administrative/read-only boundary;
- privileged API/control paths reject unauthorized identities;
- CSRF/request-integrity protections behave as intended;
- short-term authentication cookies have a reviewed `HttpOnly`, `Secure`, and intentional `SameSite` policy appropriate to the deployment;
- security logging does not expose submitted failed-login usernames, passwords, tokens, raw cookies, authorization headers, or other prohibited sensitive data;
- sensitive paths, filenames, repository locations, IP addresses, and user-agent details are minimized according to the observability contract;
- no remote analytics, advertising, tracking, third-party fonts, or unapproved remote UI runtime is required by the GoreeCloud-owned Glaze layer.

### Known source-level authentication/logging gap

The inherited HTTP authentication implementation currently records remote address and submitted username for a failed login and may record remote address plus username for successful login when request logging is enabled. The authentication cookie construction also has not yet been reconciled with the GoreeCloud explicit cookie-attribute release contract. These items remain release blockers and must be changed and regression-tested before Stable classification.

## Wardveil Security gates

Security-facing interfaces must:

- identify **Wardveil Security by GoreeCloud** where platform security identity is useful;
- preserve the identity of the underlying technical authority producing the state;
- use Glaze UI presentation and accessibility conventions;
- distinguish informational, protected-control-active, warning, degraded, denied, and error states;
- avoid using `Protected by Wardveil` unless the displayed scope has authoritative evidence supporting that statement;
- never equate vulnerability-scan success, process health, successful login, or successful snapshot creation with verified recoverability;
- make security failures actionable without exposing credentials or protected backup contents.

## Observability gates

The candidate must satisfy `docs/goreecloud/OBSERVABILITY.md`, including representative evidence for:

- process lifecycle;
- repository availability/degradation;
- snapshot lifecycle;
- restore lifecycle;
- authentication and authorization events;
- Wardveil security events;
- update failures;
- integration failures;
- monitoring and notification delivery/failure;
- privacy-safe structured context;
- explicit retention and access rules for target logs.

## Integration gates

Integrations are accepted only when they are implemented as explicit, documented, least-privilege contracts. Planned targets include GoreeCloud Manager, GoreeCloud Monitor, GoreeCloud Notify, GoreeCloud Identity where appropriate, GoreeCloud-controlled API/CLI surfaces, and repository/storage backends.

A future integration is not considered complete merely because an upstream Kopia endpoint exists. The GoreeCloud ownership, authentication/authorization, failure behavior, compatibility, privacy, and monitoring boundaries must be documented and tested.

## Recovery and target-environment gates

No source-only workflow can satisfy these gates. Before production cutover, the intended environment must demonstrate:

1. approved repository destination and independent failure domain;
2. protected source scope and exclusions;
3. application-consistent handling for databases/stateful applications where required;
4. successful scheduled backups over multiple recovery points;
5. retention behavior;
6. repository maintenance/integrity verification;
7. capacity monitoring;
8. missed-backup/failure monitoring;
9. actionable notification delivery;
10. recoverable and appropriately protected repository credentials;
11. representative file restore;
12. representative application/dataset restore where applicable;
13. ownership/permission validation after restoration;
14. rollback/recovery path for the GoreeCloud Backup application itself;
15. documented recovery evidence sufficient to repeat the operation;
16. preservation of previous production Kopia recovery points until the replacement path is proven.

## Platform packaging gates

Supported packaged applications must be tested in representative environments. The current intended desktop validation matrix includes Linux, Windows, and macOS where the inherited packaging supports them.

Packaging validation should include:

- clean installation;
- first launch;
- existing configuration migration/compatibility where applicable;
- repository connection;
- renderer sandbox/security behavior;
- certificate trust boundary;
- directory-selection IPC;
- external navigation behavior;
- update-check/update-failure behavior;
- shutdown/restart;
- uninstall or rollback behavior;
- unique product identity and artwork;
- no unexpected outbound UI dependency.

## External repository-setting gate

GitHub Dependency Review remains required by the GoreeCloud security-maintenance model where technically available. At the current checkpoint, the inherited Dependency Review workflow cannot complete because GitHub Dependency Graph is disabled for this fork. The workflow must remain present; the configuration issue must be resolved through an authorized repository-settings path rather than by deleting or weakening the gate.

## Release decision

The candidate may be described as source-hardened or source-validated only when the applicable source workflows pass on the exact candidate commit.

It may be classified **Stable** only after all release-blocking source gates and applicable target-environment gates above have evidence, no known critical security/stability/usability gap remains, and any accepted exception is explicit, bounded, reviewed, and recorded.

Until then, the pull request remains a development candidate and production Kopia remains the authoritative recovery path.
