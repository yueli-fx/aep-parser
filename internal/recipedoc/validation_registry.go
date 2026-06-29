package recipedoc

var fieldValidation = map[string]FieldMeta{
	"schema_version": {
		Validation: "Must equal the supported recipe schema version.",
	},
	"comps[].width": {
		Validation: "Pixel width must fit the composition writer's uint16 range.",
	},
	"comps[].height": {
		Validation: "Pixel height must fit the composition writer's uint16 range.",
	},
	"comps[].background_color": {
		Validation: "RGB color must contain exactly three channels in the 0..255 range.",
	},
	"comps[].layers[].type": {
		Validation: "Supported values create the corresponding layer type.",
		Enum:       []string{"solid", "text", "shape", "camera", "light", "null", "adjustment"},
	},
}
