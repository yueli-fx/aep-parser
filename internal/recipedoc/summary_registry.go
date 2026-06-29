package recipedoc

var fieldSummaries = map[string]string{
	"schema_version":               "Recipe schema version.",
	"project.name":                 "Project display name.",
	"comps[].name":                 "Composition display name.",
	"comps[].width":                "Composition width in pixels.",
	"comps[].height":               "Composition height in pixels.",
	"comps[].frame_rate":           "Composition frame rate in frames per second.",
	"comps[].duration":             "Composition duration in seconds.",
	"comps[].background_color":     "Composition background color as RGB channels.",
	"comps[].layers[].type":        "Layer creation type.",
	"comps[].layers[].name":        "Layer display name.",
	"comps[].layers[].transform":   "Layer transform block.",
	"comps[].layers[].effects[]":   "Built-in effect instance to add to the layer.",
	"expected_profile.keyframes[]": "Expected keyframes for a layer property.",
}
