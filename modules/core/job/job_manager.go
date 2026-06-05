package job

import (
	"fmt"
	"os"
	"time"

	"github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.uber.org/dig"
)

type Job struct {
	handleFn JobHandleFn
	timeout  time.Duration
}

type JobManager struct {
	jobs   map[string]Job
	logger logging.LoggerService
}

type JobManagerRegistry interface {
	Register(jobName string, handleFn JobHandleFn, timeout ...time.Duration)
}

func (this *JobManager) Register(jobName string, handleFn JobHandleFn, jobTimeout ...time.Duration) {
	if _, ok := this.jobs[jobName]; ok {
		panic(fmt.Sprintf("[Job Register] duplicate job name: %s", jobName))
	}

	job := Job{
		handleFn: handleFn,
		timeout:  constants.BackgroudJobTimeout,
	}

	if len(jobTimeout) > 0 {
		job.timeout = jobTimeout[0]
	}

	this.jobs[jobName] = job
	this.logger.Infof("[JobManager] Register job %s", jobName)
}

func (this *JobManager) HandleJob(jobName string, jobArgs *string) {
	job, ok := this.jobs[jobName]
	if !ok {
		this.logger.Errorf("[JobManager] unknown job: %s", jobName)
		os.Exit(1)
	}

	err := Handle(this.logger, jobName, job.handleFn, job.timeout, jobArgs)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

type initJobManagerResult struct {
	dig.Out

	JobManager *JobManager
	Registry   JobManagerRegistry
}

func initJobManger(logger logging.LoggerService) initJobManagerResult {
	jm := JobManager{
		jobs:   map[string]Job{},
		logger: logger,
	}

	return initJobManagerResult{
		JobManager: &jm,
		Registry:   &jm,
	}
}
