---
status: active
when_to_read: implementing per-effect typed param helpers / EnsureEffectParameter; wondering why a default effect instance exposes only the -0000 param; needing AE's param-persistence rule (value!=default, not touched-flag); splicing a materialized param tdbs into an effect sspc; needing effect point-param coordinate units (fraction-of-what) or color cdat ARGB encoding; verify JSX throws "数字结果无效（除以零？）" on a log line
applies_to: [effect-params, elision, pard, part, tdbs, param-synthesis, settext, typed-param-helper, add-effect, negative-finding, control-type, point-units, color-argb, extendscript-concat-throw]
last_updated: 2026-06-12
resolved_by:
---

# Effect param elision — value-keyed persistence + synthesis-lite path

## Signature
- symptom: `SetStaticValue cannot target an effect parameter that is not in Effect.Parameters (default-elided); a default effect instance surfaces only the "<matchname>-0000" param`
- error_type: —
- where: internal/serializer parse_properties.go collectEffects / future typed-param write path
- trigger: trying to set an effect parameter AE never persisted (its value still equals the default)

## 症状/复现

A freshly added effect (AE-native or `AddEffect` template) exposes a single
`<matchname>-0000` param; the real tunable params (`-0001`…) have **no tdbs
chunk at all** in the value tdgp, so `Property.SetStaticValue` has nothing to
write into. This blocks per-effect typed param helpers — the board's
"AddEffect 参数化" item ([[add-effect-splice-re]] finding 3 records the
surface-level symptom).

## 根因 — AE persists params by VALUE≠DEFAULT, not by touched-flag

RE probe `test_data/re_effect_param_elision.jsx` (AE 2020, fixture
`re_effect_param_elision.aep`, 3 Gaussian Blur instances; dump via
`tmp_debug/probe_effects` + `dump_chunks`):

| instance | treatment | params persisted in value tdgp |
|---|---|---|
| L1 "touched" | every param `setValue(non-default)` | **all** (`-0000` + `-0001`=25 + `-0002`=2 + `-0003`=1) |
| L2 "resetdef" | every param `setValue(its default value)` | only `-0000` (**elided**) |
| L3 "untouched" | never touched | only `-0000` |

L2 is the decisive negative finding: **`setValue(default)` does NOT
materialize a param** — persistence keys off current value ≠ default at save
time. So the "extract templates after setValue(default) on every param"
strategy is dead, and any full-param template extracted from a touched
fixture would carry non-default values needing a reset pass.

## 字节结构 — what a materialized param looks like

- **parT (definitions) is always complete**, even on an untouched instance:
  every param has its `(tdmn, pard[148B])` entry (+ optional `pdnm`). Only
  the **value tdgp** elides. Defaults therefore live in-file: each `pard`
  carries last/default fields (`parse_effect_pard.go`), and in a
  never-touched instance pard's lastValue == the default. ⚠ observed gap:
  parsed `Property.DefaultValue` came back nil for these effect params —
  the pard merge (`applyPardDefs`) needs a look before relying on it.
- **A materialized param's value entry is small and self-contained** —
  `(tdmn[40B], LIST:tdbs)` where tdbs = `tdsb(4) + tdsn(display name) +
  tdb4(124) + cdat(40)` (+ `tdum/tduM` 8B min/max for scalar). **No `tdpi`**
  — the host-layer binding only exists on the always-present `-0000` stream,
  so a spliced extra param needs no retarget (cf.
  [[add-effect-splice-re]] finding 5).
- Materialized params sit in definition (match-name) order after `-0000`,
  before the `ADBE Effect Built In Params` group.

## 修法 — synthesis-lite (SHIPPED 2026-06-11, `aep.SetEffectParam`)

Mirror AE's own semantics instead of fighting them: keep the gated
default-instance AddEffect templates, and **materialize a param on demand at
set time** — `aep.SetEffectParam(layer, fx, paramMatchName, value)`
(`internal/serializer/mutate_effect_param.go`):

1. param already in `Effect.Parameters` → plain `SetStaticValue` fast path;
2. else clone the embedded per-param `(tdmn, tdbs)` template
   (`templates/effectparam_*.bin`, extracted by
   `tmp_debug/extract_effect_params`), splice into the value tdgp at
   definition (match-name) order before Built In Params / Group End,
   re-parse via `parseLeafProperty`, write the caller's value, mirror into
   `Effect.Parameters` at the same ordinal. Atomic (snapshot + warnings
   rollback).

Output equals what AE itself writes for a touched param (value ≠ default by
construction), so no default-reset pass and no re-gate of the 29
add-templates. **AE 2020 + AE 2025 ship-gate PASS**
(`TestSetEffectParam_AEShipGate_AE20{20,25}`, verify_effect_param.jsx:
all-Go-built file, GB -0001/-0002/-0003 + Drop Shadow -0004/-0005/-0006
materialized out of order, AE reads back every value on open AND after
reopening its own resave).

**Generic per-control-type templates PROVEN (same gate)**: the value stream's
tdbs shape is control-type-keyed, not param-keyed. The Drop Shadow params
have no per-param template — they materialize from the GB-extracted
scalar/boolean streams with tdmn (match-name), tdsn (display name), and
tdum/tduM (scalar min/max) patched from the host effect's own pard
definition (parT is never elided, so the metadata is always in-file). So
**any scalar / enum / boolean param of any effect is settable today** —
no per-effect extraction sweep needed. Caveats: enum's generic template is
byte-identical to the gated GB per-param one (tdmn-patched cross-effect enum
not separately AE-gated yet). pard lastValue not refreshed (cosmetic — AE
accepted without).

## 控件类型补齐 (2026-06-12) — all 8 types gated

Touch-all fixture `test_data/re_effect_param_types.aep` (AE 2020, one
expression-control effect per missing type, each -0001 touched) yielded the
remaining generic templates via `tmp_debug/extract_effect_params`: **angle /
color / 2D point / 3D point / slider** (`effectparam_adbe_*_control_0001.bin`,
also registered per-param for the 5 expression controls). Same fixture's
untouched instance fed `tmp_debug/extract_effect_lib` → `ADBE Point3D
Control` became AddEffect template #30. **AE 2020 + AE 2025 ship-gate PASS**,
13 expects each: GB×3 per-param + Drop Shadow×5 generic (cross-effect color
-0001 + angle -0003 pard-patched from the DS sspc) + 5 expression controls,
read back on open AND after AE's own resave. Materialized color stream is
byte-identical to AE-native (verified vs fixture). Only slider's tdbs
carries tdum/tduM; angle/color/point are unbounded (no patch needed).

### Finding 3 — point-param cdat units = fraction of the layer's coord space

(RE `test_data/re_effect_param_types_units.aep`: 200×100 solid + shape layer,
Point [123,45] / Point3D [123,45,67].) 2D/3D point params store cdat as
**fractions**: layers with a source item divide by the SOURCE's w/h
(solid 200×100 → [0.615, 0.45]); source-less layers (shape/text) divide by
the COMP's w/h (1920×1080 → [0.0640625, 0.041666]). **z divides by the same
space's HEIGHT** (67/100 resp. 67/1080). Mirrors the mask-coordinate
dichotomy ([[add-mask-create-re]]) except source-less uses comp fractions,
not raw pixels. SetEffectParam passes values through raw (on-disk
StaticValue units) — callers convert; the facade doc comment records this.
Color cdat = [A,R,G,B] each 0–255 (JSX [r,g,b,a] 0–1 ↔ ×255 reorder).

### Finding 4 — ExtendScript throws on `"" + colorValue` (JSX-side trap)

The first gate run FAILed with `EXC Error: 数字结果无效（除以零？）` (invalid
numeric result / divide by zero) thrown at the *log line* of the verify JSX —
string-concatenating an AE **color** property value routes the array through
valueOf → numeric conversion → throw. `value.toString()` / `value.join(",")`
/ element access / `instanceof Array` all work; only bare `"" + v` dies.
Cost: looked exactly like a data reject (bisected the whole file before
suspecting the JSX). Any verify/RE JSX logging array values must stringify
explicitly (see `str()` in verify_effect_param.jsx).

**Repeat victim — `comp.resolutionFactor` (2026-06-14).** comp-settings showcase
logged `"" + comp.resolutionFactor` ([2,2] array) → same `数字结果无效（除以零？）`,
and the symptom got mis-recorded as "AE 侧除零深坑，SetResolutionFactor 待 RE"
(propagated into cockpit RE-candidate + showcase exclusion). It was never a real
issue: numeric-index readback (`rf[0]+"x"+rf[1]`) works; AE `resolutionFactor=[2,2]`
writes/reads fine; AE-native cdta X@0x00/Y@0x02 uint16 BE == our writer; Go-built
SetResolutionFactor(2,2) → AE2020+2025 DOM 2x2. **Lesson: an "AE throws" symptom
observed only through a `"" + arrayValue` log line is presumed-trap until re-tested
with numeric indexing — don't escalate it to an RE candidate.**

## Finding 2 — cross-version DEFAULT drift re-elides on resave (not a bug)

AE 2025 flipped Gaussian Blur "Repeat Edge Pixels" (`-0003`) default false →
true (probe logs: AE 2020 `default=0`, AE 2025 `default=1`; AE 2025 itself
also refuses to persist `setValue(true)` on it). Consequence: a param
materialized with a value that equals the OPENING AE version's default is
re-elided by that version's resave — the file drops the stream but **the
effective value survives via the default**. Gate assertions must therefore
check the reopened resave's *read-back value* (verify JSX second-open pass),
not the param stream's presence. Corollary for parsers: "param absent" can
mean different values under different AE versions — never assume absent ==
our hardcoded default across versions.

## Cases
- 2026-06-11 首次（board「AddEffect 参数化」可行性 RE → 同日 synthesis-lite
  ship；fixture + probe 落 `test_data/re_effect_param_elision.*`（manifest
  已登记）+ `re_effect_param_elision_2025.*`（AE 2025 默认值漂移对照））
- 2026-06-12 控件类型补齐（angle/color/2D/3D/slider 泛型模板 + Point3D
  Control 扩库；fixtures `re_effect_param_types.*` +
  `re_effect_param_types_units.*` 已登记 manifest；bisect 工具
  `tmp_debug/ge_setparam_bisect` + probe JSX `test_data/re_setparam_bisect.jsx`
  留存可复用）
