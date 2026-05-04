package filestorage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/samber/lo"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// _S3Adapter lưu object trên S3-compatible storage (VENDING_MACHINE_S3_STORAGE_*).
type _S3Adapter struct {
	uploader *manager.Uploader
	client   *s3.Client
	bucket   string
	logger   logging.LoggerService
}

func NewS3Adapter(cfg config.ConfigService, logger logging.LoggerService) (FileStorage, error) {
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

	return &_S3Adapter{
		uploader: uploader,
		client:   client,
		bucket:   cfg.GetStr(constants.S3StorageBucket),
		logger:   logger,
	}, nil
}

func (s *_S3Adapter) Type() string {
	return "S3"
}

func (s *_S3Adapter) Put(ctx context.Context, objectKey string, r io.Reader, size int64, opt *PutOption) error {
	in := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
		Body:   r,
	}
	if size > 0 {
		in.ContentLength = aws.Int64(size)
	}

	if opt != nil {
		in.ContentType = aws.String(opt.ContentType)
		in.Metadata = opt.Metadata
	}

	_, err := s.uploader.Upload(ctx, in)
	if err != nil {
		s.logger.Errorf("S3 kiosk-media upload failed key=%s: %v", objectKey, err)
	}
	return err
}

func (s *_S3Adapter) Open(ctx context.Context, objectKey string, r *Range) (*ObjectStream, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	}

	if r != nil {
		input.Range = aws.String(r.RangeHeader())
	}

	out, err := s.client.GetObject(ctx, input)

	if err != nil {
		s.logger.Errorf("S3 kiosk-media get object failed key=%s: %v", objectKey, err)
		return nil, err
	}
	return &ObjectStream{
		Body:          out.Body,
		ContentLength: out.ContentLength,
		ContentRange:  out.ContentRange,
	}, nil
}

func (s *_S3Adapter) Remove(ctx context.Context, objectKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		s.logger.Errorf("S3 kiosk-media delete failed key=%s: %v", objectKey, err)
	}
	return err
}

func (s *_S3Adapter) RemoveBulk(ctx context.Context, keys []string) (deletedKeys []string, failedKeys []string, err error) {
	if len(keys) == 0 {
		return []string{}, []string{}, nil
	}

	objects := lo.Map(keys, func(key string, _ int) types.ObjectIdentifier {
		return types.ObjectIdentifier{Key: aws.String(key)}
	})

	output, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
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
		s.logger.Errorf("S3 delete failed key=%s reason=%s", *e.Key, *e.Message)
		return *e.Key
	})

	return deletedKeys, failedKeys, nil
}
