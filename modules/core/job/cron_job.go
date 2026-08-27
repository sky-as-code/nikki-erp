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
			jobTimeout := constants.BackgroudJobTimeout
			if len(timeout) > 0 {
				jobTimeout = timeout[0]
			}

			_ = Handle(this.logger, jobName, job, jobTimeout, nil)
		}),
	)

	return err
}

// Start begins firing the registered jobs.
//
// Called ONCE, by the application, after every module's OnAppStarted has run. That ordering is
// load-bearing: gocron fires on the schedule from the moment it starts, so starting earlier would
// mean a job registered by a later module simply never ran until the next tick - and for an hourly
// job that is an hour of silence nobody would connect to the cause.
func (this *CronJob) Start() error {
	this.scheduler.Start()
	this.logger.Infof("CronJob server started")

	return nil
}

// Stop halts the scheduler and waits for jobs already running to finish.
//
// Shutdown rather than StopJobs: the process is going away, so there is nothing to resume, and
// Shutdown is what releases the scheduler's own goroutines. It blocks until running jobs return,
// which is the point - a sweep killed half-way through would leave whatever it was writing
// uncommitted, and the caller's shutdown budget is what bounds the wait.
func (this *CronJob) Stop() error {
	err := this.scheduler.Shutdown()
	this.logger.Infof("CronJob server stopped")
	return err
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
