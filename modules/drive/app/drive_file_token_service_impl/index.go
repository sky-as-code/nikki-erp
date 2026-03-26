package drive_file_signed_url_service_impl

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/drive/constants"
	driveFileTokenIt "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_token"
)

type driveFileSignedUrlService struct {
	logger      logging.LoggerService
	redisClient *redis.Client
}

func NewDriveFileSignedUrlService(
	logger logging.LoggerService,
	config config.ConfigService,
) driveFileTokenIt.DriveFileSignedUrlService {
	host := config.GetStr(constants.RedisHost)
	port := config.GetStr(constants.RedisPost)
	password := config.GetStr(constants.RedisPassword)
	db := config.GetInt(constants.RedisDB)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	return &driveFileSignedUrlService{
		redisClient: redisClient,
		logger:      logger,
	}
}

func (d *driveFileSignedUrlService) Create(ctx context.Context, fileId model.Id) (string, error) {
	panic("unimplemented")
}

func (d *driveFileSignedUrlService) Get(ctx context.Context, fileId model.Id) (string, error) {
	panic("unimplemented")
}

func (d *driveFileSignedUrlService) GetAndDelete(ctx context.Context, fileId model.Id) (string, error) {
	panic("unimplemented")
}

func (d *driveFileSignedUrlService) GetOrCreate(ctx context.Context, fileId model.Id) (string, error) {
	panic("unimplemented")
}

func (this *driveFileSignedUrlService) Verify(ctx context.Context, fileId model.Id, token string) (bool, error) {
	panic("unimplemented")
}
