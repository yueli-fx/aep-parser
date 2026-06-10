---
status: active
when_to_read: implementing per-effect typed param helpers / EnsureEffectParameter; wondering why a default effect instance exposes only the -0000 param; needing AE's param-persistence rule (value!=default, not touched-flag); splicing a materialized param tdbs into an effect sspc
applies_to: [effect-params, elision, pard, part, tdbs, param-synthesis, settext, typed-param-helper, add-effect, negative-finding]
last_updated: 2026-06-11
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

## 修法 — synthesis-lite design (next step, not yet built)

Mirror AE's own semantics instead of fighting them: keep the gated
default-instance AddEffect templates, and **materialize a param on demand at
set time** (`EnsureEffectParameter(fx, matchName)` or inside a typed
setter):

1. param already in `Effect.Parameters` → done;
2. else clone an embedded per-param `(tdmn, tdbs)` template, splice into the
   value tdgp at definition order, re-parse, set the user's value.

Output equals what AE itself writes for a touched param (value ≠ default by
construction — the user is setting a value), so no default-reset pass and no
re-gate of the 29 add-templates. Open design points: per-param templates
(extract from touched fixtures, ~5-10/effect) vs **generic per-control-type
tdbs template** (~7 total — needs a gate to prove tdb4 flags are not
param-specific); insertion-order strictness; whether `pard` lastValue should
also be refreshed (AE updates it on save, but it is metadata — likely
cosmetic).

## Cases
- 2026-06-11 首次（board「AddEffect 参数化」可行性 RE；fixture + probe 落
  `test_data/re_effect_param_elision.*`，manifest 已登记）
