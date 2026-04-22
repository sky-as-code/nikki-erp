package constants

const (
	// General
	AppName      ConfigName = "GENERAL.APP_NAME"
	LogLevel     ConfigName = "GENERAL.LOG_LEVEL"
	MigrationDir ConfigName = "GENERAL.MIGRATION_DIR"

	// Database
	DbDialect             ConfigName = "CORE.DB.DIALECT"
	DbHostPort            ConfigName = "CORE.DB.HOST_PORT"
	DbName                ConfigName = "CORE.DB.DATABASE"
	DbPassword            ConfigName = "CORE.DB.PASSWORD"
	DbUser                ConfigName = "CORE.DB.USER"
	DbDebugEnabled        ConfigName = "CORE.DB.ORM_DEBUG_ENABLED"
	DbTlsEnabled          ConfigName = "CORE.DB.TLS_ENABLED"
	DbMaxIdleConns        ConfigName = "CORE.DB.MAX_IDLE_CONNS"
	DbMaxOpenConns        ConfigName = "CORE.DB.MAX_OPEN_CONNS"
	DbConnMaxLifetimeSecs ConfigName = "CORE.DB.CONN_MAX_LIFETIME_SECS"

	// Redis core
	RedisHost     ConfigName = "CORE.REDIS.HOST"
	RedisPort     ConfigName = "CORE.REDIS.PORT"
	RedisPassword ConfigName = "CORE.REDIS.PASSWORD"
	RedisDB       ConfigName = "CORE.REDIS.DB"

	// Database Postgres-specific
	// DbPgSslMode ConfigName = "DB_PG_SSL_MODE"

	// CQRS pubsub
	CqrsRequestTimeoutSecs ConfigName = "CORE.CQRS.REQUEST_TIMEOUT_SECS"

	// Event Bus
	EventRequestTimeoutSecs ConfigName = "CORE.EVENT.REQUEST_TIMEOUT_SECS"

	// Event Bus Redis
	EventBusRedisHost     ConfigName = "CORE.EVENT.REDIS_HOST"
	EventBusRedisPort     ConfigName = "CORE.EVENT.REDIS_PORT"
	EventBusRedisPassword ConfigName = "CORE.EVENT.REDIS_PASSWORD"
	EventBusRedisDB       ConfigName = "CORE.EVENT.REDIS_DB"

	// HTTP Server
	HttpBasePath       ConfigName = "CORE.HTTP.BASE_PATH"
	HttpHost           ConfigName = "CORE.HTTP.HOST"
	HttpPort           ConfigName = "CORE.HTTP.PORT"
	HttpsPort          ConfigName = "CORE.HTTP.HTTPS_PORT"
	HttpsEnabled       ConfigName = "CORE.HTTP.HTTPS_ENABLED"
	HttpsRsaPublicKey  ConfigName = "CORE.HTTP.HTTPS_RSA_PUBLIC_KEY"
	HttpsRsaPrivateKey ConfigName = "CORE.HTTP.HTTPS_RSA_PRIVATE_KEY"
	MtlsEnabled        ConfigName = "CORE.HTTP.MTLS_ENABLED"
	MtlsClientCaCert   ConfigName = "CORE.HTTP.MTLS_CLIENT_CA_CERT"
	HttpCorsEnabled    ConfigName = "CORE.HTTP.CORS_ENABLED"
	HttpCorsOrigins    ConfigName = "CORE.HTTP.CORS_ORIGINS"
	HttpCorsHeaders    ConfigName = "CORE.HTTP.CORS_HEADERS"
	HttpCorsMethods    ConfigName = "CORE.HTTP.CORS_METHODS"

	// Request Guard
	RequestGuardAccessTokenEnabled          ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_ENABLED"
	RequestGuardAccessTokenDpopEnabled      ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_DPOP_ENABLED"
	RequestGuardAccessTokenAlgorithm        ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_ALGORITHM"
	RequestGuardAccessTokenAudience         ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_AUDIENCE"
	RequestGuardAccessTokenExpiryMinutes    ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_EXPIRY_MINUTES"
	RequestGuardAccessTokenIssuer           ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_ISSUER"
	RequestGuardAccessTokenHttpHeaderName   ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_HTTP_HEADER_NAME"
	RequestGuardAccessTokenHttpHeaderPrefix ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_HTTP_HEADER_PREFIX"
	RequestGuardAccessTokenSecret           ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_SECRET"
	RequestGuardAccessTokenPublicKey        ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_PUBLIC_KEY"
	RequestGuardAccessTokenPrivateKey       ConfigName = "CORE.REQUEST_GUARD.ACCESS_TOKEN_PRIVATE_KEY"
	RequestGuardRefreshTokenExpiryMinutes   ConfigName = "CORE.REQUEST_GUARD.REFRESH_TOKEN_EXPIRY_MINUTES"

	RequestGuardSessionBlacklistEnabled ConfigName = "CORE.REQUEST_GUARD.SESSION_BLACKLIST_ENABLED"

	// Token/Authentication
	TokenSecretKey   ConfigName = "CORE.TOKEN.SECRET_KEY"
	TokenExpiryHours ConfigName = "CORE.TOKEN.EXPIRY_HOURS"
)
