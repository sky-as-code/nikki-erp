package relationship

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type RelationshipService interface {
	CreateRelationship(ctx crud.Context, cmd CreateRelationshipCommand) (*CreateRelationshipResult, error)
}
