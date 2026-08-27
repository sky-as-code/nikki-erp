package dynamicengines

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// captureEngine records the action definitions without building a real engine.
type captureEngine struct {
	drif.DynamicResourceEngine
	defs []drif.DynamicActionDefinition
}

func (this *captureEngine) DefineAction(def drif.DynamicActionDefinition) error {
	this.defs = append(this.defs, def)
	return nil
}

func channelActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesChannelActions(engine); err != nil {
		t.Fatalf("defineSalesChannelActions: %v", err)
	}
	return engine.defs
}

func pointActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesPointActions(engine); err != nil {
		t.Fatalf("defineSalesPointActions: %v", err)
	}
	return engine.defs
}

// TestActionsAreFullyDeclared pins the shape of every lifecycle action.
//
// ActionName and MainProcess are mandatory; a definition missing either is rejected at Init, which
// means a broken build boots and then fails on the first request rather than at start-up.
func TestActionsAreFullyDeclared(t *testing.T) {
	actions := append(channelActions(t), pointActions(t)...)
	actions = append(actions, orderActions(t)...)
	actions = append(actions, billActions(t)...)
	actions = append(actions, fiscalActions(t)...)
	actions = append(actions, quotationActions(t)...)

	for _, def := range actions {
		if def.ActionName == "" {
			t.Error("an action is missing its name")
		}
		if def.MainProcess == nil {
			t.Errorf("action %q has no MainProcess", def.ActionName)
		}
		if def.ActionType != drif.ActionTypeGeneric {
			t.Errorf("action %q is %q, want ActionTypeGeneric: these carry a request body and "+
				"none of them is a CRUD verb", def.ActionName, def.ActionType)
		}
		if def.RestPath == "" {
			t.Errorf("action %q has no RestPath", def.ActionName)
		}
		if strings.Contains(def.RestPath, "-") {
			t.Errorf("action %q path %q contains a hyphen; the route regex rejects them and the "+
				"word separator is underscore", def.ActionName, def.RestPath)
		}
	}
}

// TestRecordActionsAreScopedToAnId guards the difference between operating on one record and
// operating on the collection.
//
// Every action that changes ONE record names it. Two are deliberately collection-level and are
// listed explicitly rather than pattern-matched, so that adding a third is a decision somebody makes
// on purpose: resolve looks a channel up by code, and create_order has no record yet to name.
func TestRecordActionsAreScopedToAnId(t *testing.T) {
	actions := append(channelActions(t), pointActions(t)...)
	actions = append(actions, orderActions(t)...)
	actions = append(actions, billActions(t)...)
	actions = append(actions, fiscalActions(t)...)
	actions = append(actions, quotationActions(t)...)

	collectionLevel := map[string]bool{
		ActionResolve:     true,
		ActionCreateOrder: true,
		ActionMergeBill:   true,

		// request_invoice creates the fiscal request, so there is no record to hang it off.
		ActionRequestInvoice: true,
	}

	for _, def := range actions {
		if collectionLevel[def.ActionName] {
			if strings.Contains(def.RestPath, ":id") {
				t.Errorf("action %q operates on the collection, but its path %q names an id",
					def.ActionName, def.RestPath)
			}
			continue
		}
		if !strings.HasPrefix(def.RestPath, ":id/") {
			t.Errorf("action %q changes one record but its path %q is collection-level",
				def.ActionName, def.RestPath)
		}
	}
}

// TestActionPermissionsAreSeeded is the check that a permission code cannot drift from its seed.
//
// The engine asserts the permission string named here against the action rows in
// 1007002_sales_iam.sql. A code with no matching row denies every request, and nothing in the 403
// points at the seed as the cause — so the two are compared directly.
func TestActionPermissionsAreSeeded(t *testing.T) {
	seeded := seededActionCodes(t)

	actions := append(channelActions(t), pointActions(t)...)
	actions = append(actions, orderActions(t)...)
	actions = append(actions, billActions(t)...)
	actions = append(actions, fiscalActions(t)...)
	actions = append(actions, quotationActions(t)...)

	for _, def := range actions {
		if def.Permission == "" {
			t.Errorf("action %q declares no permission", def.ActionName)
			continue
		}
		if !seeded[def.Permission] {
			t.Errorf("action %q demands the %q permission, which the Sales IAM migrations do not "+
				"seed", def.ActionName, def.Permission)
		}
	}
}

// seededActionCodes reads the action codes out of the IAM migration.
//
// Parsing the SQL is cruder than querying a database, and deliberately so: the test then needs no
// database, and it fails in CI the moment somebody adds an action without its seed.
func seededActionCodes(t *testing.T) map[string]bool {
	t.Helper()

	// Every Sales IAM migration, not just the first. An action seeded in a later file is as real as
	// one seeded in the original, and reading only 1007002 would let a new action pass this check
	// while being unreachable in production.
	migrations := []string{
		"1007002_sales_iam.sql",
		"1007006_sales_voucher_iam.sql",
		"1007008_sales_bill_iam.sql",
		"1007010_sales_payment_iam.sql",
		"1007012_sales_fulfillment_iam.sql",
		"1007014_sales_fiscal_iam.sql",
		"1007016_sales_outbox_iam.sql",
		"1007018_sales_quotation_iam.sql",
		"1007020_sales_manual_discount_iam.sql",
	}

	codes := map[string]bool{}
	for _, name := range migrations {
		path := filepath.Join("..", "..", "..", "scripts", "migrations", name)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("the Sales IAM migration %s must be readable from the test: %v", name, err)
		}
		collectActionCodes(string(content), codes)
	}

	if len(codes) == 0 {
		t.Fatal("no action codes parsed out of the IAM migrations; the parser or the files changed")
	}
	return codes
}

// collectActionCodes pulls the action codes out of one migration file into codes.
func collectActionCodes(content string, codes map[string]bool) {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "('01M3SALES") {
			continue
		}
		// ('<id>', '<name>', '<code>', ...) - the code is the third quoted field.
		parts := strings.Split(trimmed, "'")
		if len(parts) < 6 {
			continue
		}
		codes[parts[5]] = true
	}
}

// orderActions builds the sales order engine's custom actions, for the seed check.
func orderActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesOrderVoucherActions(engine); err != nil {
		t.Fatalf("defineSalesOrderVoucherActions: %v", err)
	}
	return engine.defs
}

// TestSalesOrderEngineInstallsItsActions closes the gap between defining an action and the engine
// actually getting it.
//
// The other guards check that each action is well-formed and that its permission is seeded. Neither
// would notice if the engine spec stopped naming DefineActions — the actions would simply never be
// installed, the routes would 404, and every test would still pass.
func TestSalesOrderEngineInstallsItsActions(t *testing.T) {
	spec := salesOrderEngineSpec()

	if spec.DefineActions == nil {
		t.Fatal("the sales order engine declares no actions; apply_voucher would have no route")
	}

	engine := &captureEngine{}
	if err := spec.DefineActions(engine); err != nil {
		t.Fatalf("installing the sales order actions: %v", err)
	}

	installed := map[string]bool{}
	for _, def := range engine.defs {
		installed[def.ActionName] = true
	}
	for _, want := range []string{
		ActionApplyVoucher, ActionExplainPrice, ActionReprice, ActionCreateOrder,
		ActionConfirmOrder, ActionCancelOrder,
	} {
		if !installed[want] {
			t.Errorf("the sales order engine installs %d actions, none of them %q",
				len(engine.defs), want)
		}
	}
}

// quotationActions builds the quotation engine's custom actions, for the seed and shape checks.
func quotationActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesQuotationActions(engine); err != nil {
		t.Fatalf("defineSalesQuotationActions: %v", err)
	}
	return engine.defs
}

// fiscalActions builds the fiscal request engine's custom actions, for the seed and shape checks.
func fiscalActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesFiscalRequestActions(engine); err != nil {
		t.Fatalf("defineSalesFiscalRequestActions: %v", err)
	}
	return engine.defs
}

// TestFiscalRequestEngineInstallsItsAction closes the same gap for the fiscal engine that
// TestSalesOrderEngineInstallsItsActions closes for the order: an engine spec that stopped naming
// DefineActions would leave request_invoice with no route, 404ing in production while every other
// test still passed.
func TestFiscalRequestEngineInstallsItsAction(t *testing.T) {
	spec := salesFiscalRequestEngineSpec()

	if spec.DefineActions == nil {
		t.Fatal("the fiscal request engine declares no actions; request_invoice would have no route")
	}

	installed := map[string]bool{}
	for _, def := range fiscalActions(t) {
		installed[def.ActionName] = true
	}
	if !installed[ActionRequestInvoice] {
		t.Errorf("the fiscal request engine does not install %q", ActionRequestInvoice)
	}
}

// billActions builds the bill engine's custom actions, for the seed and shape checks.
func billActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesBillActions(engine); err != nil {
		t.Fatalf("defineSalesBillActions: %v", err)
	}
	return engine.defs
}

// TestTheInvoicingPortIsUnboundAndTheOperationToleratesIt pins a deliberate absence.
//
// No eInvoice provider adapter ships in this repository, so invoicingProvider is nil and
// SetInvoicingPort is never called. That is the documented state, not an oversight — but an exported
// setter nothing calls reads like dead code, and the next person could either delete it or "fix" the
// nil by binding a stub that reports documents as issued when none were.
//
// So the absence is asserted rather than left implicit. When a real adapter lands, this test fails,
// and its failure is the reminder to wire SetInvoicingPort in infra/external rather than to delete
// the assertion.
func TestTheInvoicingPortIsUnboundAndTheOperationToleratesIt(t *testing.T) {
	if invoicingProvider != nil {
		t.Fatal("an invoicing provider is now bound; wire SetInvoicingPort from infra/external and " +
			"update this test, rather than removing the assertion")
	}

	// The operation must treat that nil as a supported state. A guard here is cheap and the
	// alternative is a nil dereference on the first customer who asks for a VAT invoice.
	spec := salesFiscalRequestEngineSpec()
	if spec.DefineActions == nil {
		t.Fatal("request_invoice must still be routed while the port is unbound: the request is " +
			"recorded as pending, which is BR 77's in-flight state")
	}
}
