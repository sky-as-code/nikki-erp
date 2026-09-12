package models

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The JSON documents extend the core base mixins by name, which CoreModule.RegisterModels does
// at app start-up. Without it every parse here panics on an unregistered mixin.
func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	os.Exit(m.Run())
}

// The schemas are parsed and validated at boot, so a malformed one panics the app rather than
// failing a request. Building them here turns that into a test failure instead.
func TestSchemasBuild(t *testing.T) {
	notificationSchema := NotificationSchemaBuilder().Build()
	assert.Equal(t, NotificationSchemaName, notificationSchema.Name())
	assert.Equal(t, "notification_notifications", notificationSchema.TableName())

	recipientSchema := RecipientSchemaBuilder().Build()
	assert.Equal(t, RecipientSchemaName, recipientSchema.Name())
	assert.Equal(t, "notification_recipients", recipientSchema.TableName())

	deliverySchema := DeliverySchemaBuilder().Build()
	assert.Equal(t, DeliverySchemaName, deliverySchema.Name())
	assert.Equal(t, "notification_deliveries", deliverySchema.TableName())
}

// The idempotency guard is a STRICT partial unique: rows carrying no key must stay unconstrained.
// A loose one would emit a companion "IS NULL" index and permit only one keyless notification per
// module per organization — which is almost all of them.
func TestIdempotencyGuardLeavesKeylessRowsUnconstrained(t *testing.T) {
	schema := NotificationSchemaBuilder().Build()

	strict := schema.PartialUniquesStrict()
	assert.Len(t, strict, 1)
	assert.Equal(t, NotificationFieldIdempotencyKey, strict[0].NullableField)
	assert.ElementsMatch(t,
		[]string{NotificationFieldOrgId, NotificationFieldSourceModule},
		strict[0].NotNullFields)
	assert.Empty(t, schema.PartialUniquesLoose())
}

// Read state lives on the recipient and nowhere else (BR 26), so the notification itself must not
// carry a read column that could disagree with it.
func TestNotificationCarriesNoReadState(t *testing.T) {
	schema := NotificationSchemaBuilder().Build()

	for _, name := range []string{"read_at", "is_read"} {
		_, found := schema.Field(name)
		assert.False(t, found, "notification must not declare %q", name)
	}
}

// BR 6 and BR 8: neither resource has an archive action, so neither may carry is_archived.
func TestNoArchiveColumns(t *testing.T) {
	notificationSchema := NotificationSchemaBuilder().Build()
	_, found := notificationSchema.Field(basemodel.FieldIsArchived)
	assert.False(t, found, "notification must not be archivable")

	deliverySchema := DeliverySchemaBuilder().Build()
	_, found = deliverySchema.Field(basemodel.FieldIsArchived)
	assert.False(t, found, "delivery must not be archivable")
}

func TestRecipientDeclaresTheFieldsTheStreamReads(t *testing.T) {
	schema := RecipientSchemaBuilder().Build()

	for _, name := range []string{
		RecipientFieldOrgId, RecipientFieldNotificationId, RecipientFieldRecipientUserId,
		RecipientFieldStreamSeq, RecipientFieldReadAt,
	} {
		_, found := schema.Field(name)
		assert.True(t, found, "recipient must declare %q", name)
	}
}
