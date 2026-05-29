package filestorage

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/corevision-technology/coremart/modules/vending_machine/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type s3Adapter struct {
	uploader *manager.Uploader
	client   *s3.Client
	bucket   string
	logger   logging.LoggerService
	presign  *s3.PresignClient
}

func NewS3Adapter(cfg config.ConfigService, logger logging.LoggerService) (FileStorageAdapter, error) {
	s3Config, err := aws_config.LoadDefaultConfig(context.Background(),
		aws_config.WithRegion(cfg.GetStr(constants.S3StorageRegionName)),
		aws_config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.GetStr(constants.S3StorageAccessToken),
				cfg.GetStr(constants.S3StorageSecretKey), ""),
		),
		aws_config.WithResponseChecksumValidation(aws.ResponseChecksumValidationWhenSupported),
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

	presign := s3.NewPresignClient(client)

	return &s3Adapter{
		uploader: uploader,
		client:   client,
		bucket:   cfg.GetStr(constants.S3StorageBucket),
		logger:   logger,
		presign:  presign,
	}, nil
}

func (this *s3Adapter) Put(ctx context.Context, objectKey string, r io.Reader, opts *PutOptions) error {
	in := &s3.PutObjectInput{
		Bucket:             aws.String(this.bucket),
		Key:                aws.String(objectKey),
		Body:               r,
		ACL:                types.ObjectCannedACLPrivate,
		ContentDisposition: aws.String("inline"),
	}

	if opts != nil {
		if opts.ContentLength != nil {
			in.ContentLength = aws.Int64(*opts.ContentLength)
		}

		if opts.ContentType != nil {
			in.ContentType = aws.String(*opts.ContentType)
		}

		if opts.CacheControl != nil {
			in.CacheControl = aws.String(*opts.CacheControl)
		}

		if opts.ContentDisposition != nil {
			in.ContentDisposition = aws.String(*opts.ContentDisposition)
		}
	}

	_, err := this.uploader.Upload(ctx, in)
	if err != nil {
		this.logger.Errorf("S3 upload failed key=%s: %v", objectKey, err)
	}
	return err
}

func (this *s3Adapter) Open(ctx context.Context, objectKey string, rangeHeader string) (*StreamObjectResult, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(this.bucket),
		Key:    aws.String(objectKey),
	}

	if rangeHeader != "" {
		input.Range = aws.String(rangeHeader)
	}

	out, err := this.client.GetObject(ctx, input)
	if err != nil {
		this.logger.Errorf("S3 get object failed key=%s: %v", objectKey, err)
		return nil, err
	}
	return &StreamObjectResult{
		Body:          out.Body,
		ContentLength: out.ContentLength,
		ContentRange:  out.ContentRange,
	}, nil
}

func (this *s3Adapter) Remove(ctx context.Context, objectKey string) error {
	_, err := this.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(this.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		this.logger.Errorf("S3 delete failed key=%s: %v", objectKey, err)
	}
	return err
}

func (this *s3Adapter) GeneratePresignedURL(ctx context.Context, objectKey string, expr time.Duration) (string, error) {
	req, err := this.presign.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(this.bucket),
			Key:    aws.String(objectKey),
		},
		s3.WithPresignExpires(expr),
	)
	if err != nil {
		return "", err
	}

	return req.URL, nil
}
