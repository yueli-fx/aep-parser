# Technique Portrait Corpus v1

## Goal

Make the Technique Portrait layer usable on many projects without writing
throwaway scripts.

## Scope

Extend `cmd/aeptechnique` with corpus JSONL mode:

```powershell
go run ./cmd/aeptechnique -in <dir-or-file> -mode portrait -corpus -recursive
go run ./cmd/aeptechnique -in <dir-or-file> -mode portrait -corpus -recursive -summary
go run ./cmd/aeptechnique -in <dir-or-file> -mode explain -corpus -recursive -out corpus.jsonl -summary-out summary.json
go run ./cmd/aeptechnique -in <dir-or-file> -mode explain -corpus -recursive
pwsh -NoProfile -File scripts\technique_showcase_report.ps1 -Limit 3
pwsh -NoProfile -File scripts\technique_showcase_report.ps1 -Limit 3 -Verify
pwsh -NoProfile -File scripts\technique_showcase_report.ps1 -Open
pwsh -NoProfile -File scripts\verify_technique_report.ps1 -OutDir tmp\technique_showcase_report
pwsh -NoProfile -File scripts\compare_technique_reports.ps1 -BaseDir tmp\old_report -NewDir tmp\technique_showcase_report
pwsh -NoProfile -File scripts\verify_technique_selfhost.ps1 -OutRoot tmp\technique_selfhost_gate
pwsh -NoProfile -File scripts\verify_technique_selfhost.ps1 -OutRoot tmp\technique_selfhost_gate -Open
```

For every discovered `.aep`, emit one JSON object per line:

- `path`
- `mode`
- `portrait` for `-mode portrait`
- `facts` for `-mode facts`
- `explanation` for `-mode explain`
- `error` when a project cannot be opened or profiled

With `-summary`, emit one aggregate JSON object instead of JSONL. In explain
mode the embedded portrait is used for aggregate counts:

- `project_count`
- `error_count`
- aggregate fingerprint totals
- recreation readiness counts for explain mode
- archetype counts for explain mode
- pattern counts for repeated technique combinations in explain mode
- pattern representative projects for high-signal examples in explain mode
- pattern mechanism profiles for readiness, effects, plugin effects, shapes, and text animators
- pattern recreation-step profiles for common reconstruction order
- third-party plugin effect counts for explain mode
- hint counts
- effect match counts
- shape family counts
- text animator kind counts
- layer role counts
- graph edge relation counts

The command should continue after per-file failures and return exit code `1`
when any record has an error. Usage or flag errors still return `2`.

With `-summary-out`, corpus JSONL mode writes records to `-out` and writes the
same aggregate summary to a second file in one parse pass. The showcase report
script uses this path so large corpora are not parsed twice.

## Rules

- File input with `-corpus` emits one JSONL record.
- Directory input requires `-recursive`; v1 intentionally avoids shallow
  directory ambiguity.
- Discovered paths are sorted for deterministic output.
- `-limit N` processes at most N discovered files when N is positive.
- Corpus mode does not retain parsed projects after each record is emitted.
- Summary mode aggregates the emitted records and keeps only counts.

## Showcase Report Script

`scripts/technique_showcase_report.ps1` is the self-hosted demonstration entry.
It defaults to `flightdeck\showcase` and writes generated output under
`tmp\technique_showcase_report`:

- `summary.json`
- `corpus.jsonl`
- `digest.json`
- `learning.md`
- `projects.csv`
- `patterns.csv`
- `study_queue.csv`
- `learning_actions.csv`
- `errors.csv`
- `manifest.json`
- `report.md`
- `report.html`

Passing `-Open` opens the generated HTML report after writing all files.
Passing `-Verify` runs the artifact verifier after report generation.
`scripts/verify_technique_report.ps1` validates that all generated artifacts
exist, that `summary.json`, `corpus.jsonl`, and `digest.json` agree on project
counts plus per-file errors, and that the human reports contain the expected
learning sections.

The report script continues when `aeptechnique` reports per-file corpus errors.
Those failed projects stay in `corpus.jsonl` and `report.html` as error records,
while successful projects still feed the summary, digest, CSV exports, and study
queue. Usage errors or missing output artifacts still fail the script.

The report script now uses `-mode explain`, so per-project cards include
deterministic recreation readiness, recreation steps, archetype labels, and
technique notes in addition to raw portrait counts.

`digest.json` is the compact machine-readable learning entry point for larger
corpora. `learning.md` is the compact human-readable learning entry point. They
keep representative projects per pattern, archetype, and readiness bucket so a
large run can be inspected before reading every project record. `projects.csv`
and `patterns.csv` provide spreadsheet-friendly indexes for filtering projects
and repeated technique patterns outside the HTML report. `projects.csv` includes
the stable recreation step IDs for each project; `patterns.csv` includes the
step distribution for each repeated pattern. `study_queue.csv` ranks parsed
projects by readiness and signal density so large corpora have an immediate
review order. `learning_actions.csv` turns each repeated pattern into a review
action with representative project, common steps, top mechanisms, and risk.
`errors.csv` lists failed corpus records for retry or manual inspection.
`manifest.json` records the input path, git revision, scan timing, project/error
counts, and generated artifact inventory for repeatable self-hosted runs.

`scripts/compare_technique_reports.ps1` compares two generated report
directories and writes `compare.json` plus `compare.md`. It reports scalar
changes such as project/error totals and count-map changes for readiness,
patterns, archetypes, hints, plugin effects, effects, shapes, text animators,
layer roles, and graph edges.

`scripts/verify_technique_selfhost.ps1` is the one-command self-hosted
acceptance gate. It runs the technique Go tests, generates and verifies the full
sample report, generates and verifies an intentional partial-error report,
compares a report to itself, compares the partial report to the full report, and
writes `acceptance.json` plus `acceptance.md`. It also writes
`latest_run.txt`, `latest_acceptance.md`, and `latest_index.html` at the selected
output root so the most recent result has a stable path. Passing `-Open` opens
the latest HTML index after the gate finishes.

## Non-Goals

- No persistent database.
- No clustering.
- No natural-language summaries.
- No parallel workers in v1.

## Acceptance

- CLI tests cover recursive JSONL corpus output and invalid directory usage.
- Existing single-file facts and portrait modes remain compatible.
