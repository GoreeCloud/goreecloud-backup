# GoreeCloud Backup — Feature Roadmap

**Status:** Active roadmap control  
**As of:** 2026-09-17  
**Authoritative project record:** Project Specification — Backup  
**Canonical repository:** GoreeCloud/goreecloud-backup  
**Drive control:** `GoreeCloud/Feature Roadmap/GoreeCloud Backup/FEATURE-ROADMAP.docx`

## Purpose

This file is the repository-side feature roadmap control for GoreeCloud Backup. It records current planned and recommended feature work without replacing the authoritative project record, implementation evidence, release gates, or GoreeCloud Tasks Management.

## Roadmap

| ID | Feature / obligation | Priority | Current state |
| --- | --- | --- | --- |
| FR-001 | Reconcile and maintain every current planned or recommended GoreeCloud Backup feature from the authoritative project record and verified repository evidence in this roadmap. | High | Ongoing control |
| FR-002 | Move actionable feature obligations into GoreeCloud Tasks Management when required, preserving priority, dependency, and lifecycle disposition. | High | Ongoing control |
| FR-003 | Do not mark features implemented, complete, cancelled, or superseded without authoritative evidence and synchronized repository/Drive roadmap updates. | High | Ongoing control |
| FR-004 | Reconcile the active maintained-fork Development line to Platform Contract 0.4, exactly nine Integral Platform Systems, and the current Glaze UI 1.5.1 consumer target without changing recovery-critical repository, encryption, snapshot, retention, maintenance, provider-write, or restore semantics. | High | In Progress |

## Current verified Development checkpoint

The substantive maintained-fork Development line remains Draft PR #1 (`agent/goreecloud-foundation`) at exact head `54bf969d3a761a5ecb7c5d247329d008cef2efdd`. That line contains the GoreeCloud recovery-assurance domain and bounded Backup ↔ Sync source integration foundations, but it is not globally green and is not integrated into `master`.

A bounded governance migration is isolated in stacked Draft PR #4 (`governance/platform-contract-0.4-nine-systems`) at exact head `de2dc0af4e07b5503a903343acdccb345401d292`. It adds the Contract 0.4 manifest and exact-head reusable validator, explicitly evaluates Manager, Privacy Shield, Wardveil Security, Everkeep, Glaze UI, Mesh, Identity, Policy, and Observability, and keeps GoreeCloud Sync separately governed.

The implemented owned presentation layer remains Glaze UI 1.0 at canonical source revision `d6e446fd8ef251259d16368d50aad90d9287a774`. Current mandatory Stable consumer target is Glaze UI 1.5.1 at exact Stable revision `98da57064ede0f334627b632bc16801f580331af`. PR #4 therefore records Glaze as migration-required and does not transfer V1.0 source-conformance evidence to V1.5.1.

Exact-head Platform Contract push run `35288304796` passed on `de2dc0af4e07b5503a903343acdccb345401d292`, including exact caller revision verification, Contract 0.4 manifest validation, computed conformance evaluation, result-schema/provenance validation, and evidence upload.

The broader maintained-fork line remains Development-only and nonconformant. Previously verified exact-head Build, Lint, Tests, Dependency Review, License Check, and GoreeCloud Security failures remain unresolved release blockers. Passing product-specific Protection, Product Records, UI, Compatibility, Coverage, Race Detector, HTMLUI, and Volume Shadow Copy checks do not waive those failures.

PR #4 remains Draft and stacked on PR #1. Neither candidate establishes integration into authoritative `master`, production repository cutover, runtime Platform-System acceptance, target-environment restore acceptance, protected release/signing, deployment, Release Candidate qualification, or Stable qualification. Existing production recovery points remain authoritative until a separately accepted replacement path proves recovery.

## Maintenance and synchronization

This roadmap and the corresponding Drive `FEATURE-ROADMAP.docx` must remain materially synchronized with one another and with the authoritative project or service record. Update both copies whenever feature scope, priority, dependency, implementation status, cancellation, supersession, recommendation, or verification state materially changes.

No feature may be represented as complete or Stable solely because it appears in this roadmap. Completion and lifecycle claims require the applicable authoritative implementation, validation, review, release, and production evidence.

## Reconciliation rule

At each material feature change, reconcile this roadmap against the current authoritative project record, repository implementation state, applicable platform-system requirements, and GoreeCloud Tasks Management. Missing obligations, stale status, duplicated work, roadmap drift, or undocumented disposition changes are defects to correct.
