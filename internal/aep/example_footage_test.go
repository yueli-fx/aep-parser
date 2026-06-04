package aep_test

import (
	"strings"

	"github.com/example/aep-parser/internal/aep"
)

// Doc example for docgen (compile-checked, no "// Output:").

func ExampleFootage_SetPath() {
	proj := aep.NewProject()
	for _, f := range proj.Footage {
		// batch-redirect moved assets
		if strings.HasPrefix(f.Path, `D:\old\`) {
			_ = f.SetPath(strings.Replace(f.Path, `D:\old\`, `E:\new\`, 1))
		}
	}
}
