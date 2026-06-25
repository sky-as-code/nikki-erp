package constants

import (
	core_constants "github.com/sky-as-code/nikki-erp/modules/core/constants"
)

const (
	S3StorageRegionName  core_constants.ConfigName = "CORE.S3_STORAGE.REGION_NAME"
	S3StorageAccessToken core_constants.ConfigName = "CORE.S3_STORAGE.ACCESS_TOKEN"
	S3StorageSecretKey   core_constants.ConfigName = "CORE.S3_STORAGE.SECRET_KEY"
	S3StorageEndpoint    core_constants.ConfigName = "CORE.S3_STORAGE.ENDPOINT"
	S3StorageBucket      core_constants.ConfigName = "CORE.S3_STORAGE.BUCKET"
	S3StorageBucketDrive core_constants.ConfigName = "CORE.S3_STORAGE.BUCKET_DRIVE"
)

const (
	RedisHost     core_constants.ConfigName = "DRIVE.REDIS.HOST"
	RedisPost     core_constants.ConfigName = "DRIVE.REDIS.PORT"
	RedisPassword core_constants.ConfigName = "DRIVE.REDIS.PASSWORD"
	RedisDB       core_constants.ConfigName = "DRIVE.REDIS.DB"
)
