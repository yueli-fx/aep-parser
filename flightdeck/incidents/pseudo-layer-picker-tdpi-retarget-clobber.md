---
status: active
when_to_read: implementing/debugging a pseudo or effect layer-picker / layer-reference binding that resolves to the WRONG layer (host instead of the chosen one); touching addEffectFromChunks / retargetEffectHostLayer / the effect-splice core; adding a BuildPseudoEffect control that synthesizes a tdpi; writing or reviewing a ship-gate that compares AE's 1-based layer .value/.index against the parser's 0-based Layer.Index; a pseudo/effect gate that passes for the wrong reason
applies_to: [pseudo-effect, build-pseudo-effect, layer-picker, tdpi, retarget, retargetEffectHostLayer, addEffectFromChunks, host-layer-binding, false-green, ship-gate, layer-index, zero-based, one-based, red-line-1, add-effect, ae2020, ae2025]
last_updated: 2026-06-20
resolved_by:
---

# Pseudo layer-picker tdpi clobbered by shared-core retarget; false-green gate from 0/1-based index

## Signature
- symptom: `BuildPseudoEffect layer-picker resolves to the host layer instead of the chosen layer (tdpi = host id, not the picked layer's id)`
- error_type: —
- where: `internal/serializer/mutate_effect_add.go` `addEffectFromChunks` / `retargetEffectHostLayer`; `BuildPseudoEffect`
- trigger: building a pseudo effect whose Layer control binds to a layer OTHER than the effect's host, then reading the picker back (in AE or via tdpi dump)

## Symptom / repro

A `BuildPseudoEffect` Layer-picker control bound to a non-host layer pointed at
the **host** instead. The showcase demo (Host carries the effect, picker should
target Source) wrote the picker tdpi = Host's id. `pickerprobe`: gen passes
`LayerID=14` (Source) but the written tdpi chunk = `13` (Host). The effect
header tdpi (which correctly = host) survived because retargeting it to host is
a no-op; only the picker — the one tdpi that should differ from host — was wrong.

## Root cause

Two independent bugs that cancelled each other in the ship-gate (classic
false-green, red-line-1/4):

1. **Shared-core blanket retarget.** `addEffectFromChunks` (the splice core
   behind AddEffect / ApplyPseudoEffect / BuildPseudoEffect) called
   `retargetEffectHostLayer(sspc, layer.ID)`, which rewrites **every** tdpi in
   the payload to the host. That is *correct* for AddEffect — its embedded
   templates carry the extraction fixture's foreign host id, which must be
   retargeted or AE rejects the dangling binding ("cannot find layer ID=N", see
   [[add-effect-splice-re]] Finding 5). But `BuildPseudoEffect` **synthesizes**
   its tdpi deliberately (header → host, picker → chosen layer), so the blanket
   rewrite clobbered the picker.

2. **0-based vs 1-based gate assertion.** `TestBuildPseudoEffectValueEntry`
   asserted the picker's AE-read `.value` against the parser's `Layer.Index`.
   `Layer.Index` is **0-based** (the on-disk Layr order; `parse_composition`
   passes the 0-based loop index — the old `scene_layer.go` doc comment wrongly
   said "1-based"). AE's ExtendScript `layer.index` is **1-based**. So
   AE-index = `Layer.Index + 1`.

   With both bugs present: picker → host S (AE index **1**), and the gate
   expected target T's 0-based `Index` = **1**. `1 == 1` → **PASS for the wrong
   reason** — the picker pointed at the wrong layer and the gate still went green.

## Fix

- **Move the retarget out of the shared core into `AddEffect`.** Only the
  template path has foreign tdpi; the pseudo paths own correct bindings.
  `addEffectFromChunks` is now a pure splice. `ApplyPseudoEffect` is unaffected
  (its `.ffx` transform drops all value entries → no tdpi to retarget anyway).
- **Gate expects `target2.Index + 1`** (AE's 1-based index); `Layer.Index` doc
  comment corrected to 0-based (asserted by `move_layer_test`).
- **End-to-end regression** `TestBuildPseudoEffect_LayerPickerBindsChosenLayer`
  walks the spliced layer's tdpi and asserts both a host binding AND a distinct
  chosen-layer binding exist. The isolated unit test (`synthControlValueEntry`)
  could not catch this — the clobber happens in the splice core, downstream of
  synthesis. **Lesson: a value-synth unit test does not gate a value-write; the
  integration path (synth → splice → write) needs its own assertion.**
- **Verified** AE 2020 + 2025: picker resolves to the chosen layer (T, AE index
  2); re-gated AddEffect (29-effect lib) + Wave11 (layer-ref) + all pseudo gates
  green — the retarget relocation preserves the AddEffect/layer-ref behavior.

**General lesson (false-green detection):** when a gate compares an AE-DOM index
against a parser index, the 0-based/1-based mismatch can mask a real binding
bug — make the picker target a layer whose index *differs* from the host's, and
confirm the expectation uses AE's 1-based convention.

## Cases
- 2026-06-20 first seen — BuildPseudoEffect layer-picker side-arc (continuation
  of the [[pseudo-control-label-ansi-codepage]] / pseudo value-entry work).
