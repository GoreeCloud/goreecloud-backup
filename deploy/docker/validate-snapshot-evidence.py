#!/usr/bin/env python3
import argparse
import json
import re
import sys

SNAPSHOT_ID_RE = re.compile(r"^[0-9a-f]{32}$")
ROOT_OBJECT_ID_RE = re.compile(r"^[A-Za-z][A-Za-z0-9]+$")


def fail(message):
    print(f"snapshot evidence validation failed: {message}", file=sys.stderr)
    raise SystemExit(1)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--expected-source", required=True)
    args = parser.parse_args()

    try:
        manifest = json.load(sys.stdin)
    except json.JSONDecodeError as exc:
        fail(f"invalid snapshot JSON: {exc}")

    snapshot_id = manifest.get("id")
    if not isinstance(snapshot_id, str) or not SNAPSHOT_ID_RE.fullmatch(snapshot_id):
        fail("snapshot manifest ID is missing or malformed")

    root_entry = manifest.get("rootEntry")
    if not isinstance(root_entry, dict):
        fail("snapshot rootEntry is missing")

    root_object_id = root_entry.get("obj")
    if not isinstance(root_object_id, str) or not ROOT_OBJECT_ID_RE.fullmatch(root_object_id):
        fail("snapshot root object ID is missing or malformed")

    source = manifest.get("source")
    if not isinstance(source, dict):
        fail("snapshot source metadata is missing")

    actual_source = (
        f"{source.get('userName', '')}@{source.get('host', '')}:"
        f"{source.get('path', '')}"
    )
    if actual_source != args.expected_source:
        fail(
            f"snapshot source {actual_source!r} does not match "
            f"{args.expected_source!r}"
        )

    if manifest.get("incomplete") not in (None, ""):
        fail("snapshot is marked incomplete")

    stats = manifest.get("stats")
    if not isinstance(stats, dict):
        fail("snapshot statistics are missing")

    file_count = int(stats.get("fileCount", 0))
    dir_count = int(stats.get("dirCount", 0))
    error_count = int(stats.get("errorCount", 0))
    ignored_error_count = int(stats.get("ignoredErrorCount", 0))

    if file_count + dir_count <= 0:
        fail("snapshot contains no files or directories")

    if error_count != 0 or ignored_error_count != 0:
        fail(
            "snapshot reports errors "
            f"(errorCount={error_count}, ignoredErrorCount={ignored_error_count})"
        )

    print(f"{snapshot_id}\t{root_object_id}")


if __name__ == "__main__":
    main()
