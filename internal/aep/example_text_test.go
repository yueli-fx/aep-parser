package aep_test

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// Doc example for docgen (compile-checked, no "// Output:"). Attaches to the
// TextSource.Runs field.

func ExampleTextSource_Runs() {
	var comp *aep.Composition
	l := comp.LayerByID(1)
	if l == nil || l.TextSource == nil {
		return
	}
	ts := l.TextSource
	fmt.Println(ts.Text)
	for _, r := range ts.Runs {
		fmt.Printf("%s %gpt fill=%v\n", r.FontName, r.FontSize, r.FillColor)
	}
}
