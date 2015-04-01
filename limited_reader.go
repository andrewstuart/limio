package ratelimit

import (
	"bufio"
	"io"
)

type limitedReader struct {
	br  *bufio.Reader
	c   <-chan ByteCount
	eof bool
}

func (lr *limitedReader) Read(p []byte) (written int, err error) {
	if lr.eof {
		err = io.EOF
		return
	}

	for written <= len(p) {
		var lim ByteCount
		select {
		case lim = <-lr.c:
		default:
			if written == 0 {
				lim = <-lr.c
			} else {
				return
			}
		}

		var b byte
		for i := 0; i < int(lim); i++ {
			b, err = lr.br.ReadByte()

			if err == io.EOF {
				lr.eof = true
				return
			}

			p[written] = b
			written++
		}
	}

	return
}
