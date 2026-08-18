#!/usr/bin/env python3
"""Fail-closed GoreeCloud Backup Glaze UI conformance validation.

This validator checks the application-owned presentation boundary only. It intentionally does
not validate Kopia repository formats, encryption, snapshot semantics, retention, restore, or
storage-provider behavior.
"""

from __future__ import annotations

from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
UI_ROOT = ROOT / "internal" / "server" / "goreecloud-ui"
CANONICAL_GLAZE_REVISION = "d6e446fd8ef251259d16368d50aad90d9287a774"


def read(relative: str) -> str:
    path = ROOT / relative
    if not path.is_file():
        raise AssertionError(f"missing required file: {relative}")
    return path.read_text(encoding="utf-8")


def require_markers(text: str, markers: list[str], source: str) -> None:
    missing = [marker for marker in markers if marker not in text]
    if missing:
        raise AssertionError(f"{source} missing required markers: {', '.join(missing)}")


def validate_canonical_foundation() -> None:
    glaze = read("internal/server/goreecloud-ui/glaze.css")
    accessibility = read("internal/server/goreecloud-ui/glaze.accessibility.css")
    provenance = read("internal/server/goreecloud-ui/README.md")

    require_markers(
        glaze,
        [
            "--glaze-canvas:",
            "--glaze-surface-strong:",
            "--glaze-accent:",
            "--glaze-danger:",
            "--glaze-target-min: 44px",
            "--glaze-motion-instant: 90ms",
            "--glaze-motion-fast: 160ms",
            "--glaze-motion-standard: 220ms",
            "--glaze-motion-emphasized: 320ms",
            "--glaze-radius-control:",
            "--glaze-shadow-overlay:",
            ".glaze-surface",
            ".glaze-overlay",
        ],
        "canonical Glaze CSS",
    )

    require_markers(
        accessibility,
        [
            "prefers-reduced-motion: reduce",
            "prefers-contrast: more",
            "forced-colors: active",
            "backdrop-filter",
        ],
        "canonical Glaze accessibility CSS",
    )

    require_markers(
        provenance,
        [CANONICAL_GLAZE_REVISION, "Glaze UI", "MIT"],
        "Glaze provenance record",
    )


def validate_product_adapter() -> None:
    adapter = read("internal/server/goreecloud-ui/backup.css")

    require_markers(
        adapter,
        [
            "body.goreecloud-backup",
            ".btn-primary",
            ".form-control",
            ".form-check-input",
            ".nav-tabs",
            ".dropdown-item",
            ".table-responsive",
            ".alert-success",
            ".alert-warning",
            ".alert-danger",
            ".progress-bar",
            "prefers-reduced-motion: reduce",
            "prefers-reduced-transparency: reduce",
            "prefers-contrast: more",
            "forced-colors: active",
            "max-width: 599px",
            "min-width: 600px",
            "min-width: 1024px",
            "min-width: 1440px",
            "pointer: coarse",
            "@media print",
        ],
        "GoreeCloud Backup adapter",
    )

    remote_asset_pattern = re.compile(r"(?:url|@import)\s*\(?'?\"?https?://", re.IGNORECASE)
    for css_path in sorted(UI_ROOT.glob("*.css")):
        css = css_path.read_text(encoding="utf-8")
        if remote_asset_pattern.search(css):
            raise AssertionError(f"remote UI dependency found in {css_path.relative_to(ROOT)}")


def validate_server_integration() -> None:
    integration = read("internal/server/htmlui_goreecloud.go")
    tests = read("internal/server/htmlui_goreecloud_test.go")

    require_markers(
        integration,
        [
            "GoreeCloud Backup",
            "goreecloud-backup-ui",
            "/goreecloud-ui/glaze.css",
            "/goreecloud-ui/glaze.accessibility.css",
            "/goreecloud-ui/backup.css",
            'class=\"glaze-canvas goreecloud-backup\"',
        ],
        "server presentation integration",
    )

    require_markers(
        tests,
        [
            "TestGoreeCloudAssetFileTransformsIndex",
            "TestGoreeCloudAssetFileDelegatesUpstreamAssets",
            "TestGoreeCloudAssetFileServesLocalGlazeAssets",
            "TestApplyGoreeCloudHTMLIdentityIsIdempotent",
            "TestGoreeCloudAssetFilePreservesMissingAssetError",
        ],
        "server presentation tests",
    )


def validate_documented_conformance() -> None:
    conformance = read("docs/goreecloud/GLAZE_UI_CONFORMANCE.md")
    require_markers(
        conformance,
        [
            "Glaze UI 1.0",
            CANONICAL_GLAZE_REVISION,
            "Canvas, Solid, Raised, Glaze, and Overlay",
            "Compact",
            "Medium",
            "Expanded",
            "Wide",
            "Reduced motion",
            "Reduced transparency",
            "Increased contrast",
            "Forced colors",
            "No remote UI dependencies",
            "Visual acceptance pending",
        ],
        "Glaze UI conformance record",
    )


def main() -> int:
    checks = [
        validate_canonical_foundation,
        validate_product_adapter,
        validate_server_integration,
        validate_documented_conformance,
    ]

    try:
        for check in checks:
            check()
    except AssertionError as exc:
        print(f"GoreeCloud UI validation failed: {exc}", file=sys.stderr)
        return 1

    print("GoreeCloud Backup Glaze UI source conformance checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
