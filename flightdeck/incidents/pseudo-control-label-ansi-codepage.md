---
status: active
summary: AE reads a pseudo control's label from the pard @0x10 name in the system ANSI codepage; value-entry tdsn does NOT override it (gate-disproven).
when_to_read: building/labeling pseudo-effect controls, especially CJK labels via BuildPseudoEffect
applies_to: [pseudo-effect, build-pseudo-effect, cjk, i18n, pard]
last_updated: 2026-06-20
resolved_by:
---

# Pseudo control labels are bound to AE's system ANSI codepage (CJK not portable)

## Signature
- symptom: pseudo-effect control label reads back mojibake in AE (e.g. `勾选`→`name.codes=37717,40515,8364,63`, the UTF-8 bytes misread as the system ANSI codepage; 8364=€ ⇒ cp1252)
- error_type: —
- where: serializer BuildPseudoEffect / synthControlPard (pard @0x10 name)
- trigger: BuildPseudoEffect with a non-ASCII (CJK) control Name, opened on an AE whose system ANSI codepage ≠ the label's encoding

## Symptom / repro
`BuildPseudoEffect` with CJK control names (`{Kind: PseudoCheckbox, Name: "勾选"}`). On a
ship-gate machine whose Windows system codepage is **Western (cp1252)**, AE reads the
control label back as garbage. The **effect display name** (e.g. "中文效果") round-trips
fine — that one comes from the value-group `tdsn` (a `Utf8` sub-record, byte-length, exact).

## Root cause
AE sources a control's Effect-Controls **label** from the **pard `@0x10` name** (32 bytes),
which it decodes in the **viewing machine's system ANSI codepage** — NOT UTF-8, NOT the
value-entry `tdsn`. RE proof: AE's own Pseudo Effect Maker writes the pard name as GBK
(rich-demo color pard `@0x10 = d1d5c9ab` = "颜色" in GBK), and on a Chinese (GBK) Windows it
displays correctly *because the machine's codepage is GBK*.

**Disproven hypothesis** (gate, 2026-06-20): we hoped the per-control **value-entry `tdsn`**
(Utf8, like the effect name) would override the label. We synthesized non-elided value
entries (`tdsb + tdsn(Utf8) + tdb4 + cdat`, RE'd per-kind bytes) for checkbox/color/point —
**AE accepted them** (the value-entry *mechanism* works), but the control label still came
from the pard name (cp1252 mojibake). So the value-entry tdsn does **not** drive the label.

## Fix
No portable fix — this is an AE architecture limit. **Implemented (2026-06-20):**
`synthPard`/`pardNameBytes` **GBK-encode** the pard `@0x10` name (GBK is an ASCII superset,
so ASCII labels are byte-identical). For simplified-Chinese labels the bytes are
**byte-identical to AE's native Pseudo Effect Maker output** (unit-proven:
`pardNameBytes("颜色") == d1d5c9ab`, the rich-demo color pard's bytes) and display correctly
on a GBK Windows. Costs the `golang.org/x/text` dep + a simplified-Chinese locale assumption,
and is **not ship-gateable on a Western-codepage machine** — byte-equivalence to AE-authored
output is the evidence (see `TestPardNameBytes_GBKMatchesAENative`). A non-GBK rune that GBK
can't encode falls back to raw UTF-8 (mojibakes, but nothing is dropped). On a non-GBK
viewing system CJK still mojibakes — inherent to AE.

The effect **display name** has no such limit (value-group tdsn / Utf8) — `ApplyPseudoEffectNamed`
and `BuildPseudoEffect`'s `displayName` already round-trip CJK (gate-green).

Also disproven en route: the per-control **value-entry tdsn** does NOT drive the label (AE
accepts synthesized value entries but still labels from the pard name) — that spike was
reverted.

## Cases
- 2026-06-20 first seen — BuildPseudoEffect CJK control-label gate on cp1252 ship-gate machine.
