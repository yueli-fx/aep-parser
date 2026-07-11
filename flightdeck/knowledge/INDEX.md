# Knowledge index

SUMMARY: Routing index for Flightdeck knowledge. Use this instead of bulk-reading all knowledge files.
READ WHEN: starting a task that may need project knowledge, deciding which knowledge notes to load, or adding/moving knowledge files

---

Do not bulk-read `flightdeck/knowledge/**`. Start with `cockpit.md` and
`briefing.md`, then use this index plus each file's `SUMMARY` / `READ WHEN`
header to load only the notes relevant to the task.

## High-Traffic Routes

| Need | Read |
|---|---|
| Public API, package boundaries, write/delivery rules | `workflow/project-operating-rules.md` |
| Verify before commit, choose tests, place debug artifacts | `workflow/verify.md` |
| AE ship-gate, JSX reverse-engineering, fixture regeneration | `workflow/re-fixture.md` |
| Claiming a feature is usable/shippable | `workflow/delivery-contract.md` |
| Generated evidence, `tmp/`, registry evidence placement | `workflow/folder-usage-policy.md` |
| Internal package map | `architecture/internal-codebase-map.md` and `internal/README.md` |
| Showcase generation, review-gate, visual eye-check workflow | `showcase/showcase.md` |
| JSX generator cleanup | `workflow/jsx-generator-cleanup-provenance.md` |
| Rain phenomenon recipe and plugin boundary | `techniques/build-good-rain.md` |
| Glitch phenomenon recipe and plugin boundary | `techniques/build-good-glitch.md` |
| Public/untrusted AEP upload safety | `security/untrusted-aep-input.md` |

## Domains

- `architecture/`: repository and `internal/` architecture maps.
- `workflow/`: cross-cutting process rules, gates, artifact placement, delivery
  contracts, and automation gotchas.
- `composition/` and `comp/`: composition fields, cdta behavior, recipe notes,
  frame/tick/rate/resolution semantics.
- `layer/`: layer creation/mutation, transforms, matte, camera/light, layer
  flags, and layer recipe notes.
- `shape/`: masks, shape graph, vector filters, gradients/strokes, path
  keyframes, and shape recipe notes.
- `effects/` and `effect/`: AddEffect, effect parameter templates, pseudo
  effects, expression controls, and effect recipe notes.
- `text/`: text style, btdk, kerning, variable fonts, text animator behavior.
- `keyframe-expression/`: keyframe layouts, expression enable bytes, separate
  dimensions mechanics.
- `parse-serialize/`: parser/serializer edge cases, chunk IDs, downgrader and
  project flag findings.
- `property/`: property stream/runtime behavior and concurrency caveats.
- `render-queue/`: render queue mechanics and API limits.
- `essential-graphics/`: Essential Graphics write behavior and recipes.
- `docgen/`: doc generation blind spots and documentation conventions.
- `techniques/`: high-level technique internalization, procedural FX, and sample
  analysis.
- `showcase/`: showcase contract and visual review workflow.
- `security/`: untrusted binary input, parser budgets, and service isolation.

## Loading Policy

1. Read a domain note only when its `READ WHEN` matches the current task.
2. Prefer one or two focused notes over a whole directory.
3. If a note references another note, follow it only when the current task needs
   that extra detail.
4. New durable knowledge must be self-contained and include `SUMMARY` and
   `READ WHEN`; update this index only when it adds a new route or domain.
