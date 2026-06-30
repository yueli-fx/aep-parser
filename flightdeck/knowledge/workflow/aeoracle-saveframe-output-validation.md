# ⚠ aeoracle render must validate PNG files after saveFrameToPng

SUMMARY: AE `comp.saveFrameToPng` can return without the requested PNG being present yet; `aeoracle render` must wait for each file and the CLI must validate all requested frame PNGs after AE reports `ok`.
READ WHEN: changing `scripts/ae-worker/aeoracle_render.jsx`; debugging aeoracle metadata that says frames are `rendered` but PNG files are missing; adding render gates for heavier shape/effect outputs

---

During recipe Twist validation, `scripts/ae-worker/aeoracle_render.jsx` produced:

- `aeoracle_render.done` = `ok`
- `metadata.json` with five frames marked `rendered`
- only two actual PNG files on disk

The root cause was that the JSX trusted `comp.saveFrameToPng` returning and
recorded metadata immediately. For heavier frames, the file may not be present
or non-empty by the time the script continues and eventually quits AE.

Durable fix:

- JSX waits after each `saveFrameToPng` until `<tag>.png` exists and has
  non-zero length.
- `cmd/aeoracle render` validates every requested `<tag>.png` exists and is
  non-empty after `done_path` reports `ok`.

Do not treat metadata `status: ok` as sufficient render evidence unless the
PNG files are also present. A valid render gate needs both metadata status and
actual frame artifacts.
