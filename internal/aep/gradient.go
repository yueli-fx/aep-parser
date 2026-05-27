package aep

import (
	"encoding/xml"
	"strconv"
	"strings"
)

// GradientColorStop represents a single color stop in a gradient.
type GradientColorStop struct {
	Offset   float64      // position along gradient (0.0 to 1.0)
	Midpoint float64      // interpolation midpoint to next stop (0.0 to 1.0)
	Color    [3]float64   // RGB color in 0..1 range
}

// GradientAlphaStop represents a single alpha (opacity) stop in a gradient.
type GradientAlphaStop struct {
	Offset   float64 // position along gradient (0.0 to 1.0)
	Midpoint float64 // interpolation midpoint to next stop (0.0 to 1.0)
	Alpha    float64 // opacity value (0.0 to 1.0)
}

// Gradient holds parsed gradient color data from a gradient fill/stroke
// property ("ADBE Vector Grad Colors"). The XML is stored in the cdat
// chunk as a prop.map/prop.list/prop.pair structure.
type Gradient struct {
	ColorStops []GradientColorStop
	AlphaStops []GradientAlphaStop
	Version    string
}

// ParseGradientXML parses AE gradient XML (prop.map format) into a
// *Gradient. Returns nil when parsing fails or the XML is empty.
// Exposed so tests can exercise the XML decoder without building a
// full RIFX tree.
func ParseGradientXML(xmlText string) *Gradient {
	if xmlText == "" {
		return nil
	}
	// Quick check: must contain prop.map
	if !strings.Contains(xmlText, "prop.map") {
		return nil
	}

	// Use a generic struct to decode the nested prop.map/prop.list/prop.pair XML.
	var root propMap
	if err := xml.Unmarshal([]byte(xmlText), &root); err != nil {
		return nil
	}

	g := &Gradient{Version: root.version()}

	// Navigate: prop.list > "Gradient Color Data" > prop.list
	gd := root.findPropList("Gradient Color Data")
	if gd == nil {
		return nil
	}

	// Color stops
	cs := gd.findPropList("Color Stops")
	if cs != nil {
		sl := cs.findPropList("Stops List")
		if sl != nil {
			g.ColorStops = parseColorStops(sl)
		}
	}

	// Alpha stops
	as := gd.findPropList("Alpha Stops")
	if as != nil {
		sl := as.findPropList("Stops List")
		if sl != nil {
			g.AlphaStops = parseAlphaStops(sl)
		}
	}

	return g
}

func parseColorStops(sl *propList) []GradientColorStop {
	var stops []GradientColorStop
	for i := 0; ; i++ {
		stopElem := sl.findPropList("Stop-" + strconv.Itoa(i))
		if stopElem == nil {
			break
		}
		arr := stopElem.findFloatArray("Stops Color")
		if len(arr) >= 5 {
			stops = append(stops, GradientColorStop{
				Offset:   arr[0],
				Midpoint: arr[1],
				Color:    [3]float64{arr[2], arr[3], arr[4]},
			})
		}
	}
	return stops
}

func parseAlphaStops(sl *propList) []GradientAlphaStop {
	var stops []GradientAlphaStop
	for i := 0; ; i++ {
		stopElem := sl.findPropList("Stop-" + strconv.Itoa(i))
		if stopElem == nil {
			break
		}
		arr := stopElem.findFloatArray("Stops Alpha")
		if len(arr) >= 3 {
			stops = append(stops, GradientAlphaStop{
				Offset:   arr[0],
				Midpoint: arr[1],
				Alpha:    arr[2],
			})
		}
	}
	return stops
}

// --- XML helper types for prop.map / prop.list / prop.pair ---

type propMap struct {
	XMLName xml.Name   `xml:"prop.map"`
	Version string     `xml:"version,attr"`
	Lists   []propList `xml:"prop.list"`
}

func (m *propMap) version() string {
	if m.Version != "" {
		return m.Version
	}
	return "1.0"
}

func (m *propMap) findPropList(key string) *propList {
	for i := range m.Lists {
		if found := m.Lists[i].findPropList(key); found != nil {
			return found
		}
	}
	return nil
}

type propList struct {
	Pairs []propPair `xml:"prop.pair"`
}

func (l *propList) findPropList(key string) *propList {
	for _, p := range l.Pairs {
		if p.Key == key && p.List != nil {
			return p.List
		}
	}
	return nil
}

func (l *propList) findFloatArray(key string) []float64 {
	for _, p := range l.Pairs {
		if p.Key == key && p.Array != nil {
			return p.Array.Floats
		}
	}
	return nil
}

type propPair struct {
	Key   string     `xml:"key"`
	List  *propList  `xml:"prop.list"`
	Array *floatArray `xml:"array"`
	Str   string     `xml:"string"`
}

type floatArray struct {
	Floats []float64 `xml:"float"`
}
