# Recipe Camera Layer

SUMMARY: Recipe support note for Camera Layer.
READ WHEN: working on recipe schema, compiler, profile validation, examples, or docs for layer fields related to Camera Layer

---

Context: recipe `type: "camera"` support, proven with
`examples/recipes/minimal-camera-layer.json`.

Recipe authoring:

- A layer with `type: "camera"` creates a camera layer through
  `NewCameraLayer`.
- The dedicated example creates a `Camera` layer plus a visible text layer so
  AE render acceptance has frame output to inspect.

Writer capability:

- `NewCameraLayer`

Boundary:

- This slice covers high-level camera layer creation only.
- Recipe transform validation still requires 2-value vectors, so 3D camera
  placement is intentionally deferred.
- Camera option setters such as zoom/depth-of-field/focus distance are also
  deferred to a separate recipe slice.
