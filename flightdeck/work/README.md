# Work Packages

`flightdeck/work/` contains active or resumable investigation and execution
packages.

## What Belongs Here

- Task ledgers that describe current scope, evidence, decisions, and next
  slices.
- Specs or plans that are still in progress and may change as work proceeds.
- Handoff notes needed to resume an unfinished effort.

## What Does Not Belong Here

- Durable project rules. Promote those to `flightdeck/knowledge/<domain>/`.
- Generated command output. Use `tmp/` until a finding is promoted.
- Completed large work packages. Archive them under
  `~/.flightdeck/projects/<slug>/archive/` and remove them from cockpit.

Draft work notes do not need commits while they are still exploratory. Commit
them only when they become a finalized checkpoint or durable rule.
