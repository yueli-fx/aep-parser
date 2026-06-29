package recipedoc

var capabilitiesByPath = map[string][]string{
	"comps[].background_color":                                    {"comp.set_background_color"},
	"comps[].renderer":                                            {"comp.set_renderer"},
	"comps[].layers[].type":                                       {"layer.create_text", "layer.create_shape", "layer.create_solid", "layer.create_camera", "layer.create_light", "layer.create_null", "layer.create_adjustment"},
	"comps[].layers[].transform":                                  {"layer.set_transform"},
	"comps[].layers[].effects[]":                                  {"effect.add_builtin"},
	"comps[].layers[].effects[].params[]":                         {"effect.set_param"},
	"comps[].layers[].effects[].params[].expression.source":       {"property.set_expression"},
	"comps[].layers[].transform.expressions.position.source":      {"property.set_expression"},
	"comps[].layers[].transform.expressions.anchor_point.source":  {"property.set_expression"},
	"comps[].layers[].transform.expressions.scale.source":         {"property.set_expression"},
	"comps[].layers[].transform.expressions.rotation.source":      {"property.set_expression"},
	"comps[].layers[].transform.expressions.opacity.source":       {"property.set_expression"},
	"comps[].layers[].transform.expressions.position.enabled":     {"property.set_expression_enabled"},
	"comps[].layers[].transform.expressions.anchor_point.enabled": {"property.set_expression_enabled"},
	"comps[].layers[].transform.expressions.scale.enabled":        {"property.set_expression_enabled"},
	"comps[].layers[].transform.expressions.rotation.enabled":     {"property.set_expression_enabled"},
	"comps[].layers[].transform.expressions.opacity.enabled":      {"property.set_expression_enabled"},
	"comps[].layers[].effects[].params[].expression.enabled":      {"property.set_expression_enabled"},
}

var capabilityRegistry = map[string]CapabilityMeta{
	"comp.set_background_color":       {Query: "SetBGColor", Summary: "Set composition background color."},
	"comp.set_renderer":               {Query: "SetRenderer", Summary: "Set composition renderer."},
	"layer.create_text":               {Query: "NewTextLayer", Summary: "Create a text layer."},
	"layer.create_shape":              {Query: "NewShapeLayer", Summary: "Create a shape layer."},
	"layer.create_solid":              {Query: "NewSolidLayer", Summary: "Create a solid layer."},
	"layer.create_camera":             {Query: "NewCameraLayer", Summary: "Create a camera layer."},
	"layer.create_light":              {Query: "NewLightLayer", Summary: "Create a light layer."},
	"layer.create_null":               {Query: "NewNullLayer", Summary: "Create a null layer."},
	"layer.create_adjustment":         {Query: "NewAdjustmentLayer", Summary: "Create an adjustment layer."},
	"layer.set_transform":             {Query: "SetLayerTransform", Summary: "Set layer transform values."},
	"effect.add_builtin":              {Query: "AddEffect", Summary: "Add a supported built-in effect."},
	"effect.set_param":                {Query: "SetEffectParam", Summary: "Set an effect parameter value."},
	"property.set_expression":         {Query: "Property.SetExpression", Summary: "Set a property expression."},
	"property.set_expression_enabled": {Query: "Property.SetExpressionEnabled", Summary: "Toggle property expression evaluation."},
}
