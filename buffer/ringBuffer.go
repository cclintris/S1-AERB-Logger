package buffer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

/*
*************************************************************

	VARIABLE

*************************************************************
*/

var (
	ErrIsEmpty = errors.New("ring buffer is empty")

	once sync.Once // implement singleton
)

/*
*************************************************************

	STRUCT DEFINITION

*************************************************************
*/

// RingBuffer
type RingBuffer struct {
	buf []byte // buffer

	initSize int // initial size of buffer
	size     int // dynamic size of buffer
	maxSize  int // maximum size of buffer
	extCoef  int // coefficient for extending buffer strategy

	vr int // virtual read pointer
	r  int // logical read pointer
	w  int // logical write pointer

	isEmpty bool
}

/*
*************************************************************

	MAIN API

*************************************************************
*/

// Returns a ringbuffer initialized with a given default size and a given maximum size.
func (rb *RingBuffer) Init(defaultSize int, maxSize int, extCoef int) *RingBuffer {
	once.Do(func() {
		rb.buf = make([]byte, defaultSize)
		rb.initSize = defaultSize
		rb.size = defaultSize
		rb.maxSize = maxSize
		rb.extCoef = extCoef
		rb.isEmpty = true
		rb.r = 0
		rb.w = 0
		rb.vr = 0
	})
	return rb
}

/*
Refreshes the virtual read pointer.
Note: Should be used with Virtual[*] functions.
*/
func (rb *RingBuffer) VirtualRefresh() {
	rb.r = rb.vr
	if rb.r == rb.w {
		rb.isEmpty = true
	}
}

/*
Reverts the virtual read pointer.
Note: Should be used with Virtual[*] functions.
*/
func (rb *RingBuffer) VirtualRevert() {
	rb.vr = rb.r
}

/*
Virtually reads buffer without moving read pointer.
Note: Only the virtual read pointer will be modified, logical read pointer remains stable.
Note: Should be used with Virtual[] functions.
*/
func (rb *RingBuffer) VirtualRead(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if rb.isEmpty {
		return 0, ErrIsEmpty
	}
	n = len(p)

	// write pointer ahead of read pointer
	if rb.w > rb.vr {
		if n > rb.w-rb.vr {
			n = rb.w - rb.vr
		}
		copy(p, rb.buf[rb.vr:rb.vr+n])

		rb.vr = (rb.vr + n) % rb.size
		if rb.vr == rb.w {
			rb.isEmpty = true
		}
		return n, err
	}

	// write pointer behind read pointer
	// cycle formed
	if n > rb.size-rb.vr+rb.w {
		n = rb.size - rb.vr + rb.w
	}
	if rb.vr+n <= rb.size {
		copy(p, rb.buf[rb.vr:rb.vr+n])
	} else {
		// tail
		c1 := rb.size - rb.vr
		copy(p, rb.buf[rb.vr:rb.size])
		// head
		c2 := n - c1
		copy(p[c1:], rb.buf[0:c2])
	}

	rb.vr = (rb.vr + n) % rb.size
	return n, err
}

/*
Returns the number of bytes available to read virtually.
Note: Should be used with Virtual[] functions.
*/
func (rb *RingBuffer) VirtualLength() int {
	if rb.w == rb.vr {
		if rb.isEmpty {
			return 0
		}
		return rb.size
	}

	if rb.w > rb.vr {
		return rb.w - rb.vr
	}
	return rb.size - rb.vr + rb.w
}

/*
Reads buffer content into p.
Returns the number of bytes read (0 <= n <= len(p)) and any error encountered.
Note: Both logical and virtual read pointer will be modified.
*/
func (rb *RingBuffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if rb.isEmpty {
		return 0, ErrIsEmpty
	}
	n = len(p)

	// write pointer ahead of read pointer
	if rb.w > rb.r {
		if n > rb.w-rb.r {
			n = rb.w - rb.r
		}
		copy(p, rb.buf[rb.r:rb.r+n])

		rb.r = (rb.r + n) % rb.size
		if rb.r == rb.w {
			rb.isEmpty = true
		}
		rb.vr = rb.r
		return n, err
	}

	// write pointer behind read pointer
	// cycle formed
	if n > rb.size-rb.r+rb.w {
		n = rb.size - rb.r + rb.w
	}
	if rb.r+n <= rb.size {
		copy(p, rb.buf[rb.r:rb.r+n])
	} else {
		// tail
		c1 := rb.size - rb.r
		copy(p, rb.buf[rb.r:rb.size])
		// head
		c2 := n - c1
		copy(p[c1:], rb.buf[0:c2])
	}

	rb.r = (rb.r + n) % rb.size
	if rb.r == rb.w {
		rb.isEmpty = true
	}
	rb.vr = rb.r
	return n, err
}

// Reads and returns the next byte from the buffer or returns error.
// Note: Both logical and virtual read pointer will be modified.
func (rb *RingBuffer) ReadByte() (b byte, err error) {
	if rb.isEmpty {
		return 0, ErrIsEmpty
	}

	b = rb.buf[rb.r]
	rb.r++

	if rb.r == rb.size {
		rb.r = 0
	}
	if rb.w == rb.r {
		rb.isEmpty = true
	}
	rb.vr = rb.r

	return b, err
}

// Consumes all available bytes to without returning them.
func (rb *RingBuffer) ConsumeAll() {
	rb.r = 0
	rb.w = 0
	rb.vr = 0
	rb.isEmpty = true
}

// Consumes len bytes without returning them.
func (rb *RingBuffer) Consume(len int) {
	if rb.isEmpty || len <= 0 {
		return
	}

	if len < rb.Length() {
		rb.r = (rb.r + len) % rb.size
		rb.vr = rb.r

		if rb.w == rb.r {
			rb.isEmpty = true
		}
	} else {
		rb.ConsumeAll()
	}
}

/*
Writes len(p) bytes from p to the underlying buffer.
If buffer available size not enough to write, there are two scenarios:
 1. if maximum size of buffer is not reached, allocate additional memory
 2. if maximum size of buffer reached, overwrite old data until idle buffer space is enough

Returns the number of bytes written from p (0 <= n <= len(p)) and any error encountered that caused write to stop early.
*/
func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	n = len(p)
	free := rb.Free()
	if n > free {
		if !rb.isMaximumReached() {
			// allocate additional (n - free) memory
			rb.alloc(n - free)
		} else {
			// overwrite old logs
			rb.overwrite(free, n, true)
		}
	}

	if rb.w >= rb.r {
		if rb.size-rb.w >= n {
			copy(rb.buf[rb.w:], p)
			rb.w += n
		} else {
			copy(rb.buf[rb.w:], p[:rb.size-rb.w])
			copy(rb.buf[0:], p[rb.size-rb.w:])
			rb.w += n - rb.size
		}
	} else {
		copy(rb.buf[rb.w:], p)
		rb.w += n
	}

	if rb.w == rb.size {
		rb.w = 0
	}

	rb.isEmpty = false
	return n, err
}

// Writes one byte into buffer, and returns ErrIsFull if buffer is full.
func (rb *RingBuffer) WriteByte(b byte) error {
	if rb.Free() < 1 {
		if !rb.isMaximumReached() {
			// allocate additional 1 byte memory
			rb.alloc(1)
		} else {
			// overwrite old data
			rb.overwrite(rb.Free(), 1, false)
		}
	}

	rb.buf[rb.w] = b
	rb.w++

	if rb.w == rb.size {
		rb.w = 0
	}

	rb.isEmpty = false
	return nil
}

// Returns the number of bytes available to read
func (rb *RingBuffer) Length() int {
	if rb.w == rb.r {
		if rb.isEmpty {
			return 0
		}
		return rb.size
	}

	if rb.w > rb.r {
		return rb.w - rb.r
	}
	return rb.size - rb.r + rb.w
}

// Returns the underlying size of buffer
func (rb *RingBuffer) Capacity() int {
	return rb.size
}

// Writes the contents of the string s to buffer, which accepts a slice of bytes.
func (rb *RingBuffer) WriteString(s string) (n int, err error) {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	buf := *(*[]byte)(unsafe.Pointer(&h))
	return rb.Write(buf)
}

// Returns all available read bytes. It does not move the read pointer and only copy the available data.
func (rb *RingBuffer) Bytes() []byte {
	if rb.isEmpty {
		return nil
	}

	if rb.w > rb.r {
		buf := make([]byte, rb.w-rb.r)
		copy(buf, rb.buf[rb.r:rb.w])
		return buf
	}

	buf := make([]byte, rb.size-rb.r+rb.w)
	copy(buf, rb.buf[rb.r:rb.size])
	copy(buf[rb.size-rb.r:], rb.buf[0:rb.w])
	return buf
}

// Checks if buffer is full
func (rb *RingBuffer) IsFull() bool {
	return !rb.isEmpty && rb.w == rb.r
}

// Checks if buffer is empty
func (rb *RingBuffer) IsEmpty() bool {
	return rb.isEmpty
}

// When Reset called, everything will be refreshed and reset to initial state, including the size of the buffer
func (rb *RingBuffer) Reset() {
	rb.r = 0
	rb.vr = 0
	rb.w = 0
	rb.isEmpty = true
	if rb.size > rb.initSize {
		rb.buf = make([]byte, rb.initSize)
		rb.size = rb.initSize
	}
}

func (rb *RingBuffer) String() string {
	return fmt.Sprintf("Ring Buffer: \n\tCapacity: %d\n\tReadable Bytes: %d\n\tWriteable Bytes: %d\n\tBuffer: %s\n", rb.size, rb.Length(), rb.Free(), rb.buf)
}

// Returns the length of available bytes to write.
func (rb *RingBuffer) Free() int {
	if rb.w == rb.r {
		if rb.isEmpty {
			return rb.size
		}
		return 0
	}

	if rb.w < rb.r {
		return rb.r - rb.w
	}
	return rb.size - rb.w + rb.r
}

// Overwrites old data until memory abundant to write new data.
func (rb *RingBuffer) overwrite(free int, need int, logOverriding bool) error {
	if free >= need {
		return nil
	}

	if logOverriding {
		for free < need && !rb.isEmpty {
			rd := make([]byte, 4)
			n, err := rb.Read(rd)

			if n != 4 || err != nil {
				return err
			}

			l := int(binary.LittleEndian.Uint32(rd))
			rd = make([]byte, l)
			n, err = rb.Read(rd)

			if n != l || err != nil {
				return err
			}
			free += n + 4
		}
	} else {
		for free < need && !rb.isEmpty {
			rd := make([]byte, 1)
			n, err := rb.Read(rd)

			if n != 1 || err != nil {
				return err
			}
			free += 1
		}
	}

	return nil
}

// Allocate additional memory for buffer specified by len.
func (rb *RingBuffer) alloc(len int) {
	vLen := rb.VirtualLength()
	newSize := rb.extend(rb.size + len)
	newBuf := make([]byte, newSize)
	oldLen := rb.Length()
	_, _ = rb.Read(newBuf)

	rb.w = oldLen
	rb.r = 0
	rb.vr = oldLen - vLen
	rb.size = newSize
	rb.buf = newBuf
}

/*
	Extend the buffer based on strategy: golang runtime slice append

https://github.com/golang/go/blob/ac0ba6707c1655ea4316b41d06571a0303cc60eb/src/runtime/slice.go#L125

Depending on the extending coefficient:
 1. If the expected capacity is two times larger than the current capacity, extend two times dircetly
 2. If the expected capacity is NOT two times larger than the current capacity
    a. Check if the current capacity has reached the extending coefficient
    a.1 If no, extend two times directly
    a.2 If yes, extend it to 125% of the current capacity until capacity sufficient
*/
func (rb *RingBuffer) extend(expcap int) int {
	newcap := rb.size
	doublecap := newcap + newcap
	if expcap > doublecap {
		newcap = expcap
	} else {
		if rb.size < rb.extCoef {
			newcap = doublecap
		} else {
			// Check 0 < newcap to detect overflow
			// and prevent an infinite loop
			for 0 < newcap && newcap < expcap {
				newcap += newcap / 4
			}
			// Set newcap to the requested cap when
			// the newcap calculation overflowed
			if newcap <= 0 {
				newcap = expcap
			}
		}
	}
	return newcap
}

// Checks if buffer has reached the maximum size specified.
func (rb *RingBuffer) isMaximumReached() bool {
	return rb.size >= rb.maxSize
}

/*
*************************************************************

	TEST API: Should not be used except for testing

*************************************************************
*/

// Returns the value of the write pointer.
func (rb *RingBuffer) GetW() int {
	return rb.w
}

// Returns the value of the read pointer.
func (rb *RingBuffer) GetR() int {
	return rb.r
}

// Returns the value of the virtual read pointer.
func (rb *RingBuffer) GetVR() int {
	return rb.vr
}

// Returns the buffer.
func (rb *RingBuffer) GetBuf() []byte {
	return rb.buf
}
