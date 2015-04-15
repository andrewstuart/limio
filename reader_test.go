package limio

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

type tStr struct {
	s   string
	eof bool
}

func (t *tStr) Read(p []byte) (written int, err error) {
	if t.eof {
		err = io.EOF
		return
	}

	l := len(p)

	if l > len(t.s) {
		l = len(t.s)
	}

	written = copy(p, []byte(t.s)[:l])

	if written == len(t.s) {
		err = io.EOF
		t.eof = true
	}

	t.s = t.s[l:]

	return
}

func TestLimitedReader(t *testing.T) {
	r := strings.NewReader(testText)

	c := make(chan int)

	lr := NewReader(r)
	lr.LimitChan(c)

	nBytes := 20
	c <- nBytes

	p := make([]byte, 512)

	n, err := lr.Read(p)

	if err != nil {
		t.Fatalf("message")
	}

	if n != nBytes {
		t.Errorf("Wrong number of bytes read: %d, should be %d", n, nBytes)
	}

	if !bytes.Equal(p[:nBytes], []byte(testText)[:nBytes]) {
		t.Errorf("Bytes not properly read: %s", p[:nBytes])
	}

	go func() {
		c <- nBytes / 2
		c <- nBytes / 2
	}()

	n2, err2 := lr.Read(p)

	if err2 != nil {
		t.Errorf("Error returned during second read attempt")
	}

	if n2 != nBytes {
		t.Errorf("Wrong number of bytes read second time: %d should be %d", n2, nBytes)
	}

	if !bytes.Equal(p[:nBytes], []byte(testText)[nBytes:nBytes*2]) {
		t.Errorf("Wrong bytes returned: %s", string(p[:nBytes]))
	}

	p2 := make([]byte, nBytes)

	go func() {
		c <- MB
	}()

	n3, err3 := lr.Read(p2)

	if err3 != nil {
		t.Errorf("Error during short read")
	}

	if n3 != nBytes {
		t.Errorf("Reader reported more bytes written than length of byte slice: %d, should be %d", n3, len(p2))
	}
}

func TestEOF(t *testing.T) {
	r := &tStr{s: testEOFText}

	c := make(chan int)

	lr := NewReader(r)
	lr.LimitChan(c)

	go func() {
		c <- KB
	}()

	p := make([]byte, 200)
	n, err := lr.Read(p)

	if n != len(testEOFText) {
		t.Errorf("Wrong number of bytes read: %d, should be %d", n, len(testEOFText))
	}

	if err != io.EOF {
		t.Fatalf("Read did not properly return EOF")
	}

	n2, err2 := lr.Read(p)

	if err2 != io.EOF {
		t.Error("Second read did not return EOF")
	}

	if n2 > 0 {
		t.Errorf("Bytes were READ after EOF: %d", n2)
	}
}

func TestNoLimit(t *testing.T) {
	r := NewReader(strings.NewReader(testText))
	p := make([]byte, len(testText))

	n, err := r.Read(p)

	if err != nil {
		t.Errorf("error reading: %v", err)
	}

	if n != len(testText) {
		t.Errorf("Wrong number of bytes read: %d, should be %d", n, len(testText))
	}

	if !bytes.Equal([]byte(testText), p) {
		t.Errorf("Bytes did not equal test read")
	}

	n2, err2 := r.Read(p)

	if err2 != io.EOF {
		t.Errorf("error reading EOF: %v", err2)
	}

	if n2 != 0 {
		t.Errorf("Read > 0 (%d) when should have only gotten EOF", n2)
	}

	n3, err3 := r.Read(p)

	if err3 != io.EOF {
		t.Errorf("error reading EOF: %v", err3)
	}

	if n3 != 0 {
		t.Errorf("Read > 0 (%d) when should have only gotten EOF", n3)
	}
}

func TestBasicLimit(t *testing.T) {
	r := NewReader(strings.NewReader(testText))
	r.Limit(80, 100*time.Millisecond)

	p := make([]byte, len(testText))
	n, err := r.Read(p)

	if err != nil {
		t.Errorf("error reading: %v", err)
	}

	expected := 8

	if n != expected {
		t.Errorf("Wrong number of bytes written in first window: %d, should be %d", n, expected)
	}

	n2, err2 := r.Read(p)

	if err2 != nil {
		t.Errorf("message")
	}

	if n2 != expected {
		t.Errorf("wrong number: %d", n2)
	}
}

func TestUnlimit(t *testing.T) {
	r := NewReader(strings.NewReader(testText))

	ch := make(chan int, 1)
	ch <- 20
	r.LimitChan(ch)

	p := make([]byte, len(testText))

	n, err := r.Read(p)

	if err != nil {
		t.Errorf("Got an error reading: %v", err)
	}

	if n != 20 {
		t.Errorf("Read wrong number of bytes: %d", n)
	}

	r.Unlimit()

	n, err = r.Read(p)

	if err != io.EOF {
		t.Errorf("Error reading: %v", err)
	}

	if n != len(testText)-20 {
		t.Errorf("Wrong number of bytes read: %d, should be %d", n, len(testText)-20)
	}

}

func TestClose(t *testing.T) {
	r := NewReader(strings.NewReader(testText))

	ch := make(chan int, 1)
	r.LimitChan(ch)
	err := r.Close()

	if err != nil {
		t.Fatalf("Close did not work: %v", err)
	}

	p := make([]byte, len(testText))

	n, err := r.Read(p)

	if n != len(p) {
		t.Errorf("Wrong number of bytes reported read: %d", n)
	}

	if err != nil && err != io.EOF {
		t.Errorf("Non-EOF error reported for closed limiter: %v", err)
	}
}

func TestDualLimit(t *testing.T) {
	r := NewReader(strings.NewReader(testText))

	ch := make(chan int, 1)
	ch <- 20
	done := r.LimitChan(ch)

	p := make([]byte, len(testText))

	n, err := r.Read(p)

	if err != nil {
		t.Errorf("Got an error reading: %v", err)
	}

	if n != 20 {
		t.Errorf("Read wrong number of bytes: %d", n)
	}

	ch = make(chan int, 1)
	ch <- 30
	r.LimitChan(ch)

	if _, cls := <-done; !cls {
		t.Errorf("did not close done")
	}

	n, err = r.Read(p)

	if n != 30 {
		t.Errorf("Wrong number of bytes read")
	}

	if err != nil {
		t.Errorf("Error reading bytes")
	}

}

const testEOFText = "foobarbaz"

const testText = `Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Etiam eget aliquet ipsum, vitae sodales arcu. Vivamus congue id metus eu
scelerisque. Duis fringilla felis at nunc consectetur hendrerit. Donec nec
metus nec sapien posuere euismod at rutrum felis. Vivamus aliquet nibh
sollicitudin sollicitudin vestibulum. Morbi nec felis quis nisl iaculis
fringilla sed eget eros. Ut lobortis id nulla in ultricies. Quisque nisi ipsum,
ullamcorper in metus ultrices, pretium rutrum sapien. Vivamus vestibulum, ipsum
et ultrices ultrices, massa erat maximus augue, at elementum erat leo in augue.
Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac
turpis egestas. Vestibulum ante ipsum primis in faucibus orci luctus et
ultrices posuere cubilia Curae; In dapibus, turpis id porttitor pretium, libero
elit efficitur dui, quis posuere magna quam in orci. Donec purus arcu, auctor
malesuada faucibus vel, ultrices at lacus. Aenean at ipsum in purus gravida
ultrices vitae nec ante. Nullam ut massa semper, consectetur ipsum et, pharetra
orci.

Cras ac volutpat turpis, eget pellentesque nibh. Praesent luctus tincidunt
felis, sed commodo lorem porta ac. Mauris id rhoncus massa, sed gravida dolor.
Cras tristique volutpat nibh, sed finibus justo dictum in. Praesent sodales
sapien eget augue rhoncus, sit amet rutrum quam convallis. Sed ante neque,
aliquet ut nisi nec, pulvinar dignissim mi. Morbi sit amet augue leo. Nulla
tincidunt molestie velit lobortis tristique. Aenean luctus neque nec felis
tincidunt, facilisis iaculis felis blandit. Morbi sit amet magna tortor.
Maecenas pulvinar finibus justo, eget vulputate lorem consequat malesuada. Cras
vel sem nulla.

Vestibulum id egestas sem. Integer molestie feugiat sapien at mollis. Nam sit
amet porta tortor. Praesent non dignissim tellus. Mauris congue metus leo, sed
interdum lorem rutrum in. Sed vulputate ipsum eu risus scelerisque facilisis.
Nunc at varius arcu, quis aliquam ex. Duis rhoncus nunc diam, vel mollis risus
sagittis quis. In dapibus libero in erat porta, vel porttitor augue dignissim.
Nullam libero nunc, ornare egestas bibendum et, eleifend et dui. Cras venenatis
sit amet odio vel varius. Class aptent taciti sociosqu ad litora torquent per
conubia nostra, per inceptos himenaeos. Pellentesque eget condimentum nisi.

Cras faucibus bibendum tortor ut faucibus. Cras posuere dolor urna, sit amet
aliquam ligula fringilla cursus. Praesent sodales congue augue vitae cursus.
Curabitur aliquet justo quis turpis cursus congue. Aliquam maximus tempor dui
eget suscipit. Donec facilisis dignissim augue, sit amet tempus nunc
ullamcorper in. Morbi varius nunc dapibus egestas tristique. Maecenas consequat
dui eget velit placerat elementum. Curabitur pharetra enim et eleifend viverra.
Mauris pellentesque urna non enim laoreet, at consectetur metus tincidunt.
Aliquam quis ligula congue, mattis felis vitae, mattis erat. Suspendisse nec
elementum turpis. Donec volutpat vitae libero vel placerat. Duis euismod, leo
in suscipit facilisis, ipsum risus aliquet libero, quis aliquet nulla turpis
eget leo.

Nullam bibendum ultricies ante, sit amet lobortis magna viverra id. Praesent
justo nulla, accumsan et lacinia at, bibendum id odio. Vivamus eu mi bibendum,
aliquam justo id, pellentesque urna. Duis scelerisque suscipit arcu, quis
laoreet arcu dignissim quis. Donec aliquet porta ligula et finibus. Ut
tincidunt facilisis blandit. Sed ultricies ipsum orci.`
