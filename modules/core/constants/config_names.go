package constants

const (
	// General
	AppName      ConfigName = "APP_NAME"
	LogLevel     ConfigName = "LOG_LEVEL"
	MigrationDir ConfigName = "MIGRATION_DIR"

	// Database
	DbDialect             ConfigName = "DB_DIALECT"
	DbHostPort            ConfigName = "DB_HOST_PORT"
	DbName                ConfigName = "DB_NAME"
	DbPassword            ConfigName = "DB_PASSWORD"
	DbUser                ConfigName = "DB_USER"
	DbDebugEnabled        ConfigName = "DB_DEBUG_ENABLED"
	DbTlsEnabled          ConfigName = "DB_TLS_ENABLED"
	DbMaxIdleConns        ConfigName = "DB_MAX_IDLE_CONNS"
	DbMaxOpenConns        ConfigName = "DB_MAX_OPEN_CONNS"
	DbConnMaxLifetimeSecs ConfigName = "DB_CONN_MAX_LIFETIME_SECS"

	// Database Postgres-specific
	// DbPgSslMode ConfigName = "DB_PG_SSL_MODE"

	// CQRS pubsub
	CqrsRequestTimeoutSecs ConfigName = "CQRS_REQUEST_TIMEOUT_SECS"

	// Event Bus
	EventRequestTimeoutSecs ConfigName = "EVENT_REQUEST_TIMEOUT_SECS"

	// Event Bus Redis
	EventBusRedisHost     ConfigName = "EVENT_BUS_REDIS_HOST"
	EventBusRedisPort     ConfigName = "EVENT_BUS_REDIS_PORT"
	EventBusRedisPassword ConfigName = "EVENT_BUS_REDIS_PASSWORD"
	EventBusRedisDB       ConfigName = "EVENT_BUS_REDIS_DB"

	// HTTP Server
	HttpBasePath    ConfigName = "HTTP_BASE_PATH"
	HttpHost        ConfigName = "HTTP_HOST"
	HttpPort        ConfigName = "HTTP_PORT"
	HttpCorsOrigins ConfigName = "HTTP_CORS_ORIGINS"
	HttpCorsHeaders ConfigName = "HTTP_CORS_HEADERS"
	HttpCorsMethods ConfigName = "HTTP_CORS_METHODS"
)
