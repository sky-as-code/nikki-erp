package cqrs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataRoundTripsThroughTheContext(t *testing.T) {
	ctx := WithMetadata(context.Background(), map[string]string{
		MetaIdempotencyKey:       "inventory:rebuild:2026-08-20T10:00:00Z",
		MetaSchedulerJobId:       "01M2JBJ0000000001000000000",
		MetaSchedulerExecutionId: "01M2JBE0000000001000000000",
		MetaSchedulerAttempt:     "2",
	})

	md := MetadataFrom(ctx)

	require.NotNil(t, md)
	assert.Equal(t, "inventory:rebuild:2026-08-20T10:00:00Z", md[MetaIdempotencyKey])
	assert.Equal(t, "2", md[MetaSchedulerAttempt])
}

// A caller must not be able to forge the envelope fields the bus owns. Letting one set a
// correlation id or a reply topic would let it redirect somebody else's reply to itself.
func TestReservedKeysCannotBeSuppliedByACaller(t *testing.T) {
	ctx := WithMetadata(context.Background(), map[string]string{
		MetaCorrelationId:  "forged",
		MetaReplyTopic:     "attacker-topic",
		MetaRequestTopic:   "attacker-topic",
		MetaNoReply:        "true",
		MetaIdempotencyKey: "kept",
	})

	md := MetadataFrom(ctx)

	assert.NotContains(t, md, MetaCorrelationId)
	assert.NotContains(t, md, MetaReplyTopic)
	assert.NotContains(t, md, MetaRequestTopic)
	assert.NotContains(t, md, MetaNoReply)
	assert.Equal(t, "kept", md[MetaIdempotencyKey], "unreserved keys still pass through")
}

// Successive calls merge rather than replace, so a caller can add to metadata a layer above it
// already attached without having to know what that was.
func TestMetadataAccumulatesAcrossCalls(t *testing.T) {
	ctx := WithMetadata(context.Background(), map[string]string{MetaSchedulerJobId: "job"})
	ctx = WithMetadata(ctx, map[string]string{MetaSchedulerAttempt: "3"})

	md := MetadataFrom(ctx)

	assert.Equal(t, "job", md[MetaSchedulerJobId])
	assert.Equal(t, "3", md[MetaSchedulerAttempt])
}

func TestLaterValuesWinForTheSameKey(t *testing.T) {
	ctx := WithMetadata(context.Background(), map[string]string{MetaSchedulerAttempt: "1"})
	ctx = WithMetadata(ctx, map[string]string{MetaSchedulerAttempt: "2"})

	assert.Equal(t, "2", MetadataFrom(ctx)[MetaSchedulerAttempt])
}

// The derived context must not mutate the one it came from: a retry building its own metadata
// must not rewrite what a previous attempt attached.
func TestWithMetadataDoesNotMutateItsParent(t *testing.T) {
	parent := WithMetadata(context.Background(), map[string]string{MetaSchedulerAttempt: "1"})

	_ = WithMetadata(parent, map[string]string{MetaSchedulerAttempt: "2"})

	assert.Equal(t, "1", MetadataFrom(parent)[MetaSchedulerAttempt])
}

func TestEmptyMetadataLeavesTheContextUnchanged(t *testing.T) {
	assert.Nil(t, MetadataFrom(context.Background()))
	assert.Nil(t, MetadataFrom(WithMetadata(context.Background(), nil)))
	assert.Nil(t, MetadataFrom(WithMetadata(context.Background(), map[string]string{})))
}

// Only reserved keys are dropped; a caller supplying nothing but reserved keys gets an empty map
// rather than a panic or a partially-forged envelope.
func TestOnlyReservedKeysYieldsAnEmptySet(t *testing.T) {
	ctx := WithMetadata(context.Background(), map[string]string{MetaCorrelationId: "forged"})

	assert.Empty(t, MetadataFrom(ctx))
}
