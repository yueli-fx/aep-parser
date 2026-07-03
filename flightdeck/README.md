# Flightdeck

`flightdeck/` is the project collaboration deck. It stores routing rules, active
work state, durable project knowledge, and showcase governance.

## Contents

- `briefing.md` is the entry rule sheet. Keep it short and focused on routing
  and repo policy.
- `cockpit.md` is the active work index and resumable task ledger.
- `knowledge/` contains durable project knowledge. Start with
  `knowledge/INDEX.md`; do not bulk-read the whole tree.
- `work/` contains active work packages and investigation ledgers.
- `showcase/` contains tracked showcase definitions and the showcase index.
  Heavy AEP/render outputs are ignored.
- `references/` is ignored local reference/cache material for Flightdeck use.

## What Belongs Here

- Rules and routing that agents need before work starts.
- Durable knowledge with `SUMMARY` and `READ WHEN` headers.
- Active work ledgers that are still being shaped or executed.
- Showcase source scripts, indexes, and review notes that should travel with
  the repository.

## What Does Not Belong Here

- General generated command output. Use `tmp/`.
- Public API docs. Use `docs/`.
- Long-lived machine evidence that should be queryable by tools. Use
  `registry/` or `data/reference/`.

Completed large work packages should be archived out of this repository into
`~/.flightdeck/projects/<slug>/archive/` and removed from the in-flight list.
