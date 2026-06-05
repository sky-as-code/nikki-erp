package filestorage

import (
	"context"
	"io"
	"time"
)

type FileStorageAdapter interface {
	Put(ctx context.Context, objectKey string, r io.Reader, opts *PutOptions) error
	Open(ctx context.Context, objectKey string, rangeHeader string) (*StreamObjectResult, error)
	Remove(ctx context.Context, objectKey string) error
	GeneratePresignedURL(ctx context.Context, objectKey string, expr time.Duration) (string, error)
}

type PutOptions struct {
	ContentLength      *int64
	ContentType        *string
	ContentDisposition *string
	CacheControl       *string
}

type StreamObjectResult struct {
	Body          io.ReadCloser
	ContentLength *int64
	ContentRange  *string
}
