package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/contact"
)

func NewContactServiceImpl(contactRepo it.ContactRepository) it.ContactService {
	return &ContactServiceImpl{
		contactRepo: contactRepo,
	}
}

type ContactServiceImpl struct {
	contactRepo it.ContactRepository
}

func (this *ContactServiceImpl) CreateContact(
	ctx corectx.Context, cmd it.CreateContactCommand,
) (*it.CreateContactResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Contact, *domain.Contact]{
		Action:         "create contact",
		BaseRepoGetter: this.contactRepo,
		Data:           cmd,
	})
}

func (this *ContactServiceImpl) DeleteContact(
	ctx corectx.Context, cmd it.DeleteContactCommand,
) (*it.DeleteContactResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete contact",
		DbRepoGetter: this.contactRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ContactServiceImpl) ContactExists(
	ctx corectx.Context, query it.ContactExistsQuery,
) (*it.ContactExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if contact exists",
		DbRepoGetter: this.contactRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *ContactServiceImpl) GetContact(
	ctx corectx.Context, query it.GetContactQuery,
) (*it.GetContactResult, error) {
	return corecrud.GetOne[domain.Contact](ctx, corecrud.GetOneParam{
		Action:       "get contact",
		DbRepoGetter: this.contactRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ContactServiceImpl) SearchContacts(
	ctx corectx.Context, query it.SearchContactsQuery,
) (*it.SearchContactsResult, error) {
	return corecrud.Search[domain.Contact](ctx, corecrud.SearchParam{
		Action:       "search contacts",
		DbRepoGetter: this.contactRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *ContactServiceImpl) UpdateContact(
	ctx corectx.Context, cmd it.UpdateContactCommand,
) (*it.UpdateContactResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Contact, *domain.Contact]{
		Action:       "update contact",
		DbRepoGetter: this.contactRepo,
		Data:         cmd,
	})
}
