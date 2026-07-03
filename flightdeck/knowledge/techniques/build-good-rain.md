# Build Good Rain — rain phenomenon recipe
SUMMARY: Rain recipe distilled from Motionbox Rain Day sample; separates native shape/repeater rain construction from Trapcode Particular exact-render dependencies.
READ WHEN: building or analyzing rain, falling streaks, particle rain, umbrella/water-scene motion graphics, or deciding whether a rain reference can be reproduced plugin-free
RECHECK WHEN: a new rain sample is added; native rain showcase generation lands; Particular/Unmult embed-template support lands; expression/runtime behavior support changes

---

## Sources

Machine-readable outputs:

- `tmp/technique_reference_rain/rain_summary.json`
- `tmp/technique_reference_rain/rain_explain.json`
- `tmp/technique_reference_rain/rain_explain.jsonl`
- `tmp/technique_reference_rain/package_summary.json`
- `tmp/technique_showcase_rain/rain_explain.json`
- `tmp/technique_showcase_rain/package_summary.json`

Reference project:

- `data/samples/motionbox/generative/e8vfb9qjb9n0/雨の日モーショングラフィックス.aep`

Project readme notes the source was authored for After Effects CC 2020 and uses
Trapcode Particular plus Unmult as external effects.

## Project Profile

```yaml
phenomenon: rain
sample:
  source: data/samples/motionbox/generative/e8vfb9qjb9n0/雨の日モーショングラフィックス.aep
  class: shape_controller_particle_rain_scene
  readiness: needs_plugins
  reproducibility: native_spine_plus_plugin_required_exact_route
  scale: {comps: 5, layers: 45, effects: 73, shape_operators: 173, dependencies: 101}
  top_effects:
    - ADBE Slider Control: 39
    - ADBE Fill: 6
    - ADBE Simple Choker: 6
    - ADBE Echo: 5
    - tc Particular: 4
  plugin_effects:
    - tc Particular: 4
    - KNSW Unmult: 1
  top_shape_signals:
    - adbe_vector_ellipse_size: 34
    - ellipse: 17
    - adbe_vector_stroke_width: 16
    - adbe_vector_ellipse_position: 12
    - adbe_vector_repeater_copies: 8
    - adbe_vector_trim_start/end: 7 each
  structure:
    - `rain` precomp contains keyframed shape rain strokes driven by native controls.
    - `render` assembles umbrella/body/water/rain precomps with camera and parent links.
    - shape layers, repeaters, trim paths, and effect layer references form the native rain spine.
    - Particular and Unmult are exact-render dependencies, not plugin-free claims.
```

## Recipe

Stock-AE spine:

1. Create a staged comp pipeline: rain streak precomp, water/effect precomp,
   umbrella/scene precomps, and a final render comp.
2. Build rain streaks from shape paths or narrow ellipses with fill/stroke,
   repeater copies, and optional trim-path start/end animation.
3. Drive falling motion with keyframes rather than expressions; use staggered
   offsets and varied stroke length/opacity to avoid a rigid grid.
4. Add simple native controls for density, angle, speed, and spread when the
   look needs reusable tuning.
5. Use Echo/Posterize Time selectively for persistence and cadence, but keep
   them behind the rain layer so the scene stays readable.
6. Composite scene objects and rain through precomps; use mattes/parents to
   keep umbrella, water, and background elements coordinated.

Plugin-required exact route:

1. Preserve the stock-AE spine above as the learnable fallback.
2. Treat Trapcode Particular rain/particle systems as first-class readable
   facts and future embed-template candidates.
3. Treat Unmult as an alpha/unpremultiply dependency for exact source output.
4. Do not claim plugin-free exact render unless a separate native approximation
   showcase is authored and rendered.

## Technique Atoms Used

- `shape-repeater-rain-streaks`
- `particle-emit`
- `time-evolution`
- `precomp-effect-pipeline`
- `controller-rig`
- `matte-composite`
- `final-grade`

## Boundaries

- This recipe is observed from one sample, not AE-render-validated as a
  full reference clone.
- The stock-AE shape/repeater rain spine is AE2020-render-validated by
  `flightdeck/showcase/rain`.
- Exact render depends on Trapcode Particular and Unmult.
- Shape/repeater rain can be generated natively, but Particular particle rain
  remains a plugin dependency until an embed-template or native approximation
  package is authored.
- Controller semantics are source facts unless expression/runtime support is
  added.
