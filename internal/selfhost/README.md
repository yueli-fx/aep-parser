# Selfhost Package

`internal/selfhost` powers the self-hosted learning and reporting pipeline over
local AEP sample corpora.

## What Belongs Here

- Sample-shell extraction, technique reports, effect-field inventories, and
  report comparison logic.
- Pseudo-effect behavior audits and recipe smoke/report verification helpers.
- Report models and builders used by `cmd/aepselfhost`.

## What Does Not Belong Here

- Raw sample projects. Use ignored `data/samples/`.
- Generated report output. Use `tmp/` unless a result is promoted to durable
  evidence.
- Core parser/serializer behavior. Keep that in the parser, scene, serializer,
  or technique packages.

This package may summarize sample behavior, but durable rules extracted from it
belong in `flightdeck/knowledge/` or `registry/evidence/`.
