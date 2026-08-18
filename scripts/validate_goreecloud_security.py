#!/usr/bin/env python3
"""Fail-closed GoreeCloud Backup security/source validation.

This validator intentionally focuses on GoreeCloud-owned changes instead of trying to
reinterpret the complete inherited Kopia history. It prevents new reusable-secret
material and required security-governance drift from entering the maintained fork and
keeps the Wardveil Security presentation contract attached to the actual controls.
"""

from __future__ import annotations

import pathlib
import re
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parents[1]
BASE_REF = "origin/master"

SENSITIVE_PATH_PATTERNS = (
    re.compile(r"(^|/)\.env(?:\..+)?$"),
    re.compile(r"(^|/)secrets/"),
    re.compile(r"\.key$", re.IGNORECASE),
    re.compile(r"\.pem$", re.IGNORECASE),
    re.compile(r"(^|/)(credentials|token)\.json$", re.IGNORECASE),
)

ALLOWED_SENSITIVE_PATHS = {".env.example", ".env.template"}
CONTENT_SCAN_EXCLUSIONS = {"scripts/validate_goreecloud_security.py"}

HIGH_CONFIDENCE_SECRET_PATTERNS = (
    ("private key block", re.compile(r"-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----")),
    ("GitHub token", re.compile(r"\bgh[pousr]_[A-Za-z0-9]{30,}\b")),
    ("GitHub fine-grained token", re.compile(r"\bgithub_pat_[A-Za-z0-9_]{30,}\b")),
    ("AWS access key", re.compile(r"\bAKIA[0-9A-Z]{16}\b")),
    ("Slack token", re.compile(r"\bxox[baprs]-[A-Za-z0-9-]{20,}\b")),
)

REQUIRED_GITIGNORE_LINES = {
    ".env", ".env.*", "!.env.example", "!.env.template", "secrets/", "*.key", "*.pem",
    "credentials.json", "token.json",
}
REQUIRED_SECURITY_TEXT = (
    "Reusable secrets must not be committed", "govulncheck", "go mod verify",
    "target-environment backup and representative restore evidence",
)
REQUIRED_UI_PRIVACY_TEXT = ("analytics", "tracking", "third-party fonts", "no remote ui dependencies")
REQUIRED_WARDVEIL_TEXT = (
    "wardveil security by goreecloud",
    "wardveil security presentation contract",
    "protected by wardveil",
    "does not replace the underlying backup engine",
    "must not imply that a successful snapshot",
)
REQUIRED_ELECTRON_SECURITY_TEXT = (
    "contextIsolation: true", "nodeIntegration: false", "sandbox: true", "webSecurity: true",
    "repositoryIDForSender", "requireTrustedRepositorySender", "setWindowOpenHandler", 'callback(false)',
)
REQUIRED_SHELL_SECURITY_TEXT = (
    'name=\"referrer\" content=\"no-referrer\"',
    'name=\"robots\" content=\"noindex,nofollow,noarchive\"',
    'name=\"goreecloud-security-identity\" content=\"Wardveil Security by GoreeCloud\"',
    'data-glaze-ui=\"1.0\"',
    'data-security-identity=\"wardveil\"',
    '<strong>Wardveil Security:</strong>',
)
REQUIRED_SECURITY_WORKFLOW_TEXT = (
    "Wardveil source security",
    "Record deterministic dependency evidence",
    "dependency-inputs.sha256",
    "Retain security evidence",
    "retention-days: 30",
)
REQUIRED_OBSERVABILITY_TEXT = (
    "Wardveil Security by GoreeCloud",
    "Structured logging direction",
    "Failed-login logging must not echo the submitted username",
    "Current inherited-source blocker",
    "Representative restoration remains the authoritative recovery test",
    "new GoreeCloud-owned events are structured consistently",
)
REQUIRED_READINESS_TEXT = (
    "not Stable and not approved for production replacement",
    "Evidence freshness and exact-candidate rule",
    "Known source-level authentication/logging gap",
    "GitHub Dependency Graph is disabled",
    "representative file restore",
    "Wardveil Security by GoreeCloud",
)


def run_git(*args: str) -> str:
    result = subprocess.run(["git", *args], cwd=ROOT, check=True, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    return result.stdout


def changed_files() -> list[str]:
    try:
        output = run_git("diff", "--name-only", "--diff-filter=ACMR", f"{BASE_REF}...HEAD")
    except subprocess.CalledProcessError as exc:
        print(f"security validation could not compare against {BASE_REF}: {exc.stderr.strip()}", file=sys.stderr)
        raise
    return [line.strip() for line in output.splitlines() if line.strip()]


def is_sensitive_path(path: str) -> bool:
    return path not in ALLOWED_SENSITIVE_PATHS and any(pattern.search(path) for pattern in SENSITIVE_PATH_PATTERNS)


def validate_paths(paths: list[str], failures: list[str]) -> None:
    for path in paths:
        if is_sensitive_path(path):
            failures.append(f"tracked GoreeCloud change uses secret-bearing path pattern: {path}")


def validate_content(paths: list[str], failures: list[str]) -> None:
    for path in paths:
        if path in CONTENT_SCAN_EXCLUSIONS:
            continue
        candidate = ROOT / path
        if not candidate.is_file():
            continue
        try:
            text = candidate.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        for label, pattern in HIGH_CONFIDENCE_SECRET_PATTERNS:
            if pattern.search(text):
                failures.append(f"possible {label} in changed file: {path}")


def validate_gitignore(failures: list[str]) -> None:
    lines = {line.strip() for line in (ROOT / ".gitignore").read_text(encoding="utf-8").splitlines() if line.strip() and not line.lstrip().startswith("#")}
    missing = sorted(REQUIRED_GITIGNORE_LINES - lines)
    if missing:
        failures.append(".gitignore is missing GoreeCloud secret exclusions: " + ", ".join(missing))


def validate_security_policy(failures: list[str]) -> None:
    security = (ROOT / "SECURITY.md").read_text(encoding="utf-8")
    for marker in REQUIRED_SECURITY_TEXT:
        if marker not in security:
            failures.append(f"SECURITY.md missing required contract marker: {marker!r}")


def validate_ui_security_contract(failures: list[str]) -> None:
    conformance = (ROOT / "docs/goreecloud/GLAZE_UI_CONFORMANCE.md").read_text(encoding="utf-8").lower()
    for marker in REQUIRED_UI_PRIVACY_TEXT:
        if marker not in conformance:
            failures.append(f"Glaze UI conformance record missing privacy marker: {marker!r}")
    for marker in REQUIRED_WARDVEIL_TEXT:
        if marker not in conformance:
            failures.append(f"Glaze UI conformance record missing Wardveil contract marker: {marker!r}")


def validate_operational_contracts(failures: list[str]) -> None:
    observability = (ROOT / "docs/goreecloud/OBSERVABILITY.md").read_text(encoding="utf-8")
    for marker in REQUIRED_OBSERVABILITY_TEXT:
        if marker not in observability:
            failures.append(f"Observability contract missing required marker: {marker!r}")

    readiness = (ROOT / "docs/goreecloud/PRODUCTION_READINESS.md").read_text(encoding="utf-8")
    for marker in REQUIRED_READINESS_TEXT:
        if marker not in readiness:
            failures.append(f"Production-readiness contract missing required marker: {marker!r}")


def validate_shell_security_contract(failures: list[str]) -> None:
    shell = (ROOT / "internal/server/htmlui_goreecloud.go").read_text(encoding="utf-8")
    for marker in REQUIRED_SHELL_SECURITY_TEXT:
        if marker not in shell:
            failures.append(f"GoreeCloud shell missing security/privacy marker: {marker!r}")
    if "Protected by Wardveil" in shell:
        failures.append("GoreeCloud shell must not claim 'Protected by Wardveil' without authoritative dynamic protection evidence")
    if "recoverable backup and restore" in shell.lower():
        failures.append("GoreeCloud shell must not claim recoverability before restore acceptance proves it")


def validate_electron_security_contract(failures: list[str]) -> None:
    electron = (ROOT / "app/public/electron.js").read_text(encoding="utf-8")
    for marker in REQUIRED_ELECTRON_SECURITY_TEXT:
        if marker not in electron:
            failures.append(f"Electron shell missing required security marker: {marker!r}")
    if 'app.name = "GoreeCloud Backup"' not in electron:
        failures.append("Electron shell must expose the GoreeCloud Backup product identity")


def validate_workflow_permissions(failures: list[str]) -> None:
    for relative in (".github/workflows/goreecloud-ui.yml", ".github/workflows/goreecloud-security.yml"):
        text = (ROOT / relative).read_text(encoding="utf-8")
        if "permissions:\n  contents: read" not in text:
            failures.append(f"{relative} must default to read-only repository contents")

    security_workflow = (ROOT / ".github/workflows/goreecloud-security.yml").read_text(encoding="utf-8")
    for marker in REQUIRED_SECURITY_WORKFLOW_TEXT:
        if marker not in security_workflow:
            failures.append(f"GoreeCloud Security workflow missing evidence marker: {marker!r}")


def main() -> int:
    failures: list[str] = []
    paths = changed_files()
    validate_paths(paths, failures)
    validate_content(paths, failures)
    validate_gitignore(failures)
    validate_security_policy(failures)
    validate_ui_security_contract(failures)
    validate_operational_contracts(failures)
    validate_shell_security_contract(failures)
    validate_electron_security_contract(failures)
    validate_workflow_permissions(failures)
    if failures:
        print("GoreeCloud security validation failed:", file=sys.stderr)
        for failure in failures:
            print(f"- {failure}", file=sys.stderr)
        return 1
    print(f"GoreeCloud security validation passed for {len(paths)} changed files with Wardveil Security contract enforced.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
