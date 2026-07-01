# History — aep-understanding-generation

This file holds detailed progress history moved out of index.md so the topic index stays usable as a recovery board.

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
  `scripts/ae-worker/aeoracle_render.jsx` for sentinel frame planning, render sidecars,
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
  `cmd/aeoracle render` and `scripts/ae-worker/aeoracle_render.jsx`.
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
- Text-style recipe/profile coverage now also includes run auto-leading,
  leading, baseline shift, horizontal scale, vertical scale, and tsume. Profile
  `TextStyleRun` preserves these parsed run fields, `minimal-text-style.json`
  asserts them through `expected_profile.text_styles[]`, and whole-recipe
  profile verification passes with 127 recipes and 161 covered profile paths.
- Text-style recipe/profile coverage now also includes the run `no_break` and
  `stroke_over_fill` switches. Scene JSON and profile output preserve both
  parsed flags, `minimal-text-style.json` asserts them through
  `expected_profile.text_styles[]`, and whole-recipe profile verification
  passes with 127 recipes and 163 covered profile paths.
- Text-style recipe/profile coverage now also includes run enum options:
  `caps_option`, `baseline_option`, `auto_kern_type`, `line_join_type`, and
  `digit_set`. Recipe/profile use stable snake_case values, scene JSON exports
  the same contract, `minimal-text-style.json` asserts all five fields, and
  whole-recipe profile verification passes with 127 recipes and 168 covered
  profile paths.
- Text-style recipe/profile coverage now also includes paragraph controls:
  first/start/end indent, space before/after, auto-hyphenation, leading type,
  hanging Roman punctuation, and paragraph direction. Scene JSON and profile
  output preserve the parsed paragraph fields, `minimal-text-style.json`
  asserts them through `expected_profile.text_styles[]`, and whole-recipe
  profile verification passes with 127 recipes and 177 covered profile paths.
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
- Recipe effect params now support first-slice Essential Graphics exposure
  through `essential_graphics` and `AddEssentialProperty`. The supported recipe
  shape is static `value` plus optional controller name; keyframes, expressions,
  and layer-reference params remain outside this slice. `expected_profile`
  can assert controller name/type through `essential_graphics[]`, covered by
  `minimal-essential-graphics-controller.json`,
  `minimal-essential-graphics-checkbox-controller.json`, and
  `minimal-essential-graphics-color-controller.json`.
- Recipe comp settings now support `draft_3d`. Profile composition output now
  exposes `draft_3d` from the cdta flag, and `expected_profile.draft_3d` can
  assert it. This remains a roundtrip/profile contract because AE 2025 DOM
  readback does not reliably reflect the Draft 3D preview switch.
- Recipe comp settings now support `motion_graphics_template_name` via
  `Composition.SetMotionGraphicsTemplateName`. Profile composition output
  already exposes the parsed Essential Graphics template name, and
  `expected_profile.motion_graphics_template_name` asserts it. The minimal
  coverage recipe is `examples/recipes/minimal-motion-graphics-template-name.json`.
- Recipe project settings now support `project.bits_per_channel` via
  `Project.SetBitsPerChannel`. Profile metadata already exposes
  `meta.bits_per_channel`, `expected_profile.bits_per_channel` asserts the
  generated project color depth, and
  `examples/recipes/minimal-project-bits-per-channel.json` covers the field.
- Recipe project linear color flags now support `project.linear_blending` and
  `project.linearize_working_space` via their Project setters. Profile metadata
  exposes `meta.linear_blending` and `meta.linearize_working_space`,
  `expected_profile` asserts both flags, and
  `examples/recipes/minimal-project-linear-color.json` covers the pair.
- Recipe project display settings now support `time_display_type`,
  `frames_count_type`, `frames_use_feet_frames`, `feet_frames_film_type`, and
  `footage_timecode_display_start_type`. Profile metadata exposes stable
  snake-case values for the same settings, `expected_profile` asserts them,
  and `examples/recipes/minimal-project-display-settings.json` covers the group.
- Recipe project display scalar settings now also support
  `timecode_default_base` and `transparency_grid_thumbnails`; the existing
  display settings recipe asserts both through profile metadata.
- Recipe project preferences now support `expression_engine`,
  `audio_sample_rate`, `working_gamma`, and
  `compensate_for_scene_referred_profiles`. Profile metadata exposes the same
  values, `expected_profile` asserts them, and
  `examples/recipes/minimal-project-preferences.json` covers the group.
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
- Added `go run ./cmd/aepverify recipe-profiles` as the one-command recipe profile
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
- Technique Facts v1 is implemented as `internal/technique` plus
  `cmd/aeptechnique`. It derives comp/layer/effect/text-animator/shape-operator,
  dependency, and unknown facts from `profile.Profile` only, and the CLI emits
  stable JSON for an input `.aep`.
- Technique Portrait v1 is implemented over facts. `technique.BuildPortrait`
  produces fingerprint, signal-layer, mechanism, graph, unknown, and conservative
  technique-hint summaries; `cmd/aeptechnique -mode portrait` and `-portrait`
  emit the portrait JSON.
- `cmd/aeptechnique` corpus mode is implemented:
  `-in <dir-or-file> -mode portrait -corpus -recursive [-limit N]` emits
  deterministic JSONL records, one per `.aep`, and continues across per-file
  build failures. `-summary` emits aggregate counts instead of JSONL.
  `-summary-out <path>` can now write the same aggregate summary while JSONL
  records are written to `-out`, so large explain-mode report runs parse each
  project only once.

Current:
- Phase 6 render-compare loop and minimal recipe IR are implemented and proven
  with AE 2025 render gates.
- `cmd/aeptechnique -in <file.aep>` now provides the first machine-readable
  AE-understanding layer for downstream project explanation, corpus search, and
  later technique extraction. It is descriptive only; recipe generation and
  natural-language interpretation remain later layers.
- `cmd/aeptechnique -mode portrait -in <file.aep>` now provides a compact
  corpus-facing project image suitable for comparing many projects without
  retaining the full fact dump.
- `cmd/aeptechnique -mode explain` now adds deterministic project explanations
  with recreation readiness, archetypes, repeated pattern catalog entries,
  deterministic recreation steps, representative projects, mechanism profiles,
  pattern-level recreation-step profiles, and plugin-effect blockers.
- `go run ./cmd/aepselfhost technique-report` is the self-hosted corpus dashboard.
  It runs explain-mode corpus analysis once with `-summary-out`, then writes
  `summary.json`, `corpus.jsonl`, `digest.json`, `learning.md`, `projects.csv`,
  `patterns.csv`, `study_queue.csv`, `errors.csv`, `manifest.json`, `report.md`,
  and self-contained `report.html` under the selected output directory.
  Per-file parse/profile errors no longer abort the report; they remain visible
  as error records while successful projects still feed the learning artifacts.
- `tmp\technique_samples_report` currently demonstrates the 90-project
  `data\samples` corpus: 90 parsed projects, 0 errors, 4 repeated pattern
  families, and a compact `learning.md` for selecting high-signal reference
  projects before reading the full JSONL.
- `report.html` now links all generated artifacts, shows scan timing, supports
  path/readiness/pattern/effect/plugin filtering, and displays per-project top
  effects, plugin effects, shape families, text animator mechanisms, and
  deterministic recreation steps. Pattern playbooks now include the common
  recreation-step distribution for each repeated technique pattern, and the
  study queue ranks projects by readiness and signal density.
- `go run ./cmd/aepselfhost compare-reports` compares two generated report
  directories and writes `compare.json` / `compare.md` for scalar totals and
  count-map deltas across readiness, patterns, effects, plugin effects, shapes,
  text animators, layer roles, and graph edges.
- `go run ./cmd/aepselfhost verify` is the one-command self-hosted
  acceptance gate. It runs the technique Go tests, full sample report, partial
  error report, self-compare, partial-to-full compare, and writes
  `acceptance.json` / `acceptance.md`.
- Versioned AEP migration is now scoped as a separate product line from normal
  recipe generation. `versioned-aep-migration-spec.md` defines assess/convert/
  verify modes, version capability ledger semantics, and preservation/loss
  reporting for AE2020/AE2022/AE2025 targets.
- Versioned AEP migration assess slice is implemented as `internal/aepmigrate`
  and `cmd/aepmigrate assess`. It detects/normalizes source and target version
  labels, builds a stable profile, reports first-slice compatibility entries,
  and blocks AE2025 explicit matte downgrades to AE2020/AE2022.
- Versioned AEP migration convert first slice is implemented as
  `cmd/aepmigrate convert`. It retargets no-layer composition skeleton projects
  to AE2020/AE2022/AE2025, preserving comp name/size/frame-rate/duration through
  target-version writers. It now also preserves stable profile-visible comp
  settings for no-layer comps: background color, resolution factor, pixel
  aspect, display start time, work area, comp flags, motion-blur settings,
  renderer, Motion Graphics template name, label, and comment. Source projects
  with unsupported layers are intentionally blocked before output so conversion
  cannot silently drop layer content. The first layer-bearing slices are default
  null layers, default solid layers, default adjustment layers, default camera
  layers, default light layers, default text layers, default empty shape layers,
  single rect+fill shape layers, and default precomp layers. Default
  null/adjustment/camera/light
  conversion preserves the source profile's visible default transform surface
  when needed; default solid conversion uses profile-visible footage item
  details for source dimensions/color; default text conversion uses
  `aep.NewTextLayer` + `Layer.SetText` and relies on profile diff for text
  document/style fidelity; default empty shape conversion uses `aep.NewShapeLayer`;
  single parametric graphic/filter shape conversion rebuilds the parametric
  primitive, including rect, ellipse, and polystar fields, plus Trim Paths, Round
  Corners radius, Offset Paths fields, ZigZag
  fields, Pucker & Bloat amount, Twist angle/center, Wiggle Paths fields,
  Wiggle Transform fields, Repeater fields, Merge Paths type, Gradient Fill
  stops/type/ramp, fill, stroke, one Dash 1/Gap 1 pair, percent-mode stroke taper, and
  wavelength-mode stroke wave from
  stable profile properties while still excluding stroke offset, additional
  dash/gap pairs, wave units/cycles, gradient stroke, and other filter shape content;
  static layer transform reconstruction now preserves
  profile-visible Anchor/Position/Scale/Rotation/Opacity values while converting
  profile scale/opacity unit values back to writer percent units; layer timing
  reconstruction now preserves start_time/in_point/out_point/stretch for all
  supported layer paths; supported non-shape layer reconstruction now preserves
  layer label/comment metadata, text-layer common switches/quality/blend,
  advanced text-layer switches, and same-comp text parent refs; default precomp
  conversion rebuilds all target comps before resolving composition source refs
  into recreated target comps. Go-writer
  and recipe fixtures for the supported migration slices plus the
  static-transform text and layer-timing fixtures now pass profile diff.
  Successful convert reopens the target and
  runs source-vs-target `profilediff`; unexpected profile diffs are recorded in
  the migration report and block success. AE2025 open smoke has passed for the
  converted no-layer comp object profile output and recipe default null, solid,
  adjustment, camera, light, text, static-transform text, layer timing, empty
  shape, rect+fill shape, rect+stroke shape, rect+dashed-stroke shape,
  ellipse+tapered-stroke shape, ellipse+waved-stroke shape, gradient-fill
  shape, trim-paths shape, round-corners shape, offset-paths shape, zigzag
  shape, pucker-bloat shape, twist shape, wiggle-paths shape,
  wiggle-transform shape, repeater shape, merge-paths shape, polystar shape,
  precomp, layer label/comment, text common-switch/quality/blend, text advanced
  switch, and text-parent outputs.
  `cmd/aepmigrate convert -ae-open -ae <AfterFX.exe>` writes AE open gate status
  into the migration report.
- Versioned AEP migration matrix gate is implemented as
  `cmd/aepmigrate matrix`. It batches recipe fixtures across source writer
  labels and target writer labels, writes `matrix.json` plus per-case source,
  target, compile report, and convert report files, and can auto-discover
  After Effects 2020 through 2025 hosts under an install root such as
  `E:\adobe`. Current writer targets remain AE2020/AE2022/AE2025; AE2021,
  AE2023, and AE2024 host installs are recorded for visibility but requested
  source/target writer cases are marked `skipped` until native writer templates
  exist. The CLI now defaults `-ae-open` to `-max-ae-open-cases 25` so broad
  recipe matrices cannot accidentally launch hundreds of AE open gates; set the
  cap higher or `0` only for an intentional long run. This replaces manual
  one-fixture-at-a-time migration smoke runs for supported slices.
- `profilediff` now compares profile-visible layer metadata, quality/blending/
  auto-orient/light kind, all current layer flags, stable layer refs by
  name/index, layer properties, expression status, layer-ref params, keyframe
  interpolation/tangent/ease details, and footage item details for solid source
  dimensions/color. This is a validation prerequisite for opening layer-bearing
  migration slices; it does not by itself claim broad layer conversion support.
- Recipe IR coverage for the current comp and layer strategy matrices is
  complete: comp settings have object-level and field-level profile checks,
  authored-layer examples carry `expected_profile.layers[]`, camera/light
  options have consolidated baselines, transform keyframes have a five-channel
  baseline plus the dedicated ease baseline, and text style has a dedicated
  `minimal-text-style.json` baseline instead of relying only on the larger
  text/shape recipe. Use
  `go run ./cmd/aepverify recipe-profiles` for whole-recipe
- validation/compile/profile-check coverage instead of manual per-field
  spot-checking. For any new comp or layer work, first update the relevant
  execution strategy with the new writer/profile evidence boundary. Do not
  start automated correction loops.
