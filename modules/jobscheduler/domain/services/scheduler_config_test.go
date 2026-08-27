package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/constants"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/models"
)

// fakeConfigService returns whatever the test put in, so a test can prove a value travelled
// from configuration into the scheduler rather than being a literal somewhere in between.
type fakeConfigService struct {
	values map[c.ConfigName]any
}

func newFakeConfig() *fakeConfigService {
	return &fakeConfigService{values: map[c.ConfigName]any{
		constants.DefaultMaxAttempts:         3,
		constants.DefaultRetryIntervalSecs:   10,
		constants.ExpBackoffMaxIntervalSecs:  300,
		constants.DefaultAttemptTimeoutSecs:  60,
		constants.LeaseSafetyMarginSecs:      30,
		constants.ReconciliationIntervalSecs: 60,
		constants.WorkerConcurrency:          8,
		constants.ClaimBatchSize:             20,
		constants.DefaultConcurrencyPolicy:   models.ConcurrencyForbidOverlap,
		constants.DefaultMisfirePolicy:       models.MisfireRunOnce,
		constants.DefaultJobEnabled:          true,
		constants.HistoryRetentionDays:       30,
		constants.MisfireThresholdSecs:       120,
	}}
}

func (this *fakeConfigService) with(name c.ConfigName, val any) *fakeConfigService {
	this.values[name] = val
	return this
}

func (this *fakeConfigService) Init() error           { return nil }
func (this *fakeConfigService) GetAppVersion() string { return "test" }

func (this *fakeConfigService) GetStr(name c.ConfigName, _ ...any) string {
	val, _ := this.values[name].(string)
	return val
}

func (this *fakeConfigService) GetStrArr(c.ConfigName, ...any) []string { return nil }

func (this *fakeConfigService) GetDuration(c.ConfigName, ...any) time.Duration { return 0 }

func (this *fakeConfigService) GetBool(name c.ConfigName, _ ...any) bool {
	val, _ := this.values[name].(bool)
	return val
}

func (this *fakeConfigService) GetUint(c.ConfigName, ...any) uint     { return 0 }
func (this *fakeConfigService) GetUint64(c.ConfigName, ...any) uint64 { return 0 }

func (this *fakeConfigService) GetInt(name c.ConfigName, _ ...any) int {
	val, _ := this.values[name].(int)
	return val
}

func (this *fakeConfigService) GetInt32(c.ConfigName, ...any) int32     { return 0 }
func (this *fakeConfigService) GetInt64(c.ConfigName, ...any) int64     { return 0 }
func (this *fakeConfigService) GetFloat32(c.ConfigName, ...any) float32 { return 0 }

// AC-4 through AC-12: every default and system limit is read from Application Configuration
// rather than hard-coded. Distinctive values are used so a literal left behind in the code
// would show up as a mismatch here.
func TestEverySettingIsReadFromConfiguration(t *testing.T) {
	cfg := newFakeConfig().
		with(constants.DefaultMaxAttempts, 7).
		with(constants.DefaultRetryIntervalSecs, 11).
		with(constants.ExpBackoffMaxIntervalSecs, 77).
		with(constants.DefaultAttemptTimeoutSecs, 41).
		with(constants.LeaseSafetyMarginSecs, 13).
		with(constants.ReconciliationIntervalSecs, 90).
		with(constants.WorkerConcurrency, 5).
		with(constants.ClaimBatchSize, 17).
		with(constants.DefaultConcurrencyPolicy, models.ConcurrencyAllowOverlap).
		with(constants.DefaultMisfirePolicy, models.MisfireSkip).
		with(constants.DefaultJobEnabled, false).
		with(constants.HistoryRetentionDays, 45).
		with(constants.MisfireThresholdSecs, 99)

	loaded, err := LoadSchedulerConfig(cfg)

	require.NoError(t, err)
	assert.Equal(t, 7, loaded.DefaultMaxAttempts, "AC-8")
	assert.Equal(t, 11, loaded.DefaultRetryIntervalSecs, "AC-9")
	assert.Equal(t, 77, loaded.ExpBackoffMaxIntervalSecs, "AC-4")
	assert.Equal(t, 41, loaded.AttemptTimeoutSecs, "AC-5")
	assert.Equal(t, 13, loaded.LeaseSafetyMarginSecs)
	assert.Equal(t, 90*time.Second, loaded.ReconciliationInterval)
	assert.Equal(t, 5, loaded.WorkerConcurrency, "AC-7")
	assert.Equal(t, 17, loaded.ClaimBatchSize, "AC-6")
	assert.Equal(t, models.ConcurrencyAllowOverlap, loaded.DefaultConcurrencyPolicy, "AC-10")
	assert.Equal(t, models.MisfireSkip, loaded.DefaultMisfirePolicy, "AC-11")
	assert.False(t, loaded.DefaultJobEnabled)
	assert.Equal(t, 45, loaded.HistoryRetentionDays, "AC-12")
	assert.Equal(t, 99, loaded.MisfireThresholdSecs)
}

func TestDerivedDurationsComposeTheConfiguredSeconds(t *testing.T) {
	cfg := newFakeConfig().
		with(constants.DefaultAttemptTimeoutSecs, 60).
		with(constants.LeaseSafetyMarginSecs, 30)

	loaded, err := LoadSchedulerConfig(cfg)
	require.NoError(t, err)

	assert.Equal(t, 60*time.Second, loaded.AttemptTimeout())
	assert.Equal(t, 90*time.Second, loaded.LeaseDuration(),
		"the lease must outlast the attempt it covers, or a running worker is reaped")
	assert.Greater(t, loaded.LeaseDuration(), loaded.AttemptTimeout())
}

func TestLoadRejectsValuesThatWouldBreakTheSchedulerSilently(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*fakeConfigService)
	}{
		{"zero workers", func(cfg *fakeConfigService) {
			cfg.with(constants.WorkerConcurrency, 0)
		}},
		{"zero claim batch", func(cfg *fakeConfigService) {
			cfg.with(constants.ClaimBatchSize, 0)
		}},
		{"zero reconciliation interval", func(cfg *fakeConfigService) {
			cfg.with(constants.ReconciliationIntervalSecs, 0)
		}},
		{"zero max attempts", func(cfg *fakeConfigService) {
			cfg.with(constants.DefaultMaxAttempts, 0)
		}},
		{"retry interval below the business floor", func(cfg *fakeConfigService) {
			cfg.with(constants.DefaultRetryIntervalSecs, 4)
		}},
		{"negative lease margin", func(cfg *fakeConfigService) {
			cfg.with(constants.LeaseSafetyMarginSecs, -1)
		}},
		{"unknown concurrency policy", func(cfg *fakeConfigService) {
			cfg.with(constants.DefaultConcurrencyPolicy, "sometimes")
		}},
		{"unknown misfire policy", func(cfg *fakeConfigService) {
			cfg.with(constants.DefaultMisfirePolicy, "catch_up_all")
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newFakeConfig()
			tc.mutate(cfg)

			_, err := LoadSchedulerConfig(cfg)

			assert.Error(t, err,
				"a bad value must fail at startup, next to the change that caused it")
		})
	}
}

// The configured default must not undercut the floor the API enforces on a job's own
// override, or every job that declines to override would violate the rule.
func TestConfiguredRetryDefaultRespectsTheBusinessFloor(t *testing.T) {
	cfg := newFakeConfig().with(constants.DefaultRetryIntervalSecs, constants.MinRetryIntervalSeconds)

	loaded, err := LoadSchedulerConfig(cfg)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, loaded.DefaultRetryIntervalSecs, constants.MinRetryIntervalSeconds)
}

func TestOversizedClaimBatchIsReportedButNotFatal(t *testing.T) {
	cfg := newFakeConfig().
		with(constants.WorkerConcurrency, 2).
		with(constants.ClaimBatchSize, 100)

	loaded, err := LoadSchedulerConfig(cfg)

	require.NoError(t, err, "a disproportionate batch still works; it just wastes lease time")
	assert.True(t, loaded.IsClaimBatchOversized())
}

func TestProportionateClaimBatchIsNotFlagged(t *testing.T) {
	cfg := newFakeConfig().
		with(constants.WorkerConcurrency, 8).
		with(constants.ClaimBatchSize, 20)

	loaded, err := LoadSchedulerConfig(cfg)

	require.NoError(t, err)
	assert.False(t, loaded.IsClaimBatchOversized())
}

func TestInstanceIdIsStableWithinTheProcess(t *testing.T) {
	assert.NotEmpty(t, InstanceId())
	assert.Equal(t, InstanceId(), InstanceId(),
		"the id must not change under a running instance, or its own leases become unrecognizable")
}
