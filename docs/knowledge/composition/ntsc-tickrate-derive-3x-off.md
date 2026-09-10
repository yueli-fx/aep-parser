# ⚠ deriveTickRate 3× off for NTSC modern-AE files (×1000/scale was bogus)

deriveTickRate 3× off for NTSC modern-AE files (×1000/scale was bogus)

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

## Second finding — from-scratch NTSC keyframes 1.281× too fast → ROOT-CAUSED to tdb4 @0x0C (FIXED 2026-06-22)

`NewComposition(..., 29.97, ...)` then lowering from-scratch keyframes made AE evaluate
every keyframe `30720/23976 = 1.281×` too fast (booyah comp ⑤ precomp Position/Opacity
slid early; the original's 0.9009 s kf-span played over 0.7032 s). Symptom: clone kf read
back by AE's `keyTime` at `tick / 30720` while the original (and AE-native files) read the
**same tick** at `tick / 23976`.

**Disproven hypothesis (the original "second finding"):** that cdta `scale@0xA8` marks
legacy(2997)/modern(1) and drives the keyframe base. **Falsified by ground truth** — both
the legacy original (`@0xA8=2997`) AND a fresh AE-2025-authored 29.97 comp (`@0xA8=1`) are
read at 23976; writing 2997 on our (modern) clone changed nothing, and an AE-native fresh
29.97 comp writes `@0xA8=1` with keyframe tick = `1.0 s × 23976`. cdta (every timing
field), lhd3, and the keyframe ldat blocks were **byte-identical** between our clone and
both AE files, yet AE evaluated them differently → the discriminator is **not** in
cdta/lhd3/ldat/head-fingerprint (all bisected out).

**True root cause:** the per-property descriptor **`tdb4 @0x0C..0x0F` is the keyframe
TIME BASE (= comp TickRate)**, not the "observed constant" `00 00 78 00` (30720). AE
divides each keyframe's stored tick (ldat block @0x00) by `tdb4@0x0C` to get its seconds.
`30720 = 1024×30` was right only because every probed fixture was 30 fps; a 29.97 comp's
native tdb4 holds `00 00 5D A8` (23976). Our `makeTdb4` hardcoded 30720, and the embedded
shape/transform template's tdb4 (cloned by `injectAnimatedStream`) carried its 30-fps
source value — both unpatched. **This stayed hidden the same way deriveTickRate did:** at
30 fps `tdb4@0x0C == TickRate == 30720`, so the bug only bites a fractional-fps comp.

**Fix (commit pending):** `makeTdb4(layout, tickRate)` writes `@0x0C = round(tickRate)`;
`injectAnimatedStream(..., tickRate)` stamps the cloned template's tdb4 `@0x0C` to the
comp's TickRate (alongside the existing static→animated flag patches @0x05/@0x44/@0x4f).
Both keyframe-lowering paths now thread `ctx.tickRate`. Verified end-to-end: booyah comp ⑤
clone L0 Position (2 kf) + L1 Opacity (13 kf) `keyTime` now match the original to 4 dp via
AE 2025 DOM (`verify_precomp_timing.jsx`); 30-fps ship-gates unchanged (tickRate=30720 ⇒
identical bytes). Probe `tools/debug/dump_kf_ticks` (tdb4 + ldat tick dump). The earlier
"build the clone at whole 30 fps" workaround is **retired** — clones now run true 29.97.

Related: [tickrate-per-composition](tickrate-per-composition.md) (the per-comp gotcha this corrects), [cdta-0xB0-shutter-ref-not-duration](cdta-0xB0-shutter-ref-not-duration.md), [comp-setframerate-no-duration-rescale](comp-setframerate-no-duration-rescale.md).

## Cases
- 2026-06-21 first seen — booyah-clone comp① mid-frame fidelity RE; root-caused via AE valueAtTime/keyTime ground truth; deriveTickRate fixed; NewComposition fractional-fps left open.
