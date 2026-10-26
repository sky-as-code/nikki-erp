package services

import (
	"bufio"
	"io"

	"github.com/gabriel-vasile/mimetype"
	"go.bryk.io/pkg/errors"
)

// sniffLength is how much of the content the signature check reads before the upload starts.
const sniffLength = 512

var errSizeMismatch = errors.New("the uploaded content does not match the declared size_bytes")

// sniffMime peeks at the head of body without consuming it and returns the detected media type
// together with a reader that still yields every byte. Nothing is seeked or buffered beyond the
// head, so the upload stays a stream.
func sniffMime(body io.Reader) (string, io.Reader, error) {
	buffered := bufio.NewReaderSize(body, sniffLength)
	head, err := buffered.Peek(sniffLength)
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		return "", nil, errors.Wrap(err, "sniff upload content")
	}
	return mimetype.Detect(head).String(), buffered, nil
}

// exactSizeReader fails the stream as soon as it yields more bytes than declared, or ends with
// fewer, so storage never keeps an object whose size the row would misreport.
type exactSizeReader struct {
	inner    io.Reader
	expected int64
	count    int64
	mismatch bool
}

func newExactSizeReader(body io.Reader, expected int64) *exactSizeReader {
	// One byte past the declared size is enough to prove an overrun without draining the rest.
	return &exactSizeReader{inner: io.LimitReader(body, expected+1), expected: expected}
}

func (this *exactSizeReader) Read(buffer []byte) (int, error) {
	read, err := this.inner.Read(buffer)
	this.count += int64(read)
	if this.count > this.expected {
		this.mismatch = true
		return read, errSizeMismatch
	}
	if err == io.EOF && this.count != this.expected {
		this.mismatch = true
		return read, errSizeMismatch
	}
	return read, err
}

// verifyDrained catches a consumer that stopped at the declared length: storage may read exactly
// ContentLength bytes, which would otherwise hide content beyond it.
func (this *exactSizeReader) verifyDrained() bool {
	if this.mismatch || this.count != this.expected {
		return false
	}
	probe := make([]byte, 1)
	read, _ := this.inner.Read(probe)
	return read == 0
}
