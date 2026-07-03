# Registry Package

`internal/registry` is the Go implementation behind registry audits,
ownership checks, coverage reports, cleanup policy, and migration state updates.

## What Belongs Here

- Code that reads, validates, audits, or updates files under `registry/`.
- Asset ownership, layout, cleanup, and evidence governance logic.
- Coverage and current-state report builders used by `cmd/aepregistry`.

## What Does Not Belong Here

- The registry data files themselves. Keep those under top-level `registry/`.
- Public API documentation generation. Use `internal/apidoc`,
  `internal/capindex`, or `docs/`.
- One-off cleanup scripts. Use `scripts/` or `tmp/` until the workflow is
  stable.

Changes here should usually be paired with an audit command or test that proves
the registry data still loads and validates.
