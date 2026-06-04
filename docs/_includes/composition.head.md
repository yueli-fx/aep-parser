<!-- Hand-authored lead. Per-symbol docs below are generated from internal/aep
     doc comments. -->

A `Composition` is an AE comp — resolution, frame rate, duration, work area,
motion-blur sampling, background color, its layer list, and comp-level markers.
Access them via `project.Compositions[i]` or `project.CompositionByID(id)`;
create one with [`NewComposition`](project.md#newcomposition).

> **`TickRate` is per-composition** — don't assume a global 8000. It's derived
> from `cdta`, and all keyframe / marker time conversions use it (modern comps
> use ticks/sec; legacy 29.97 NTSC comps use 8000).
