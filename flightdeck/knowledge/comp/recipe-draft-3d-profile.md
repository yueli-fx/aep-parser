# ⚠ Recipe comp `draft_3d` is profile/roundtrip evidence, not AE DOM evidence

Recipe `comp.draft_3d` writes the cdta Draft 3D flag and `profile.Composition.draft_3d` reads that flag back, but AE 2025 DOM readback is not a reliable oracle for this setting.

Recipe syntax:

```json
{
  "name": "Main",
  "width": 1920,
  "height": 1080,
  "frame_rate": 30,
  "duration": 1,
  "draft_3d": true
}
```

The compiler calls `Composition.SetDraft3D`, which writes cdta byte `0x8A` bit
0. `internal/profile` reports `composition.draft_3d` by reading the same cdta
flag from the parsed composition bytes, so `expected_profile.draft_3d` is a
roundtrip/profile contract.

Do not upgrade this to an AE DOM acceptance claim without new evidence. Existing
notes in `internal/aep_test/comp_settings_shipgate_test.go` document that AE 2025's
`comp.draft3d` DOM readback did not reflect the written flag reliably.
