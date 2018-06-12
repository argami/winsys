// +build go1.10

// ideas came from github.com/silentred/gid

package winsys

import (
	"runtime"
	_ "runtime/cgo"
	"unsafe"
)

func funcPC(f interface{}) uintptr {
	return *(*[2]*uintptr)(unsafe.Pointer(&f))[1]
}

type libcall struct {
	fn   uintptr
	n    uintptr // number of parameters
	args uintptr // parameters
	r1   uintptr // return values
	r2   uintptr
	err  uintptr // error number
}

func Syscall(fn, nargs uintptr, args ...uintptr) (r1, r2, err uintptr) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	c := &libcall{} //&getg().m.syscall
	c.fn = fn
	c.n = nargs
	c.args = uintptr(unsafe.Pointer(&args[0]))
	_cgo_runtime_cgocall(unsafe.Pointer(funcPC(asmstdcall)), uintptr(unsafe.Pointer(c)))
	return c.r1, c.r2, c.err
}

func asmstdcall(fn unsafe.Pointer)

//go:linkname _cgo_runtime_cgocall runtime.cgocall
func _cgo_runtime_cgocall(unsafe.Pointer, uintptr) int32
