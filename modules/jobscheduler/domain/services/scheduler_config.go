package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/constants"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// Fallbacks used only if a configuration key is missing entirely. They match
// config.default.yaml. They exist because the configuration service type-asserts its fallback
// argument to a string and panics without one, so every read needs a literal even though the
// YAML should always supply the real value.
const (
	defaultMaxAttempts         = "3"
	defaultRetryIntervalSecs   = "10"
	defaultExpBackoffMaxSecs   = "300"
	defaultAttemptTimeoutSecs  = "60"
	defaultLeaseSafetyMargin   = "30"
	defaultReconcileIntervalS  = "60"
	defaultWorkerConcurrency   = "8"
	defaultClaimBatchSize      = "20"
	defaultConcurrencyPolicy   = models.ConcurrencyForbidOverlap
	defaultMisfirePolicy       = models.MisfireRunOnce
	defaultJobEnabled          = "true"
	defaultHistoryRetentionDay = "30"
	defaultMisfireThresholdS   = "120"
)

// claimBatchWarnFactor is how far the claim batch may exceed the worker pool before it is
// worth warning about. Some excess is healthy - it keeps the pool fed - but a large multiple
// means claiming work this instance cannot start, which burns lease time for nothing.
const claimBatchWarnFactor = 4

// SchedulerConfig is the Application Configuration the scheduler runs on, read once at
// startup.
//
// Reading once rather than per tick is deliberate. It is the same contract the rest of this
// codebase uses - a deployment that changes a value restarts to pick it up - and it is what
// lets an execution keep the configuration it was created with, so that retuning the
// scheduler cannot change the behaviour of work already in flight.
type SchedulerConfig struct {
	DefaultMaxAttempts        int
	DefaultRetryIntervalSecs  int
	ExpBackoffMaxIntervalSecs int
	AttemptTimeoutSecs        int
	LeaseSafetyMarginSecs     int
	ReconciliationInterval    time.Duration
	WorkerConcurrency         int
	ClaimBatchSize            int
	DefaultConcurrencyPolicy  string
	DefaultMisfirePolicy      string
	DefaultJobEnabled         bool
	HistoryRetentionDays      int
	MisfireThresholdSecs      int
}

// LoadSchedulerConfig reads every scheduler setting and validates the result.
//
// It returns an error rather than correcting a bad value, because each of these silently
// breaks the scheduler in a way that is hard to trace back: a zero worker pool accepts work
// and never runs it, a zero claim batch claims nothing, and an unrecognized policy string
// falls through every switch to a default the operator did not choose. Failing at startup
// puts the mistake next to the change that caused it.
func LoadSchedulerConfig(cfg config.ConfigService) (SchedulerConfig, error) {
	loaded := SchedulerConfig{
		DefaultMaxAttempts:        cfg.GetInt(constants.DefaultMaxAttempts, defaultMaxAttempts),
		DefaultRetryIntervalSecs:  cfg.GetInt(constants.DefaultRetryIntervalSecs, defaultRetryIntervalSecs),
		ExpBackoffMaxIntervalSecs: cfg.GetInt(constants.ExpBackoffMaxIntervalSecs, defaultExpBackoffMaxSecs),
		AttemptTimeoutSecs:        cfg.GetInt(constants.DefaultAttemptTimeoutSecs, defaultAttemptTimeoutSecs),
		LeaseSafetyMarginSecs:     cfg.GetInt(constants.LeaseSafetyMarginSecs, defaultLeaseSafetyMargin),
		WorkerConcurrency:         cfg.GetInt(constants.WorkerConcurrency, defaultWorkerConcurrency),
		ClaimBatchSize:            cfg.GetInt(constants.ClaimBatchSize, defaultClaimBatchSize),
		DefaultConcurrencyPolicy:  cfg.GetStr(constants.DefaultConcurrencyPolicy, defaultConcurrencyPolicy),
		DefaultMisfirePolicy:      cfg.GetStr(constants.DefaultMisfirePolicy, defaultMisfirePolicy),
		DefaultJobEnabled:         cfg.GetBool(constants.DefaultJobEnabled, defaultJobEnabled),
		HistoryRetentionDays:      cfg.GetInt(constants.HistoryRetentionDays, defaultHistoryRetentionDay),
		MisfireThresholdSecs:      cfg.GetInt(constants.MisfireThresholdSecs, defaultMisfireThresholdS),
	}

	reconcileSecs := cfg.GetInt(constants.ReconciliationIntervalSecs, defaultReconcileIntervalS)
	loaded.ReconciliationInterval = time.Duration(reconcileSecs) * time.Second

	if err := loaded.validate(reconcileSecs); err != nil {
		return SchedulerConfig{}, err
	}
	return loaded, nil
}

func (this SchedulerConfig) validate(reconcileSecs int) error {
	positives := []struct {
		name  string
		value int
	}{
		{string(constants.DefaultMaxAttempts), this.DefaultMaxAttempts},
		{string(constants.ExpBackoffMaxIntervalSecs), this.ExpBackoffMaxIntervalSecs},
		{string(constants.DefaultAttemptTimeoutSecs), this.AttemptTimeoutSecs},
		{string(constants.WorkerConcurrency), this.WorkerConcurrency},
		{string(constants.ClaimBatchSize), this.ClaimBatchSize},
		{string(constants.ReconciliationIntervalSecs), reconcileSecs},
		{string(constants.HistoryRetentionDays), this.HistoryRetentionDays},
	}
	for _, item := range positives {
		if item.value < 1 {
			return errors.Errorf("%s must be at least 1, got %d", item.name, item.value)
		}
	}

	if this.LeaseSafetyMarginSecs < 0 {
		return errors.Errorf("%s must not be negative, got %d",
			constants.LeaseSafetyMarginSecs, this.LeaseSafetyMarginSecs)
	}
	if this.MisfireThresholdSecs < 0 {
		return errors.Errorf("%s must not be negative, got %d",
			constants.MisfireThresholdSecs, this.MisfireThresholdSecs)
	}

	// The floor is a business constraint, so a configured default below it would let every
	// job that does not override the interval violate the rule the API enforces on the ones
	// that do.
	if this.DefaultRetryIntervalSecs < constants.MinRetryIntervalSeconds {
		return errors.Errorf("%s must be at least %d, got %d",
			constants.DefaultRetryIntervalSecs, constants.MinRetryIntervalSeconds,
			this.DefaultRetryIntervalSecs)
	}

	switch this.DefaultConcurrencyPolicy {
	case models.ConcurrencyForbidOverlap, models.ConcurrencyAllowOverlap:
	default:
		return errors.Errorf("%s must be %q or %q, got %q",
			constants.DefaultConcurrencyPolicy, models.ConcurrencyForbidOverlap,
			models.ConcurrencyAllowOverlap, this.DefaultConcurrencyPolicy)
	}

	switch this.DefaultMisfirePolicy {
	case models.MisfireRunOnce, models.MisfireSkip:
	default:
		return errors.Errorf("%s must be %q or %q, got %q",
			constants.DefaultMisfirePolicy, models.MisfireRunOnce,
			models.MisfireSkip, this.DefaultMisfirePolicy)
	}

	return nil
}

// AttemptTimeout is how long one attempt may run.
func (this SchedulerConfig) AttemptTimeout() time.Duration {
	return time.Duration(this.AttemptTimeoutSecs) * time.Second
}

// LeaseDuration is how long a claim holds. It covers the attempt timeout plus a margin, so
// that a worker still legitimately running is never reaped out from under itself.
func (this SchedulerConfig) LeaseDuration() time.Duration {
	return time.Duration(this.AttemptTimeoutSecs+this.LeaseSafetyMarginSecs) * time.Second
}

// MisfireThreshold is how late an occurrence may be and still run normally.
func (this SchedulerConfig) MisfireThreshold() time.Duration {
	return time.Duration(this.MisfireThresholdSecs) * time.Second
}

// IsClaimBatchOversized reports whether the claim batch is disproportionate to the worker
// pool. The caller warns rather than failing: the configuration still works, it just wastes
// lease time on executions the pool cannot start promptly.
func (this SchedulerConfig) IsClaimBatchOversized() bool {
	return this.ClaimBatchSize > this.WorkerConcurrency*claimBatchWarnFactor
}
