# Technique Portrait Corpus v1

## Goal

Make the Technique Portrait layer usable on many projects without writing
throwaway scripts.

## Scope

Extend `cmd/aeptechnique` with corpus JSONL mode:

```powershell
go run ./cmd/aeptechnique -in <dir-or-file> -mode portrait -corpus -recursive
go run ./cmd/aeptechnique -in <dir-or-file> -mode portrait -corpus -recursive -summary
go run ./cmd/aeptechnique -in <dir-or-file> -mode explain -corpus -recursive
pwsh -NoProfile -File scripts\technique_showcase_report.ps1 -Limit 3
pwsh -NoProfile -File scripts\technique_showcase_report.ps1 -Open
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
- hint counts
- effect match counts
- shape family counts
- text animator kind counts
- layer role counts
- graph edge relation counts

The command should continue after per-file failures and return exit code `1`
when any record has an error. Usage or flag errors still return `2`.

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
- `report.md`
- `report.html`

Passing `-Open` opens the generated HTML report after writing all files.

The report script now uses `-mode explain`, so per-project cards include
deterministic technique notes in addition to raw portrait counts.

## Non-Goals

- No persistent database.
- No clustering.
- No natural-language summaries.
- No parallel workers in v1.

## Acceptance

- CLI tests cover recursive JSONL corpus output and invalid directory usage.
- Existing single-file facts and portrait modes remain compatible.
