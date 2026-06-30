# Technique Portrait Corpus v1

## Goal

Make the Technique Portrait layer usable on many projects without writing
throwaway scripts.

## Scope

Extend `cmd/aeptechnique` with corpus JSONL mode:

```powershell
go run ./cmd/aeptechnique -in <dir-or-file> -mode portrait -corpus -recursive
```

For every discovered `.aep`, emit one JSON object per line:

- `path`
- `mode`
- `portrait` for `-mode portrait`
- `facts` for `-mode facts`
- `error` when a project cannot be opened or profiled

The command should continue after per-file failures and return exit code `1`
when any record has an error. Usage or flag errors still return `2`.

## Rules

- File input with `-corpus` emits one JSONL record.
- Directory input requires `-recursive`; v1 intentionally avoids shallow
  directory ambiguity.
- Discovered paths are sorted for deterministic output.
- `-limit N` processes at most N discovered files when N is positive.
- Corpus mode does not retain parsed projects after each record is emitted.

## Non-Goals

- No persistent database.
- No clustering.
- No natural-language summaries.
- No parallel workers in v1.

## Acceptance

- CLI tests cover recursive JSONL corpus output and invalid directory usage.
- Existing single-file facts and portrait modes remain compatible.
