package main

import (
	"fmt"
	"os"
	"github.com/yueli-fx/aep-parser/internal/aep"
)

func main() {
	p, err := aep.Open(os.Args[1])
	if err != nil { fmt.Println(err); os.Exit(1) }
	for _, c := range p.Compositions { fmt.Printf("COMP    id=%d name=%q\n", c.ID, c.Name) }
	for _, f := range p.Footage     { fmt.Printf("FOOTAGE id=%d name=%q\n", f.ID, f.Name) }
	for _, fo := range p.Folders    { fmt.Printf("FOLDER  id=%d name=%q\n", fo.ID, fo.Name) }
}
