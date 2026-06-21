---
status: active
when_to_read: from-scratch keyframe replication renders shapes at wrong (3× too late) times; a NTSC/29.97 comp's kf.Time / marker seconds read 3× too large; copying keyframe seconds from one comp into a freshly NewComposition'd comp; touching deriveTickRate / cdta @0x08 / @0xA8; NewComposition for fractional (29.97/59.94) fps produces inconsistent cdta time-base
applies_to: [tickrate, deriveTickRate, cdta, 0x08, 0xA8, keyframe-time, ntsc, 29-97, from-scratch, new-composition, fractional-fps, false-green, roundtrip-cancels, booyah-clone, valueAtTime, ae-ground-truth]
last_updated: 2026-06-21
resolved_by: d03101c (deriveTickRate fix); NewComposition fractional-fps still open
---

# deriveTickRate 3× off for NTSC modern-AE files (×1000/scale was bogus)

## Signature
- symptom: from-scratch clone shape keyframes play over 2.7 s instead of the original's 0.9 s visible window; NTSC comp kf.Time reads 3× too large (0.8 s where AE says 0.2669 s)
- error_type: —
- where: `internal/serializer/parse_composition.go` `deriveTickRate`
- trigger: parse a 29.97-fps comp with cdta scale@0xA8 > 1, then read kf seconds / write them into a fresh comp

## Symptom / repro

booyah-clone comp① (`シェイイイイプ！！！`, 4 animated rects, layer visible `[0, 0.901]`):
clone shapes rendered spread across the whole 2.7 s keyframe span, so at comp-time
0.3 s the clone showed one growing rect while the original showed four (the original
compresses its 2.7 s of keyframes into the 0.9 s window). Pixel diff clone-vs-orig at
t=0.3/0.5/0.8 was a gross structural mismatch.

**Ground truth (AE DOM, `valueAtTime` / `keyTime`):** original RECT0 Size keys at
frames 0/8/11 → **0.0 / 0.2669 / 0.3670 s** (`= frame / 29.97`). Our parser read them as
**0.0 / 0.8 / 1.1 s** — exactly **×2.997 (= 29.97/10)** too large.

## Root cause

cdta for every comp in the file: `tpf@0x06=800, rate@0x08=23976, scale@0xA8=2997`
(29.97 fps; `23976 = 800 × 29.97`). `deriveTickRate` did, for `scale > 1`:

    TickRate = rate × 1000 / scale = 23976 × 1000 / 2997 = 8000

so kf seconds = `ticks / 8000`. AE actually evaluates against **`rate@0x08` directly**
(23976): frame-8 kf = `8 × 800 = 6400` ticks → `6400 / 23976 = 0.2669 s` ✓, vs the
parser's `6400 / 8000 = 0.8 s`. **`@0xA8` is a display-time divisor, not a kf-rate
scale.** The `×1000/scale` "legacy NTSC → 8000" rule was py-aep-derived and verified only
against hex / py-aep, **never against AE's rendered keyframe times** (red-line: py-aep
parity ≠ AE truth). The repo test `composition_tickrate_test.go` enshrined it
(`legacy_29_97 … want 8000`).

**Why it stayed hidden:** under read-modify-write the keyframe *ticks* are preserved
(length-preserving write re-emits the same tick value), so a wrong TickRate cancels —
every existing round-trip / keyframe ship-gate is blind to it. It only bites when you
read kf **seconds** out of comp A and lower them into a **different** comp B
(from-scratch replication): the seconds are 3× inflated, so B's keyframes land 3× too
late.

## Fix

`deriveTickRate` returns `rate@0x08` directly (fallback `aeLegacyTimeBase`=8000 only when
`rate == 0`); dropped the `scale` branch. Verified end-to-end: after the fix the parser
reads orig kf times = AE's (0.2669 …), and booyah-clone comp① renders **pixel-identical**
to the original at t=0.3/0.5/0.8 (diff 0/0/32 px; the 32 px is sub-pixel 29.97-vs-30
frame-snap). `composition_tickrate_test.go` `legacy_29_97` expectation 8000 → 23976
(renamed `ntsc_29_97_scale_is_display_divisor`). Commit `d03101c`.

Scope note: AE-verified for this one 29.97 file (AE 2025). The modern (`scale=1`) cases
already returned `rate` and are unchanged; broader NTSC-fps AE re-verification (50/59.94)
is advisable if such fixtures appear.

## Second finding — NewComposition fractional-fps cdta is inconsistent (OPEN)

`NewComposition(..., 29.97, ...)` writes a cdta with a **modern marker** `scale@0xA8=1`
but **legacy** `tpf@0x06=800 / rate@0x08=23976`. AE, seeing `scale=1`, uses a modern
tickrate of `1024 × round(fps) = 30720`, while we lower keyframes at 23976 — compressing
every keyframe by `23976/30720`. Clone kf then read back (by us) at the right seconds but
rendered by AE at `× 0.781`. **Workaround:** build the clone at a whole 30 fps —
self-consistent (`30720`), and because keyframe *seconds* survive a matching tickrate, AE
evaluates them at the exact original comp-times (the ~0.1% 29.97-vs-30 snap is sub-pixel).
The proper fix (make NewComposition emit a consistent fractional-fps time-base — either
`tpf=1024` modern bytes or a `scale=fps×100` legacy triple) is a separate writer task.

Related: [[tickrate-per-composition]] (the per-comp gotcha this corrects), [[cdta-0xB0-shutter-ref-not-duration]], [[comp-setframerate-no-duration-rescale]].

## Cases
- 2026-06-21 first seen — booyah-clone comp① mid-frame fidelity RE; root-caused via AE valueAtTime/keyTime ground truth; deriveTickRate fixed; NewComposition fractional-fps left open.
