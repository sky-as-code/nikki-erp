package job

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

var jobManagerSingleton *JobManager
var cronjobSingleton *CronJob

func InitSubModule() error {
	err := deps.Register(
		func(logger logging.LoggerService) initCronJobResult {
			return initCronJob(logger)
		},
		func(logger logging.LoggerService) initJobManagerResult {
			return initJobManger(logger)
		},
	)

	return err
}

func GetJobManger() *JobManager {
	if jobManagerSingleton != nil {
		return jobManagerSingleton
	}

	var jobManager *JobManager
	if err := deps.Invoke(func(jm *JobManager) { jobManager = jm }); err != nil {
		panic(err)
	}

	jobManagerSingleton = jobManager
	return jobManagerSingleton
}

func GetCronjob() *CronJob {
	if cronjobSingleton != nil {
		return cronjobSingleton
	}

	var cronJob *CronJob
	if err := deps.Invoke(func(cj *CronJob) { cronJob = cj }); err != nil {
		panic(err)
	}

	cronjobSingleton = cronJob
	return cronjobSingleton
}
