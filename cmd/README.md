# cmd tools

This directory contains repo-local command-line tools. Run them from the
repository root with `go run ./cmd/<name> ...` unless the command says
otherwise.

## Main workflow tools

| Command | Purpose |
|---|---|
| `aep` | Unified product CLI for headless `inspect`, `profile`, `diff`, `migrate`, and `capabilities`; prefer this for new external workflows. |
| `aeprecipe` | Compile/inspect recipe-level JSON into generated AEP projects. This is the main from-scratch generation entry point for recipe experiments. |
| `aepmigrate` | Convert AEPs across supported AE writer versions and run version-matrix migration checks. Uses `internal/aepmigrate`. |
| `aepregistry` | Registry governance tool for asset locations, capability atoms, evidence, version coverage, cleanup policy, and gates. It owns commands such as `audit`, `coverage`, `current`, `gate`, `layout`, `ownership`, and `cleanup`. |
| `aepverify` | Verify recipes/projects against host or profile expectations. Used for recipe profile summaries and validation reports. |
| `aepselfhost` | Self-hosted learning/research pipeline over sample projects: sample shell extraction, effect-field inventory/understanding, pseudo behavior proofs, technique reports, and run monitoring. |
| `aeptechnique` | Analyze AEP profiles into technique-level facts and explanations for the technique-internalization work. |

## Analysis and reverse-engineering tools

| Command | Purpose |
|---|---|
| `aepdissect` | Structured teardown of an AEP: effects, dependencies, precomp nesting, layer details, changed parameters, and third-party plugin use. Good first pass for understanding a reference project. |
| `aepdiff` | Compare two AEPs through semantic profiles and report differences. |
| `aepgaps` | Turn AEP/profile/render differences into gap reports. Subcommands include `diff` and `render`. |
| `aepslices` | Plan and diagnose smaller clone/debug slices from a larger AEP. Subcommands include `plan` and `diagnose`. |
| `aepsearch` | Search one AEP or build a small corpus index for facts such as effects, sources, properties, expressions, and fonts. |
| `aeoracle` | Host-assisted visual oracle workflow: plan render requests, render through AE, and compare image outputs. Subcommands include `plan`, `clone-request`, `render`, `compare`, and `compare-set`. |

## Service and demo tools

| Command | Purpose |
|---|---|
| `aepserver` | HTTP server wrapper around the internal server package. Useful for local API experiments. |
| `aepdemo` | Small from-scratch demo generator that writes a feature demonstration AEP. Mostly useful as a smoke/demo command, not as the main workflow. |

## Documentation and index generators

| Command | Purpose |
|---|---|
| `capindex` | Generate and query `docs/capabilities.json` and `docs/capabilities.md` from `internal/aep` capability tags. Also supports drift checking. |
| `docgen` | Generate public API docs from doc comments using `docs/docgen.json`. |
| `recipedocgen` | Generate recipe schema/docs into `docs/` from recipe metadata. |

## Practical notes

- Prefer `aepregistry gate -scope asset-policy` before deleting or moving repo
  data. Cleanup policy is registry-driven, not ad-hoc.
- Prefer `aeprecipe`, `aepmigrate`, and `aepregistry` for the main generation /
  migration / verification loop.
- Treat `aeoracle` and `aepselfhost` as host-aware tools: parts of those flows
  may require installed After Effects versions and the AE worker scripts.
- Generator commands (`capindex`, `docgen`, `recipedocgen`) update tracked docs;
  run the matching tests/checks before committing generated output.
- Existing focused commands remain compatibility and advanced-maintenance
  surfaces; new user workflows should enter through `cmd/aep` unless they need
  a capability the unified CLI has not adopted yet.
