package scene

// scene_essential_graphics.go — Essential Graphics panel (EGP) model. AE
// exposes composition properties as .mogrt controllers; the panel definition
// lives in the comp's Item-level LIST:CIF3. Read-only (mirrors py-aep's
// EssentialGraphicsController). Decoded in parse_essential_graphics.go.

// EGControllerType is an Essential Graphics controller's type code, as stored
// in its CTyp chunk. Values mirror py-aep's controller_type.
type EGControllerType uint32

const (
	EGCheckbox         EGControllerType = 1
	EGSlider           EGControllerType = 2
	EGColor            EGControllerType = 4
	EGPoint            EGControllerType = 5
	EGText             EGControllerType = 6
	EGComment          EGControllerType = 8
	EGMultiDimensional EGControllerType = 9
	EGGroup            EGControllerType = 10
	EGDropdown         EGControllerType = 13
)

// String renders the controller type as AE's panel label (or "unknown").
func (t EGControllerType) String() string {
	switch t {
	case EGCheckbox:
		return "checkbox"
	case EGSlider:
		return "slider"
	case EGColor:
		return "color"
	case EGPoint:
		return "point"
	case EGText:
		return "text"
	case EGComment:
		return "comment"
	case EGMultiDimensional:
		return "multidimensional"
	case EGGroup:
		return "group"
	case EGDropdown:
		return "dropdown"
	default:
		return "unknown"
	}
}

// EssentialGraphicsController is one controller in the Essential Graphics
// panel — an exposed property / .mogrt control.
type EssentialGraphicsController struct {
	// Name is the controller's display name (from its CpS2 localized string).
	Name string

	// Type is the controller kind (checkbox / slider / color / ...).
	Type EGControllerType

	// UUID is the controller's unique identifier (the CCtl's Utf8 child),
	// used by AE to link the controller to its source property.
	UUID string
}

// MotionGraphicsTemplateControllerCount returns the number of Essential
// Graphics controllers — mirrors AE's CompItem.motionGraphicsTemplateControllerCount.
func (c *Composition) MotionGraphicsTemplateControllerCount() int {
	return len(c.EssentialGraphicsControllers)
}

// MotionGraphicsTemplateControllerNames returns the controller display names
// in panel order — mirrors AE's CompItem.motionGraphicsTemplateControllerName(i).
func (c *Composition) MotionGraphicsTemplateControllerNames() []string {
	if len(c.EssentialGraphicsControllers) == 0 {
		return nil
	}
	names := make([]string, len(c.EssentialGraphicsControllers))
	for i, ctrl := range c.EssentialGraphicsControllers {
		names[i] = ctrl.Name
	}
	return names
}
