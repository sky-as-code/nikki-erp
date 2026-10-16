package app

import (
	"time"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

const (
	cronPendingPoller    = "* * * * *"
	jobNamePendingPoller = "paymentinvoice-pending-poller"

	pendingPollWindow = 5 * time.Minute
)

type PendingPoller struct {
	orders   *services.OrderDomainService
	notifier *ResultNotifier
	logger   logging.LoggerService
	now      func() time.Time
}

func NewPendingPoller(
	orders *services.OrderDomainService,
	notifier *ResultNotifier,
	logger logging.LoggerService,
) *PendingPoller {
	return &PendingPoller{
		orders:   orders,
		notifier: notifier,
		logger:   logger,
		now:      time.Now,
	}
}

func (this *PendingPoller) RegisterJobs(registry job.CronjobRegistry) error {
	return registry.Register(cronPendingPoller, jobNamePendingPoller,
		job.ScopeSweep(this.Poll, jobNamePendingPoller))
}

func (this *PendingPoller) Poll(ctx corectx.Context) error {
	awaiting, err := services.FindAwaitingOrders(ctx, this.now().Add(-pendingPollWindow))
	if err != nil {
		return errors.Wrap(err, jobNamePendingPoller)
	}
	if len(awaiting) == 0 {
		return nil
	}
	this.logger.Infof("%s: %d order(s) awaiting payment", jobNamePendingPoller, len(awaiting))

	for _, order := range awaiting {
		this.pollOne(ctx, order)
	}
	return nil
}

func (this *PendingPoller) pollOne(ctx corectx.Context, order services.StaleOrder) {
	verdict, err := this.orders.ReconcileStaleOrder(ctx, order)
	if err != nil {
		this.logger.Warnf("%s: order '%s' could not be checked: %s",
			jobNamePendingPoller, order.OrderId, err.Error())
		return
	}
	if !verdict.Settled || !verdict.Applied {
		return
	}

	this.logger.Infof("%s: order '%s' settled by polling, paid=%t",
		jobNamePendingPoller, order.OrderId, verdict.Paid)
	this.notifier.Notify(ctx, NotifyTarget{
		Pk:        order.Pk,
		OrderId:   order.OrderId,
		ReturnUrl: order.ReturnUrl,
	}, verdictStatus(verdict.Paid))
}
