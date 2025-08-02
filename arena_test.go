package garena

import (
	"bytes"
	"runtime"
	"runtime/debug"
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

func Test_Segfault(t *testing.T) {
	var (
		a     *Arena
		value *byte
	)

	a = New(1)

	value = Alloc[byte](a)
	*value = 0xCC
	assert(0xCC == *value)

	a.FreeAll()

	// use after free is undefined behavior, for now the memory is zeroed,
	// but you can end up changing another allocated value
	assert(0x00 == *value)

	a.Destroy()

	// uncomment to get a segfault
	// assert(0x00 == *value)
}

func Test_MaySegfaultOrUB(t *testing.T) {
	type testData struct {
		data []byte
	}
	var (
		a  *Arena
		td *testData
	)

	a = New(1)
	defer a.Destroy()

	f := func(a *Arena, td **testData) {
		*td = Alloc[testData](a)
		// There's no guarantee that the GC won't free the allocated
		// bytes, since we are allocating outside Go heap
		(*td).data = make([]byte, 100<<20)
		for i := range (*td).data {
			(*td).data[i] = 0xCC
		}
	}

	maySegfaultOrUB := func() {
		for i := range 1_000 {
			f(a, &td)
			runtime.GC()
			debug.FreeOSMemory()
			t.Log(i)
			// MAY segfault
			assert(td.data[0] == 0xCC)
			a.FreeAll()
		}
	}

	debug.SetGCPercent(1)
	debug.SetMemoryLimit(1 << 20)

	// uncomment to maybe crash the program
	// maySegfaultOrUB()
	_ = maySegfaultOrUB

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
