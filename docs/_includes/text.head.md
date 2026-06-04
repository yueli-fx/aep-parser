<!-- Hand-authored lead. Per-symbol docs below are generated from internal/aep
     doc comments. -->

Text layers carry a decoded `TextSource` — content, fonts, per-run character
styling, paragraphs, and justification. Reach it via `layer.TextSource` (nil
for non-text layers; nil too when the underlying `btdk` PostScript dict failed
to parse, in which case the failure lands on `Project.Warnings`).

Text **mutation** lives on `Layer`, not here: `SetText` plus the per-run
setters (`SetRunFontSize`, `SetRunFontIndex`, `AddFont`, `SetRunFillColor`,
`SetRunStrokeColor`, …) are documented in [layer.md](layer.md). The
`TextJustification` enum is catalogued in [constants.md](constants.md).
