package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	p := aep.NewProject(aep.TargetAE2025)
	main, err := p.NewComposition("Main", 1920, 1080, 29.97, 10)
	if err != nil { panic(err) }
	main.SetBGColor([3]uint8{20, 30, 40})
	if _, err := p.NewComposition("BG_loop", 1920, 1080, 30, 5); err != nil { panic(err) }
	out, _ := os.Create("test_data/v2_smoke.aep")
	defer out.Close()
	if err := p.WriteAEP(out); err != nil { panic(err) }
	fmt.Println("wrote test_data/v2_smoke.aep")
}
