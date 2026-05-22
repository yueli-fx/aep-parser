package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	target := aep.TargetAE2025
	outPath := "test_data/v2_smoke.aep"
	if len(os.Args) >= 2 && os.Args[1] == "2020" {
		target = aep.TargetAE2020
		outPath = "test_data/v2_smoke_ae2020.aep"
	}

	p := aep.NewProject(target)
	main, err := p.NewComposition("Main", 1920, 1080, 29.97, 10)
	if err != nil {
		panic(err)
	}
	main.SetBGColor([3]uint8{20, 30, 40})
	if _, err := p.NewComposition("BG_loop", 1920, 1080, 30, 5); err != nil {
		panic(err)
	}
	out, _ := os.Create(outPath)
	defer out.Close()
	if err := p.WriteAEP(out); err != nil {
		panic(err)
	}
	fmt.Printf("wrote %s (target=AE%d)\n", outPath, int(target))
}
