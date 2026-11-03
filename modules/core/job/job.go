package job

import (
	"context"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type JobHandleFn func(ctx context.Context, jobArgs *string) error

// Handle runs one job under its own root span and idempotency id. A panic is contained here,
// where the run's context still exists to log it with.
func Handle(
	logger logging.LoggerService, jobName string, jobFn JobHandleFn, timeout time.Duration, jobArgs *string,
) (err error) {
	ctx, cc := context.WithTimeout(context.Background(), timeout)
	defer cc()
	ctx, span := StartRun(ctx, jobName)
	defer func() { EndRun(span, err) }()
	ctx = corectx.WithRequestMeta(ctx, corectx.NewRequestMeta(corectx.RequestSourceJob))
	defer func() {
		if recovered := recover(); recovered != nil {
			err = ft.RecoverPanicFailedTo(recovered, "run job "+jobName)
			logger.Error(ctx, "[JOB] panicked", err)
		}
	}()

	err = jobFn(ctx, jobArgs)
	if err != nil {
		logger.Error(ctx, "[JOB] error", err)
	}
	return err
}
