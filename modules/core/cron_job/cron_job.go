package cron_job

import (
	"context"

	"github.com/go-co-op/gocron/v2"
	coreConst "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.uber.org/dig"
)

type JobFn func(ctx context.Context) error

type JobRegistry interface {
	Register(crontab string, jobName string, job JobFn) error
}

type CronJob struct {
	logger    logging.LoggerService
	scheduler gocron.Scheduler
}

func (this *CronJob) Register(crontab string,
	jobName string,
	job JobFn) error {
	_, err := this.scheduler.NewJob(
		gocron.CronJob(crontab, false),
		gocron.NewTask(func() {
			defer func() {
				// fallback handling if job is panicked
				if r := recover(); r != nil {
					this.logger.Errorf("[CRONJOB] %s panicked: %v", jobName, r)
				}
			}()

			this.logger.Infof("[CRONJOB] %s perform", jobName)
			bgCtx, cc := context.WithTimeout(context.Background(), coreConst.BackgroundTimeout)
			defer cc()

			job(bgCtx)
			this.logger.Infof("[CRONJOB] %s done", jobName)
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
	Registry JobRegistry
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
