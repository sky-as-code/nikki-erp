package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/contact"
)

func NewContactDomainServiceImpl(contactRepo it.ContactRepository) it.ContactDomainService {
	return &ContactDomainServiceImpl{
		contactRepo: contactRepo,
	}
}

type ContactDomainServiceImpl struct {
	contactRepo it.ContactRepository
}

func (this *ContactDomainServiceImpl) CreateContact(
	ctx corectx.Context, cmd it.CreateContactCommand,
) (*it.CreateContactResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.Contact, *models.Contact]{
		Action:         "create contact",
		BaseRepoGetter: this.contactRepo,
		Data:           cmd,
	})
}

func (this *ContactDomainServiceImpl) DeleteContact(
	ctx corectx.Context, cmd it.DeleteContactCommand,
) (*it.DeleteContactResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete contact",
		DbRepoGetter: this.contactRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ContactDomainServiceImpl) ContactExists(
	ctx corectx.Context, query it.ContactExistsQuery,
) (*it.ContactExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if contact exists",
		DbRepoGetter: this.contactRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *ContactDomainServiceImpl) GetContact(
	ctx corectx.Context, query it.GetContactQuery,
) (*it.GetContactResult, error) {
	return corecrud.GetOne[models.Contact](ctx, corecrud.GetOneParam{
		Action:       "get contact",
		DbRepoGetter: this.contactRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *ContactDomainServiceImpl) SearchContacts(
	ctx corectx.Context, query it.SearchContactsQuery,
) (*it.SearchContactsResult, error) {
	return corecrud.Search[models.Contact](ctx, corecrud.SearchParam{
		Action:       "search contacts",
		DbRepoGetter: this.contactRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *ContactDomainServiceImpl) UpdateContact(
	ctx corectx.Context, cmd it.UpdateContactCommand,
) (*it.UpdateContactResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.Contact, *models.Contact]{
		Action:       "update contact",
		DbRepoGetter: this.contactRepo,
		Data:         cmd,
	})
}
