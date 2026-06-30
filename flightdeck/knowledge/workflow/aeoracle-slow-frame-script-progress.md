# aeoracle slow frames need request-level frame timeout and progress-dialog ignore

SUMMARY: `cmd/aeoracle render` can have a large CLI timeout while the JSX per-frame PNG wait is much shorter; slow AE frames also keep an `Executing Script ...` progress dialog visible long enough to trip unknown-modal handling unless it is explicitly ignored.
READ WHEN: `aeoracle render` reports `rendered frame missing or empty`; AE shows an `Executing Script ...` dialog for more than 30 seconds; adding render-heavy recipe examples or shape filters; handling the `Scripting plugin is not installed` dialog

---

Root cause from the Wiggle Paths recipe slice:

- `scripts/ae-worker/aeoracle_render.jsx` waited only 30 seconds for each
  `saveFrameToPng` output even when `cmd/aeoracle render -timeout-sec` was much
  larger.
- For slow frames, AE keeps a modal progress dialog titled like
  `Executing Script aeoracle_render.jsx...`. The dialog is normal progress, not
  a failure, but the generic modal watcher used to classify it as unknown after
  the grace window.
- A separate interrupted AE gate exposed `Unable to execute script. The
  Scripting plugin is not installed.` This appears before JSX can write `.done`;
  it must be an `Abort` rule, not a blind OK/dismiss rule.

Current behavior:

- `internal/aeoracle.RenderRequest` includes `frame_timeout_ms`.
- `aeoracle.NewRenderRequest` defaults `frame_timeout_ms` to `120000`.
- `scripts/ae-worker/aeoracle_render.jsx` reads `req.frame_timeout_ms` and falls back to
  `120000` for older request files.
- `scripts/ae-worker/ae_dialog_rules.json` ignores the long-running
  `Executing Script ...` progress dialog.
- The same rules file aborts on `Scripting plugin is not installed` with a
  clear operator-facing message.
- Dialog rule `ocrMatch` values are OR-matched. Keep the scripting-plugin rule
  on distinctive phrases such as `Scripting plugin is not installed`; do not use
  broad tokens such as a single word for script, or unrelated script alerts can
  be misclassified as environment aborts.

If a frame still fails after the longer wait, do not keep increasing the wait
blindly. First check whether the recipe is stacking too many expensive path
filters in one shape group. For acceptance recipes, prefer a dedicated focused
example over a single kitchen-sink recipe.
