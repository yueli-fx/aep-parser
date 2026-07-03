# Build Good Glitch — glitch phenomenon recipe
SUMMARY: Good glitch recipe distilled from Motionbox Booyah and GlitchText samples; separates stock-AE native glitch from plugin-heavy glitch while preserving reusable technique atoms.
READ WHEN: building or analyzing glitch, CRT, cyberpunk text, datamosh-like transitions, RGB split looks, noisy displacement cuts, or deciding whether a glitch reference can be reproduced plugin-free
RECHECK WHEN: a new glitch sample is added; Twitch/PEDG/Colorama embed-template support lands; showcase/glitch is user-reviewed; expression runtime support changes

---

## Sources

Machine-readable outputs:

- `tmp/technique_reference_glitch/glitch_summary.json`
- `tmp/technique_reference_glitch/booyah_explain.json`
- `tmp/technique_reference_glitch/glitchtext_explain.json`
- `tmp/technique_reference_glitch/package_summary.json`

Reference projects:

- `data/samples/motionbox/glitch/booyah-glitch/Booyah Glitch.aep`
- `data/samples/motionbox/glitch/glitchtext/GlitchText_è«É¼.aep`

## Project Profiles

```yaml
phenomenon: glitch
samples:
  booyah:
    source: data/samples/motionbox/glitch/booyah-glitch/Booyah Glitch.aep
    class: procedural_text_glitch
    readiness: analysis_ready
    reproducibility: stock_ae_native
    scale: {comps: 12, layers: 61, effects: 42, dependencies: 110}
    top_effects:
      - ADBE Displacement Map: 12
      - ADBE Geometry2: 7
      - ADBE Glo2: 5
      - ADBE Fractal Noise: 4
    structure:
      - RGB split precomp with R/G/B text copies and staggered opacity keyframes
      - dense adjustment-layer slices with masks and Transform/Displacement Map
      - Fractal Noise maps referenced by Displacement Map layers
      - native glow, blur, exposure, Venetian Blinds, and shape accents
  glitchtext:
    source: data/samples/motionbox/glitch/glitchtext/GlitchText_è«É¼.aep
    class: plugin_heavy_text_glitch
    readiness: needs_plugins
    reproducibility: readable_and_template_supportable_later
    scale: {comps: 7, layers: 35, effects: 68, dependencies: 129}
    top_native_effects:
      - ADBE Displacement Map: 20
      - ADBE Fractal Noise: 5
      - ADBE Posterize Time: 4
      - ADBE Tint: 4
    plugin_effects:
      - PEDG: 2
      - PEQCAGL: 2
      - Videocopilot Twitch: 2
      - APC Colorama: 1
      - KNSW Unmult: 1
      - VIDEOCOPILOT VIBRANCE: 1
```

## Recipe

Stock-AE route:

1. Create a staged precomp pipeline: editable source text, RGB split pass,
   noise-map pass, distortion/glow pass, final comp.
2. Build RGB split from three copies of the source text, each filled R/G/B and
   opacity-staggered over short windows.
3. Generate blocky motion maps with Fractal Noise solids.
4. Apply Displacement Map on short-lived adjustment slices; horizontal
   displacement carries the glitch tear, vertical displacement stays secondary.
5. Add Transform scale-height punches on masked adjustment slices for digital
   stretch cuts.
6. Add Glow/Blur/Exposure as finishing passes.
7. Use Venetian Blinds or Grid only when the recipe wants CRT scanlines; Booyah
   proves scanlines exist in the reference, but prior showcase review found they
   can darken a clean text look.

Plugin-heavy route:

1. Preserve the native route above as the fallback spine.
2. Treat Twitch/PEDG/PEQCAGL/Colorama/Vibrance as first-class readable facts,
   not as skipped data.
3. Do not claim plugin-free equivalence. Exact render requires installed
   plugins or later embed-template support from real sample chunks.

## Technique Atoms Used

- `rgb-channel-split`
- `displacement-distortion`
- `noise-as-material`
- `temporal-glitch`
- `emissive-glow`
- `scanlines-crt`
- `controller-rig`
- `precomp-effect-pipeline`

No new cross-domain atom was added in this pass. The value is a phenomenon
recipe that orders existing atoms and separates plugin-free vs plugin-required
routes.

## Boundaries

- Booyah is the clean stock-AE learning sample.
- GlitchText is useful for plugin parameter reading and future embed-template
  sampling, but it is not plugin-free render-equivalent.
- Expression and controller semantics are source facts unless runtime
  expression support is added.
- A recipe is not validated until a generated showcase/render gate proves the
  constructed look.
