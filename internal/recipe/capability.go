package recipe

type CapabilityStatus string

const (
	CapabilitySupported   CapabilityStatus = "supported"
	CapabilityUnsupported CapabilityStatus = "unsupported"
	CapabilityUnknown     CapabilityStatus = "unknown"
)

type CapabilityLookup struct {
	Query    string           `json:"query"`
	Status   CapabilityStatus `json:"status"`
	Symbol   string           `json:"symbol,omitempty"`
	Domain   string           `json:"domain,omitempty"`
	Tier     string           `json:"tier,omitempty"`
	Verify   string           `json:"verify,omitempty"`
	MinVer   string           `json:"minver,omitempty"`
	Boundary string           `json:"boundary,omitempty"`
	Gate     []string         `json:"gate,omitempty"`
}

type CapabilityIndex interface {
	Lookup(query string) CapabilityLookup
}

type StaticCapabilities map[string]CapabilityStatus

func (s StaticCapabilities) Lookup(query string) CapabilityLookup {
	if status, ok := s[query]; ok {
		return CapabilityLookup{Query: query, Status: status}
	}
	return CapabilityLookup{Query: query, Status: CapabilityUnknown}
}

type unknownCapabilities struct{}

func (unknownCapabilities) Lookup(query string) CapabilityLookup {
	return CapabilityLookup{Query: query, Status: CapabilityUnknown}
}
