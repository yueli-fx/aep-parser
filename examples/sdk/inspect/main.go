// Inspect prints a project inventory without starting After Effects.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	aep "github.com/yueli-fx/aep-parser"
)

func main() {
	input := flag.String("in", "test_data/fixtures/v2_smoke.aep", "input AEP")
	flag.Parse()
	doc, err := aep.OpenWithLimits(*input, aep.Limits{MaxInputBytes: 64 << 20})
	if err != nil {
		log.Fatal(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(doc.Inspect()); err != nil {
		log.Fatal(err)
	}
}
