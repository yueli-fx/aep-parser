// Snapshot prints ProjectJSON v2, including property records and write targets.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"os"

	aep "github.com/yueli-fx/aep-parser"
)

func main() {
	input := flag.String("in", "test_data/fixtures/transform_unseparated.aep", "input AEP")
	flag.Parse()
	doc, err := aep.Open(*input)
	if err != nil {
		log.Fatal(err)
	}
	data, err := doc.ProjectJSON()
	if err != nil {
		log.Fatal(err)
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		log.Fatal(err)
	}
	pretty.WriteByte('\n')
	if _, err := pretty.WriteTo(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
