package file_storage

import (
	"context"
	"io"
	"mime/multipart"
)

type BucketKey string

const (
	BucketDefault BucketKey = ""
	BucketDrive   BucketKey = "drive"
)

type FileStorageAdapter interface {
	Upload(ctx context.Context, objectKey string, file multipart.File) error
	UploadBucket(ctx context.Context, bucketKey BucketKey, objectKey string, file multipart.File) error
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DownloadBucket(ctx context.Context, bucketKey BucketKey, objectKey string) (io.ReadCloser, error)
	Delete(ctx context.Context, objectKey string) error
	DeleteBucket(ctx context.Context, bucketKey BucketKey, objectKey string) error
	DeleteBulk(ctx context.Context, keys []string) (deletedKeys []string, failedKeys []string, err error)
}
