// +build go1.10

package winsys

import (
	_ "runtime/cgo"
	"sync"
	"syscall"
)

type DLL struct {
	*syscall.DLL
	Name   string
	Handle syscall.Handle
}

func (d *DLL) FindProc(name string) (*Proc, error) {
	//fmt.Printf("FindProc: %+v %s \n", d, name)

	p, err := d.DLL.FindProc(name)
	//fmt.Printf("FindProc ret: %+v %+v \n", p, err)
	if err != nil {
		return nil, err
	}
	p2 := &Proc{
		Dll:  p.Dll,
		Name: p.Name,
		addr: p.Addr(),
	}
	//fmt.Printf("Proc ret: %+v \n", p2)
	return p2, nil
}

type LazyDLL struct {
	*syscall.LazyDLL
	mu   sync.Mutex
	dll  *DLL // non nil once DLL is loaded
	Name string
}

func NewLazyDLL(name string) *LazyDLL {
	return &LazyDLL{LazyDLL: &syscall.LazyDLL{Name: name}, Name: name}
}

// Load loads DLL file d.Name into memory. It returns an error if fails.
// Load will not try to load DLL, if it is already loaded into memory.
func (d *LazyDLL) Load() error {
	// Non-racy version of:
	if d.dll == nil {
		// if atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&d.dll))) == nil {
		d.mu.Lock()
		defer d.mu.Unlock()
		//fmt.Printf("Load: %+v\n", d)

		if d.dll == nil {
			dll, e := syscall.LoadDLL(d.Name)
			//fmt.Printf("Load: %+v %+v \n", dll, e)
			if e != nil {
				return e
			}
			// Non-racy version of:
			d.dll = &DLL{Name: dll.Name, Handle: dll.Handle, DLL: dll}
			// atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&d.dll)), unsafe.Pointer(dll))
		}
	}
	return nil
}

func (d *LazyDLL) NewProc(name string) *LazyProc {
	ret := &LazyProc{LazyProc: d.LazyDLL.NewProc(name), l: d, Name: name}
	//fmt.Printf("NewProc LazyDLL2 %+v", ret)
	return ret
}

type LazyProc struct {
	*syscall.LazyProc
	Name string
	l    *LazyDLL
	proc *Proc
	mu   sync.Mutex
}

func (p *LazyProc) Find() error {
	// Non-racy version of:
	//fmt.Println("Finding PROC")
	if p.proc == nil {
		// if atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&p.proc))) == nil {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.proc == nil {
			e := p.l.Load()
			if e != nil {
				return e
			}
			proc, e := p.l.dll.FindProc(p.Name)
			//fmt.Printf("Finding PROC ret: %+v %+v\n", e, proc)
			if e != nil {
				return e
			}
			// Non-racy version of:
			p.proc = proc
			//fmt.Printf("Finding PROC: %+v %+v\n", p, proc)
			// atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&p.proc)), unsafe.Pointer(proc))
		}
	}
	return nil
}

func (p *LazyProc) Call(a ...uintptr) (r1, r2 uintptr, lastErr error) {
	//fmt.Println("LazyProc Call")
	e := p.Find()
	if e != nil {
		//fmt.Println("LazyProc Call panic")
		panic(e)
	}
	//fmt.Println("Calling")

	return p.proc.Call(a...)
}

type Proc struct {
	*syscall.Proc
	Dll  *syscall.DLL
	Name string
	addr uintptr
}

func (p *Proc) Addr() uintptr {
	return p.addr
}

func (p *Proc) Call(a ...uintptr) (r1, r2 uintptr, lastErr error) {
	//fmt.Println("Proc Call")
	//fmt.Printf("Proc: %+v\n", p)
	//fmt.Printf("nparams: %d\n", len(a))
	//fmt.Printf("params: %+v\n", *(*[10]uintptr)(unsafe.Pointer(&a[0])))
	//fmt.Printf("addr: %d\n", p.Addr())

	a1, a2, err := Syscall(p.Addr(), uintptr(len(a)), a...)
	return a1, a2, syscall.Errno(err)
	// return 1, 2, nil
}
