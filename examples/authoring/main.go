// Run from the repository root: go run ./examples/authoring -out tmp/hello.aep
// This example uses the repository's internal authoring API.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func must[T any](value T, err error) T {
	check(err)
	return value
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func build() *aep.Project {
	// 1. Create a five-second composition and three layers.
	project := aep.NewProject(aep.TargetAE2020)
	comp := must(aep.NewComposition(project, "Hello AEP", 1920, 1080, 30, 5))
	must(aep.NewTextLayer(comp, "Title"))
	must(aep.NewSolidLayer(comp, "Card", 640, 360, [3]float64{0.1, 0.8, 0.7}))
	must(aep.NewSolidLayer(comp, "Draft", 100, 100, [3]float64{1, 0, 0}))

	// 2. Reopen in memory to enable property and structural edits.
	project = must(aep.Reopen(project))
	comp = project.Compositions[0]
	title := comp.LayerByName("Title")
	check(title.SetText("HELLO, AEP"))
	card := comp.LayerByName("Card")
	blur := must(aep.AddEffect(card, aep.EffectGaussianBlur))
	must(aep.SetEffectParam(card, blur, "Blurriness", 30.0))

	// 3. Animate blur from 30 to 0, then delete the temporary solid.
	must(aep.AnimateEffectParam(card, blur, "Blurriness",
		[]aep.ScalarKeyframe{{Time: 0, Value: 30}, {Time: 1, Value: 0}}))
	for index, layer := range comp.Layers {
		if layer.Name == "Draft" {
			check(aep.DeleteLayer(comp, index)) // Zero-based index; AV layer deletion.
			break
		}
	}
	return project
}

func main() {
	out := flag.String("out", "tmp/hello.aep", "new output file (must not exist)")
	flag.Parse()
	project := build()
	var buffer bytes.Buffer
	check(project.WriteAEP(&buffer))
	reopened := must(aep.FromReader(bytes.NewReader(buffer.Bytes())))
	if len(reopened.Compositions) != 1 || len(reopened.Compositions[0].Layers) != 2 {
		panic("unexpected composition or layer count after writing")
	}
	check(os.MkdirAll(filepath.Dir(*out), 0o755))
	file := must(os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644))
	defer file.Close()
	_, err := file.Write(buffer.Bytes())
	check(err)
	check(file.Close())
	fmt.Printf("Created %s: 1920x1080, 30 fps, 5 seconds, 2 layers, animated Gaussian Blur\n", *out)
}
