// Package usagecheck asks the modules that reference a resource whether it is still in use,
// before the module owning it deletes it.
//
// Three mechanisms guard a cross-module reference and none replaces the others:
//
//	CheckUsage        business-level validation — "may this be deleted", answered by each
//	                  dependant against its own rules
//	Foreign key       database referential integrity — the final authority, which also settles
//	                  the race between a check and a concurrent insert
//	FK normalization  turns the database's refusal into the same client error the check produces
//
// The check runs over the command bus rather than through a direct port call, so a dependant
// that later moves into its own process changes its transport and nothing else. The owning
// module knows whom to ask from its hard-coded dependant list (modules.InCodeModuleDependants),
// never by inspecting the dependency graph — that graph runs the other way.
package usagecheck

import (
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

// SubmoduleName and ActionCheckUsage compose the request type a dependant subscribes to. The
// command bus allows one handler per request type, so the type is addressed to the DEPENDANT:
// "sales_usage.checkResourceUsage" is handled by sales alone, and the owning module sends one
// request per dependant rather than broadcasting.
const (
	SubmoduleName    = "usage"
	ActionCheckUsage = "checkResourceUsage"
)

// RequestTypeFor builds the request type addressed to one dependant module.
func RequestTypeFor(dependantModule string) cqrs.RequestType {
	return cqrs.RequestType{
		Module:    dependantModule,
		Submodule: SubmoduleName,
		Action:    ActionCheckUsage,
	}
}

// The resource names the checks are keyed by. They are declared here, in the package both sides
// import, because the owning module and every dependant must spell them identically: a mismatch
// resolves no checker and the handler answers with an error, which blocks every delete of that
// resource until someone notices.
const (
	ResourceUom            = "uom"
	ResourceProductVariant = "product_variant"

	// Owned by Inventory, alongside the variant.
	ResourceWarehouse         = "warehouse"
	ResourceInventoryLocation = "inventory_location"

	// ResourceStockQuant is a balance another module may name directly rather than looking it up
	// from a place — a vending slot stores the id of the balance it holds, so the row is that
	// slot's contents and deleting it would leave the slot pointing at nothing.
	ResourceStockQuant = "stock_quant"

	// Owned by Sales.
	//
	// Nothing dispatches these three yet: Sales declares no Dependants() and calls no
	// AssertDependantsSubscribed, so a checker registered against them answers a question that is
	// never asked. That is deliberate - registering them now means the owner side becomes a
	// one-line change on Sales' side alone, and until then deleting a sales point still succeeds.
	// Do not read a registered checker as protection: the OWNER decides when to ask.
	ResourceSalesPoint  = "sales_point"
	ResourceSalesOrder  = "sales_order"
	ResourceFulfillment = "fulfillment"
)

// IdentifierKeyId and IdentifierKeyOrgId are the identifier entries a checker reads.
const (
	IdentifierKeyId    = "id"
	IdentifierKeyOrgId = "org_id"
)

// NewIdentifier builds the identifier for a single-column primary key, carrying org scope when
// the resource has one. An org-scoped resource must be checked within its organization: a row in
// another organization referencing the same id is a different tenant's business, and letting it
// block this delete would leak the existence of that data.
//
// A resource with a composite key builds the map directly instead.
func NewIdentifier(id string, orgId string) map[string]string {
	identifier := map[string]string{IdentifierKeyId: id}
	if orgId != "" {
		identifier[IdentifierKeyOrgId] = orgId
	}
	return identifier
}

// ResourceRef names one resource instance to check.
//
// Identifier is a map rather than a single id because a primary key may be composite, and
// because org and tenant scoping travel with the identity: a reference held by another
// organization must not make the resource look used to this one.
type ResourceRef struct {
	ResourceName string            `json:"resource_name"`
	Identifier   map[string]string `json:"identifier"`
}

// CheckResourceUsageCommand asks one dependant module whether it still uses the listed
// resources.
//
// TargetModule is what the bus addresses the command with, so it is not merely descriptive: the
// same struct reaches a different handler for each dependant. It is also checked by the handler
// on arrival, so a command that reaches the wrong module is refused rather than answered for
// somebody else.
type CheckResourceUsageCommand struct {
	RequestId    string        `json:"request_id"`
	SourceModule string        `json:"source_module"`
	TargetModule string        `json:"target_module"`
	Resources    []ResourceRef `json:"resources"`
}

// CqrsRequestType addresses the command to its target module. The bus derives the topic from
// this, which is why TargetModule must be set before the command is sent; the dispatcher does
// that for every dependant it fans out to.
func (this CheckResourceUsageCommand) CqrsRequestType() cqrs.RequestType {
	return RequestTypeFor(this.TargetModule)
}

// ResourceUsageItem is one answer: whether this dependant considers this instance in use.
// The identifier is echoed so a caller can match answers to requests without relying on order.
type ResourceUsageItem struct {
	ResourceName string            `json:"resource_name"`
	Identifier   map[string]string `json:"identifier"`
	IsUsed       bool              `json:"is_used"`
	// UsedBy optionally names what makes it used - a resource name within the answering
	// module, never a record id. It sharpens the refusal message without exposing the
	// dependant's data.
	UsedBy string `json:"used_by,omitempty"`
}

// CheckResourceUsageResult is one dependant's reply. Results carries one entry per requested
// resource; a caller treats a missing entry as unanswered, which blocks the delete the same way
// an error does.
type CheckResourceUsageResult struct {
	RequestId string              `json:"request_id"`
	Module    string              `json:"module"`
	Results   []ResourceUsageItem `json:"results"`
}

// UsedItems returns the entries this dependant reported as in use.
func (this CheckResourceUsageResult) UsedItems() []ResourceUsageItem {
	used := make([]ResourceUsageItem, 0, len(this.Results))
	for _, item := range this.Results {
		if item.IsUsed {
			used = append(used, item)
		}
	}
	return used
}
