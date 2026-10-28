package job

import (
	"context"
	"time"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type JobHandleFn func(ctx context.Context, jobArgs *string) error

func Handle(logger logging.LoggerService, jobName string, jobFn JobHandleFn, timeout time.Duration, jobArgs *string) error {
	ctx, cc := context.WithTimeout(context.Background(), timeout)
	defer cc()
	// One id per run, so every entry the job writes reads as one execution.
	ctx = corectx.WithRequestMeta(ctx, corectx.NewRequestMeta(corectx.RequestSourceJob))

	// random, err := util.GenerateRandomStr("1234567890", 10)
	// if err != nil {
	// 	logger.Error("GenerateRandomStr err", err)
	// 	return err
	// }

	// logger.Infof("[JOB] %s(%s) perform", jobName, random)
	err := jobFn(ctx, jobArgs)
	if err != nil {
		// logger.Infof("[JOB] %s(%s) error", jobName, random)
		logger.Error("[JOB] error", err)
		return err
	}

	// logger.Infof("[JOB] %s(%s) done", jobName, random)
	return nil
}
