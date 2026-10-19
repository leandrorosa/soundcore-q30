// IOBluetooth backend.

package rfcomm

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework IOBluetooth
#include <stdlib.h>
#include "rfcomm_darwin.h"
*/
import "C"

import (
	"fmt"
	"io"
	"runtime"
	"time"
	"unsafe"
)

// IOBluetooth delivers channel callbacks on the run loop of the thread that
// opened the channel, so every call must stay on the main OS thread.
func init() { runtime.LockOSThread() }

func check(op string, r C.int) error {
	if r != 0 {
		return fmt.Errorf("%s: IOReturn %#x", op, uint32(r))
	}
	return nil
}

type Conn struct{ h unsafe.Pointer }

func Dial(addr string) (*Conn, error) {
	a := C.CString(addr)
	defer C.free(unsafe.Pointer(a))
	var h unsafe.Pointer
	if err := check("open rfcomm", C.q30_open(a, &h)); err != nil {
		return nil, err
	}
	return &Conn{h}, nil
}

func (c *Conn) Write(b []byte) error {
	return check("write", C.q30_write(c.h, unsafe.Pointer(&b[0]), C.int(len(b))))
}

// Read returns the next chunk of received bytes, or nil once timeout elapses.
func (c *Conn) Read(timeout time.Duration) ([]byte, error) {
	buf := make([]byte, 1024)
	n := C.q30_read(c.h, unsafe.Pointer(&buf[0]), C.int(len(buf)), C.double(timeout.Seconds()))
	if n < 0 {
		return nil, io.EOF
	}
	return buf[:n], nil
}

func (c *Conn) Close() { C.q30_close(c.h) }

func withAddr(addr string, f func(*C.char) C.int) C.int {
	a := C.CString(addr)
	defer C.free(unsafe.Pointer(a))
	return f(a)
}

func Connect(addr string) error {
	return check("connect", withAddr(addr, func(a *C.char) C.int { return C.q30_connect(a) }))
}

func Disconnect(addr string) error {
	return check("disconnect", withAddr(addr, func(a *C.char) C.int { return C.q30_disconnect(a) }))
}

func Connected(addr string) bool {
	return withAddr(addr, func(a *C.char) C.int { return C.q30_connected(a) }) != 0
}

// Find returns the address of the first paired device whose name contains name.
func Find(name string) (string, error) {
	n := C.CString(name)
	defer C.free(unsafe.Pointer(n))
	a := C.q30_find(n)
	if a == nil {
		return "", fmt.Errorf("%w matching %q", ErrNotFound, name)
	}
	defer C.free(unsafe.Pointer(a))
	return C.GoString(a), nil
}
