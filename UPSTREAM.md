# GoreeCloud Backup Upstream Record

## Upstream project

- Project: Kopia
- Upstream repository: `kopia/kopia`
- GoreeCloud repository: `GoreeCloud/goreecloud-backup`
- Upstream license: Apache License 2.0
- Initial GoreeCloud fork branch baseline: `master`
- Recorded baseline commit: `5f9fe81ec47c6dd524af5abe37803566ff640fb8`
- Baseline recorded: 2026-08-18

The recorded commit is the fork state inspected when GoreeCloud-controlled development began. Later upstream synchronization must be documented separately rather than silently changing this baseline record.

## Maintained-fork model

GoreeCloud Backup is a maintained derivative of Kopia, not an attempt to erase Kopia's provenance. The project initially preserves the mature backup engine and repository behavior while GoreeCloud develops its own product identity, management layers, verification model, integrations, and user experience.

The amount of upstream code replaced is not a success metric. Recovery correctness, maintainability, security, privacy, portability, and independently controlled operation are the governing measures.

## Upstream synchronization rules

Upstream changes must be reviewed before integration. Reviews should consider at least:

- repository-format compatibility;
- encryption and key-management behavior;
- snapshot creation and restoration behavior;
- storage backend changes;
- retention and maintenance behavior;
- data-integrity and corruption-handling changes;
- CLI, API, server, and UI compatibility;
- dependency and vulnerability changes;
- build and release workflow changes;
- licensing, notice, attribution, and trademark implications;
- any reintroduction of upstream branding into GoreeCloud-controlled product surfaces.

Upstream commits must not be merged solely because they are newer.

## Recovery compatibility principle

Where practical, GoreeCloud Backup should retain compatibility with standard Kopia repositories during the early maintained-fork phases. Compatibility provides an independent recovery path if a GoreeCloud-specific UI or management layer becomes unavailable.

Any future change that intentionally breaks repository compatibility requires an explicit migration design, rollback plan, representative restore validation, and preserved access to required historical recovery points.

## Production boundary

The existence of this fork does not authorize replacement of an existing production Kopia installation or repository. Production migration requires a separate controlled acceptance process with successful restoration evidence.
