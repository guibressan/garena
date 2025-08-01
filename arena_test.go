package garena

import (
	"bytes"
	"testing"
)

func Test_ArenaAlloc(t *testing.T) {
	const n = 3
	var (
		a Arena
		b *byte
		u *int64
	)

	a = ArenaInit(1 << 20)
	defer ArenaFreeAll(&a)

	for range n {
		b = ArenaAlloc[byte](&a)
		*b = 0xCC
	}

	assert(*b == 0xCC)
	assert(a.len == n)
	assert(a.cap == 1<<20)
	assert(bytes.Equal(a.mem[:4], []byte{0xCC, 0xCC, 0xCC, 0x00}))

	u = ArenaAlloc[int64](&a)
	// 3 + pad 5 + 8 == 16
	*u = ^0
	assert(a.len == 16)
}

func Test_ArenaAllocSlice(t *testing.T) {
	var (
		a Arena
		p []byte
	)

	a = ArenaInit(10 << 20)
	defer ArenaFreeAll(&a)

	p = ArenaAllocSlice[byte](&a, 1, 2)

	p[0] = 0xCC
	p = append(p, 0xDD)

	assert(len(p) == 2)
	assert(cap(p) == 2)
	assert(bytes.Equal(p, []byte{0xCC, 0xDD}))
}

func Test_ptrAlign(t *testing.T) {
	tests := []struct {
		ptr      uintptr
		align    uintptr
		expected uintptr
	}{
		{1, 8, 8},
		{8, 8, 8},
		{1, 2, 2},
		{3, 16, 16},
	}

	for _, v := range tests {
		t.Run("", func(t *testing.T) {
			assert(v.expected == ptrAlign(v.ptr, v.align))
		})
	}
}

func BenchmarkArena(b *testing.B) {
	const size = 5 << 20
	var (
		a Arena
		s []byte
	)

	_ = s

	a = ArenaInit(size)

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
