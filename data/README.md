# Data

`data/` holds local source material and durable reference artifacts used to
understand or validate AEP behavior.

## Contents

- `samples/` is the approved local project corpus. It is ignored because it may
  contain large or user-provided AEPs, but it is not a cleanup target.
- `reference/` contains small, tracked artifacts that knowledge notes, reverse
  engineering plans, or tests cite as durable evidence.
- `effects-dict/` contains effect dictionary source material used by the effect
  dictionary scripts and generated indexes.

## What Belongs Here

- Curated sample projects that are useful across multiple investigations.
- Small reference AEPs, reports, or extracted facts that future work must be
  able to cite.
- Source dictionaries or stable vendor-like inputs that are required to
  regenerate project indexes.

## What Does Not Belong Here

- Temporary command output, debug dumps, or one-off probes. Use `tmp/` or
  `_tmp_debug/`.
- Machine-governed registry evidence. Use `registry/evidence/<topic>/`.
- New large binary corpora without an explicit reason and a known owner.
