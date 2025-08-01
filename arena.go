package garena

import (
	"syscall"
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

// New allocates at least size bytes from the system and initializes the Arena.
//
// Call Destroy to release the memory pages to the OS.
func New(size uintptr) *Arena {
	var (
		a   Arena
		cap uintptr
		err error
	)

	cap = uintptr(syscall.Getpagesize())
	cap = ptrAlign(size, cap)
	a.mem, err = syscall.Mmap(
		-1,
		0,
		int(cap),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE,
	)
	if err != nil {
		panic(err)
	}
	a.cap = cap
	a.base = uintptr(unsafe.Pointer(&a.mem[0]))

	return &a
}

// Destroy releases the resources to the OS, we assume there are no live pointers
// allocated, otherwise the behavior is undefined
func (a *Arena) Destroy() {
	clear(a.mem)
	syscall.Munmap(a.mem)
	*a = Arena{}
}

// FreeAll clears the arena for reuse
func (a *Arena) FreeAll() {
	clear(a.mem)
	a.len = 0
}

// Alloc reserves data to store sizeof(T) and returns the aligned pointer
func Alloc[T any](a *Arena) (val *T) {
	return (*T)(unsafe.Pointer(
		alloc(a, unsafe.Sizeof(*val), unsafe.Alignof(*val))),
	)
}

// AllocSlice reserves data to store sizeof(T * max(len, cap)) and returns the
// aligned pointer
func AllocSlice[T any](a *Arena, len, cap uintptr) []T {
	var (
		tdest T
		sdest internalSlice
	)

	cap = max(len, cap)
	sdest.len = len
	sdest.cap = cap
	sdest.ptr = alloc(
		a, unsafe.Sizeof(tdest)*cap, unsafe.Alignof(tdest),
	)

	return *((*[]T)(unsafe.Pointer(&sdest)))
}

func ptrAlign(ptr, align uintptr) uintptr {
	return (ptr + align - 1) & ^(align - 1)
}

func alloc(a *Arena, size, align uintptr) uintptr {
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
