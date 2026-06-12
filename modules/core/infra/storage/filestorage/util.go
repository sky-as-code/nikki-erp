package filestorage

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/util"
)

func NewPutOptions(contentType string, size int64) *PutOptions {
	opts := &PutOptions{
		ContentType: util.ToPtr(contentType),
	}
	if size > 0 {
		opts.ContentLength = util.ToPtr(size)
	}
	return opts
}

func GeneratePresignedBulk(ctx context.Context, a FileStorageAdapter, keys []string, expr time.Duration) (map[string]string, error) {
	res := make(map[string]string, len(keys))
	var err error

	for _, key := range keys {
		res[key], err = a.GeneratePresignedUrl(ctx, key, expr)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
