# Index — aep-understanding-generation

## State

This work package defines the next long-running direction: turn `aep-parser`
from a parser/writer plus one-off showcase generator into an AEP understanding,
replication-diagnostics, and future generation system.

The current codebase already has the key raw materials:
- `cmd/aepdissect -json` can emit a shallow project profile.
- `internal/scene.WriteJSON` can export a broad parsed scene view.
- `flightdeck/showcase/booyah-clone` contains real replication and render-diff
  tooling, but much of it is bespoke to one project.
- `docs/capabilities.md` is the write-capability truth source.

The gap is not "the project cannot do this"; the gap is that the pieces are not
yet a stable, reusable pipeline.

## Next

Review `design.md`. If accepted, write an implementation plan that starts with
Phase 0 export-surface audit and contract lock before extracting
`internal/profile` or expanding generation.

## Read now

- `design.md` — full spec and phased direction.
- `phase0-audit.md` — current export-surface audit.
- `profile-coverage.md` — coverage matrix and Phase 1 field admission list.
- `profile-contract.md` — evidence ladder, stable path object, and Phase 1
  entry gate.
- `plan.md` — Phase 0 execution plan.
- `phase1-plan.md` — Phase 1 profile extraction plan and verification record.
- `phase2-plan.md` — Phase 2 structural/semantic diff plan.
- `phase3-plan.md` — Phase 3 render oracle harness plan.
- `phase4-plan.md` — Phase 4 gap ledger and capability coupling plan.
- `phase5-plan.md` — Phase 5 slice replication workflow plan.
- `phase5-booyah-acceptance.md` — real Booyah source vs current clone Phase 5
  acceptance run.
- `phase5-non-booyah-acceptance.md` — non-Booyah Motionbox sample Phase 5
  acceptance run.
- `phase6-plan.md` — Phase 6 source-vs-clone render frame-set compare plan
  and recipe-IR readiness gate.
- `phase6-booyah-render-compare.md` — real Booyah source-vs-clone AE 2025
  frame-set render compare result.
- `phase6-recipe-ir-plan.md` — minimal recipe IR/compiler implementation plan,
  gated by the render compare result.
- `phase6-recipe-ir-acceptance.md` — minimal recipe compile + AE 2025 render
  acceptance result.
- `flightdeck/knowledge/techniques/understand-a-project.md` — existing
  reference-project internalization workflow.
- `flightdeck/knowledge/techniques/fx-techniques.md` — current technique
  library.
- `docs/capabilities.md` — current write capability matrix.

## Progress

Done:
- Direction scoped from current codebase.
- Existing support and gaps identified.
- External review feedback triaged into the spec: Phase 0 audit, evidence
  ladder, stable path contract, profile/export boundary, render oracle limits,
  and generation readiness gate.
- Phase 0 audit artifacts written: export-surface audit, profile coverage
  matrix, and profile contract.
- Phase 1 implemented `internal/profile` and moved `aepdissect -json` onto the
  stable profile builder.
- Phase 1 self-review fixed profile gaps for track matte refs and parsed
  in/out points.
- Phase 2 implemented `internal/profilediff` and `cmd/aepdiff` with JSON ignore
  rules and deterministic path-level reports.
- Phase 3 implemented `internal/aeoracle`, `cmd/aeoracle`, and
  `scripts/aeoracle_render.jsx` for sentinel frame planning, render sidecars,
  PNG comparison, and AE dry-run invocation.
- Phase 4 implemented `internal/gapledger` and `cmd/aepgaps` for structured
  diff/render gap reports with stable IDs, evidence, severity, action type,
  and context fields.
- Phase 5 implemented `internal/sliceworkflow` and `cmd/aepslices` for generic
  slice planning and combined profile/render gap workflow reports.
- Phase 5 Booyah acceptance run processed the real source project at
  `data/samples/motionbox/glitch/booyah-glitch/Booyah Glitch.aep` against the current
  clone and produced actionable write gaps after fixing layer-index fallback in
  `profilediff`.
- Phase 5 non-Booyah acceptance run processed
  `data/samples/motionbox/motion-graphics/circle-animation/circle animation.aep`,
  selected representative slices, produced zero self-diff gaps, produced
  cross-project gap output against `seabox.aep`, generated a render-oracle
  sentinel plan, and passed a real AE 2025 hard render gate after hardening
  `cmd/aeoracle render` and `scripts/aeoracle_render.jsx`.
- Phase 6 render-compare plan written. The first implementation slice is
  source-vs-clone frame-set compare (`clone-request` + `compare-set`) before
  recipe IR or automated correction loops.
- Phase 6 render-compare implementation executed on Booyah: both source and
  clone rendered 8 frames for `グリッチテキスト` in AE 2025, `compare-set`
  produced a valid report with 2 matching frames and 6 differing frames, and
  `aepslices diagnose -render-set` merged 141 profile gaps with 6 render gaps.
- Minimal recipe IR plan written; automated correction loops remain blocked.
- Minimal recipe IR implemented: `internal/recipe`, `cmd/aeprecipe`, and
  `examples/recipes/minimal-text-shape.json` compile a one-comp text/shape
  recipe into an AEP. The generated AEP passed AE 2025 render oracle.
- Recipe capability reporting now uses reusable `internal/capindex` loaded from
  `docs/capabilities.json`; `cmd/aeprecipe` reports used capabilities and
  downgrades, and refuses effect requests instead of silently dropping them.
- Recipe effect materialization now supports adding built-in effects from
  `SupportedEffects()` by compiling the base project, reopening it, then calling
  `AddEffect` on parsed layers. The `minimal-text-effect` recipe passed profile
  inspection and AE 2025 render oracle.
- Recipe effect params now support static scalar/bool/numeric-array values via
  `SetEffectParam`; the Gaussian Blur example passed profile inspection and AE
  2025 render oracle with params `25`, `2`, and `1`.
- Recipes now support embedded `expected_profile` checks. `cmd/aeprecipe
  compile` builds a profile from the exact AEP bytes it is about to write,
  reports `profile_checks`, and refuses final output on mismatch. Both recipe
  examples now carry self-contained profile contracts; the effect-contract
  output passed AE 2025 render oracle.
- Shape recipe stroke controls are implemented for static color/width/opacity.
  `expected_profile.properties[]` can assert static layer/shape property values,
  and the updated shape example passed `go test ./...`, `go vet ./...`, profile
  checks, and an AE 2025 render oracle gate.
- Additional shape detail controls are implemented for local primitive position,
  rectangle roundness, and fill opacity. The updated shape example again passed
  `go test ./...`, `go vet ./...`, profile checks, and an AE 2025 render oracle
  gate.
- First text-style recipe slice is implemented for font size, tracking, and
  paragraph justification, with `expected_profile.text_styles[]` checks. The
  updated shape/text example passed `go test ./...`, `go vet ./...`, profile
  checks, and an AE 2025 render oracle gate.
- Additional text-style recipe controls are implemented for fill color, faux
  bold/italic, apply stroke, stroke color, and stroke width, with
  `expected_profile.text_styles[]` checks. The updated shape/text example
  passed `go test ./...`, `go vet ./...`, profile checks, and an AE 2025
  render oracle gate.
- Recipe embedded profile contracts now support `expected_profile.keyframes[]`
  for property keyframe count/time/value checks. The updated shape/text example
  animates the title Position across three keyframes, passed `go test ./...`,
  `go vet ./...`, profile checks, and an AE 2025 render oracle gate.
- Recipe transform keyframes now also support `transform.opacity_keyframes[]`
  in percent authoring units, with validation and `ADBE Opacity`
  `expected_profile.keyframes[]` checks against profile unit opacity values.
  The updated shape/text example passed `go test ./...`, `go vet ./...`,
  profile checks, and an AE 2025 render oracle gate.
- Recipe transform keyframes now also support `transform.scale_keyframes[]` in
  2D percent authoring units, with validation and `ADBE Scale`
  `expected_profile.keyframes[]` checks against profile 3D unit scale values.
  The updated shape/text example passed `go test ./...`, `go vet ./...`,
  profile checks, and an AE 2025 render oracle gate.
- Recipe transform keyframes now also support `transform.rotation_keyframes[]`
  in degree authoring units, with validation and `ADBE Rotate Z`
  `expected_profile.keyframes[]` checks against profile degree values. The
  updated shape/text example passed `go test ./...`, `go vet ./...`, profile
  checks, and an AE 2025 render oracle gate.
- Recipe transform keyframes now also support
  `transform.anchor_point_keyframes[]` in 2D authoring units, with validation
  and `ADBE Anchor Point` `expected_profile.keyframes[]` checks against profile
  3D `[x,y,0]` values. The updated shape/text example passed `go test ./...`,
  `go vet ./...`, profile checks, and an AE 2025 render oracle gate.
- Recipe transform keyframes now accept optional `in_ease` / `out_ease`
  objects on position, anchor point, scale, rotation, and opacity keyframes.
  The compiler preserves the old linear path when no ease is supplied and uses
  `AddKeyframeWithEase` only for eased keyframes. Focused recipe tests reopen
  the compiled AEP and verify Bezier interpolation plus temporal ease speed /
  influence on Position and Opacity.
- Recipe transform properties now accept `transform.expressions` for position,
  anchor point, scale, rotation, and opacity. Compilation writes the base
  transform first, reopens the generated project to obtain parsed property
  backrefs, then applies `Property.SetExpression` and optional
  `Property.SetExpressionEnabled`. `expected_profile.properties[]` can now
  assert property expression source strings.
- First shape-filter recipe slice is implemented for `shape.trim` static
  start/end/offset controls. The updated shape/text example asserts
  `ADBE Vector Trim Start/End/Offset`, passed `go test ./...`, `go vet ./...`,
  profile checks, and an AE 2025 render oracle gate.
- Second shape-filter recipe slice is implemented for
  `shape.round_corners.radius`. The updated shape/text example asserts
  `ADBE Vector RoundCorner Radius`, passed `go test ./...`, `go vet ./...`,
  profile checks, and an AE 2025 render oracle gate.
- Third shape-filter recipe slice is implemented for `shape.offset_paths`
  amount, line join, miter limit, copies, and copy offset. The updated
  shape/text example asserts all five `ADBE Vector Offset ...` properties,
  passed `go test ./...`, `go vet ./...`, profile checks, and an AE 2025
  render oracle gate.
- Fourth shape-filter recipe slice is implemented for `shape.zigzag` size,
  detail, and points. The updated shape/text example asserts all three
  `ADBE Vector Zigzag ...` properties, passed `go test ./...`,
  `go vet ./...`, profile checks, and an AE 2025 render oracle gate.
- Fifth shape-filter recipe slice is implemented for
  `shape.pucker_bloat.amount`. The updated shape/text example asserts
  `ADBE Vector PuckerBloat Amount`, passed `go test ./...`,
  `go vet ./...`, profile checks, and an AE 2025 render oracle gate.
- Sixth shape-filter recipe slice is implemented for `shape.twist` angle and
  center. The updated shape/text example asserts `ADBE Vector Twist Angle` and
  `ADBE Vector Twist Center`, passed `go test ./...`, `go vet ./...`, profile
  checks, and an AE 2025 render oracle gate after hardening the generic
  aeoracle renderer to wait for and validate PNG outputs.
- Seventh shape-filter recipe slice is implemented for `shape.wiggle_paths`
  size, detail, wiggles per second, random seed, points, correlation, temporal
  phase, and spatial phase. The dedicated
  `examples/recipes/minimal-shape-wiggle-paths.json` example asserts all eight
  `ADBE Vector Roughen ...` / modulation properties, passed profile checks, and
  passed an AE 2025 render oracle gate.
- Eighth shape-filter recipe slice is implemented for `shape.wiggle_transform`
  anchor, position, scale, rotation, wiggles per second, random seed,
  correlation, temporal phase, and spatial phase. The dedicated
  `examples/recipes/minimal-shape-wiggle-transform.json` example asserts all
  nine `ADBE Vector Wiggler ...` / modulation properties, passed profile
  checks, and passed an AE 2025 render oracle gate.
- Ninth shape-filter recipe slice is implemented for `shape.repeater` copies,
  offset, order, transform anchor/position/scale/rotation, and start/end
  opacity. The dedicated `examples/recipes/minimal-shape-repeater.json`
  example asserts all nine `ADBE Vector Repeater ...` properties, passed
  profile checks, and passed an AE 2025 render oracle gate.
- Tenth shape-filter recipe slice is implemented for `shape.merge_paths.type`.
  The dedicated `examples/recipes/minimal-shape-merge-paths.json` example
  asserts `ADBE Vector Merge Type`, passed profile checks, and passed an AE
  2025 render oracle gate. Current recipe IR still has one primitive per shape
  layer; visible multi-path boolean recipes need a later multi-shape or
  vector-group structure.
- Eleventh shape recipe slice is implemented for `shape.kind: star` /
  `polygon` plus polystar points, position, rotation, inner/outer radius, and
  inner/outer roundness. The dedicated
  `examples/recipes/minimal-shape-polystar.json` example asserts all eight
  `ADBE Vector Star ...` properties, passed profile checks, and passed an AE
  2025 render oracle gate.
- Twelfth shape recipe slice is implemented for `shape.stroke.dashes` dash/gap
  support. The dedicated `examples/recipes/minimal-shape-stroke-dashes.json`
  example asserts `ADBE Vector Stroke Dash 1` and `ADBE Vector Stroke Gap 1`,
  passed profile checks, and passed an AE 2025 render oracle gate.
- Thirteenth shape recipe slice is implemented for `shape.stroke` line cap,
  line join, and miter limit. The dedicated
  `examples/recipes/minimal-shape-stroke-style.json` example asserts the three
  `ADBE Vector Stroke ...` style properties, passed profile checks, and passed
  an AE 2025 render oracle gate.
- Fourteenth shape recipe slice is implemented for `shape.stroke.taper`
  percent-mode controls. The dedicated
  `examples/recipes/minimal-shape-stroke-taper.json` example asserts all six
  `ADBE Vector Taper ...` properties, passed profile checks, and passed an AE
  2025 render oracle gate.
- Fifteenth shape recipe slice is implemented for `shape.stroke.wave`
  wavelength-mode controls. The dedicated
  `examples/recipes/minimal-shape-stroke-wave.json` example asserts all three
  `ADBE Vector Taper Wave...` / wavelength properties, passed profile checks,
  and passed an AE 2025 render oracle gate.
- First comp-level recipe settings slice is implemented for
  `comp.background_color`. The dedicated
  `examples/recipes/minimal-comp-background-color.json` example compiles a
  no-layer colored comp, passes profile count checks, and passes an AE 2025
  render oracle gate.
- Second comp-level recipe settings slice is implemented for
  `comp.motion_blur` shutter/sample settings. The dedicated
  `examples/recipes/minimal-comp-motion-blur.json` example asserts all four
  profile-visible motion blur settings and passes an AE 2025 render oracle
  gate.
- Third comp-level recipe settings slice is implemented for `comp.work_area`
  second-based start/end settings. The dedicated
  `examples/recipes/minimal-comp-work-area.json` example asserts both
  profile-visible work area fields and passes an AE 2025 render oracle gate.
- Fourth comp-level recipe settings slice is implemented for `comp.renderer`.
  The dedicated `examples/recipes/minimal-comp-renderer.json` example asserts
  the profile-visible renderer match name and passes an AE 2025 render oracle
  gate.
- Fifth comp-level recipe settings slice is implemented for
  `comp.resolution_factor`. The dedicated
  `examples/recipes/minimal-comp-resolution-factor.json` example compiles a
  half-resolution preview comp, passes compiled AEP readback, and passes an AE
  2025 render oracle gate.
- Sixth comp-level recipe settings slice is implemented for `comp.pixel_aspect`.
  The dedicated `examples/recipes/minimal-comp-pixel-aspect.json` example
  compiles a non-square pixel comp, passes compiled AEP readback, and passes an
  AE 2025 render oracle gate.
- Seventh comp-level recipe settings slice is implemented for
  `comp.display_start_time`. The dedicated
  `examples/recipes/minimal-comp-display-start-time.json` example compiles a
  shifted timeline origin comp, passes compiled AEP readback, and passes an AE
  2025 render oracle gate.
- Eighth comp-level recipe settings slice is implemented for
  `comp.frame_blending`. The dedicated
  `examples/recipes/minimal-comp-frame-blending.json` example compiles a comp
  master frame-blending switch, passes compiled AEP `cdta` flag readback, and
  passes an AE 2025 render oracle gate.
- Ninth comp-level recipe settings slice is implemented for
  `comp.hide_shy_layers`. The dedicated
  `examples/recipes/minimal-comp-hide-shy-layers.json` example compiles a comp
  master Hide Shy Layers switch, passes compiled AEP `cdta` flag readback, and
  passes an AE 2025 render oracle gate.
- Tenth comp-level recipe settings slice is implemented for
  `comp.preserve_nested_frame_rate`. The dedicated
  `examples/recipes/minimal-comp-preserve-nested-frame-rate.json` example
  compiles the nested-frame-rate preservation switch, passes compiled AEP
  `cdta` flag readback, and passes an AE 2025 render oracle gate.
- Eleventh comp-level recipe settings slice is implemented for
  `comp.preserve_nested_resolution`. The dedicated
  `examples/recipes/minimal-comp-preserve-nested-resolution.json` example
  compiles the nested-resolution preservation switch, passes compiled AEP
  `cdta` flag readback, and passes an AE 2025 render oracle gate.
- Twelfth comp-level recipe settings slice is implemented for
  `comp.motion_blur.enabled`. The dedicated
  `examples/recipes/minimal-comp-motion-blur-enabled.json` example compiles the
  comp motion-blur master switch, passes compiled AEP `cdta` flag readback, and
  passes an AE 2025 render oracle gate.

Current:
- Phase 6 render-compare loop and minimal recipe IR are implemented and proven
  with AE 2025 render gates.
- Next: broaden recipe coverage in narrow, evidence-gated slices. Strong
  candidates are remaining comp settings, shape filters, expression support,
  transform keyframe/ease where writer support exists, or more profile-visible
  text style. Do not start automated correction loops.

## Open questions

- How much capability linking can be automated from current profile paths
  before a richer path-to-capability index exists.
- What the smallest Phase 5 replication slice should be, so it exercises both
  structural and render gaps without forcing a full-project rebuild first.
- Which recipe field family should enter next: additional shape filters,
  expression support, transform keyframe/ease, or more profile-visible text
  style.
