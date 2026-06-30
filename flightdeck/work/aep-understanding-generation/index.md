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
- `comp-recipe-execution-strategy.md` — object-level comp recipe/profile
  execution strategy and field matrix.
- `layer-recipe-execution-strategy.md` — object-level layer recipe/profile
  execution strategy and field matrix.
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
- Recipe profiles now expose property expression enabled state, and
  `expected_profile` can assert it for both layer properties and effect
  parameters.
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
- Recipe effect params now accept optional `expression` objects after their
  `value` materializes the parameter with `SetEffectParam`. The compiler writes
  `Property.SetExpression` and optional `Property.SetExpressionEnabled` on the
  returned effect-param property, and `expected_profile.effects[].params[]` can
  assert expression source strings.
- Recipe effect params now support scalar `keyframes[]` through
  `AnimateEffectParam`. `minimal-effect-param-keyframes.json` asserts Gaussian
  Blur blurriness keyframes through `expected_profile.keyframes[]`.
- Recipe effect param `keyframes[]` now also support 2/3/4 component vector
  values through `AnimateEffectParamVec`. The field remains a single
  scalar-or-vector keyframe list, and
  `minimal-effect-param-vector-keyframes.json` asserts Point Control keyframes
  through `expected_profile.keyframes[]`.
- Recipe effect params now support layer-reference parameters through
  `target_layer` and `SetEffectLayerParam`. Profile output exposes
  `properties[].layer_ref`, and `expected_profile.effects[].params[]` can
  assert the target layer name. `minimal-effect-layer-param.json` covers the
  Set Matte layer picker.
- Recipe comp settings now support `draft_3d`. Profile composition output now
  exposes `draft_3d` from the cdta flag, and `expected_profile.draft_3d` can
  assert it. This remains a roundtrip/profile contract because AE 2025 DOM
  readback does not reliably reflect the Draft 3D preview switch.
- Comp-level recipe/profile work now has an object-level execution strategy:
  field order, evidence level, profile target, test shape, verification, and
  commit rules are captured in `comp-recipe-execution-strategy.md`.
- Comp display settings now have object-level profile contracts:
  `background_color`, `resolution_factor`, `pixel_aspect`, and
  `display_start_time` are exposed on `profile.Composition`, asserted through
  `expected_profile`, and covered by `minimal-comp-object-profile.json`.
- Comp flag settings now have object-level profile contracts:
  `frame_blending`, `hide_shy_layers`, `preserve_nested_frame_rate`,
  `preserve_nested_resolution`, and `motion_blur.enabled` are exposed through
  `profile.Composition`, asserted through `expected_profile`, and covered by
  `minimal-comp-flag-profile.json`.
- Comp item metadata now has an object-level profile contract: `label` and
  `comment` are exposed directly on `profile.Composition`, asserted through
  `expected_profile`, and covered by `minimal-comp-metadata-profile.json`.
- The consolidated `minimal-comp-object-profile.json` now exercises the
  completed comp display, flag, motion-blur, work-area, renderer, and metadata
  profile contracts together as a no-layer object-level baseline.
- The consolidated comp object baseline now also asserts constructor-backed
  identity/geometry/timing fields through `expected_profile`: `name`, `width`,
  `height`, `frame_rate`, and `duration`.
- Standalone comp recipes for background color, label, comment, resolution
  factor, pixel aspect, display start time, frame blending, hide shy layers,
  preserve nested frame rate, preserve nested resolution, and motion blur
  enabled now assert their own profile-visible fields instead of count-only
  smoke checks.
- Added `scripts/verify_recipe_profiles.ps1` as the one-command recipe profile
  verifier. It builds `cmd/aeprecipe` once, validates and compiles all selected
  recipes, summarizes `profile_checks` coverage/failures, and exits non-zero on
  validation, compile, invalid-report, or profile-check failures.
- Layer object profile contracts now support `expected_profile.layers[]` for
  layer identity, metadata, timing, and flags. `profile.Layer` exposes
  `label`/`comment`, and `minimal-layer-object-profile.json` asserts one text
  layer's name, type, label, comment, start/in/out times, and common flags as a
  single object-level baseline.
- Layer object profile contracts now also include `quality`, `blending_mode`,
  and `auto_orient` as recipe-friendly strings on `profile.Layer` and
  `expected_profile.layers[]`; `minimal-layer-object-profile.json` covers them
  in the same baseline.
- Parent refs now have an object-level recipe/profile contract:
  `expected_profile.layers[].parent` asserts `profile.Layer.parent_ref.name`,
  and `minimal-layer-parent.json` covers the two-layer parent relationship.
- Classic track matte mode is now recipe-owned through `layers[].track_matte`.
  `expected_profile.layers[].track_matte` asserts the profile matte mode and
  `expected_profile.layers[].matte` asserts the inferred classic matte source;
  `minimal-layer-track-matte.json` covers the two-layer alpha matte baseline.
- Layer render/sampling flags now flow through `scene.JSONLayer`,
  `profile.Layer.flags`, and `expected_profile.layers[].flags`.
  `minimal-layer-object-profile.json` asserts `frame_blend_pixel_motion` and
  `sampling_bicubic` alongside the existing object-level layer flags.
- Camera and light base recipes now assert layer object identity through
  `expected_profile.layers[]`; option values remain covered by their
  `expected_profile.properties[]` recipe family.
- Camera option coverage now has a consolidated object/profile baseline:
  `minimal-camera-object-profile.json` sets zoom, DOF/focus/aperture/blur, and
  iris controls together and asserts all profile-visible properties.
- Camera single-option recipes now also assert camera/text layer identity
  through `expected_profile.layers[]` alongside their property checks.
- Light option coverage now has a consolidated object/profile baseline:
  `minimal-light-object-profile.json` sets light kind, intensity, color,
  shadow, falloff, and cone controls together and asserts all profile-visible
  layer/property fields.
- Light single-option recipes now also assert light/text layer identity through
  `expected_profile.layers[]` alongside their property checks.
- The standalone `minimal-light-kind.json` recipe now also asserts light layer
  identity and `expected_profile.layers[].light_kind`, clearing the remaining
  count-only recipe profile contract.
- Light source refs now have a profile contract: `profile.Layer` exposes
  `light_source_ref`, and `minimal-light-source.json` asserts
  `expected_profile.layers[].light_source`.
- Generic layer source refs now have a profile contract:
  `minimal-layer-source-ref.json` asserts a solid layer's footage
  `source_ref` through `expected_profile.layers[].source` and `source_kind`.
- Layer object flags now include `markers_locked`; the consolidated
  `minimal-layer-object-profile.json` asserts it through
  `expected_profile.layers[].flags.markers_locked`.
- Layer object timing baseline now also asserts parsed `duration` and
  `stretch`; with `in_point: 0.1` and `out_point: 0.9`, the expected parsed
  span is `duration: 0.8`.
- Null and adjustment base layer recipes now assert layer object identity and
  type flags through `expected_profile.layers[]`; `minimal-null-layer.json`
  also checks the child parent ref, while `minimal-adjustment-layer.json`
  keeps its effect contract.
- Standalone layer recipes for label, comment, timing, common switches, motion
  blur, shy, advanced switches, quality/blending, auto-orient, and null flag
  now assert object-level `expected_profile.layers[]` contracts instead of
  count-only smoke checks.
- All recipe examples that author layers now include
  `expected_profile.layers[]` object identity checks. A recipe test guard fails
  any future authored-layer example that omits the layer object contract or
  asserts fewer layer entries than it authored.
- Recipe examples that author transform keyframes now must include
  `expected_profile.keyframes[]` contracts. The auto-orient recipe now asserts
  the Position keyframes it uses for `along_path`, preventing animation streams
  from becoming visual-only setup with no profile check.
- Recipe examples with nested content now have a profile-family guard:
  shape/camera/light property content requires `expected_profile.properties[]`,
  text style content requires `expected_profile.text_styles[]`, and effect
  content requires `expected_profile.effects[]`.
- Shape filter coverage now has dedicated minimal recipes for trim paths, round
  corners, offset paths, zigzag, pucker/bloat, and twist. The recipes assert
  their parsed filter properties directly instead of relying only on the larger
  `minimal-text-shape.json` baseline.
- First mask recipe slice is implemented for static vector masks on non-camera
  / non-light layers. `layers[].masks[]` now compiles through the existing
  Reopen + `AddMask` path, supports mode and inverted controls, and
  `expected_profile.masks[]` asserts mask name, mode, inverted, closed, and
  vertex count through the parsed profile.
- Mask recipe/profile coverage now also includes byte-level mask options:
  locked state, timeline color, motion-blur override, and feather-falloff mode.
  `profile.Mask` and scene JSON expose those parsed values, and
  `minimal-layer-mask.json` asserts them through `expected_profile.masks[]`.
- Mask recipe/profile coverage now includes mask property-style options:
  opacity, feather, and expansion. The same `minimal-layer-mask.json` baseline
  compiles these through stable mask setters and asserts their parsed profile
  values, raising the recipe profile verifier coverage to 20 checks for that
  recipe.
- Mask path keyframes are now recipe/profile covered. `layers[].masks[]`
  accepts `path_keyframes[]`, compiles them through `SetMaskPathKeyframes`, and
  `expected_profile.masks[].path_keyframes[]` asserts keyframe count, time, and
  vertex count through `minimal-layer-mask.json`, raising that recipe's profile
  coverage to 25 checks.
- Stable profile output now includes comp/layer marker payloads with parsed
  time, duration, label, comment, chapter, URL, frame target, cue-point name,
  paths, and evidence. `re_compmarker.aep` covers comp-marker profile output;
  recipe marker authoring remains out of scope until the AddMarker seed-template
  boundary is designed.
- Stable profile output now includes comp guide payloads with parsed
  orientation, position, paths, and evidence. `guides.aep` covers the
  multi-guide comp profile output; guide authoring remains out of recipe scope.
- Stable profile output now includes Essential Graphics template names and
  controller summaries with parsed name, type, UUID, paths, and evidence.
  `eg_multiple_controllers.aep` covers multi-controller profile output; EG
  authoring remains out of recipe scope.
- Stable profile output now includes render queue item summaries with parsed
  comp name, status, label/comment, time span, output-module count, paths, and
  evidence. `rq_numitems_1.aep` covers the queue summary; full render/output
  module settings remain detail-only.
- Render queue profile summaries now include output module name/file-template
  entries with paths and evidence, still excluding full output module settings
  and format options.
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
- Explicit AE2025 source matte recipe support is implemented. Recipes can set
  `project.target_version: "AE2025"` and `layers[].matte` with a non-none
  `layers[].track_matte`; `minimal-layer-explicit-matte.json` asserts both
  the matte mode and profile `matte_ref`, and the verifier passes for that
  example.
- First text animator recipe slice is implemented for static opacity Range
  Selector animators through `layers[].text_animators[]`. The dedicated
  `examples/recipes/minimal-text-animator-opacity.json` example compiles an
  opacity text animator, asserts the `ADBE Text Opacity` profile property, and
  passes recipe profile verification.
- Text animator position recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-position.json` example compiles a 3D
  position text animator, asserts the `ADBE Text Position 3D` profile property,
  and passes recipe profile verification.
- Text animator scale recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-scale.json` example compiles a 3D
  scale text animator, asserts the `ADBE Text Scale 3D` profile property, and
  passes recipe profile verification.
- Text animator rotation recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-rotation.json` example compiles a
  rotation text animator, asserts the `ADBE Text Rotation` profile property,
  and passes recipe profile verification.
- Text animator fill-color recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-color.json` example compiles a fill
  color text animator and asserts the `ADBE Text Fill Color` profile property.
- Text animator tracking recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-tracking.json` example compiles a
  tracking text animator and asserts the `ADBE Text Tracking Amount` profile
  property.
- Text animator character-offset recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-character-offset.json` example
  compiles a character offset text animator and asserts the
  `ADBE Text Character Offset` profile property.
- Text animator fill-opacity recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-fill-opacity.json` example compiles a
  fill opacity text animator and asserts the `ADBE Text Fill Opacity` profile
  property.
- Text animator stroke-opacity recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-stroke-opacity.json` example compiles
  a stroke opacity text animator and asserts the `ADBE Text Stroke Opacity`
  profile property.
- Text animator stroke-width recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-stroke-width.json` example compiles a
  stroke width text animator and asserts the `ADBE Text Stroke Width` profile
  property.
- Text animator skew recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-skew.json` example compiles a skew
  text animator and asserts the `ADBE Text Skew` profile property.
- Text animator rotation-x/y recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-rotation-x.json` and
  `examples/recipes/minimal-text-animator-rotation-y.json` examples compile 3D
  rotation text animators and assert the `ADBE Text Rotation X/Y` profile
  properties.
- Text animator stroke-color recipe support is implemented. The dedicated
  `examples/recipes/minimal-text-animator-stroke-color.json` example compiles a
  stroke color text animator and asserts the `ADBE Text Stroke Color` profile
  property.
- Text animator Range Selector Offset keyframe recipe support is implemented.
  The dedicated
  `examples/recipes/minimal-text-animator-range-offset.json` example compiles
  `range_offset_keyframes`, asserts `ADBE Text Percent Offset` keyframes through
  `expected_profile.keyframes[]`, and passes recipe profile verification.
- Text animator opacity value-keyframe recipe support is implemented. The
  dedicated
  `examples/recipes/minimal-text-animator-opacity-value-keyframes.json` example
  compiles `value_keyframes` and asserts `ADBE Text Opacity` keyframes through
  `expected_profile.keyframes[]`.
- Text animator scalar value-keyframe recipe support now also covers rotation,
  tracking, and character offset. The dedicated `*-value-keyframes.json`
  examples assert `ADBE Text Rotation`, `ADBE Text Tracking Amount`, and
  `ADBE Text Character Offset` keyframes through `expected_profile.keyframes[]`.
- Text animator vector value-keyframe recipe support is implemented for
  position, scale, and fill color. `value_keyframes[].value` now accepts a
  number or numeric array according to the selected text animator property.

Current:
- Phase 6 render-compare loop and minimal recipe IR are implemented and proven
  with AE 2025 render gates.
- Next: broaden recipe coverage in narrow, evidence-gated slices. Strong
  candidates are remaining comp settings or the next recipe-owned family with
  an existing parser/profile contract. Transform keyframe coverage now has a
  complete five-channel baseline in `minimal-transform-keyframes.json`, plus
  the existing ease baseline. Text style coverage now has a dedicated
  `minimal-text-style.json` baseline instead of relying only on the larger
  text/shape recipe. Use
  `scripts/verify_recipe_profiles.ps1` for whole-recipe
  validation/compile/profile-check coverage instead of manual per-field
  spot-checking. For comp work, follow `comp-recipe-execution-strategy.md`
  instead of asking for per-field direction; standalone comp examples now carry
  field-level profile checks. For layer work, follow
  `layer-recipe-execution-strategy.md`; explicit AE2025 source mattes now have
  a target-version-gated recipe slice, while the consolidated layer object
  baseline includes render/sampling flags, base null/adjustment and standalone
  layer recipes assert object-level layer contracts, every authored-layer
  recipe now has `expected_profile.layers[]`, and camera/light options now have
  consolidated baselines for profile-visible properties. Do not start automated
  correction loops.

## Open questions

- How much capability linking can be automated from current profile paths
  before a richer path-to-capability index exists.
- What the smallest Phase 5 replication slice should be, so it exercises both
  structural and render gaps without forcing a full-project rebuild first.
- Which recipe field family should enter next: a new recipe-owned family with
  a proven parser/profile surface, or a comp setting still missing a dedicated
  recipe/profile slice.
