#!/usr/bin/env python3
"""Validate GoreeCloud Backup cross-platform icon and product-identity wiring.

The normal source gate permits the canonical artwork to remain absent while the product is
in active development, but makes that state explicit. Once the canonical source appears,
all required web/desktop derivatives become mandatory in the same change so a partially
rebranded product cannot silently pass.

Use --release to require approved canonical artwork and all currently supported derivatives.
"""

from __future__ import annotations

import argparse
import json
import struct
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CANONICAL = ROOT / "branding/goreecloud-backup-icon.svg"
MANIFEST = ROOT / "app/public/manifest.json"
INDEX = ROOT / "app/public/index.html"
ICON_CONTRACT = ROOT / "docs/goreecloud/ICON_ASSETS.md"

WEB_PNGS = {
    ROOT / "app/public/icon-16.png": (16, 16),
    ROOT / "app/public/icon-32.png": (32, 32),
    ROOT / "app/public/icon-48.png": (48, 48),
    ROOT / "app/public/logo192.png": (192, 192),
    ROOT / "app/public/logo512.png": (512, 512),
    ROOT / "app/public/icon-maskable-512.png": (512, 512),
}

DESKTOP_REQUIRED = [
    ROOT / "app/assets/icon.png",
    ROOT / "app/assets/icon.icns",
    ROOT / "app/assets/icon.ico",
]

INHERITED_COMPATIBILITY_ASSETS = [
    ROOT / "app/public/favicon.ico",
    ROOT / "app/public/logo192.png",
    ROOT / "app/public/logo512.png",
    ROOT / "app/assets/icon.icns",
    ROOT / "app/assets/icon.ico",
]


def fail(message: str) -> None:
    print(f"ERROR: {message}", file=sys.stderr)
    raise SystemExit(1)


def require_text(path: Path, snippets: list[str]) -> str:
    if not path.is_file():
        fail(f"required file is missing: {path.relative_to(ROOT)}")

    text = path.read_text(encoding="utf-8")
    for snippet in snippets:
        if snippet not in text:
            fail(f"{path.relative_to(ROOT)} is missing required marker: {snippet!r}")
    return text


def png_dimensions(path: Path) -> tuple[int, int]:
    with path.open("rb") as stream:
        header = stream.read(24)

    if len(header) < 24 or header[:8] != b"\x89PNG\r\n\x1a\n" or header[12:16] != b"IHDR":
        fail(f"not a valid PNG with an IHDR header: {path.relative_to(ROOT)}")

    return struct.unpack(">II", header[16:24])


def validate_manifest() -> None:
    if not MANIFEST.is_file():
        fail("app/public/manifest.json is missing")

    data = json.loads(MANIFEST.read_text(encoding="utf-8"))

    if data.get("name") != "GoreeCloud Backup":
        fail("manifest name must be 'GoreeCloud Backup'")
    if data.get("short_name") != "Backup":
        fail("manifest short_name must be 'Backup'")

    serialized = json.dumps(data)
    for forbidden in ("Create React App", "React App", "KopiaUI"):
        if forbidden in serialized:
            fail(f"manifest still contains inherited/default product identity: {forbidden}")

    icons = {(entry.get("src"), entry.get("sizes")) for entry in data.get("icons", [])}
    required = {
        ("favicon.ico", "16x16 24x24 32x32 48x48 64x64"),
        ("logo192.png", "192x192"),
        ("logo512.png", "512x512"),
    }
    missing = required - icons
    if missing:
        fail(f"manifest is missing required current icon declarations: {sorted(missing)}")


def validate_shell_identity() -> None:
    text = require_text(
        INDEX,
        [
            'name="application-name" content="GoreeCloud Backup"',
            "GoreeCloud Backup - %REACT_APP_SHORT_VERSION_INFO%",
            'name="referrer" content="no-referrer"',
            'name="robots" content="noindex,nofollow,noarchive"',
        ],
    )

    for forbidden in ("Web site created using create-react-app", "Create React App Sample"):
        if forbidden in text:
            fail(f"editable frontend shell still contains default framework identity: {forbidden}")


def validate_contract() -> None:
    require_text(
        ICON_CONTRACT,
        [
            "GoreeCloud Backup — Canonical Icon and Asset Contract",
            "branding/goreecloud-backup-icon.svg",
            "Wardveil Security by GoreeCloud",
            "app/assets/icon.png",
            "Android APK/AAB outputs",
        ],
    )


def validate_complete_asset_set() -> None:
    if not CANONICAL.is_file():
        fail("release requires branding/goreecloud-backup-icon.svg")

    canonical = CANONICAL.read_text(encoding="utf-8")
    if "<svg" not in canonical or "viewBox" not in canonical:
        fail("canonical icon source must be an SVG with a viewBox")

    lowered = canonical.lower()
    for forbidden in ("http://", "https://", "<script", "@font-face"):
        if forbidden in lowered:
            fail(f"canonical SVG contains forbidden remote/script/font marker: {forbidden}")

    for path, expected in WEB_PNGS.items():
        if not path.is_file():
            fail(f"canonical artwork exists but required derivative is missing: {path.relative_to(ROOT)}")
        actual = png_dimensions(path)
        if actual != expected:
            fail(
                f"{path.relative_to(ROOT)} has dimensions {actual[0]}x{actual[1]}, "
                f"expected {expected[0]}x{expected[1]}"
            )

    for path in DESKTOP_REQUIRED:
        if not path.is_file():
            fail(f"canonical artwork exists but desktop derivative is missing: {path.relative_to(ROOT)}")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--release",
        action="store_true",
        help="require the canonical artwork and complete currently supported derivative set",
    )
    args = parser.parse_args()

    validate_contract()
    validate_manifest()
    validate_shell_identity()

    if CANONICAL.is_file():
        validate_complete_asset_set()
        print("GoreeCloud Backup canonical icon source and required derivatives are present.")
        return

    if args.release:
        fail("canonical GoreeCloud Backup artwork is not supplied; release icon gate cannot pass")

    missing_compatibility = [p for p in INHERITED_COMPATIBILITY_ASSETS if not p.is_file()]
    if missing_compatibility:
        fail(
            "canonical artwork is pending and inherited compatibility assets unexpectedly disappeared: "
            + ", ".join(str(p.relative_to(ROOT)) for p in missing_compatibility)
        )

    print(
        "GoreeCloud Backup icon integration contract is valid. "
        "Canonical artwork remains an explicit release blocker; run with --release for the Stable gate."
    )


if __name__ == "__main__":
    main()
