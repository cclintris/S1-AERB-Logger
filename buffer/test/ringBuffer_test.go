package buffer_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"testing"

	. "gitlab-smartgaia.sercomm.com/s1util/logger/buffer"
)

func TestRingBuffer_interface(t *testing.T) {
	rb := &RingBuffer{}
	rb.Init(1, 1, 1024)
	var _ io.Writer = rb
	var _ io.Reader = rb
	var _ io.StringWriter = rb
	var _ io.ByteReader = rb
	var _ io.ByteWriter = rb
}

func TestRingBuffer_Write(t *testing.T) {
	rb := &RingBuffer{}
	rb.Init(64, 1024, 1024)

	// check empty or full
	if !rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is true but got false")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}

	// check retrieve
	n, err := rb.Write([]byte(strings.Repeat("abcd", 2)))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if n != 8 {
		t.Fatalf("expect write 8 bytes but got %d", n)
	}
	if !bytes.Equal(rb.Bytes(), []byte(strings.Repeat("abcd", 2))) {
		t.Fatalf("expect 8 abcdabcd but got %s. r.w=%d, r.r=%d", rb.Bytes(), rb.GetW(), rb.GetR())
	}
	rb.Consume(5)
	if rb.Length() != 3 {
		t.Fatalf("expect len 3 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if !bytes.Equal(rb.Bytes(), []byte(strings.Repeat("bcd", 1))) {
		t.Fatalf("expect 1 bcd but got %s. r.w=%d, r.r=%d", rb.Bytes(), rb.GetW(), rb.GetR())
	}
	_, err = rb.Write([]byte(strings.Repeat("abcd", 15)))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if rb.Capacity() != 64 {
		t.Fatalf("expect capacity 64 bytes but got %d. r.w=%d, r.r=%d", rb.Capacity(), rb.GetW(), rb.GetR())
	}
	if rb.Length() != 63 {
		t.Fatalf("expect len 63 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if !bytes.Equal(rb.Bytes(), []byte("bcd"+strings.Repeat("abcd", 15))) {
		t.Fatalf("expect 63 ... but got %s. buf %s. r.w=%d, r.r=%d", rb.Bytes(), rb.GetBuf(), rb.GetW(), rb.GetR())
	}
	rb.ConsumeAll()

	// write 4 * 4 = 16 bytes
	n, err = rb.Write([]byte(strings.Repeat("abcd", 4)))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if n != 16 {
		t.Fatalf("expect write 16 bytes but got %d", n)
	}
	if rb.Length() != 16 {
		t.Fatalf("expect len 16 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if !bytes.Equal(rb.Bytes(), []byte(strings.Repeat("abcd", 4))) {
		t.Fatalf("expect 4 abcd but got %s. r.w=%d, r.r=%d", rb.Bytes(), rb.GetW(), rb.GetR())
	}

	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// write 48 bytes, should full
	n, err = rb.Write([]byte(strings.Repeat("abcd", 12)))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if n != 48 {
		t.Fatalf("expect write 48 bytes but got %d", n)
	}
	if rb.Length() != 64 {
		t.Fatalf("expect len 64 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if rb.GetW() != 0 {
		t.Fatalf("expect r.w=0 but got %d. r.r=%d", rb.GetW(), rb.GetR())
	}
	if !bytes.Equal(rb.Bytes(), []byte(strings.Repeat("abcd", 16))) {
		t.Fatalf("expect 16 abcd but got %s. r.w=%d, r.r=%d", rb.Bytes(), rb.GetW(), rb.GetR())
	}

	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if !rb.IsFull() {
		t.Fatalf("expect IsFull is true but got false")
	}

	// write more 4 bytes, should reject
	_, _ = rb.Write([]byte(strings.Repeat("abcd", 1)))
	if rb.Length() != 68 {
		t.Fatalf("expect len 68 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}

	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// reset this ringbuffer and set a long slice
	rb.Reset()
	n, _ = rb.Write([]byte(strings.Repeat("abcd", 20)))
	if n != 80 {
		t.Fatalf("expect write 80 bytes but got %d", n)
	}
	if rb.Length() != 80 {
		t.Fatalf("expect len 80 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if rb.GetW() != 80 {
		t.Fatalf("expect r.w=0 but got %d. r.r=%d", rb.GetW(), rb.GetR())
	}

	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	if !bytes.Equal(rb.Bytes(), []byte(strings.Repeat("abcd", 20))) {
		t.Fatalf("expect 20 abcd but got %s. r.w=%d, r.r=%d", rb.Bytes(), rb.GetW(), rb.GetR())
	}
}

func TestRingBuffer_Read(t *testing.T) {
	rb := &RingBuffer{}
	rb.Init(64, 1024, 1024)

	// check empty or full
	if !rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is true but got false")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}

	// read empty
	buf := make([]byte, 1024)
	n, err := rb.Read(buf)
	if err == nil {
		t.Fatalf("expect an error but got nil")
	}
	if err != ErrIsEmpty {
		t.Fatalf("expect ErrIsEmpty but got nil")
	}
	if n != 0 {
		t.Fatalf("expect read 0 bytes but got %d", n)
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if rb.GetR() != 0 {
		t.Fatalf("expect r.r=0 but got %d. r.w=%d", rb.GetR(), rb.GetW())
	}

	// write 16 bytes to read
	_, _ = rb.Write([]byte(strings.Repeat("abcd", 4)))
	n, err = rb.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if n != 16 {
		t.Fatalf("expect read 16 bytes but got %d", n)
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 bytes but got %d. r.w=%d, r.r=%d, r.isEmpy=%t", rb.Length(), rb.GetW(), rb.GetR(), rb.IsEmpty())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if rb.GetR() != 16 {
		t.Fatalf("expect r.r=16 but got %d. r.w=%d", rb.GetR(), rb.GetW())
	}

	// write long slice to  read
	_, _ = rb.Write([]byte(strings.Repeat("abcd", 20)))
	n, err = rb.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if n != 80 {
		t.Fatalf("expect read 80 bytes but got %d", n)
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if rb.GetR() != 80 {
		t.Fatalf("expect r.r=16 but got %d. r.w=%d", rb.GetR(), rb.GetW())
	}
}

func TestRingBuffer_Peek(t *testing.T) {
	rb := &RingBuffer{}
	rb.Init(16, 16, 1024)

	buf := make([]byte, 8)
	// write 16 bytes to read
	_, _ = rb.Write([]byte(strings.Repeat("abcd", 4)))
	n, err := rb.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if n != 8 {
		t.Fatalf("expect read 8 bytes but got %d", n)
	}
	if rb.Length() != 8 {
		t.Fatalf("expect len 8 bytes but got %d. r.w=%d, r.r=%d, r.isEmpy=%t", rb.Length(), rb.GetW(), rb.GetR(), rb.IsEmpty())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d bytes but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if rb.GetR() != 8 {
		t.Fatalf("expect r.r=8 but got %d. r.w=%d", rb.GetR(), rb.GetW())
	}
}

func TestRingBuffer_ByteInterface(t *testing.T) {
	rb := &RingBuffer{}
	rb.Init(2, 2, 1024)

	// write one
	err := rb.WriteByte('a')
	if err != nil {
		t.Fatalf("WriteByte failed: %v", err)
	}
	if rb.Length() != 1 {
		t.Fatalf("expect len 1 byte but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d byte but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if !bytes.Equal(rb.Bytes(), []byte{'a'}) {
		t.Fatalf("expect a but got %s. r.w=%d, r.r=%d", rb.Bytes(), rb.GetW(), rb.GetR())
	}
	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// write to, isFull
	err = rb.WriteByte('b')
	if err != nil {
		t.Fatalf("WriteByte failed: %v", err)
	}
	if rb.Length() != 2 {
		t.Fatalf("expect len 2 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d byte but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if !bytes.Equal(rb.Bytes(), []byte{'a', 'b'}) {
		t.Fatalf("expect a but got %s. r.w=%d, r.r=%d", rb.Bytes(), rb.GetW(), rb.GetR())
	}
	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if !rb.IsFull() {
		t.Fatalf("expect IsFull is true but got false")
	}

	// write
	_ = rb.WriteByte('c')
	if rb.Length() != 2 {
		t.Fatalf("expect len 2 bytes but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Capacity() != 2 {
		t.Fatalf("expect Capacity 2 bytes but got %d. r.w=%d, r.r=%d", rb.Capacity(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d byte but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if !bytes.Equal(rb.Bytes(), []byte{'b', 'c'}) {
		t.Fatalf("expect a but got %s. r.w=%d, r.r=%d", rb.Bytes(), rb.GetW(), rb.GetR())
	}
	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if !rb.IsFull() {
		t.Fatalf("expect IsFull is true but got false")
	}

	// read one
	b, err := rb.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte failed: %v", err)
	}
	if b != 'b' {
		t.Fatalf("expect a but got %c. r.w=%d, r.r=%d", b, rb.GetW(), rb.GetR())
	}
	if rb.Length() != 1 {
		t.Fatalf("expect len 2 byte but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d byte but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	if !bytes.Equal(rb.Bytes(), []byte{'c'}) {
		t.Fatalf("expect b but got %s. r.w=%d, r.r=%d", rb.Bytes(), rb.GetW(), rb.GetR())
	}
	// check empty or full
	if rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is false but got true")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// read two, empty
	b, err = rb.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte failed: %v", err)
	}
	if b != 'c' {
		t.Fatalf("expect b but got %c. r.w=%d, r.r=%d", b, rb.GetW(), rb.GetR())
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 byte but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d byte but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}

	// read three, error
	_, _ = rb.ReadByte()
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 byte but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d byte but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	// check empty or full
	if !rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is true but got false")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}

	// read four, error
	_, err = rb.ReadByte()
	if err == nil {
		t.Fatalf("expect ErrIsEmpty but got nil")
	}
	if rb.Length() != 0 {
		t.Fatalf("expect len 0 byte but got %d. r.w=%d, r.r=%d", rb.Length(), rb.GetW(), rb.GetR())
	}
	if rb.Free() != len(rb.GetBuf())-rb.Length() {
		t.Fatalf("expect Free %d byte but got %d. r.w=%d, r.r=%d", len(rb.GetBuf())-rb.Length(), rb.Free(), rb.GetW(), rb.GetR())
	}
	// check empty or full
	if !rb.IsEmpty() {
		t.Fatalf("expect IsEmpty is true but got false")
	}
	if rb.IsFull() {
		t.Fatalf("expect IsFull is false but got true")
	}
}

func TestRingBuffer_VirtualFunc(t *testing.T) {
	rb := &RingBuffer{}
	rb.Init(10, 10, 1024)

	_, err := rb.Write([]byte("abcd1234"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	buf := make([]byte, 4)
	_, err = rb.Read(buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !bytes.Equal(buf, []byte("abcd")) {
		t.Fatal()
	}

	buf = make([]byte, 2)
	_, err = rb.VirtualRead(buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !bytes.Equal(buf, []byte("12")) {
		t.Fatal()
	}
	if rb.Length() != 4 {
		t.Fatal()
	}
	if rb.VirtualLength() != 2 {
		t.Fatal()
	}
	rb.VirtualRefresh()
	if rb.Length() != 2 {
		t.Fatal()
	}
	if rb.VirtualLength() != 2 {
		t.Fatal()
	}

	_, err = rb.VirtualRead(buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if !bytes.Equal(buf, []byte("34")) {
		t.Fatal()
	}
	if rb.Length() != 2 {
		t.Fatal()
	}
	if rb.VirtualLength() != 0 {
		t.Fatal()
	}
	rb.VirtualRevert()
	if rb.Length() != 2 {
		t.Fatal()
	}
	if rb.VirtualLength() != 2 {
		t.Fatal()
	}
}

func TestRingBuffer_Bytes(t *testing.T) {
	var (
		size = 64
		rb   *RingBuffer
		n    int
		err  error
		data []byte
	)

	rb = &RingBuffer{}
	rb.Init(size, size, 1024)

	n, err = rb.Write([]byte(strings.Repeat("abcd", 2)))
	if n != 8 || err != nil {
		t.Fatal()
	}
	data = make([]byte, 4)
	n, err = rb.Read(data)
	if n != 4 || err != nil {
		t.Fatal()
	}
	n, err = rb.Write([]byte(strings.Repeat("efgh", 15)))
	if n != 60 || err != nil {
		t.Fatal()
	}
	if rb.GetR() != 4 || rb.GetW() != 4 || !rb.IsFull() {
		t.Fatal()
	}
	expect := strings.Repeat("abcd", 1) + strings.Repeat("efgh", 15)
	actual := string(rb.Bytes())
	if expect != actual {
		t.Fatalf("except %s, but got %s", expect, actual)
	}
}

func TestRingBuffer_WriteLog(t *testing.T) {
	var (
		size = 64
		rb   *RingBuffer
		n    int
		err  error
		logs []string
		data []byte
	)

	rb = &RingBuffer{}
	rb.Init(size, size, 1024)

	logs = []string{
		"[1]: log w/ resource, debug1 level",
		"[1]: log w/ resource, info1 level",
		"[1]: log w/ resource, trace1 level",
		"[1]: log w/ resource, error1 level",
		"[1]: log w/ resource, panic1 level",
	}

	for _, log := range logs {
		fmt.Println("-----------------------------")
		l := len(log)
		bn := make([]byte, 4)
		binary.LittleEndian.PutUint32(bn, uint32(l))
		n, err = rb.Write(bn)
		if n != 4 || err != nil {
			t.Fatal()
		}

		n, err = rb.Write([]byte(log))
		if n != l || err != nil {
			t.Fatal()
		}

		data = make([]byte, 4)
		n, err = rb.VirtualRead(data)
		if n != 4 || err != nil {
			t.Fatal()
		}

		l = int(binary.LittleEndian.Uint32(data))
		fmt.Println("prefix integer: ", l)

		data = make([]byte, l)
		n, err = rb.VirtualRead(data)
		if n != l || err != nil {
			t.Fatal()
		}
		fmt.Println("log read: ", bytes.NewBuffer(data).String())
		fmt.Println("cur capacity: ", rb.Capacity())
		fmt.Println("cur length: ", rb.Length())
	}
}

func TestRingBuffer_WriteLogOverflow(t *testing.T) {
	var (
		size = 64
		rb   *RingBuffer
		n    int
		err  error
		logs []string
		data []byte
	)

	rb = &RingBuffer{}
	rb.Init(size, size, 1024)

	logs = []string{
		"[1]: log w/ resource, debug1 level",
		"[1]: log w/ resource, info1 level",
		"[1]: log w/ resource, trace1 level",
		"[1]: log w/ resource, error1 level",
		"[1]: log w/ resource, panic1 level",
		"[2]: log w/ resource, debug2 level",
		"[2]: log w/ resource, info2 level",
		"[2]: log w/ resource, trace2 level",
	}

	for _, log := range logs {
		fmt.Println("-----------------------------")
		l := len(log)
		bn := make([]byte, 4)
		binary.LittleEndian.PutUint32(bn, uint32(l))
		n, err = rb.Write(bn)
		if n != 4 || err != nil {
			t.Fatal()
		}
		fmt.Println("cur W: ", rb.GetW())
		fmt.Println("cur R: ", rb.GetR())

		n, err = rb.Write([]byte(log))
		if n != l || err != nil {
			t.Fatal()
		}
		fmt.Println("cur W: ", rb.GetW())
		fmt.Println("cur R: ", rb.GetR())
		fmt.Println("cur capacity: ", rb.Capacity())
		fmt.Println("cur length: ", rb.Length())
	}

	fmt.Println("-----------------------------")
	data = make([]byte, 4)
	n, err = rb.VirtualRead(data)
	if n != 4 || err != nil {
		t.Fatal()
	}

	l := int(binary.LittleEndian.Uint32(data))
	fmt.Println("prefix integer: ", l)

	data = make([]byte, l)
	n, err = rb.VirtualRead(data)
	if n != l || err != nil {
		t.Fatal()
	}
	fmt.Println("log read: ", bytes.NewBuffer(data).String())
	fmt.Println("cur capacity: ", rb.Capacity())
	fmt.Println("cur length: ", rb.Length())
}
