#!/usr/bin/env python3
"""Validate GoreeCloud Backup's mandatory repository product records."""

from __future__ import annotations

from pathlib import Path
import sys


ROOT = Path(__file__).resolve().parents[1]


class ValidationError(RuntimeError):
    """Raised when a product-record contract is violated."""


def read_required(path: str) -> str:
    target = ROOT / path
    if not target.is_file():
        raise ValidationError(f"required product record is missing: {path}")

    text = target.read_text(encoding="utf-8")
    if not text.strip():
        raise ValidationError(f"required product record is empty: {path}")

    return text


def require_fragments(path: str, text: str, fragments: tuple[str, ...]) -> None:
    for fragment in fragments:
        if fragment not in text:
            raise ValidationError(f"{path} is missing required contract text: {fragment!r}")


def validate() -> None:
    competitive = read_required("COMPETITIVE-OBJECTIVES.md")
    features = read_required("FEATURES.md")
    benefits = read_required("BENEFITS.md")
    readme = read_required("README.md")

    require_fragments(
        "COMPETITIVE-OBJECTIVES.md",
        competitive,
        (
            "## Primary competitors and benchmarks",
            "## Capabilities worth matching",
            "## Capabilities GoreeCloud intends to exceed",
            "## Capabilities GoreeCloud intentionally rejects",
            "## Privacy and security objectives",
            "## Ownership, self-hosting, and independence objectives",
            "## User-experience and accessibility objectives",
            "## Performance and reliability objectives",
            "## Interoperability and administrative-control objectives",
            "## Data portability and recovery objectives",
            "## GoreeCloud differentiators",
            "Competitive objectives are product-development targets.",
        ),
    )

    require_fragments(
        "FEATURES.md",
        features,
        (
            "## Current implemented source capabilities",
            "## Experimental or partial features",
            "## Planned features",
            "## Production-acceptance boundary",
            "not approved to replace the existing production Kopia deployment",
            "snapshot/recovery-point success alone producing `Protected` or `Restore Verified`",
        ),
    )

    require_fragments(
        "BENEFITS.md",
        benefits,
        (
            "## Current supportable benefits",
            "## Planned benefits not yet claimable as current",
            "not approved to replace the existing production Kopia deployment",
            "Benefits are limited to what the current source, architecture, and validated controls can support.",
        ),
    )

    require_fragments(
        "README.md",
        readme,
        (
            "[Competitive Objectives](COMPETITIVE-OBJECTIVES.md)",
            "[Features](FEATURES.md)",
            "[Benefits](BENEFITS.md)",
            "not Stable and is not approved to replace an existing production Kopia deployment",
        ),
    )


def main() -> int:
    try:
        validate()
    except (OSError, UnicodeError, ValidationError) as exc:
        print(f"GoreeCloud product-record validation failed: {exc}", file=sys.stderr)
        return 1

    print("GoreeCloud product-record validation passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
