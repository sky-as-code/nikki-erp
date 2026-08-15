package computed

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// EvalPlan is the per-request evaluation plan: which computed fields a read must produce, which
// physical columns the projection must carry for them, and which batched source reads the
// related fields need. Built from the finalized SchemaPlan; contains no per-row state.
type EvalPlan struct {
	SchemaName string
	// Wanted lists the computed fields to evaluate, dependencies first. It includes computed
	// dependencies of requested fields even when not requested themselves.
	Wanted []string
	// ExtraFields are the physical columns to append to an explicit projection so evaluation
	// has its operands. Empty when the request had no explicit projection.
	ExtraFields []string
	// RelatedReads describe one batched source read per edge, covering every wanted related
	// field on that edge.
	RelatedReads []RelatedRead

	schemaPlan *SchemaPlan
}

// RelatedRead is one batched read: fetch the source rows of one to-one edge for a whole page,
// then copy each leaf onto the owning rows.
type RelatedRead struct {
	SchemaName string
	// FkColumn on the owning row holds the key; RefColumn on the source schema matches it.
	FkColumn  string
	RefColumn string
	// Leaves maps computed field name -> source leaf column.
	Leaves map[string]string
}

// BuildEvalPlan decides what a read request needs. requested is the client's explicit field
// projection; empty means the default field set, which includes every computed field. A nil,
// nil return means nothing computed is wanted and the read proceeds untouched.
func BuildEvalPlan(schemaName string, requested []string) (*EvalPlan, ft.ClientErrors) {
	schemaPlan := PlanFor(schemaName)
	if schemaPlan == nil {
		return nil, nil
	}
	wanted := wantedFields(schemaPlan, requested)
	if len(wanted) == 0 {
		return nil, nil
	}
	if limit := ActiveLimits().MaxComputedFieldsPerRequest; len(wanted) > limit {
		return nil, ft.ClientErrors{*ft.NewValidationError("fields",
			ft.ErrorKey("err_too_many_computed_fields"),
			"request evaluates too many computed fields")}
	}

	plan := &EvalPlan{SchemaName: schemaName, Wanted: wanted, schemaPlan: schemaPlan}
	plan.ExtraFields = missingOperands(schemaPlan, wanted, requested)
	plan.RelatedReads = groupRelatedReads(schemaPlan, wanted)
	return plan, nil
}

// wantedFields intersects the request with the schema's computed fields and closes over their
// computed dependencies, keeping the schema plan's dependency-safe order.
func wantedFields(schemaPlan *SchemaPlan, requested []string) []string {
	include := map[string]bool{}
	if len(requested) == 0 {
		for name := range schemaPlan.Fields {
			include[name] = true
		}
	} else {
		for _, name := range requested {
			if _, ok := schemaPlan.Fields[name]; ok {
				markWithDeps(schemaPlan, name, include)
			}
		}
	}

	wanted := make([]string, 0, len(include))
	for _, name := range schemaPlan.EvalOrder {
		if include[name] {
			wanted = append(wanted, name)
		}
	}
	return wanted
}

func markWithDeps(schemaPlan *SchemaPlan, name string, include map[string]bool) {
	if include[name] {
		return
	}
	include[name] = true
	for _, dep := range schemaPlan.Fields[name].ComputedDeps {
		markWithDeps(schemaPlan, dep, include)
	}
}

// missingOperands lists the physical operands of the wanted fields that the explicit projection
// does not already carry. With no explicit projection every column comes back anyway.
func missingOperands(schemaPlan *SchemaPlan, wanted []string, requested []string) []string {
	if len(requested) == 0 {
		return nil
	}
	present := map[string]bool{}
	for _, name := range requested {
		present[name] = true
	}
	var extra []string
	for _, name := range wanted {
		for _, operand := range schemaPlan.Fields[name].PhysicalOperands {
			if !present[operand] {
				present[operand] = true
				extra = append(extra, operand)
			}
		}
	}
	return extra
}

// groupRelatedReads merges the wanted related fields per edge, so N fields copied from one edge
// cost one batched read.
func groupRelatedReads(schemaPlan *SchemaPlan, wanted []string) []RelatedRead {
	byEdge := map[string]*RelatedRead{}
	var order []string
	for _, name := range wanted {
		fieldPlan := schemaPlan.Fields[name]
		if fieldPlan.Def.Kind != ComputeRelated {
			continue
		}
		read, ok := byEdge[fieldPlan.RelatedEdge]
		if !ok {
			read = &RelatedRead{
				SchemaName: fieldPlan.RelatedSchemaName,
				FkColumn:   fieldPlan.RelatedFkColumn,
				RefColumn:  fieldPlan.RelatedRefColumn,
				Leaves:     map[string]string{},
			}
			byEdge[fieldPlan.RelatedEdge] = read
			order = append(order, fieldPlan.RelatedEdge)
		}
		read.Leaves[name] = fieldPlan.RelatedLeaf
	}

	reads := make([]RelatedRead, 0, len(order))
	for _, edge := range order {
		reads = append(reads, *byEdge[edge])
	}
	return reads
}
