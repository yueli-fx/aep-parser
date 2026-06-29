package aep_test

import (
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// Doc example for docgen (compile-checked, no "// Output:").

func ExampleProject_WriteAEP() {
	proj, err := aep.Open("my-project.aep")
	if err != nil {
		return
	}
	// open → mutate → save
	if c := proj.CompositionByName("Main"); c != nil {
		_ = c.SetName("Final")
	}
	out, err := os.Create("modified.aep")
	if err != nil {
		return
	}
	defer out.Close()
	_ = proj.WriteAEP(out)
}
