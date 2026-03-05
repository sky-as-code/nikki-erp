package file_storage

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/samber/lo"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive/constants"
)

var configKeyForBucket = map[BucketKey]c.ConfigName{
	BucketDefault: constants.S3StorageBucket,
	BucketDrive:   constants.S3StorageBucketDrive,
}

type S3StorageAdapterImpl struct {
	uploader   *manager.Uploader
	downloader *manager.Downloader
	client     *s3.Client
	bucket     string
	cfg        config.ConfigService
	logger     logging.LoggerService
}

func NewS3StorageService(cfg config.ConfigService, logger logging.LoggerService) (FileStorageAdapter, error) {
	s3Config, err := aws_config.LoadDefaultConfig(context.Background(),
		aws_config.WithRegion(cfg.GetStr(constants.S3StorageRegionName)),
		aws_config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.GetStr(constants.S3StorageAccessToken),
				cfg.GetStr(constants.S3StorageSecretKey), ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(s3Config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.GetStr(constants.S3StorageEndpoint))
		o.UsePathStyle = true
	})

	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024
		u.Concurrency = 3
	})

	downloader := manager.NewDownloader(client, func(d *manager.Downloader) {
		d.PartSize = 10 * 1024 * 1024
		d.Concurrency = 3
	})

	return &S3StorageAdapterImpl{
		uploader:   uploader,
		downloader: downloader,
		client:     client,
		bucket:     cfg.GetStr(constants.S3StorageBucket),
		cfg:        cfg,
		logger:     logger,
	}, nil
}

func (s *S3StorageAdapterImpl) resolveBucket(key BucketKey) string {
	if key == BucketDefault || key == "" {
		return s.bucket
	}
	configKey, ok := configKeyForBucket[key]
	if !ok {
		return s.bucket
	}
	name := s.cfg.GetStr(configKey)
	if name != "" {
		return name
	}
	return s.bucket
}

func (s *S3StorageAdapterImpl) Upload(ctx context.Context, objectKey string, file multipart.File) error {
	return s.UploadBucket(ctx, BucketDefault, objectKey, file)
}

func (s *S3StorageAdapterImpl) UploadBucket(ctx context.Context, bucket BucketKey, objectKey string, file multipart.File) error {
	b := s.resolveBucket(bucket)
	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	return err
}

func (s *S3StorageAdapterImpl) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	return s.DownloadBucket(ctx, BucketDefault, objectKey)
}

func (s *S3StorageAdapterImpl) DownloadBucket(ctx context.Context, bucket BucketKey, objectKey string) (io.ReadCloser, error) {
	b := s.resolveBucket(bucket)
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *S3StorageAdapterImpl) Delete(ctx context.Context, objectKey string) error {
	return s.DeleteBucket(ctx, BucketDefault, objectKey)
}

func (s *S3StorageAdapterImpl) DeleteBucket(ctx context.Context, bucket BucketKey, objectKey string) error {
	b := s.resolveBucket(bucket)
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(objectKey),
	})
	return err
}

func (r *S3StorageAdapterImpl) DeleteBulk(ctx context.Context, keys []string) (
	deletedKeys []string, failedKeys []string, err error) {
	objects := lo.Map(keys, func(key string, _ int) types.ObjectIdentifier {
		return types.ObjectIdentifier{Key: aws.String(key)}
	})

	output, err := r.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(r.bucket),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false),
		},
	})
	if err != nil {
		return nil, nil, err
	}

	deletedKeys = lo.Map(output.Deleted, func(d types.DeletedObject, _ int) string {
		return *d.Key
	})

	failedKeys = lo.Map(output.Errors, func(e types.Error, _ int) string {
		r.logger.Errorf("S3 delete failed key=%s reason=%s", *e.Key, *e.Message)
		return *e.Key
	})

	return deletedKeys, failedKeys, nil
}
