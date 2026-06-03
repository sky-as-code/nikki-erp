package job

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type JobHandleFn func(ctx context.Context, jobArgs *string) error

func Handle(logger logging.LoggerService, jobName string, jobFn JobHandleFn, timeout time.Duration, jobArgs *string) error {
	ctx, cc := context.WithTimeout(context.Background(), timeout)
	defer cc()

	random, err := util.GenerateRandomStr("1234567890", 10)
	if err != nil {
		logger.Error("GenerateRandomStr err", err)
		return err
	}

	logger.Infof("[JOB] %s(%s) perform", jobName, random)
	err = jobFn(ctx, jobArgs)
	if err != nil {
		logger.Infof("[JOB] %s(%s) error", jobName, random)
		logger.Error("[JOB] error", err)
		return err
	}

	logger.Infof("[JOB] %s(%s) done", jobName, random)
	return nil
}
