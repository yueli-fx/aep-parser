# Nexrender / aerender preflight

Use the unified `aep` CLI before a project enters a render worker. The preflight process does not replace Nexrender or aerender; it rejects unreadable inputs, records a normalized profile, and optionally migrates the project to the worker's writer target.

## 1. Inspect the source

```powershell
aep inspect -in source.aep > source.inspect.json
```

Exit `0` means the project was parsed. Treat exit `1` as a failed preflight and exit `2` as invalid command configuration.

## 2. Capture a stable profile

```powershell
aep profile -in source.aep > source.profile.json
```

Store this artifact with the render job. It provides a deterministic input fingerprint for later incident analysis.

## 3. Migrate when the worker target differs

```powershell
aep migrate -in source.aep -target AE2025 -out worker-input.aep > migration.json
```

Submit `worker-input.aep` only when the command exits `0`. A blocked report is an explicit compatibility failure, not a reason to silently render the original file.

## 4. Detect template drift

```powershell
aep diff -expected approved-template.aep -actual worker-input.aep > template.diff.json
```

A diff exits `1`, allowing the job producer to route the project for review before consuming render capacity.

## Pipeline placement

```text
upload/job -> aep inspect/profile -> optional aep migrate -> policy decision -> nexrender/aerender worker
```

Keep AE execution in the existing render worker. The preflight node can run on Linux, macOS, or Windows and requires no AE installation.
