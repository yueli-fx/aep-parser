package sample_test

import "github.com/yueli-fx/aep-parser/cmd/docgen/testdata/sample"

func ExampleWidget_SetName() {
	w := &sample.Widget{}
	_ = w.SetName("hi")
	// Output:
}
