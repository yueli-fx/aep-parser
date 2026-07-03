# Scripts

`scripts/` contains repository automation that is useful outside a single Go
package or command.

## Contents

- `ae-worker/` contains After Effects host automation scripts and PowerShell
  wrappers used by AE-assisted gates.
- `effects-dict/` contains scripts for extracting and matching effect
  dictionaries.
- `fixtures/` contains fixture regeneration and ship-gate helpers.
- `migration/` contains coverage, matrix, and checkpoint automation for AEP
  version migration work.

## What Belongs Here

- Reusable automation that coordinates external tools, AE, fixtures, or registry
  updates.
- Scripts with a documented command path and expected output location.
- Small helper libraries that are shared by more than one script in this tree.

## What Does Not Belong Here

- One-off probes. Use `_tmp_debug/` or `tmp/`.
- Main repo CLIs. Use `cmd/<name>/`.
- Package-specific Go logic. Keep it in the owning `internal/<package>/`.
