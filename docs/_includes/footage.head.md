<!-- Hand-authored lead. Per-symbol docs below are generated from internal/aep
     doc comments. -->

Footage items are the project's source media — external files, AE solids, and
placeholders. Access them via `project.Footage[i]` or `project.FootageByName`.

`Path` is the source path on disk; `SetPath` is the redirect API (it rewrites
the path chunk — the one write that is **not** length-preserving). `MainSource()`
returns the typed source metadata (`*FileSource` / `*SolidSource` /
`*PlaceholderSource`).
