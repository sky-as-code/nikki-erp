package role

import (
	"encoding/json"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// DescribeRolesMaxIds caps how many roles a single describe call may resolve. Each role fans
// out into an entitlement search plus scope-name lookups, so an unbounded list would let one
// request do arbitrary work.
const DescribeRolesMaxIds = 20

var describeRolesQueryType = cqrs.RequestType{Module: "iam", Submodule: "role", Action: "describeRoles"}

// DescribeRolesQuery resolves roles into a human-presentable shape: every entitlement carries
// its resource name, action name and scope name rather than the raw expression triple.
type DescribeRolesQuery struct {
	RoleIds []model.Id `json:"role_ids" query:"role_id"`
}

func (DescribeRolesQuery) CqrsRequestType() cqrs.RequestType { return describeRolesQueryType }

func (DescribeRolesQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.describe_roles_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(basemodel.DefineFieldIdArr("role_ids").
					Rule(dmodel.FieldRuleArrayLength(0, DescribeRolesMaxIds)))
		},
	)
}

// DescribedEntitlement is one entitlement with every id already resolved to a display name.
// ScopeId and ScopeName are only set for the org and orgunit scopes; domain and private
// entitlements carry no target, and the caller labels them from a static translation key.
type DescribedEntitlement struct {
	Id           model.Id  `json:"id"`
	ResourceId   *model.Id `json:"resource_id"`
	ResourceName *string   `json:"resource_name"`
	ActionId     *model.Id `json:"action_id"`
	ActionName   *string   `json:"action_name"`
	Scope        *string   `json:"scope"`
	ScopeId      *model.Id `json:"scope_id"`
	ScopeName    *string   `json:"scope_name"`
}

type DescribedRole struct {
	Id           model.Id               `json:"id"`
	Name         *string                `json:"name"`
	Entitlements []DescribedEntitlement `json:"entitlements"`
}

type DescribeRolesResultData struct {
	Items []DescribedRole `json:"items"`
}

type DescribeRolesResult = dyn.OpResult[DescribeRolesResultData]

const UpdateRoleFieldEntitlements = "entitlements"

// RoleEntitlementsDeltaMaxItems caps each list of the delta: every add costs a resource lookup
// and an insert inside one transaction.
const RoleEntitlementsDeltaMaxItems = 100

// RoleEntitlementGrant is one entitlement to create on the role. A nil ActionId grants every
// action of the resource (`*`). The caller does not choose the scope: the grant takes the
// resource's max_scope, and an `org` scope is bound to the role's own org.
type RoleEntitlementGrant struct {
	ResourceId  model.Id  `json:"resource_id"`
	ActionId    *model.Id `json:"action_id"`
	Description *string   `json:"description"`
}

// RoleEntitlementUpdate changes the description of one entitlement the role already holds. The
// description is the only editable column: every other one feeds the stored expression.
type RoleEntitlementUpdate struct {
	Id          model.Id `json:"id"`
	Description *string  `json:"description"`
}

// RoleEntitlementsDelta rides a role update as `entitlements: {add, remove, update}`. It is not
// a role column.
type RoleEntitlementsDelta struct {
	Add    []RoleEntitlementGrant  `json:"add"`
	Remove []model.Id              `json:"remove"`
	Update []RoleEntitlementUpdate `json:"update"`
}

func (this RoleEntitlementsDelta) IsEmpty() bool {
	return len(this.Add) == 0 && len(this.Remove) == 0 && len(this.Update) == 0
}

// HasEntitlementsDelta reports whether a role update carries a non-empty delta, without
// changing the command. A malformed delta counts: it must still be authorized as one.
func HasEntitlementsDelta(fields dmodel.DynamicFields) bool {
	raw, found := fields[UpdateRoleFieldEntitlements]
	if !found || raw == nil {
		return false
	}
	delta, err := decodeRoleEntitlementsDelta(raw)
	return err != nil || !delta.IsEmpty()
}

// LiftEntitlementsDelta removes `entitlements` from a role update, so the role schema never sees
// it, and returns the decoded delta. malformed reports a value that is not a delta.
func LiftEntitlementsDelta(fields dmodel.DynamicFields) (delta *RoleEntitlementsDelta, malformed bool) {
	raw, found := fields[UpdateRoleFieldEntitlements]
	if !found {
		return nil, false
	}
	delete(fields, UpdateRoleFieldEntitlements)
	if raw == nil {
		return nil, false
	}
	delta, err := decodeRoleEntitlementsDelta(raw)
	if err != nil {
		return nil, true
	}
	return delta, false
}

func decodeRoleEntitlementsDelta(raw any) (*RoleEntitlementsDelta, error) {
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	delta := &RoleEntitlementsDelta{}
	if err := json.Unmarshal(encoded, delta); err != nil {
		return nil, err
	}
	return delta, nil
}
