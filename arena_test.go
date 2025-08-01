package garena

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"
)

func Test_Alloc(t *testing.T) {
	const n = 3
	var (
		a *Arena
		b *byte
		u *int64
	)

	a = New(1 << 20)
	defer a.Destroy()

	for range n {
		b = Alloc[byte](a)
		*b = 0xCC
	}

	assert(*b == 0xCC)
	assert(a.len == n)
	assert(a.cap == 1<<20)
	assert(bytes.Equal(a.mem[:4], []byte{0xCC, 0xCC, 0xCC, 0x00}))

	u = Alloc[int64](a)
	// 3 + pad 5 + 8 == 16
	*u = ^0
	assert(a.len == 16)
}

func Test_AllocSlice(t *testing.T) {
	var (
		a *Arena
		p []byte
	)

	a = New(10 << 20)
	defer a.Destroy()

	p = AllocSlice[byte](a, 1, 2)

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

func BenchmarkArenaAlloc(b *testing.B) {
	const size = 5 << 20
	var (
		a *Arena
		s []byte
	)

	_ = s

	a = New(size)
	defer a.Destroy()

	for b.Loop() {
		s = AllocSlice[byte](a, size, size)
		a.FreeAll()
	}
}

func BenchmarkGCAlloc(b *testing.B) {
	const size = 5 << 20
	var s []byte

	_ = s

	for b.Loop() {
		s = make([]byte, size, size)
	}
}

func BenchmarkArenaStress(b *testing.B) {
	const (
		npad   = 32
		nnodes = 26843545
	)

	type node struct {
		parent *node
		pad    [npad]byte
	}

	var (
		a *Arena
	)

	a = New(unsafe.Sizeof(node{}) * nnodes)
	defer a.Destroy()

	for b.Loop() {
		var (
			tail *node
			swap *node
		)
		for range nnodes {
			swap = Alloc[node](a)
			*swap = node{parent: tail}
			tail = swap
		}
		tail = nil
		a.FreeAll()
	}
}

func BenchmarkGCStress(b *testing.B) {
	const (
		npad   = 32
		nnodes = 26843545
	)

	type node struct {
		parent *node
		pad    [npad]byte
	}

	for b.Loop() {
		var (
			tail *node
		)
		for range nnodes {
			tail = &node{parent: tail}
		}
		tail = nil
		runtime.GC()
	}
}

func assert(cond bool) {
	if cond {
		return
	}
	panic("assertion failure")
}
