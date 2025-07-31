package garena

import (
	"testing"
	"bytes"
)

func Test_ArenaAlloc(t *testing.T) {
	const n = 3
	var (
		a Arena
		b *byte
	)

	ArenaInit(&a, n+1)
	defer ArenaFreeAll(&a)

	for range n {
		b = ArenaAlloc[byte](&a)
		*b = 0xCC
	}

	assert(*b == 0xCC)
	assert(a.len == n)
	assert(a.cap == n+1)
	assert(bytes.Equal(a.mem, []byte{0xCC, 0xCC, 0xCC, 0x00}))
}

func Test_ArenaAllocSlice(t *testing.T) {
	var (
		a Arena
		p []byte
	)

	ArenaInit(&a, 10 << 20)
	defer ArenaFreeAll(&a)

	p = ArenaAllocSlice[byte](&a, 1, 2)

	p[0] = 0xCC
	p = append(p, 0xDD);

	assert(len(p) == 2)
	assert(cap(p) == 2)
	assert(bytes.Equal(p, []byte{0xCC, 0xDD}))
}

func BenchmarkArena(b *testing.B) {
	const size = 5 << 20
	var (
		a Arena
		s []byte
	)

	_ = s
	
	ArenaInit(&a, size)

	for b.Loop() {
		s = ArenaAllocSlice[byte](&a, size, size)
		ArenaFreeAll(&a)
	}
}

func BenchmarkGC(b *testing.B) {
	const size = 5 << 20
	var s []byte

	_ = s

	for b.Loop() {
		s = make([]byte, size, size)
	}
}

func assert(cond bool) {
	if cond {
		return
	}
	panic("assertion failure")
}
