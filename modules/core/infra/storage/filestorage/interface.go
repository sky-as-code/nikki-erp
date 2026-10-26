package filestorage

import (
	"context"
	"fmt"
	"io"
	"time"
)

type FileStorageAdapter interface {
	Put(ctx context.Context, objectKey string, r io.Reader, opts *PutOptions) error
	Open(ctx context.Context, objectKey string, rangeHeader string) (*StreamObjectResult, error)
	Remove(ctx context.Context, objectKey string) error
	// SetTags replaces every tag of the object with tags.
	SetTags(ctx context.Context, objectKey string, tags map[string]string) error
	// Copy duplicates srcKey into dstKey server-side, replacing the tags with opts.Tags.
	Copy(ctx context.Context, srcKey string, dstKey string, opts *CopyOptions) error
	GeneratePresignedUrl(ctx context.Context, objectKey string, expr time.Duration) (string, error)
}

type PutOptions struct {
	ContentLength      *int64
	ContentType        *string
	ContentDisposition *string
	CacheControl       *string
	// Tags are written with the object, so a lifecycle rule matching them applies from the start.
	Tags map[string]string
}

type CopyOptions struct {
	Tags map[string]string
}

type StreamObjectResult struct {
	Body          io.ReadCloser
	ContentLength *int64
	ContentRange  *string
}

// ErrRangeNotSatisfiable is returned by Open when the Range header lies wholly past the end of
// the object. Size is the object's length, which a 416 answer reports as Content-Range: bytes */Size.
type ErrRangeNotSatisfiable struct {
	Size int64
}

func (this ErrRangeNotSatisfiable) Error() string {
	return fmt.Sprintf("the requested range is not satisfiable for an object of %d bytes", this.Size)
}
