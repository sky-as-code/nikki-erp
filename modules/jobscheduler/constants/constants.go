package constants

// JobSchedulerModuleName is the module's name in the loader and the dependency graph. It is
// also the schema-name prefix, which is why schemaPrefixesOf in cmd/application.go needs no
// entry for this module: "jobscheduler_" is the default it derives from this name.
const JobSchedulerModuleName = "jobscheduler"

// Business constraints. These are capabilities of the scheduler rather than tunable defaults,
// so they are constants here and deliberately not application configuration: changing one
// would change what the scheduler is, not how it is tuned.
const (
	// MinRetryIntervalSeconds is the floor on a job's retry interval. A request below it is
	// rejected rather than rounded up, so a caller is never silently given a schedule
	// different from the one they asked for.
	MinRetryIntervalSeconds = 5

	// CronFieldCount is the number of fields in a supported cron expression. There is no
	// seconds field and no year field.
	CronFieldCount = 5
)
