package garena

import (
	"unsafe"
)

type Arena struct {
	mem []byte
	len uintptr
	cap uintptr
}

func ArenaInit(a *Arena, size uintptr) {
	*a = Arena{
		mem: make([]byte, size),
		cap: size,
	}
}

func ArenaAlloc[T any](a *Arena) *T {
	var (
		size uintptr
		dest *T
	)
	
	size = unsafe.Sizeof(*dest)

	if a.cap - a.len < size {
		panic("ARENA FULL")
	}
	
	dest = (*T)(unsafe.Pointer(&a.mem[a.len]))
	a.len += size
	return dest
}

type internalSlice struct {
	ptr unsafe.Pointer
	len uintptr
	cap uintptr
}

 func ArenaAllocSlice[T any](a *Arena, len, cap uintptr) []T {
 	var (
 		size uintptr
 		tdest T
 		sdest internalSlice
		tmp []T
 	)
 	
 	size = (unsafe.Sizeof(tdest) * cap)
 
 	if a.cap - a.len < size {
 		panic("ARENA FULL")
 	}
 
 	sdest.len = len
 	sdest.cap = cap
 	sdest.ptr = unsafe.Pointer(&a.mem[a.len])

	tmp = *((*[]T)(unsafe.Pointer(&sdest)))
 	
 	a.len += size
 	return tmp
 }

func ArenaFreeAll(a *Arena) {
	a.len = 0;
}
