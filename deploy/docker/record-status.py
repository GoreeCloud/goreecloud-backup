#!/usr/bin/env python3
import argparse
import datetime as dt
import fcntl
import json
import os
import re
import tempfile
import uuid

SCHEMA_VERSION = 1
MAX_EVENTS = 512
OPERATIONS = {
    "backup",
    "acceptance-recovery",
    "connect-repository",
    "repository-status",
    "validate",
}
STATES = {"started", "succeeded", "failed", "technical-restore"}
SNAPSHOT_ID_RE = re.compile(r"^[0-9a-f]{32}$")
CATEGORY_RE = re.compile(r"^[a-z0-9._-]{1,64}$")


def now_utc():
    return (
        dt.datetime.now(dt.timezone.utc)
        .replace(microsecond=0)
        .isoformat()
        .replace("+00:00", "Z")
    )


def atomic_write(path, data, mode=0o644):
    directory = os.path.dirname(path) or "."
    os.makedirs(directory, mode=0o755, exist_ok=True)

    fd, temp_path = tempfile.mkstemp(prefix=".goreecloud-backup-", dir=directory)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            handle.write(data)
            handle.flush()
            os.fsync(handle.fileno())
        os.chmod(temp_path, mode)
        os.replace(temp_path, path)
    finally:
        try:
            os.unlink(temp_path)
        except FileNotFoundError:
            pass


def load_status(path):
    if not os.path.exists(path):
        return {
            "schemaVersion": SCHEMA_VERSION,
            "component": "goreecloud-backup-vps",
            "protectionState": "Configured",
            "recoveryEvidenceState": "InsufficientForRestoreVerified",
            "monitoringIntegrationState": "NotYetAccepted",
            "notificationIntegrationState": "NotYetAccepted",
        }

    with open(path, encoding="utf-8") as handle:
        data = json.load(handle)

    if data.get("schemaVersion") != SCHEMA_VERSION:
        raise SystemExit("unsupported status schema version")
    if data.get("component") != "goreecloud-backup-vps":
        raise SystemExit("unexpected status component")

    return data


def load_event_lines(path):
    if not os.path.exists(path):
        return []

    with open(path, encoding="utf-8") as handle:
        lines = [line.rstrip("\n") for line in handle if line.strip()]

    for line in lines:
        event = json.loads(line)
        if event.get("schemaVersion") != SCHEMA_VERSION:
            raise SystemExit("unsupported event schema version")

    return lines[-(MAX_EVENTS - 1) :]


def event_name(state):
    if state == "technical-restore":
        return "goreecloud.backup.vps.recovery.technical_restore_completed"
    return f"goreecloud.backup.vps.operation.{state}"


def make_event(state, operation, operation_id, timestamp, snapshot_id=None, category=None):
    event = {
        "schemaVersion": SCHEMA_VERSION,
        "eventName": event_name(state),
        "severity": "error" if state == "failed" else "info",
        "component": "goreecloud-backup-vps",
        "operation": operation,
        "operationId": operation_id,
        "occurredAt": timestamp,
        "resultCategory": category or state,
    }
    if snapshot_id:
        event["snapshotId"] = snapshot_id
    return event


def validate_args(args):
    if args.operation not in OPERATIONS:
        raise SystemExit("unsupported operation")
    if args.state not in STATES:
        raise SystemExit("unsupported state")

    if args.snapshot_id and not SNAPSHOT_ID_RE.fullmatch(args.snapshot_id):
        raise SystemExit("snapshot ID is malformed")

    if args.failure_category and not CATEGORY_RE.fullmatch(args.failure_category):
        raise SystemExit("failure category is malformed")

    if args.state == "started":
        if args.operation_id:
            raise SystemExit("started state generates its own operation ID")
    else:
        if not args.operation_id:
            raise SystemExit("terminal state requires an operation ID")
        try:
            uuid.UUID(args.operation_id)
        except ValueError as exc:
            raise SystemExit("operation ID is malformed") from exc

    if args.state == "failed" and not args.failure_category:
        raise SystemExit("failed state requires a failure category")

    if args.state == "succeeded" and args.operation == "backup" and not args.snapshot_id:
        raise SystemExit("successful backup requires a snapshot ID")

    if args.state == "technical-restore" and not args.snapshot_id:
        raise SystemExit("technical restore requires a snapshot ID")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--status-file", required=True)
    parser.add_argument("--events-file", required=True)
    parser.add_argument("--operation", required=True)
    parser.add_argument("--operation-id")
    parser.add_argument("--snapshot-id")
    parser.add_argument("--failure-category")
    parser.add_argument("state")
    args = parser.parse_args()
    validate_args(args)

    status_dir = os.path.dirname(args.status_file) or "."
    os.makedirs(status_dir, mode=0o755, exist_ok=True)
    lock_path = args.status_file + ".lock"

    with open(lock_path, "a+", encoding="utf-8") as lock:
        os.chmod(lock_path, 0o600)
        fcntl.flock(lock.fileno(), fcntl.LOCK_EX)

        status = load_status(args.status_file)
        event_lines = load_event_lines(args.events_file)
        timestamp = now_utc()

        if args.state == "started":
            current = status.get("currentOperation")
            if isinstance(current, dict) and current.get("id"):
                interrupted_id = current.get("id")
                interrupted_name = current.get("name", "unknown")
                implicit_event = make_event(
                    "failed",
                    interrupted_name,
                    interrupted_id,
                    timestamp,
                    category="interrupted_or_unrecorded_completion",
                )
                event_lines.append(
                    json.dumps(implicit_event, sort_keys=True, separators=(",", ":"))
                )
                status["lastFailure"] = {
                    "at": timestamp,
                    "operation": interrupted_name,
                    "operationId": interrupted_id,
                    "category": "interrupted_or_unrecorded_completion",
                }
                status["protectionState"] = "Degraded"

            operation_id = str(uuid.uuid4())
            status["currentOperation"] = {
                "id": operation_id,
                "name": args.operation,
                "state": "running",
                "startedAt": timestamp,
            }
            status["updatedAt"] = timestamp
            event = make_event("started", args.operation, operation_id, timestamp)
            print(operation_id)
        else:
            current = status.get("currentOperation")
            if not isinstance(current, dict):
                raise SystemExit("no current operation exists for terminal update")
            if current.get("id") != args.operation_id:
                raise SystemExit("operation ID does not match current operation")
            if current.get("name") != args.operation:
                raise SystemExit("operation name does not match current operation")

            status["lastOperation"] = {
                "id": args.operation_id,
                "name": args.operation,
                "state": args.state,
                "startedAt": current.get("startedAt"),
                "finishedAt": timestamp,
            }
            status.pop("currentOperation", None)
            status["updatedAt"] = timestamp

            if args.state == "failed":
                status["protectionState"] = "Degraded"
                status["lastFailure"] = {
                    "at": timestamp,
                    "operation": args.operation,
                    "operationId": args.operation_id,
                    "category": args.failure_category,
                }
            else:
                # A successful operation does not establish full recoverability.
                # Monitoring, notifications, and application-level recovery
                # acceptance remain separate gates.
                status["protectionState"] = "Configured"
                if args.snapshot_id:
                    status["lastSuccessfulBackup"] = {
                        "at": timestamp,
                        "snapshotId": args.snapshot_id,
                    }

                if args.state == "technical-restore":
                    status["lastTechnicalRestore"] = {
                        "at": timestamp,
                        "snapshotId": args.snapshot_id,
                    }
                    status["recoveryEvidenceState"] = (
                        "TechnicalRestoreCompletedPendingApplicationValidation"
                    )

            event = make_event(
                args.state,
                args.operation,
                args.operation_id,
                timestamp,
                snapshot_id=args.snapshot_id,
                category=args.failure_category,
            )

        event_lines.append(json.dumps(event, sort_keys=True, separators=(",", ":")))
        event_lines = event_lines[-MAX_EVENTS:]

        atomic_write(
            args.status_file,
            json.dumps(status, indent=2, sort_keys=True) + "\n",
        )
        atomic_write(args.events_file, "\n".join(event_lines) + "\n")


if __name__ == "__main__":
    main()
