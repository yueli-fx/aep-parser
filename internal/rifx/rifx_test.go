package rifx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

func TestParseRejectsContainerSmallerThanFormType(t *testing.T) {
	data := chunkBytes(IDRifx, 3, []byte{0, 0, 0})
	_, err := Parse(bytes.NewReader(data))
	var formatErr *FormatError
	if !errors.As(err, &formatErr) {
		t.Fatalf("error = %v, want FormatError", err)
	}
}

func TestReadChunkRejectsDeclaredSizeBeforeAllocation(t *testing.T) {
	data := chunkBytes(ChunkID{'d', 'a', 't', 'a'}, 1024, nil)
	_, err := ReadChunkWithLimits(bytes.NewReader(data), Limits{MaxInputBytes: 4096, MaxChunkBytes: 16})
	var limitErr *LimitError
	if !errors.As(err, &limitErr) || limitErr.Resource != "chunk bytes" {
		t.Fatalf("error = %v, want chunk-bytes LimitError", err)
	}
}

func TestParseRejectsChildOutsideParent(t *testing.T) {
	var body bytes.Buffer
	body.Write(IDEgg[:])
	childID := ChunkID{'d', 'a', 't', 'a'}
	body.Write(childID[:])
	_ = binary.Write(&body, binary.BigEndian, uint32(10))
	data := chunkBytes(IDRifx, uint32(body.Len()), body.Bytes())
	_, err := Parse(bytes.NewReader(data))
	var formatErr *FormatError
	if !errors.As(err, &formatErr) {
		t.Fatalf("error = %v, want FormatError", err)
	}
}

func TestParseEnforcesDepthNodeAndAllocationLimits(t *testing.T) {
	t.Run("depth", func(t *testing.T) {
		leaf := &Chunk{ID: ChunkID{'d', 'a', 't', 'a'}}
		for range 4 {
			leaf = &Chunk{ID: IDList, FormType: ChunkID{'N', 'e', 's', 't'}, Children: []*Chunk{leaf}}
		}
		root := &Chunk{ID: IDRifx, FormType: IDEgg, Children: []*Chunk{leaf}}
		_, err := ParseWithLimits(bytes.NewReader(writeChunkBytes(t, root)), Limits{MaxDepth: 2})
		assertLimitResource(t, err, "chunk depth")
	})

	t.Run("nodes", func(t *testing.T) {
		root := &Chunk{ID: IDRifx, FormType: IDEgg, Children: []*Chunk{
			{ID: ChunkID{'o', 'n', 'e', ' '}},
			{ID: ChunkID{'t', 'w', 'o', ' '}},
		}}
		_, err := ParseWithLimits(bytes.NewReader(writeChunkBytes(t, root)), Limits{MaxNodes: 2})
		assertLimitResource(t, err, "chunk nodes")
	})

	t.Run("allocated payload", func(t *testing.T) {
		root := &Chunk{ID: IDRifx, FormType: IDEgg, Children: []*Chunk{
			{ID: ChunkID{'d', 'a', 't', 'a'}, Data: make([]byte, 8)},
		}}
		_, err := ParseWithLimits(bytes.NewReader(writeChunkBytes(t, root)), Limits{MaxAllocatedBytes: 4})
		assertLimitResource(t, err, "allocated payload bytes")
	})
}

func TestChunkAccessorsRejectNegativeOffsets(t *testing.T) {
	chunk := &Chunk{Data: make([]byte, 8)}
	if _, err := chunk.U8(-1); err == nil {
		t.Fatal("U8(-1) succeeded")
	}
	if _, err := chunk.U16(-1); err == nil {
		t.Fatal("U16(-1) succeeded")
	}
	if _, err := chunk.U32(-1); err == nil {
		t.Fatal("U32(-1) succeeded")
	}
}

func FuzzParseNeverPanics(f *testing.F) {
	root := &Chunk{ID: IDRifx, FormType: IDEgg, Children: []*Chunk{{ID: ChunkID{'d', 'a', 't', 'a'}, Data: []byte("seed")}}}
	f.Add(writeChunkBytes(f, root))
	f.Add([]byte("RIFX"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseWithLimits(bytes.NewReader(data), Limits{
			MaxInputBytes:     1 << 20,
			MaxChunkBytes:     1 << 20,
			MaxAllocatedBytes: 1 << 20,
			MaxNodes:          10_000,
			MaxDepth:          64,
		})
	})
}

type testHelper interface {
	Helper()
	Fatalf(string, ...any)
}

func writeChunkBytes(t testHelper, chunk *Chunk) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := chunk.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	return buf.Bytes()
}

func chunkBytes(id ChunkID, size uint32, payload []byte) []byte {
	var buf bytes.Buffer
	buf.Write(id[:])
	_ = binary.Write(&buf, binary.BigEndian, size)
	buf.Write(payload)
	return buf.Bytes()
}

func assertLimitResource(t *testing.T, err error, resource string) {
	t.Helper()
	var limitErr *LimitError
	if !errors.As(err, &limitErr) || limitErr.Resource != resource {
		t.Fatalf("error = %v, want %s LimitError", err, resource)
	}
}
