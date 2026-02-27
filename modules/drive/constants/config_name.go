package constants

import (
	core_constants "github.com/sky-as-code/nikki-erp/modules/core/constants"
)

const (
	S3StorageRegionName  core_constants.ConfigName = "DRIVE_S3_STORAGE_REGION_NAME"
	S3StorageAccessToken core_constants.ConfigName = "DRIVE_S3_STORAGE_ACCESS_TOKEN"
	S3StorageSecretKey   core_constants.ConfigName = "DRIVE_S3_STORAGE_SECRET_KEY"
	S3StorageEndpoint    core_constants.ConfigName = "DRIVE_S3_STORAGE_ENDPOINT"
	S3StorageBucket      core_constants.ConfigName = "DRIVE_S3_STORAGE_BUCKET"
	S3StorageBucketDrive core_constants.ConfigName = "DRIVE_S3_STORAGE_BUCKET_DRIVE"

	CrontabDeleteTrashedFile core_constants.ConfigName = "DRIVE_CRONTAB_DELETE_TRASHED_FILE"
)

const (
	RedisHost     core_constants.ConfigName = "DRIVE_REDIS_HOST"
	RedisPost     core_constants.ConfigName = "DRIVE_REDIS_PORT"
	RedisPassword core_constants.ConfigName = "DRIVE_REDIS_PASSWORD"
	RedisDB       core_constants.ConfigName = "DRIVE_REDIS_DB"
)
