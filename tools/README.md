# Tools

`tools/` contains tracked developer utilities that are useful for inspection,
debugging, or reverse engineering but are not part of the public CLI surface.

## Contents

- `debug/` contains small Go tools for dumping chunks, inspecting AEP internals,
  and running focused probes.

## What Belongs Here

- Reusable local utilities that are worth keeping under version control.
- Debug tools with a narrow purpose and a clear package boundary.
- Utilities that support investigation but should not become shipped commands.

## What Does Not Belong Here

- User-facing or workflow-level commands. Use `cmd/<name>/`.
- Disposable experiments. Use `_tmp_debug/` or `tmp/`.
- Generated output from debug runs.
