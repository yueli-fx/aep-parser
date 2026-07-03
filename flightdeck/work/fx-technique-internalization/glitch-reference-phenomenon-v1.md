# Glitch Reference Phenomenon v1

## Purpose

Process the existing Motionbox glitch samples as the next reference phenomenon
for the technique internalization pipeline.

This package turns two real `.aep` projects into machine-readable technique
outputs and a reusable glitch phenomenon recipe. It must not drift into
single-effect or single-field verification.

## Inputs

- `data/samples/motionbox/glitch/booyah-glitch/Booyah Glitch.aep`
- `data/samples/motionbox/glitch/glitchtext/GlitchText_è«É¼.aep`
- `cmd/aeptechnique`
- `cmd/aepdissect`
- `flightdeck/knowledge/techniques/fx-techniques.md`

## Outputs

Generated JSON artifacts:

- `tmp/technique_reference_glitch/glitch_summary.json`
- `tmp/technique_reference_glitch/booyah_explain.json`
- `tmp/technique_reference_glitch/glitchtext_explain.json`
- `tmp/technique_reference_glitch/package_summary.json`

Persistent knowledge:

- `flightdeck/knowledge/techniques/build-good-glitch.md`

## Acceptance

- Both glitch samples parse through `aeptechnique`.
- Corpus summary records 2 projects, 0 errors.
- Booyah is classified as stock-AE/native analysis-ready.
- GlitchText is classified as plugin-dependent and not plugin-free render
  equivalent.
- The recipe records shared glitch techniques and separates plugin-free vs
  plugin-required routes.
- `fx-techniques.md` points the glitch phenomenon row to the recipe.
- Verification passes:

```powershell
go run ./cmd/aeptechnique -in data\samples\motionbox\glitch -corpus -recursive -mode explain -out tmp\technique_reference_glitch\glitch_explain.jsonl -summary -summary-out tmp\technique_reference_glitch\glitch_summary.json
go run ./cmd/aeptechnique -in "data\samples\motionbox\glitch\booyah-glitch\Booyah Glitch.aep" -mode explain -out tmp\technique_reference_glitch\booyah_explain.json
go run ./cmd/aeptechnique -in "data\samples\motionbox\glitch\glitchtext\GlitchText_è«É¼.aep" -mode explain -out tmp\technique_reference_glitch\glitchtext_explain.json
go test ./internal/technique ./cmd/aeptechnique -count=1
git diff --check
```

## Boundaries

- This package is understanding/recipe work, not a new generator.
- No expression runtime reconstruction.
- No third-party effect equivalence claim. Plugin-heavy glitch is readable and
  template-supportable later, but render equivalence needs the plugins.
