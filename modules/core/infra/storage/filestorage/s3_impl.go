package filestorage

import (
	"context"
	stdErr "errors"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/constants"
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
	// An empty bucket otherwise surfaces as the storage's NoSuchBucket, which names no setting.
	if this.bucket == "" {
		return errors.Errorf("file storage has no bucket: set %s", constants.S3StorageBucket)
	}
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

		if len(opts.Tags) > 0 {
			in.Tagging = aws.String(encodeTags(opts.Tags))
		}
	}

	_, err := this.uploader.Upload(ctx, in)
	if err != nil {
		return errors.Wrap(err, "s3 uploader Upload error")
	}

	return nil
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
		if isApiErrorCode(err, "InvalidRange") {
			return nil, this.rangeNotSatisfiable(ctx, objectKey, err)
		}
		return nil, errors.Wrapf(err, "s3 client GetObject failed with input=%v", input)
	}

	return &StreamObjectResult{
		Body:          out.Body,
		ContentLength: out.ContentLength,
		ContentRange:  out.ContentRange,
	}, nil
}

// rangeNotSatisfiable reads the object's size, which the 416 answer must report.
func (this *s3Adapter) rangeNotSatisfiable(ctx context.Context, objectKey string, cause error) error {
	head, err := this.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(this.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return errors.Wrapf(cause, "s3 client GetObject: range not satisfiable, and HeadObject failed: %v", err)
	}
	return ErrRangeNotSatisfiable{Size: aws.ToInt64(head.ContentLength)}
}

func isApiErrorCode(err error, code string) bool {
	var apiErr smithy.APIError
	return stdErr.As(err, &apiErr) && apiErr.ErrorCode() == code
}

func (this *s3Adapter) SetTags(ctx context.Context, objectKey string, tags map[string]string) error {
	tagSet := make([]types.Tag, 0, len(tags))
	for _, key := range sortedKeys(tags) {
		tagSet = append(tagSet, types.Tag{Key: aws.String(key), Value: aws.String(tags[key])})
	}
	_, err := this.client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket:  aws.String(this.bucket),
		Key:     aws.String(objectKey),
		Tagging: &types.Tagging{TagSet: tagSet},
	})
	return errors.Wrapf(err, "s3 client PutObjectTagging failed for key=%s", objectKey)
}

// Copy runs server-side, so the bytes never pass through this process. Objects above 5 GiB
// would need a multipart copy, which uploads capped far below that do not.
func (this *s3Adapter) Copy(ctx context.Context, srcKey string, dstKey string, opts *CopyOptions) error {
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(this.bucket),
		Key:        aws.String(dstKey),
		CopySource: aws.String((&url.URL{Path: this.bucket + "/" + srcKey}).EscapedPath()),
		ACL:        types.ObjectCannedACLPrivate,
	}
	if opts != nil && len(opts.Tags) > 0 {
		input.TaggingDirective = types.TaggingDirectiveReplace
		input.Tagging = aws.String(encodeTags(opts.Tags))
	}
	_, err := this.client.CopyObject(ctx, input)
	return errors.Wrapf(err, "s3 client CopyObject failed from key=%s to key=%s", srcKey, dstKey)
}

// encodeTags renders tags as the URL query string the Tagging header expects.
func encodeTags(tags map[string]string) string {
	pairs := make([]string, 0, len(tags))
	for _, key := range sortedKeys(tags) {
		pairs = append(pairs, url.QueryEscape(key)+"="+url.QueryEscape(tags[key]))
	}
	return strings.Join(pairs, "&")
}

func sortedKeys(tags map[string]string) []string {
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (this *s3Adapter) Remove(ctx context.Context, objectKey string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(this.bucket),
		Key:    aws.String(objectKey),
	}
	_, err := this.client.DeleteObject(ctx, input)
	if err != nil {
		return errors.Wrapf(err, "s3 client DeleteObject failed with input=%v", input)
	}

	return err
}

func (this *s3Adapter) GeneratePresignedUrl(ctx context.Context, objectKey string, expr time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(this.bucket),
		Key:    aws.String(objectKey),
	}
	req, err := this.presign.PresignGetObject(
		ctx,
		input,
		s3.WithPresignExpires(expr),
	)
	if err != nil {
		return "", errors.Wrapf(err, "s3 presign client PresignGetObject failed with input=%v", input)
	}

	return req.URL, nil
}
