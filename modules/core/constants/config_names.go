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
	HttpBasePath    ConfigName = "CORE.HTTP.BASE_PATH"
	HttpHost        ConfigName = "CORE.HTTP.HOST"
	HttpPort        ConfigName = "CORE.HTTP.PORT"
	HttpCorsOrigins ConfigName = "CORE.HTTP.CORS_ORIGINS"
	HttpCorsHeaders ConfigName = "CORE.HTTP.CORS_HEADERS"
	HttpCorsMethods ConfigName = "CORE.HTTP.CORS_METHODS"
)
