# Recipe comp Motion Graphics template name

SUMMARY: Recipe `comp.motion_graphics_template_name` writes the Essential Graphics template name and `expected_profile.motion_graphics_template_name` checks the parsed comp profile.
READ WHEN: adding Essential Graphics recipe fields; deciding whether template-name support implies `AddEssentialProperty`; debugging Motion Graphics template profile checks

---

Recipe syntax:

```json
{
  "name": "Main",
  "width": 1920,
  "height": 1080,
  "frame_rate": 30,
  "duration": 1,
  "motion_graphics_template_name": "Lower Third Pack"
}
```

The compiler calls `Composition.SetMotionGraphicsTemplateName`. The field is
omitted when empty; the underlying setter rejects an empty template name.

Profile readback uses the parsed composition's Essential Graphics panel data, so
`expected_profile.motion_graphics_template_name` checks the value from the AEP
bytes the recipe is about to write.

This slice only names the Motion Graphics template. It does not mean recipe IR
can expose Essential Graphics controllers yet; `AddEssentialProperty` remains a
separate, larger slice because it binds layer/effect parameters into the EG panel.
