# aep-parser

**Create, edit, and generate real `.aep` projects in Go, without launching After Effects.**

[中文](README.md) · **English** · [日本語](README.ja.md)

## Build an animated AE project in a few lines

These are the core steps from the [complete runnable example](examples/authoring/main.go). Run it inside this repository using the `internal/aep` authoring API. The example defines its own `must` / `check` error helpers, shown below.

### Create a project, add text and solids

```go
import "github.com/yueli-fx/aep-parser/internal/aep"

project := aep.NewProject(aep.TargetAE2020)
comp := must(aep.NewComposition(project, "Hello AEP", 1920, 1080, 30, 5))

must(aep.NewTextLayer(comp, "Title"))
must(aep.NewSolidLayer(comp, "Card", 640, 360, [3]float64{0.1, 0.8, 0.7}))
must(aep.NewSolidLayer(comp, "Draft", 100, 100, [3]float64{1, 0, 0}))
```

A **1080p, 30 fps, five-second** composition with three layers. Now change the text and add an effect.

`must` accepts `(value, error)`, checks the error, and returns the value. `check` handles operations that return only an `error`. Define these helpers outside `main`; they are neither Go built-ins nor library APIs:

```go
// Stop the example on an error instead of continuing with a broken project.
func must[T any](value T, err error) T {
    check(err)
    return value
}

func check(err error) {
    if err != nil {
        panic(err)
    }
}
```

### Add Gaussian Blur, edit text and effect parameters

```go
// Reparse in memory to enable edits on newly created layers. No AE needed.
project = must(aep.Reopen(project))
comp = project.Compositions[0]
check(comp.LayerByName("Title").SetText("HELLO, AEP"))

card := comp.LayerByName("Card")
blur := must(aep.AddEffect(card, aep.EffectGaussianBlur))
must(aep.SetEffectParam(card, blur, "Blurriness", 30.0))
```

This sets the blur amount to **30**. Use parameter names such as `"Blurriness"` or `"模糊度"`; the library resolves the internal identifier. The authoring interface accepts parameter names only, not raw match-names. Gaussian Blur is one of [231 addable effect templates](internal/serializer/mutate_effect_add.go). The [effects example](showcase/effects/gen.go) includes parameter settings for shadows, strokes, color effects, and more.

### Animate the blur, delete the draft layer, save

```go
// Go from blur 30 to 0 during the first second.
must(aep.AnimateEffectParam(card, blur, "Blurriness",
    []aep.ScalarKeyframe{{Time: 0, Value: 30}, {Time: 1, Value: 0}}))

for index, layer := range comp.Layers {
    if layer.Name == "Draft" {
        check(aep.DeleteLayer(comp, index)) // Zero-based index; deleting a solid here.
        break
    }
}

file := must(os.Create("hello.aep"))
check(project.WriteAEP(file))
check(file.Close())
```

The result is an `.aep` with text, a solid, and animated blur. **Creating, editing, and writing the file requires no AE installation.** Use AE for visual rendering and further editing. The complete program also reparses its output and refuses to overwrite existing files.

### Run it now

```sh
git clone https://github.com/yueli-fx/aep-parser.git
cd aep-parser
go run ./examples/authoring -out tmp/hello.aep
go run ./cmd/aep inspect -in tmp/hello.aep
```

Requires Go 1.25.13 or a newer patch in the same series. The authoring API above is for programs within this repository. For your own Go module, use workflows such as `Open`, `Export`, and `Compile` from the [public SDK examples](examples/sdk/README.md).

## Go beyond a blurred rectangle

- **[Procedural fire](showcase/procedural-fx/gen.go)**: three layers of turbulent noise, temperature-based coloring, feathered masks, additive compositing, and Glow in an animated fire project. [Acceptance record](showcase/procedural-fx/INDEX.md)
- **[3D camera scene](showcase/3d-camera/gen.go)**: create a camera and cards at different depths, with Z parallax and Y-axis perspective rotation. [Acceptance record](showcase/3d-camera/INDEX.md)
- **[Vector shape animation](showcase/shape-filters/INDEX.md)**: eleven filter families including Trim Paths, Repeater, ZigZag, and Wiggle, with combinations.
- **[Expression-driven animation](showcase/expressions/INDEX.md)**: cross-layer references, loopOut, wiggle, and slider controls.

[Explore all 25 showcases →](showcase/README.md) · [Lower third, animated title, and progress bar recipes →](examples/projects/README.md)

You can also generate projects directly from JSON:

```sh
go run ./cmd/aeprecipe compile -recipe examples/projects/animated-title.json -out tmp/title.aep -json
```

## Read, compare, and migrate existing projects

```sh
# Extract composition, layer, effect, and keyframe information
go run ./cmd/aep profile -in project.aep

# Compare projects before and after editing
go run ./cmd/aep diff -expected before.aep -actual after.aep

# Migrate supported project semantics to AE2025
go run ./cmd/aep migrate -in source.aep -target AE2025 -out tmp/migrated.aep
```

For batch inspection, property snapshots, and preserving edits to numeric properties, see the [five public SDK programs](examples/sdk/README.md). To understand how a complex project works, generate an [HTML technique report](docs/self-hosted-reports.md) with per-layer effect stacks, recreation steps, and study tasks.

## How much is implemented?

**500 function/method capability entries · 231 effect templates · 16 domains · Six writer targets, AE2020–AE2025.**

The implementation spans RIFX containers, nested chunks, and byte layouts through layer structure, 2D/3D transforms, text styles, shapes and gradients, keyframes and easing, expressions, effects, and masks. **151 public knowledge notes** document reverse-engineering findings and failures, alongside **151 atomic Recipe examples**.

Counts come from the [capability index](docs/capabilities.json) and [effect registry](internal/serializer/mutate_effect_add.go): 382 methods + 118 functions, with 438 marked stable and 62 alpha; 266 tagged ae-accept and 163 render-pixel. These count underlying capabilities and existing verification records. See the [capability matrix](docs/capabilities.md) for individual interfaces and boundaries. The [migration regression record](registry/evidence/versioned-aep-migration/migration_matrix_verify/smoke_all/ledger.md) covers 906 combinations: 900 passed, six skipped, zero failed.

## Explore further

| Goal | Start here |
| --- | --- |
| Integrate with Go / HTTP | [Usage guide](docs/usage.en.md) · [SDK examples](examples/sdk/README.md) |
| Generate projects from JSON | [Recipe examples](examples/recipes) · [Field reference](docs/recipe.md) · [Schema](docs/recipe_schema.json) |
| Study AEP reverse engineering | [Knowledge base](docs/knowledge/README.md) · [API reference](docs/README.md) |
| Check support and compatibility | [Capability matrix](docs/capabilities.md) · [Usage and limitations](docs/usage.en.md) |
| Contribute | [Contributing](CONTRIBUTING.md) · [Verification workflow](docs/knowledge/workflow/verify.md) · [Publication audit](docs/open-source-audit.md) |

The project is prerelease. The parsing and generation core is pure Go and builds on Windows, macOS, and Linux. Knowledge notes are mostly Chinese; generated API / Recipe references are mostly English. New combined examples have passed generation and reparse checks but have not had separate AE visual acceptance. See each showcase for its host verification record.

## License

**[PolyForm Noncommercial 1.0.0](LICENSE)**. Permitted noncommercial uses are free. Commercial use, resale, and commercial integration outside the free license’s permitted purposes require a separate written commercial license.

This is a **source-available** project. See **[Usage and commercial licensing](LICENSING.md#english)** for permitted uses, organizational exceptions, license inquiries, and generated content. Third-party notices: [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md). Security reporting: [SECURITY.md](SECURITY.md).

This project is not affiliated with Adobe. Adobe, After Effects, and related trademarks belong to their respective owners.
