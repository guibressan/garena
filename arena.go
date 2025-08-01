package garena

import (
	"unsafe"
)

type Arena struct {
	base uintptr
	mem  []byte
	len  uintptr
	cap  uintptr
}

type internalSlice struct {
	ptr uintptr
	len uintptr
	cap uintptr
}

func ArenaInit(size uintptr) Arena {
	var a Arena

	a.mem = make([]byte, size)
	a.cap = size
	a.base = uintptr(unsafe.Pointer(&a.mem[0]))

	return a
}

func ArenaAlloc[T any](a *Arena) (val *T) {
	return (*T)(unsafe.Pointer(
		arenaAlloc(a, unsafe.Sizeof(*val), unsafe.Alignof(*val))),
	)
}

func ArenaAllocSlice[T any](a *Arena, len, cap uintptr) []T {
	var (
		tdest T
		sdest internalSlice
	)

	sdest.len = len
	sdest.cap = cap
	sdest.ptr = arenaAlloc(
		a, unsafe.Sizeof(tdest)*cap, unsafe.Alignof(tdest),
	)

	return *((*[]T)(unsafe.Pointer(&sdest)))
}

func ArenaFreeAll(a *Arena) {
	clear(a.mem)
	a.len = 0
}

func ptrAlign(ptr, align uintptr) uintptr {
	return (ptr + align - 1) & ^(align - 1)
}

func arenaAlloc(a *Arena, size, align uintptr) uintptr {
	var (
		ptr           uintptr
		effectiveSize uintptr
	)

	ptr = ptrAlign(a.base+a.len, align)
	effectiveSize = ptr - a.base - a.len + size

	if a.cap-a.len < effectiveSize {
		panic("ARENA FULL")
	}

	a.len += effectiveSize

	return ptr
}
