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

// EncodeGradientXML renders a *Gradient back into AE's prop.map XML form
// (the inverse of ParseGradientXML). The byte layout — element order
// (Alpha Stops before Color Stops), the 6-float color array
// [offset, midpoint, r, g, b, 1], the 3-float alpha array
// [offset, midpoint, alpha], the trailing "Gradient Colors" = "1.0" marker,
// and flat newline-separated lines with no indentation — mirrors an
// AE 25.6-saved gradient fill verbatim so AE re-parses it without complaint.
//
// Exposed so the serializer (lowerGradientFillNode) and tests can produce the
// Utf8 chunk payload. The output round-trips through ParseGradientXML.
func EncodeGradientXML(g *Gradient) string {
	var b strings.Builder
	b.WriteString("<?xml version='1.0'?>\n")
	b.WriteString("<prop.map version='4'>\n")
	b.WriteString("<prop.list>\n")
	b.WriteString("<prop.pair>\n")
	b.WriteString("<key>Gradient Color Data</key>\n")
	b.WriteString("<prop.list>\n")

	// Alpha Stops first.
	writeStopGroup(&b, "Alpha Stops", len(g.AlphaStops), func(i int) {
		s := g.AlphaStops[i]
		writeStopArray(&b, "Stops Alpha", s.Offset, s.Midpoint, s.Alpha)
	})
	// Color Stops second.
	writeStopGroup(&b, "Color Stops", len(g.ColorStops), func(i int) {
		s := g.ColorStops[i]
		writeStopArray(&b, "Stops Color", s.Offset, s.Midpoint, s.Color[0], s.Color[1], s.Color[2], 1)
	})

	b.WriteString("</prop.list>\n") // end Gradient Color Data
	b.WriteString("</prop.pair>\n")
	b.WriteString("<prop.pair>\n")
	b.WriteString("<key>Gradient Colors</key>\n")
	b.WriteString("<string>1.0</string>\n")
	b.WriteString("</prop.pair>\n")
	b.WriteString("</prop.list>\n")
	b.WriteString("</prop.map>\n")
	return b.String()
}

// writeStopGroup emits one "Alpha Stops" / "Color Stops" prop.pair: a
// "Stops List" holding n "Stop-i" entries (each filled by emit) plus a
// trailing "Stops Size" int.
func writeStopGroup(b *strings.Builder, groupKey string, n int, emit func(i int)) {
	b.WriteString("<prop.pair>\n")
	b.WriteString("<key>" + groupKey + "</key>\n")
	b.WriteString("<prop.list>\n")
	b.WriteString("<prop.pair>\n")
	b.WriteString("<key>Stops List</key>\n")
	b.WriteString("<prop.list>\n")
	for i := 0; i < n; i++ {
		b.WriteString("<prop.pair>\n")
		b.WriteString("<key>Stop-" + strconv.Itoa(i) + "</key>\n")
		b.WriteString("<prop.list>\n")
		emit(i)
		b.WriteString("</prop.list>\n")
		b.WriteString("</prop.pair>\n")
	}
	b.WriteString("</prop.list>\n") // end Stops List
	b.WriteString("</prop.pair>\n")
	b.WriteString("<prop.pair>\n")
	b.WriteString("<key>Stops Size</key>\n")
	b.WriteString("<int type='unsigned' size='32'>" + strconv.Itoa(n) + "</int>\n")
	b.WriteString("</prop.pair>\n")
	b.WriteString("</prop.list>\n") // end group
	b.WriteString("</prop.pair>\n")
}

// writeStopArray emits one "Stops Alpha" / "Stops Color" prop.pair holding a
// float array.
func writeStopArray(b *strings.Builder, key string, vals ...float64) {
	b.WriteString("<prop.pair>\n")
	b.WriteString("<key>" + key + "</key>\n")
	b.WriteString("<array>\n")
	b.WriteString("<array.type><float/></array.type>\n")
	for _, v := range vals {
		b.WriteString("<float>" + fmtGradFloat(v) + "</float>\n")
	}
	b.WriteString("</array>\n")
	b.WriteString("</prop.pair>\n")
}

// fmtGradFloat formats a gradient scalar the way AE's XML does: integral
// values as bare ints ("1", "0"), everything else as a minimal decimal.
func fmtGradFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
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
