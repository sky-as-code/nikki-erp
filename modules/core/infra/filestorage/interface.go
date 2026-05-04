package filestorage

import (
	"context"
	"fmt"
	"io"
)

type FileStorage interface {
	Put(ctx context.Context, objectKey string, r io.Reader, size int64, opt *PutOption) error
	Open(ctx context.Context, objectKey string, r *Range) (*ObjectStream, error)
	Remove(ctx context.Context, objectKey string) error
	RemoveBulk(ctx context.Context, objectKey []string) (deletedKeys []string, failedKeys []string, err error)
	Type() string
}

type PutOption struct {
	ContentType string
	Metadata    map[string]string
}

type Range struct {
	From int64
	End  *int64
}

func (r *Range) RangeHeader() string {
	if r != nil {
		if r.End != nil {
			return fmt.Sprintf("bytes=%d-%d", r.From, *r.End)
		}

		return fmt.Sprintf("bytes=%d-", r.From)
	}

	return ""
}

type ObjectStream struct {
	Body          io.ReadCloser
	ContentLength *int64
	ContentRange  *string
}
