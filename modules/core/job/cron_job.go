package job

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.uber.org/dig"
)

type CronjobRegistry interface {
	Register(crontab string, jobName string, job JobHandleFn, timeout ...time.Duration) error
}

type CronJob struct {
	logger    logging.LoggerService
	scheduler gocron.Scheduler
}

func (this *CronJob) Register(crontab string,
	jobName string,
	job JobHandleFn, timeout ...time.Duration) error {
	_, err := this.scheduler.NewJob(
		gocron.CronJob(crontab, false),
		gocron.NewTask(func() {
			defer func() {
				if r := recover(); r != nil {
					this.logger.Errorf("[CRONJOB] %s panicked: %v", jobName, r)
				}
			}()
			jobTimeout := constants.BackgroundJobTimeout
			if len(timeout) > 0 {
				jobTimeout = timeout[0]
			}

			_ = Handle(this.logger, jobName, job, jobTimeout, nil)
		}),
	)

	return err
}

func (this *CronJob) Start() error {
	this.scheduler.Start()
	this.logger.Infof("CronJob server started")

	return nil
}

type initCronJobResult struct {
	dig.Out

	CronJob  *CronJob
	Registry CronjobRegistry
}

func initCronJob(logger logging.LoggerService) initCronJobResult {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logger.Errorf("Fail to init CronJob")
		panic(err)
	}

	cronJob := &CronJob{
		logger:    logger,
		scheduler: scheduler,
	}

	return initCronJobResult{
		CronJob:  cronJob,
		Registry: cronJob,
	}
}
