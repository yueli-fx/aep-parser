<!-- Hand-authored lead. Per-symbol docs below are generated from internal/aep
     doc comments. -->

Markers are timeline annotations. The same `Marker` type covers both
**composition** markers (`comp.Markers`, the chapter strip at the top of the
timeline) and **layer** markers (`layer.Markers`); they differ only in scope.
Each marker carries a time, an optional duration, a label-color index, and the
five annotation strings from AE's Marker dialog (comment / chapter / URL /
frame-target / cue-point name).
