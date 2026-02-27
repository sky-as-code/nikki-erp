package file_storage

import (
	"context"
	"io"
	"mime/multipart"
)

type BucketKey string

const (
	BucketDefault BucketKey = ""      // default bucket from config
	BucketDrive   BucketKey = "drive" // bucket for drive module
)

type StorageService interface {
	Upload(ctx context.Context, objectKey string, file multipart.File) error
	UploadBucket(ctx context.Context, bucketKey BucketKey, objectKey string, file multipart.File) error
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DownloadBucket(ctx context.Context, bucketKey BucketKey, objectKey string) (io.ReadCloser, error)
}
