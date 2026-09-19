#!/usr/bin/env python3
import argparse
import json
import os
import re
import sys

TARGET_RE = re.compile(r"^/source/[A-Za-z0-9][A-Za-z0-9._-]*$")


def fail(message):
    print(f"source-scope validation failed: {message}", file=sys.stderr)
    raise SystemExit(1)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--require-host-readable",
        action="store_true",
        help="require every host bind source to exist and be readable/traversable",
    )
    args = parser.parse_args()

    try:
        config = json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        fail(f"invalid Compose JSON: {exc}")

    service = config.get("services", {}).get("backup")
    if not isinstance(service, dict):
        fail("Compose service 'backup' is missing")

    source_binds = []
    seen_targets = set()

    for volume in service.get("volumes", []):
        if not isinstance(volume, dict):
            continue

        target = volume.get("target")
        if not isinstance(target, str) or not (
            target == "/source" or target.startswith("/source/")
        ):
            continue

        if volume.get("type") != "bind":
            fail(f"{target} must be a bind mount")

        if not TARGET_RE.fullmatch(target):
            fail(
                f"{target!r} must use the explicit /source/<simple-name> form"
            )

        if target in seen_targets:
            fail(f"duplicate backup target {target}")

        source = volume.get("source")
        if not isinstance(source, str) or not os.path.isabs(source):
            fail(f"{target} has a missing or non-absolute host source")

        if "REPLACE_WITH" in source or "REPLACE_WITH" in target:
            fail(f"{target} still contains a deployment placeholder")

        if os.path.normpath(source) == os.path.sep:
            fail("host root '/' is not an approved explicit backup source")

        if volume.get("read_only") is not True:
            fail(f"{target} must be read-only")

        if args.require_host_readable:
            if not os.path.exists(source):
                fail(f"host source for {target} does not exist")
            required_mode = os.R_OK | (os.X_OK if os.path.isdir(source) else 0)
            if not os.access(source, required_mode):
                fail(f"host source for {target} is not readable/traversable")

        seen_targets.add(target)
        source_binds.append((source, target))

    if not source_binds:
        fail("no explicit read-only /source/<name> bind mounts are configured")

    for _, target in source_binds:
        print(target)


if __name__ == "__main__":
    main()
