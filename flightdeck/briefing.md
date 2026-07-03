# flightdeck briefing — aep-parser

## Start Here

- Read `flightdeck/cockpit.md` for active work and next steps.
- Do not bulk-read `flightdeck/knowledge/**`. Use
  `flightdeck/knowledge/INDEX.md`, then load only notes whose `READ WHEN`
  matches the task.
- Keep durable knowledge self-contained: future action rules go in
  `flightdeck/knowledge/<domain>/`, not in temporary specs or work notes.

## Repository Rules

- Work on the current branch/mainline directly. Do not create routine feature
  branches unless the user asks.
- Never push to a remote.
- `data/samples/` is approved local corpus, not a cleanup target.
- `tmp/` is disposable. Durable generated evidence belongs in tracked truth
  locations such as `registry/evidence/<topic>/`, `data/reference/<topic>/`,
  committed fixtures, or embedded templates.
- Large plans use total -> slices -> total: define the control ledger first,
  execute slices, then summarize status and gaps back into the ledger.
- Completed large work efforts are archived out of `flightdeck/work/` to
  `~/.flightdeck/projects/<slug>/archive/` and removed from cockpit's in-flight
  list.

## Commit Policy

- Commit code, tests, fixtures, generated docs, or a meaningful completed
  checkpoint.
- Do not commit exploratory markdown churn. Draft specs, plans, knowledge
  sketches, and cockpit/index wording can stay uncommitted while they are still
  being discussed or shaped.
- Commit documentation-only work when it is a key checkpoint: the user approves
  the direction, says to start executing, asks to commit, or the document is a
  finalized rule/spec/knowledge update needed by future work.
- Keep commits atomic by landed unit. Code plus its required tests, fixtures,
  docs, and Flightdeck sync can be one commit; unrelated cleanup should be
  separate.
- Use the global commit checklist in `knowledge/git/commits.md` before staging
  or writing commit messages.

## Public API And Docs

- Public API comments on exported `internal/aep` symbols are the doc source.
  They are English source text for `cmd/docgen`; generated docs are not edited
  by hand.
- Public API changes must sync docs/docgen registration where relevant and
  align with capindex.
- Capability truth source: `go run ./cmd/capindex -q "<term>"`.
- API stability, package boundaries, write semantics, comments, and delivery
  rules live in `flightdeck/knowledge/workflow/project-operating-rules.md`.

## Verification Routes

- General verification and artifact placement:
  `flightdeck/knowledge/workflow/verify.md`.
- AE ship-gates, fixture regeneration, and JSX reverse-engineering:
  `flightdeck/knowledge/workflow/re-fixture.md`.
- Delivery/shippability claims:
  `flightdeck/knowledge/workflow/delivery-contract.md`.
- Showcase generation and visual review:
  `flightdeck/knowledge/showcase/showcase.md`.

## Project Commit Types

The global commit conventions apply, with these project-specific additions:

- `re`: pure reverse-engineering finding or fixture evidence before a shipped
  setter/API exists.
- Common scopes: `layer`, `comp`, `text`, `mask`, `keyframe`, `shape`,
  `property`, `marker`, `footage`, `project`, `aep`, `flightdeck`.
- `feat` means new shipped field R/W, setter, structure API, or fixture-backed
  capability.
- `fix` means wrong read/write bytes, broken setter behavior, or round-trip /
  AE-acceptance regression.
- `docs` means public API docs or Flightdeck documentation.

## Subscriptions

<!-- Global knowledge subscriptions, one ~/.flightdeck-relative path per line. -->
knowledge/coding/comments.md
knowledge/git/commits.md
knowledge/agents/subagent-guide.md
