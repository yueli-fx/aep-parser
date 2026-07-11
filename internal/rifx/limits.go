package rifx

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Limits bounds the resources consumed while parsing one RIFX input.
type Limits struct {
	MaxInputBytes     uint64
	MaxChunkBytes     uint64
	MaxAllocatedBytes uint64
	MaxNodes          int
	MaxDepth          int
}

// DefaultLimits is suitable for local project files while still rejecting
// declarations that could exhaust a process before the input is validated.
var DefaultLimits = Limits{
	MaxInputBytes:     1 << 30,
	MaxChunkBytes:     512 << 20,
	MaxAllocatedBytes: 1 << 30,
	MaxNodes:          1_000_000,
	MaxDepth:          256,
}

// LimitError reports which parser resource budget was exceeded.
type LimitError struct {
	Resource string
	Limit    uint64
	Actual   uint64
}

func (e *LimitError) Error() string {
	return fmt.Sprintf("rifx: %s limit exceeded: got %d, limit %d", e.Resource, e.Actual, e.Limit)
}

// FormatError reports invalid chunk structure at a byte offset.
type FormatError struct {
	Offset  int64
	Message string
}

func (e *FormatError) Error() string {
	return fmt.Sprintf("rifx: invalid format at offset %d: %s", e.Offset, e.Message)
}

type parser struct {
	r              io.ReadSeeker
	limits         Limits
	inputEnd       int64
	nodes          int
	allocatedBytes uint64
}

func newParser(r io.ReadSeeker, limits Limits) (*parser, error) {
	limits = normalizeLimits(limits)
	start, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("rifx: get input position: %w", err)
	}
	end, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("rifx: get input size: %w", err)
	}
	if _, err := r.Seek(start, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rifx: restore input position: %w", err)
	}
	if end < start {
		return nil, &FormatError{Offset: start, Message: "input end precedes current position"}
	}
	inputBytes := uint64(end - start)
	if inputBytes > limits.MaxInputBytes {
		return nil, &LimitError{Resource: "input bytes", Limit: limits.MaxInputBytes, Actual: inputBytes}
	}
	return &parser{r: r, limits: limits, inputEnd: end}, nil
}

func normalizeLimits(limits Limits) Limits {
	if limits.MaxInputBytes == 0 {
		limits.MaxInputBytes = DefaultLimits.MaxInputBytes
	}
	if limits.MaxChunkBytes == 0 {
		limits.MaxChunkBytes = DefaultLimits.MaxChunkBytes
	}
	if limits.MaxAllocatedBytes == 0 {
		limits.MaxAllocatedBytes = DefaultLimits.MaxAllocatedBytes
	}
	if limits.MaxNodes <= 0 {
		limits.MaxNodes = DefaultLimits.MaxNodes
	}
	if limits.MaxDepth <= 0 {
		limits.MaxDepth = DefaultLimits.MaxDepth
	}
	return limits
}

func (p *parser) readChunk(depth int, parentEnd int64) (*Chunk, error) {
	offset, err := p.r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	if depth > p.limits.MaxDepth {
		return nil, &LimitError{Resource: "chunk depth", Limit: uint64(p.limits.MaxDepth), Actual: uint64(depth)}
	}
	p.nodes++
	if p.nodes > p.limits.MaxNodes {
		return nil, &LimitError{Resource: "chunk nodes", Limit: uint64(p.limits.MaxNodes), Actual: uint64(p.nodes)}
	}
	if parentEnd-offset < 8 {
		return nil, &FormatError{Offset: offset, Message: "chunk header exceeds containing payload"}
	}

	var id ChunkID
	if _, err := io.ReadFull(p.r, id[:]); err != nil {
		return nil, fmt.Errorf("read chunk ID: %w", err)
	}
	var size uint32
	if err := binary.Read(p.r, binary.BigEndian, &size); err != nil {
		return nil, fmt.Errorf("chunk %q: read size: %w", id, err)
	}
	if uint64(size) > p.limits.MaxChunkBytes {
		return nil, &LimitError{Resource: "chunk bytes", Limit: p.limits.MaxChunkBytes, Actual: uint64(size)}
	}
	payloadStart := offset + 8
	payloadEnd := payloadStart + int64(size)
	pad := int64(size & 1)
	if id == IDRifx {
		pad = 0
	}
	if payloadEnd < payloadStart || payloadEnd+pad > parentEnd || payloadEnd+pad > p.inputEnd {
		return nil, &FormatError{Offset: offset, Message: fmt.Sprintf("chunk %q payload size %d exceeds containing payload", id, size)}
	}

	chunk := &Chunk{ID: id, Size: size}
	if id == IDList || id == IDRifx {
		if size < 4 {
			return nil, &FormatError{Offset: offset, Message: fmt.Sprintf("container %q size %d is smaller than form type", id, size)}
		}
		if _, err := io.ReadFull(p.r, chunk.FormType[:]); err != nil {
			return nil, fmt.Errorf("chunk %q: read form type: %w", id, err)
		}
		if opaqueListTypes[chunk.FormType] {
			dataLen := uint64(size - 4)
			if err := p.reserve(dataLen); err != nil {
				return nil, err
			}
			chunk.Data = make([]byte, int(dataLen))
			if _, err := io.ReadFull(p.r, chunk.Data); err != nil {
				return nil, fmt.Errorf("chunk %q/%q: read opaque payload: %w", id, chunk.FormType, err)
			}
		} else {
			for {
				pos, err := p.r.Seek(0, io.SeekCurrent)
				if err != nil {
					return nil, err
				}
				if pos == payloadEnd {
					break
				}
				if pos > payloadEnd || payloadEnd-pos < 8 {
					return nil, &FormatError{Offset: pos, Message: fmt.Sprintf("container %q/%q has an incomplete child header", id, chunk.FormType)}
				}
				child, err := p.readChunk(depth+1, payloadEnd)
				if err != nil {
					return nil, fmt.Errorf("child of %q/%q: %w", id, chunk.FormType, err)
				}
				chunk.Children = append(chunk.Children, child)
			}
		}
	} else {
		if err := p.reserve(uint64(size)); err != nil {
			return nil, err
		}
		chunk.Data = make([]byte, int(size))
		if _, err := io.ReadFull(p.r, chunk.Data); err != nil {
			return nil, fmt.Errorf("chunk %q: read data: %w", id, err)
		}
	}
	if pad != 0 {
		if _, err := p.r.Seek(pad, io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("chunk %q: skip padding: %w", id, err)
		}
	}
	return chunk, nil
}

func (p *parser) reserve(size uint64) error {
	maxInt := uint64(^uint(0) >> 1)
	if size > maxInt {
		return &LimitError{Resource: "addressable payload bytes", Limit: maxInt, Actual: size}
	}
	actual := p.allocatedBytes + size
	if actual < p.allocatedBytes || actual > p.limits.MaxAllocatedBytes {
		return &LimitError{Resource: "allocated payload bytes", Limit: p.limits.MaxAllocatedBytes, Actual: actual}
	}
	p.allocatedBytes = actual
	return nil
}
