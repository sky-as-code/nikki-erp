package services

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var pngHead = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

func TestSniffMime_KeepsEveryByte(t *testing.T) {
	content := append(append([]byte{}, pngHead...), bytes.Repeat([]byte{1}, 2000)...)

	detected, reader, err := sniffMime(bytes.NewReader(content))
	assert.NoError(t, err)
	assert.Equal(t, "image/png", detected)
	read, _ := io.ReadAll(reader)
	assert.Equal(t, content, read)
}

func TestExactSizeReader(t *testing.T) {
	exact := newExactSizeReader(bytes.NewReader([]byte("abcd")), 4)
	read, err := io.ReadAll(exact)
	assert.NoError(t, err)
	assert.Equal(t, "abcd", string(read))
	assert.True(t, exact.verifyDrained())

	longer := newExactSizeReader(bytes.NewReader([]byte("abcdef")), 4)
	_, err = io.ReadAll(longer)
	assert.ErrorIs(t, err, errSizeMismatch)
	assert.True(t, longer.mismatch)

	shorter := newExactSizeReader(bytes.NewReader([]byte("ab")), 4)
	_, err = io.ReadAll(shorter)
	assert.ErrorIs(t, err, errSizeMismatch)

	// A consumer that stops at the declared length must still see the overrun.
	stopped := newExactSizeReader(bytes.NewReader([]byte("abcdef")), 4)
	_, _ = io.ReadFull(stopped, make([]byte, 4))
	assert.False(t, stopped.verifyDrained())
}
