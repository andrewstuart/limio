package ratelimit

import (
	"bytes"
	"strings"
	"testing"
)

type testLimiter struct {
	c chan ByteCount
}

func (tl *testLimiter) GetLimit() <-chan ByteCount {
	return tl.c
}

func TestLimitedReader(t *testing.T) {
	r := strings.NewReader(testText)

	tl := &testLimiter{
		c: make(chan ByteCount),
	}

	lr := NewReader(r, tl)

	nBytes := 20
	go func() {
		tl.c <- ByteCount(nBytes)
	}()

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
}

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
