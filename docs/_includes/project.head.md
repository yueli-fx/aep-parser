<!-- Hand-authored lead. Per-symbol docs below are generated from internal/aep
     doc comments. -->

`Project` is the fully parsed contents of an `.aep` file — the root of the
object graph. Parse one with `aep.Open` / `aep.FromReader`, or build a fresh
one with `aep.NewProject` (see [Functions](#functions)). From there reach
`Compositions`, `Footage`, and `Folders`; mutate via the typed setters; and
serialize back with `WriteAEP` (binary) or `WriteJSON` (one-way export).

> Concurrency: a `Project` and everything reachable from it is **not** safe for
> concurrent mutation — `Set*` calls patch shared chunk bytes in place. Wrap
> mutation in your own synchronization.
