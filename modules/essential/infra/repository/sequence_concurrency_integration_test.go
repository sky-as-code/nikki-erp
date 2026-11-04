//go:build integration

// Concurrency acceptance tests for Sequence issuance (BR §13 AC10, AC11, AC14, AC43) against a real
// Postgres. They run the statements the repositories run - buildReserveIssuanceSql and
// buildIncrementCounterSql, plus the definition share lock and the completing update - in the order
// NextValue runs them, one transaction per request.
//
// Run: SEQ_IT_DB_URL='postgres://nikki_admin:nikki_password@localhost:5432/nikki_erp?sslmode=disable' \
//
//	go test -tags integration -run TestSequenceConcurrency ./modules/essential/infra/repository/
//
// The test borrows an existing (tenant, org) pair, creates its own definitions with a unique code and
// deletes every row it wrote (issuances, counters, definitions) on cleanup, so it leaves nothing that
// could disturb the seed files.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

const (
	itDefinitionTable = "essential_sequence_definitions"
	itCounterTable    = "essential_sequence_counters"
	itIssuanceTable   = "essential_sequence_issuances"
)

type seqHarness struct {
	t        *testing.T
	db       *sql.DB
	tenantId string
	orgId    string
	tenant   tenantScope
}

func newSeqHarness(t *testing.T) *seqHarness {
	url := os.Getenv("SEQ_IT_DB_URL")
	if url == "" {
		t.Skip("SEQ_IT_DB_URL is not set; skipping the Sequence concurrency integration test")
	}
	db, err := sql.Open("postgres", url)
	require.NoError(t, err)
	db.SetMaxOpenConns(40)
	t.Cleanup(func() { db.Close() })
	harness := &seqHarness{t: t, db: db}
	err = db.QueryRow("SELECT tenant_id, org_id FROM " + itDefinitionTable +
		" WHERE org_id IS NOT NULL LIMIT 1").Scan(&harness.tenantId, &harness.orgId)
	require.NoError(t, err, "the database needs at least one seeded org-scoped definition to borrow its tenant and org")
	harness.tenant = tenantScope{column: "tenant_id", value: harness.tenantId}
	return harness
}

func newUlid(t *testing.T) string {
	id, err := model.NewId()
	require.NoError(t, err)
	return string(*id)
}

func newUuid(t *testing.T) string {
	id, err := corectx.NewIdempotencyId()
	require.NoError(t, err)
	return id
}

// createDefinition inserts a scratch increment definition and registers its cleanup.
func (this *seqHarness) createDefinition() string {
	id := newUlid(this.t)
	now := time.Now().UTC()
	_, err := this.db.Exec("INSERT INTO "+itDefinitionTable+" (id, tenant_id, configuration_scope, org_id, "+
		"owner_module, code, name, scope_type, format_owner, algorithm, padding, increment, initial_value, "+
		"reset_period, timezone, issuance_retention_days, is_archived, created_at, updated_at, etag) VALUES "+
		"($1, $2, 'org', $3, 'robot', $4, '{\"en-US\": \"Integration\"}'::jsonb, 'org', 'sequence', 'increment', "+
		"5, 1, 1, 'never', 'UTC', 90, false, $5, $5, '1')",
		id, this.tenantId, this.orgId, "it.seq."+fmt.Sprint(now.UnixNano()), now)
	require.NoError(this.t, err)
	this.t.Cleanup(func() { this.deleteDefinition(id) })
	return id
}

func (this *seqHarness) deleteDefinition(id string) {
	for _, table := range []string{itIssuanceTable, itCounterTable, itDefinitionTable} {
		column := "definition_id"
		if table == itDefinitionTable {
			column = "id"
		}
		_, err := this.db.Exec("DELETE FROM "+table+" WHERE "+column+" = $1 AND tenant_id = $2", id, this.tenantId)
		assert.NoError(this.t, err, "cleanup of %s", table)
	}
}

type issueRequest struct {
	definitionId  string
	scopeKey      string
	idempotencyId string
	operationKey  string
}

type issueResult struct {
	value    int64
	replayed bool
}

// issue runs one NextValue in its own transaction: reserve the request key, share-lock the
// definition, move the counter, complete the issuance. A key already taken is replayed instead.
func (this *seqHarness) issue(ctx context.Context, req issueRequest) (issueResult, error) {
	tx, err := this.db.BeginTx(ctx, nil)
	if err != nil {
		return issueResult{}, err
	}
	defer tx.Rollback()
	issuanceId := newUlidOrPanic()
	orgId := model.Id(this.orgId)
	query, args := buildReserveIssuanceSql(itIssuanceTable, this.tenant, itSeq.ReserveIssuanceParam{
		NewId: model.Id(issuanceId), DefinitionId: model.Id(req.definitionId), OrgId: &orgId,
		IdempotencyId: req.idempotencyId, OperationKey: req.operationKey, RequestFingerprint: "it-fingerprint",
		CallerModule: "robot", AcceptedAt: time.Now().UTC(),
	}, nil)
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return issueResult{}, err
	}
	reserved := rows.Next()
	rows.Close()
	if !reserved {
		var value int64
		err = tx.QueryRowContext(ctx, "SELECT allocated_value FROM "+itIssuanceTable+
			" WHERE definition_id = $1 AND idempotency_id = $2::uuid AND operation_key = $3 AND tenant_id = $4",
			req.definitionId, req.idempotencyId, req.operationKey, this.tenantId).Scan(&value)
		return issueResult{value: value, replayed: true}, err
	}
	if _, err = tx.ExecContext(ctx, "SELECT id FROM "+itDefinitionTable+" WHERE id = $1 AND tenant_id = $2 FOR SHARE",
		req.definitionId, this.tenantId); err != nil {
		return issueResult{}, err
	}
	query, args, err = buildIncrementCounterSql(itCounterTable, this.tenant, itSeq.IncrementCounterParam{
		NewId: model.Id(newUlidOrPanic()), DefinitionId: model.Id(req.definitionId), OrgId: &orgId,
		ScopeKey: req.scopeKey, PeriodKey: "NONE", InitialValue: 1, Increment: 1,
	}, nil)
	if err != nil {
		return issueResult{}, err
	}
	var counterId string
	var value int64
	var params []byte
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&counterId, &value, &params); err != nil {
		return issueResult{}, err
	}
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, "UPDATE "+itIssuanceTable+" SET counter_id = $1, allocated_value = $2, "+
		"formatted_value = $3, effective_at = $4, issued_at = $4, expires_at = $5 WHERE id = $6 AND tenant_id = $7",
		counterId, value, fmt.Sprintf("IT%05d", value), now, now.Add(90*24*time.Hour), issuanceId, this.tenantId); err != nil {
		return issueResult{}, err
	}
	return issueResult{value: value}, tx.Commit()
}

func newUlidOrPanic() string {
	id, err := model.NewId()
	if err != nil {
		panic(err)
	}
	return string(*id)
}

// issueParallel fires every request at once and collects the results in request order.
func (this *seqHarness) issueParallel(requests []issueRequest) []issueResult {
	results := make([]issueResult, len(requests))
	errs := make([]error, len(requests))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range requests {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			results[index], errs[index] = this.issue(context.Background(), requests[index])
		}(i)
	}
	close(start)
	wg.Wait()
	for i, err := range errs {
		require.NoError(this.t, err, "request %d", i)
	}
	return results
}

func (this *seqHarness) counterState(definitionId string) (rows int, lastNumber int64) {
	err := this.db.QueryRow("SELECT COUNT(*), COALESCE(MAX(last_number), 0) FROM "+itCounterTable+
		" WHERE definition_id = $1 AND tenant_id = $2", definitionId, this.tenantId).Scan(&rows, &lastNumber)
	require.NoError(this.t, err)
	return rows, lastNumber
}

func (this *seqHarness) issuanceCount(definitionId string) int {
	var count int
	err := this.db.QueryRow("SELECT COUNT(*) FROM "+itIssuanceTable+" WHERE definition_id = $1 AND tenant_id = $2",
		definitionId, this.tenantId).Scan(&count)
	require.NoError(this.t, err)
	return count
}

func (this *seqHarness) scopeKey(suffix string) string {
	return "tenant:" + this.tenantId + "/org:" + this.orgId + suffix
}

// AC10 + AC14: 1000 distinct requests racing on one brand-new counter create it once and receive
// 1000 distinct values; the counter advances by exactly 1000 steps.
func TestSequenceConcurrency_DistinctKeysGetDistinctValues(t *testing.T) {
	harness := newSeqHarness(t)
	definitionId := harness.createDefinition()
	const total = 1000
	requests := make([]issueRequest, total)
	for i := range requests {
		requests[i] = issueRequest{definitionId: definitionId, scopeKey: harness.scopeKey(""),
			idempotencyId: newUuid(t), operationKey: fmt.Sprintf("it:%d", i)}
	}
	results := harness.issueParallel(requests)

	seen := make(map[int64]bool, total)
	for _, result := range results {
		assert.False(t, result.replayed)
		assert.False(t, seen[result.value], "value %d issued twice", result.value)
		seen[result.value] = true
	}
	for value := int64(1); value <= total; value++ {
		assert.True(t, seen[value], "value %d was never issued", value)
	}
	rows, last := harness.counterState(definitionId)
	assert.Equal(t, 1, rows, "AC14: the first-counter race must create a single counter")
	assert.Equal(t, int64(total), last, "AC10: the counter advances by exactly N steps")
	assert.Equal(t, total, harness.issuanceCount(definitionId))
}

// AC11: 100 concurrent requests with the same idempotency id and operation key produce one issuance
// and one increment, and all of them answer the same value.
func TestSequenceConcurrency_SameKeyIssuesOnce(t *testing.T) {
	harness := newSeqHarness(t)
	definitionId := harness.createDefinition()
	idempotencyId := newUuid(t)
	requests := make([]issueRequest, 100)
	for i := range requests {
		requests[i] = issueRequest{definitionId: definitionId, scopeKey: harness.scopeKey(""),
			idempotencyId: idempotencyId, operationKey: "it:same"}
	}
	results := harness.issueParallel(requests)

	fresh := 0
	for _, result := range results {
		assert.Equal(t, int64(1), result.value, "every response carries the one issued value")
		if !result.replayed {
			fresh++
		}
	}
	assert.Equal(t, 1, fresh, "exactly one request issues; the others replay")
	rows, last := harness.counterState(definitionId)
	assert.Equal(t, 1, rows)
	assert.Equal(t, int64(1), last, "one increment only")
	assert.Equal(t, 1, harness.issuanceCount(definitionId))
}

// AC43: an open transaction holding one counter's row lock does not block another counter.
func TestSequenceConcurrency_CountersDoNotSerialize(t *testing.T) {
	harness := newSeqHarness(t)
	definitionA := harness.createDefinition()
	definitionB := harness.createDefinition()
	ctx := context.Background()

	holder, err := harness.db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer holder.Rollback()
	orgId := model.Id(harness.orgId)
	query, args, err := buildIncrementCounterSql(itCounterTable, harness.tenant, itSeq.IncrementCounterParam{
		NewId: model.Id(newUlid(t)), DefinitionId: model.Id(definitionA), OrgId: &orgId,
		ScopeKey: harness.scopeKey(""), PeriodKey: "NONE", InitialValue: 1, Increment: 1,
	}, nil)
	require.NoError(t, err)
	_, err = holder.ExecContext(ctx, query, args...)
	require.NoError(t, err, "counter A is now locked by an open transaction")

	bounded, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	started := time.Now()
	result, err := harness.issue(bounded, issueRequest{definitionId: definitionB, scopeKey: harness.scopeKey(""),
		idempotencyId: newUuid(t), operationKey: "it:b"})
	require.NoError(t, err, "counter B must not wait for counter A's transaction")
	assert.Equal(t, int64(1), result.value)
	assert.Less(t, time.Since(started), 2*time.Second)

	// Same definition, another scope object: a separate counter row, so it does not wait either.
	result, err = harness.issue(bounded, issueRequest{definitionId: definitionA,
		scopeKey: harness.scopeKey("/it_scope:" + newUlid(t)), idempotencyId: newUuid(t), operationKey: "it:a2"})
	require.NoError(t, err, "another counter of the same definition must not wait either")
	assert.Equal(t, int64(1), result.value)
	require.NoError(t, holder.Rollback())
}
