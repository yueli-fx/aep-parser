package recipe

type CapabilityStatus string

const (
	CapabilitySupported   CapabilityStatus = "supported"
	CapabilityUnsupported CapabilityStatus = "unsupported"
	CapabilityUnknown     CapabilityStatus = "unknown"
)

type CapabilityIndex interface {
	Lookup(query string) CapabilityStatus
}

type StaticCapabilities map[string]CapabilityStatus

func (s StaticCapabilities) Lookup(query string) CapabilityStatus {
	if status, ok := s[query]; ok {
		return status
	}
	return CapabilityUnknown
}

type unknownCapabilities struct{}

func (unknownCapabilities) Lookup(string) CapabilityStatus {
	return CapabilityUnknown
}
