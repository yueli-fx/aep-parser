# Versioned AEP Migration Spec

## Goal

Support a future workflow where a user can provide one `.aep` from a known or
detected AE version, choose a target version such as `AE2025` or `AE2020`, and
receive a migrated project plus an explicit preservation/loss report.

The product goal is not "save a file with a different label". The product goal
is near-lossless upgrade or honest downgrade:

- preserve everything the parser/writer can represent;
- translate constructs that have a proven equivalent in the target version;
- refuse or report constructs that cannot be represented safely;
- verify the result with profile diff and, when available, AE open/render gates.

## Relationship To Recipe

Recipe `project.target_version` is a compile target. It tells the recipe compiler
which AE skeleton and version-gated writer behavior to use.

Version migration is a larger contract:

- it has a source project and a detected source version;
- it has one target version or a set of candidate target versions;
- it compares source and target capabilities;
- it emits a migration report that explains preservation, translations,
  approximations, drops, and unknowns;
- it may use recipe generation internally, but recipe JSON is not the whole
  migration model.

The `examples/recipes/*.json` files should continue to declare
`project.target_version` explicitly. They are examples of generation targets,
not proof that arbitrary source projects can be downgraded without loss.

## Supported Version Labels

The first public version set should match the current library targets:

- `AE2020`
- `AE2022`
- `AE2025`

Use these labels in every user-facing report, recipe, and CLI flag. Internally
they map to `aep.TargetAE2020`, `aep.TargetAE2022`, and `aep.TargetAE2025`.

When reading a source project, the tool should record both:

- `source_version_label`: normalized bucket such as `AE2022`, if known;
- `source_version_raw`: the exact parsed or AE-reported version string, when
  available.

If source detection is not reliable, report `source_version_label: "unknown"`
instead of guessing. A user may override it with an explicit flag.

## Migration Modes

### Assess

Analyze a source project against one or more target versions without writing a
new `.aep`.

Expected command shape:

```powershell
go run ./cmd/aepmigrate assess -in source.aep -target AE2020 -out assess.json
go run ./cmd/aepmigrate assess -in source.aep -targets AE2020,AE2025 -out assess.json
```

Assess mode produces:

- source version detection;
- target version contract;
- profile fingerprint;
- version compatibility summary;
- blockers and warnings;
- estimated migration classes by profile path and capability key.

### Convert

Produce a migrated `.aep` for one target version and a report.

Expected command shape:

```powershell
go run ./cmd/aepmigrate convert -in source.aep -target AE2025 -out migrated.aep -report report.json
go run ./cmd/aepmigrate convert -in source.aep -target AE2020 -out downgraded.aep -report report.json
```

Convert mode must never silently drop unsupported constructs. If a target cannot
preserve a construct, the report records the decision before the output is
considered acceptable.

### Verify

Validate a migrated project against the source project.

Expected command shape:

```powershell
go run ./cmd/aepmigrate verify -source source.aep -target migrated.aep -target-version AE2020 -out verify.json
```

Verification should use:

- `profilediff` with version-aware ignore rules;
- migration report checks, so every allowed difference has a known reason;
- optional AE open/render checks when an AE executable is available.

## Migration Classes

Every relevant source construct should be assigned exactly one class in the
migration report:

| Class | Meaning |
| --- | --- |
| `preserved` | The target project keeps the same semantic construct and profile value. |
| `retargeted` | The construct is preserved but rewritten through a target-version writer/template. |
| `translated` | The target uses a proven equivalent representation. |
| `approximated` | The target cannot match exactly but an intentional fallback was used. |
| `dropped` | The construct is intentionally omitted because the target version cannot represent it. |
| `blocked` | Conversion cannot continue without manual action or a new writer/parser feature. |
| `unknown` | The source contains data the current parser cannot classify well enough to migrate. |

`dropped`, `blocked`, and `unknown` must include a reason and a stable source
path when possible.

## Version Capability Ledger

The migration system needs a machine-readable ledger independent from prose
docs. This ledger should not duplicate the whole schema. It should only record
version and migration semantics that reflection/profile data cannot infer.

Suggested fields:

```json
{
  "path": "comps[].layers[].matte",
  "kind": "profile_path",
  "min_source_version": "AE2025",
  "target_support": {
    "AE2020": "translate_to_classic_track_matte_or_drop",
    "AE2022": "translate_to_classic_track_matte_or_drop",
    "AE2025": "preserve"
  },
  "capability_key": "layer.set_track_matte_source",
  "default_policy": "blocked",
  "notes": "Explicit source matte is AE2025-only in the current writer contract."
}
```

Rules:

- `path` is a stable public identifier, usually a profile path or recipe path.
- `capability_key` uses stable capability query keys, not Go symbol names.
- `target_support` records semantic support per target version.
- Unknown paths are not silently compatible.

## Pipeline

### 1. Inspect Source

- Open the source `.aep`.
- Detect raw and normalized source version.
- Build `profile.Profile`.
- Build technique facts/explanation when useful for human report context.
- Record parser warnings and unknown chunks/features.

### 2. Plan Migration

- Load the version capability ledger.
- Match profile paths and technique facts against ledger entries.
- Classify each relevant construct for the requested target.
- Produce an assess report before writing.

### 3. Transform

The first implementation should favor conservative reconstruction over blind
binary patching:

- create a new target-version project skeleton;
- rebuild supported project, comp, layer, property, effect, mask, shape, text,
  and dependency constructs through existing writer APIs;
- use recipe generation only for the subset that has recipe coverage;
- keep unsupported source constructs out of the target unless there is a proven
  target-version translation.

This means early downgrade output may be incomplete but honest. It is better to
ship a smaller migrated project with a precise loss report than a corrupted
"almost copied" project.

### 4. Diff And Gate

- Build a target profile from the migrated `.aep`.
- Diff source profile to target profile.
- Apply version-aware allowed-difference rules from the migration report.
- Fail if any unreported difference remains.
- When AE is available, run an AE open gate for the target version.
- When frame comparison is configured, render sentinel frames and attach
  render-gap evidence.

## Report Shape

The report should be JSON-first, with optional Markdown/HTML renderers later.

Minimum JSON fields:

```json
{
  "schema_version": 1,
  "source": {
    "path": "source.aep",
    "version_label": "AE2022",
    "version_raw": "22.x"
  },
  "target": {
    "version_label": "AE2020",
    "path": "downgraded.aep"
  },
  "summary": {
    "status": "blocked",
    "preserved": 120,
    "retargeted": 12,
    "translated": 3,
    "approximated": 0,
    "dropped": 1,
    "blocked": 2,
    "unknown": 4
  },
  "entries": [
    {
      "path": "comps[0].layers[2].matte_ref",
      "class": "blocked",
      "target_version": "AE2020",
      "reason": "Explicit matte source requires AE2025 writer support.",
      "capability_key": "layer.set_track_matte_source"
    }
  ],
  "verification": {
    "profile_diff_status": "pass",
    "profile_diff_count": 0,
    "ae_open_status": "not_run",
    "render_status": "not_run"
  }
}
```

Status rules:

- `pass`: all differences are preserved, retargeted, or translated.
- `warn`: approximations or drops exist, but the user allowed lossy output.
- `blocked`: at least one required construct cannot be safely migrated.
- `error`: parsing, writing, or verification failed unexpectedly.

For `convert`, `profile_diff_status` should be `pass` before a report can be
considered successful. If source-vs-target profile diff finds unexpected
differences, record them under `verification.profile_diffs`, set
`profile_diff_status: "fail"`, and block success.

## Upgrade Policy

Upgrade means source `AE2020`/`AE2022` to target `AE2025`.

Default policy:

- preserve semantics, do not introduce new AE2025-only features automatically;
- retarget project skeleton and target-gated chunk templates where necessary;
- keep the output compatible with the target version's writer capabilities;
- use AE2025 validation to prove the output opens.

Optional future policy:

- `-modernize` may intentionally translate older constructs into newer target
  constructs, but only with explicit report entries.

## Downgrade Policy

Downgrade means source `AE2025`/`AE2022` to target `AE2020`.

Default policy:

- block when the source uses constructs without proven AE2020 representation;
- translate only when a tested equivalent exists;
- preserve older/common constructs through target-version writers;
- require a loss report before writing any lossy output.

Optional future policy:

- `-allow-lossy` can produce a downgraded file with `approximated` and `dropped`
  entries, but the output status cannot be `pass`.

## First Useful Slice

Do not start with arbitrary full-project migration.

The first implementation slice should be:

1. `assess` only, no output `.aep`;
2. source version detection plus target version parsing;
3. profile build;
4. version capability ledger with a few known rules:
   - `project.target_version` / project skeleton target;
   - explicit AE2025 matte source;
   - common AE2020-safe comp/layer/shape/text/effect constructs already covered
     by recipe profile verification;
5. JSON report with migration classes;
6. tests using existing minimal recipe outputs and `minimal-layer-explicit-matte`
   as the first downgrade blocker.

After assess is stable, add conservative convert for projects whose assess
report is all `preserved` / `retargeted` / `translated`.

## Acceptance Gates

For `assess`:

```powershell
go test ./internal/aepmigrate ./cmd/aepmigrate -count=1
go run ./cmd/aepmigrate assess -in examples-or-fixture.aep -target AE2020 -out tmp\migration_assess.json
```

For `convert`:

```powershell
go run ./cmd/aepmigrate convert -in source.aep -target AE2020 -out tmp\downgraded.aep -report tmp\migration_report.json
go run ./cmd/aepmigrate verify -source source.aep -target tmp\downgraded.aep -target-version AE2020 -out tmp\migration_verify.json
```

When AE is available, target AE open gates are required before claiming
commercial-grade migration.

## Non-Goals

- No claim of 1:1 migration before profile diff and AE validation pass.
- No blind chunk copying across target versions.
- No automatic "modernization" unless the user explicitly asks for it.
- No hiding lossy downgrade behind a successful exit status.
- No dependency on Windows-only automation for assess mode.

## Open Questions

- Which source version signal is most reliable from parsed project bytes for
  older real-world files?
- Should the migration ledger live under `internal/aepmigrate` as Go data first,
  or under `data/compatibility` as JSON consumed by tools?
- Should `convert` use recipe IR as the main reconstruction surface, or a lower
  scene-to-scene retargeter that can emit recipe only as a side artifact?
- What is the first real customer-facing downgrade target: AE2025 to AE2020, or
  AE2022 to AE2020?
