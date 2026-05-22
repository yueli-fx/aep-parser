// Verify Camera / Light typed accessors work against a real AE fixture.
// Usage: go run ./tmp_debug/demo_cameralight test_data/re_cameralight.aep
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func dump(label string, p *aep.Property) {
	if p == nil {
		fmt.Printf("  %s: <nil>\n", label)
		return
	}
	fmt.Printf("  %s: matchname=%q components=%d static=%v keyframes=%d\n",
		label, p.MatchName, p.Components, p.StaticValue, len(p.Keyframes))
}

func main() {
	p, err := aep.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, c := range p.Compositions {
		if c.Name != "RE_CL" {
			continue
		}
		for _, l := range c.Layers {
			fmt.Printf("[%s] %s (type=%s)\n", c.Name, l.Name, l.Type)
			if l.Type == aep.LayerTypeCamera {
				dump("CameraZoom", l.CameraZoom())
				dump("CameraAperture", l.CameraAperture())
				dump("CameraFocusDistance", l.CameraFocusDistance())
				dump("CameraBlurLevel", l.CameraBlurLevel())
				dump("CameraDepthOfField", l.CameraDepthOfField())
			}
			if l.Type == aep.LayerTypeLight {
				dump("LightType", l.LightType())
				dump("LightColor", l.LightColor())
				dump("LightIntensity", l.LightIntensity())
				dump("LightConeAngle", l.LightConeAngle())
				dump("LightConeFeather", l.LightConeFeather())
				dump("LightCastsShadows", l.LightCastsShadows())
				dump("LightShadowDarkness", l.LightShadowDarkness())
			}
		}
	}
}
