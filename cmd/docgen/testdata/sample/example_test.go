package sample_test

import "github.com/example/aep-parser/cmd/docgen/testdata/sample"

func ExampleWidget_SetName() {
	w := &sample.Widget{}
	_ = w.SetName("hi")
	// Output:
}
