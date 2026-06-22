package group

import (
	"github.com/sky-as-code/nikki-erp/common/orm"
)

// Types below support the legacy Ent group repository implementation.

type DeleteParam = DeleteGroupCommand
type AddRemoveUsersParam = ManageGroupUsersCommand
type ExistsParam = GroupExistsQuery

type FindByNameParam struct {
	Name string
}

type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
