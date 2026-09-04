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

// ActionName and MainProcess are mandatory: a definition missing either is rejected at Init, so a
// broken build boots and then fails on the first request.
func TestActionsAreFullyDeclared(t *testing.T) {
	actions := append(channelActions(t), pointActions(t)...)
	actions = append(actions, orderActions(t)...)
	actions = append(actions, billActions(t)...)
	actions = append(actions, fiscalActions(t)...)
	actions = append(actions, quotationActions(t)...)
	actions = append(actions, partyActions(t)...)

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

// Every action that changes one record must name it. The collection-level exceptions are listed
// explicitly rather than pattern-matched, so adding another is a deliberate decision.
func TestRecordActionsAreScopedToAnId(t *testing.T) {
	actions := append(channelActions(t), pointActions(t)...)
	actions = append(actions, orderActions(t)...)
	actions = append(actions, billActions(t)...)
	actions = append(actions, fiscalActions(t)...)
	actions = append(actions, quotationActions(t)...)
	actions = append(actions, partyActions(t)...)

	collectionLevel := map[string]bool{
		ActionResolve:     true,
		ActionCreateOrder: true,
		ActionMergeBill:   true,

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

// The engine asserts each permission string against the action rows in the IAM migration; a code
// with no matching row denies every request with no hint in the 403.
func TestActionPermissionsAreSeeded(t *testing.T) {
	seeded := seededActionCodes(t)

	actions := append(channelActions(t), pointActions(t)...)
	actions = append(actions, orderActions(t)...)
	actions = append(actions, billActions(t)...)
	actions = append(actions, fiscalActions(t)...)
	actions = append(actions, quotationActions(t)...)
	actions = append(actions, partyActions(t)...)

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

// seededActionCodes parses the SQL rather than querying a database, so the test needs no database
// and fails in CI the moment an action is added without its seed.
func seededActionCodes(t *testing.T) map[string]bool {
	t.Helper()

	// Every Sales IAM migration, not just the first: reading only 1007002 would let an action seeded
	// later pass this check while being unreachable in production.
	//
	// Matched by pattern rather than listed by name, because the incremental per-submodule migrations
	// get folded back into 1007002_sales_iam.sql. A hardcoded list breaks on the consolidation and,
	// worse, silently stops covering a file that is renamed rather than deleted.
	pattern := filepath.Join("..", "..", "..", "scripts", "migrations", "*_sales*_iam.sql")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("listing the Sales IAM migrations: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no Sales IAM migration matched %s; the migrations moved or were renamed", pattern)
	}

	codes := map[string]bool{}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("the Sales IAM migration %s must be readable from the test: %v", path, err)
		}
		collectActionCodes(string(content), codes)
	}

	if len(codes) == 0 {
		t.Fatal("no action codes parsed out of the IAM migrations; the parser or the files changed")
	}
	return codes
}

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

func orderActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesOrderVoucherActions(engine); err != nil {
		t.Fatalf("defineSalesOrderVoucherActions: %v", err)
	}
	return engine.defs
}

func partyActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesOrderPartyActions(engine); err != nil {
		t.Fatalf("defineSalesOrderPartyActions: %v", err)
	}
	return engine.defs
}

// The other guards would not notice if the engine spec stopped naming DefineActions: the actions
// would never be installed, the routes would 404, and every test would still pass.
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

func quotationActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesQuotationActions(engine); err != nil {
		t.Fatalf("defineSalesQuotationActions: %v", err)
	}
	return engine.defs
}

func fiscalActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesFiscalRequestActions(engine); err != nil {
		t.Fatalf("defineSalesFiscalRequestActions: %v", err)
	}
	return engine.defs
}

// Same gap as for the order engine: a spec that stopped naming DefineActions would leave
// request_invoice with no route, 404ing in production while every other test passed.
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

func billActions(t *testing.T) []drif.DynamicActionDefinition {
	t.Helper()
	engine := &captureEngine{}
	if err := defineSalesBillActions(engine); err != nil {
		t.Fatalf("defineSalesBillActions: %v", err)
	}
	return engine.defs
}

// No eInvoice adapter ships here, so invoicingProvider is nil and SetInvoicingPort is never called.
// The absence is asserted so nobody deletes the setter as dead code or binds a stub that reports
// documents as issued when none were. When a real adapter lands, this test fails as a reminder to
// wire SetInvoicingPort in infra/external.
func TestTheInvoicingPortIsUnboundAndTheOperationToleratesIt(t *testing.T) {
	if invoicingProvider != nil {
		t.Fatal("an invoicing provider is now bound; wire SetInvoicingPort from infra/external and " +
			"update this test, rather than removing the assertion")
	}

	// Nil must stay a supported state; the alternative is a nil dereference on the first customer
	// who asks for a VAT invoice.
	spec := salesFiscalRequestEngineSpec()
	if spec.DefineActions == nil {
		t.Fatal("request_invoice must still be routed while the port is unbound: the request is " +
			"recorded as pending, which is BR 77's in-flight state")
	}
}
