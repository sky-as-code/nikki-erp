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
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/file_storage/constants"
)

var configKeyForBucket = map[BucketKey]c.ConfigName{
	BucketDefault: constants.S3StorageBucket,
	BucketDrive:   constants.S3StorageBucketDrive,
}

type S3StorageServiceImpl struct {
	uploader   *manager.Uploader
	downloader *manager.Downloader
	client     *s3.Client
	bucket     string
	cfg        config.ConfigService
}

func NewS3StorageService(cfg config.ConfigService) (StorageService, error) {
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

	return &S3StorageServiceImpl{
		uploader:   uploader,
		downloader: downloader,
		client:     client,
		bucket:     cfg.GetStr(constants.S3StorageBucket),
		cfg:        cfg,
	}, nil
}

func (s *S3StorageServiceImpl) resolveBucket(key BucketKey) string {
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

func (s *S3StorageServiceImpl) Upload(ctx context.Context, objectKey string, file multipart.File) error {
	return s.UploadBucket(ctx, BucketDefault, objectKey, file)
}

func (s *S3StorageServiceImpl) UploadBucket(ctx context.Context, bucket BucketKey, objectKey string, file multipart.File) error {
	b := s.resolveBucket(bucket)
	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	return err
}

func (s *S3StorageServiceImpl) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	return s.DownloadBucket(ctx, BucketDefault, objectKey)
}

func (s *S3StorageServiceImpl) DownloadBucket(ctx context.Context, bucket BucketKey, objectKey string) (io.ReadCloser, error) {
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
