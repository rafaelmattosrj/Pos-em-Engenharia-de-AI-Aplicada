package splitter

import (
	"strings"
	"testing"
)

func TestRecursive_SingleShortText(t *testing.T) {
	chunks := Recursive("hello world", 1000, 200)
	if len(chunks) != 1 || chunks[0] != "hello world" {
		t.Fatalf("esperava 1 chunk com o texto original, obteve %v", chunks)
	}
}

func TestRecursive_RespectsChunkSize(t *testing.T) {
	text := strings.Repeat("palavra ", 500) // 4000 chars
	chunks := Recursive(text, 1000, 200)

	if len(chunks) < 2 {
		t.Fatalf("esperava multiplos chunks para texto de %d chars, obteve %d", len(text), len(chunks))
	}
	for i, c := range chunks {
		if len(c) > 1000 {
			t.Errorf("chunk %d tem %d chars, excede chunkSize=1000", i, len(c))
		}
	}
}

func TestRecursive_OverlapBetweenChunks(t *testing.T) {
	text := strings.Repeat("a", 50) + " " + strings.Repeat("b", 50) + " " + strings.Repeat("c", 50)
	chunks := Recursive(text, 60, 20)

	if len(chunks) < 2 {
		t.Fatalf("esperava multiplos chunks, obteve %d", len(chunks))
	}

	// O início do segundo chunk deve conter uma porção do final do primeiro.
	tail := chunks[0][len(chunks[0])-20:]
	if !strings.HasPrefix(chunks[1], tail) {
		t.Errorf("chunk[1] nao comeca com o overlap esperado do chunk[0]:\nchunk[0] tail=%q\nchunk[1] head=%q", tail, chunks[1][:min(20, len(chunks[1]))])
	}
}

func TestRecursive_NoSeparatorsHardSplits(t *testing.T) {
	text := strings.Repeat("x", 2500) // uma "palavra" gigante, sem espaços
	chunks := Recursive(text, 1000, 0)

	if len(chunks) != 3 {
		t.Fatalf("esperava 3 chunks (1000+1000+500), obteve %d: %v", len(chunks), lens(chunks))
	}
}

func lens(ss []string) []int {
	out := make([]int, len(ss))
	for i, s := range ss {
		out[i] = len(s)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
