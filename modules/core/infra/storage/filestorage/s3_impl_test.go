package filestorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeTags_IsASortedQueryString(t *testing.T) {
	assert.Equal(t, "a=1&itstorage-state=pending", encodeTags(map[string]string{"itstorage-state": "pending", "a": "1"}))
	assert.Equal(t, "k=a+b%26c", encodeTags(map[string]string{"k": "a b&c"}))
}

func TestErrRangeNotSatisfiable_Message(t *testing.T) {
	assert.Contains(t, ErrRangeNotSatisfiable{Size: 10}.Error(), "10 bytes")
}
