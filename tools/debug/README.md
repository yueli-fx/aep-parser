# Debug Tools

`tools/debug/` contains small Go programs for inspecting AEP structures during
reverse engineering and serializer debugging.

## What Belongs Here

- Focused dumpers, scanners, and probes that help inspect binary chunks or scene
  model behavior.
- Utilities that are still useful after the original investigation.
- Tests for shared helper code inside a debug utility.

## What Does Not Belong Here

- Long-running workflow CLIs. Promote those to `cmd/<name>/`.
- Temporary probes that are useful for only one run. Use `_tmp_debug/`.
- Output files from a run; write them under `tmp/` or another ignored output
  directory.

Keep each tool narrow. If a debug utility becomes part of normal verification or
automation, move the stable entry point into `cmd/`.
