package serializer

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// PreservationDigest fingerprints the complete parsed RIFX tree while
// redacting only byte ranges explicitly owned by a mutation. Equal digests
// prove that every other byte, opaque tail, and chunk-tree relationship is
// unchanged.
type PreservationDigest struct {
	SHA256        string
	NodeCount     int
	DataBytes     int
	RedactedBytes int
}

type preservationRange struct {
	start int
	end   int
}

// StaticValuePreservationDigest returns a byte-sensitive digest of project
// after redacting the cdat/otda ranges written by Property.SetStaticValue for
// properties. It is intentionally serializer-owned: only this package knows
// the concrete back-references and therefore the exact claimed byte ranges.
func StaticValuePreservationDigest(project *scene.Project, properties []*scene.Property) (PreservationDigest, error) {
	back := projectBack(project)
	if back == nil || back.root == nil {
		return PreservationDigest{}, fmt.Errorf("preservation: project has no RIFX backing")
	}
	redactions := make(map[*rifx.Chunk][]preservationRange)
	seen := make(map[*scene.Property]bool)
	for _, property := range properties {
		if property == nil || seen[property] {
			continue
		}
		seen[property] = true
		propertyBacking := propertyBack(property)
		if propertyBacking == nil || propertyBacking.cdat == nil {
			return PreservationDigest{}, fmt.Errorf("preservation: property %q has no static-value backing", property.MatchName)
		}
		claimedBytes := property.Components * 8
		if claimedBytes <= 0 || len(propertyBacking.cdat.Data) < claimedBytes {
			return PreservationDigest{}, fmt.Errorf("preservation: property %q has invalid cdat claim", property.MatchName)
		}
		redactions[propertyBacking.cdat] = append(redactions[propertyBacking.cdat], preservationRange{end: claimedBytes})
		if propertyBacking.otda != nil {
			if len(propertyBacking.otda.Data) < claimedBytes {
				return PreservationDigest{}, fmt.Errorf("preservation: property %q has invalid otda claim", property.MatchName)
			}
			redactions[propertyBacking.otda] = append(redactions[propertyBacking.otda], preservationRange{end: claimedBytes})
		}
	}

	h := sha256.New()
	result := PreservationDigest{}
	writePreservationChunk(h, back.root, redactions, &result)
	result.SHA256 = fmt.Sprintf("%x", h.Sum(nil))
	return result, nil
}

func writePreservationChunk(
	h hash.Hash,
	chunk *rifx.Chunk,
	redactions map[*rifx.Chunk][]preservationRange,
	result *PreservationDigest,
) {
	if chunk == nil {
		writePreservationUint64(h, ^uint64(0))
		return
	}
	result.NodeCount++
	_, _ = h.Write(chunk.ID[:])
	_, _ = h.Write(chunk.FormType[:])
	writePreservationUint64(h, uint64(len(chunk.Data)))
	writePreservationUint64(h, uint64(len(chunk.Children)))

	data := chunk.Data
	result.DataBytes += len(data) + len(chunk.Trailing)
	if ranges := redactions[chunk]; len(ranges) > 0 {
		redacted := append([]byte(nil), data...)
		claimed := make([]bool, len(redacted))
		for _, item := range ranges {
			start, end := item.start, item.end
			if start < 0 {
				start = 0
			}
			if end > len(redacted) {
				end = len(redacted)
			}
			for index := start; index < end; index++ {
				redacted[index] = 0
				if !claimed[index] {
					claimed[index] = true
					result.RedactedBytes++
				}
			}
		}
		data = redacted
	}
	_, _ = h.Write(data)
	writePreservationUint64(h, uint64(len(chunk.Trailing)))
	_, _ = h.Write(chunk.Trailing)
	for _, child := range chunk.Children {
		writePreservationChunk(h, child, redactions, result)
	}
}

func writePreservationUint64(h hash.Hash, value uint64) {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], value)
	_, _ = h.Write(data[:])
}
