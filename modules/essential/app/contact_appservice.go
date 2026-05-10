package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/contact"
)

func NewContactApplicationServiceImpl(contactSvc it.ContactDomainService) it.ContactAppService {
	return &ContactApplicationServiceImpl{contactSvc: contactSvc}
}

type ContactApplicationServiceImpl struct {
	contactSvc it.ContactDomainService
}

func (this *ContactApplicationServiceImpl) CreateContact(ctx corectx.Context, cmd it.CreateContactCommand) (*it.CreateContactResult, error) {
	return this.contactSvc.CreateContact(ctx, cmd)
}

func (this *ContactApplicationServiceImpl) DeleteContact(ctx corectx.Context, cmd it.DeleteContactCommand) (*it.DeleteContactResult, error) {
	return this.contactSvc.DeleteContact(ctx, cmd)
}

func (this *ContactApplicationServiceImpl) ContactExists(ctx corectx.Context, query it.ContactExistsQuery) (*it.ContactExistsResult, error) {
	return this.contactSvc.ContactExists(ctx, query)
}

func (this *ContactApplicationServiceImpl) GetContact(ctx corectx.Context, query it.GetContactQuery) (*it.GetContactResult, error) {
	return this.contactSvc.GetContact(ctx, query)
}

func (this *ContactApplicationServiceImpl) SearchContacts(ctx corectx.Context, query it.SearchContactsQuery) (*it.SearchContactsResult, error) {
	return this.contactSvc.SearchContacts(ctx, query)
}

func (this *ContactApplicationServiceImpl) UpdateContact(ctx corectx.Context, cmd it.UpdateContactCommand) (*it.UpdateContactResult, error) {
	return this.contactSvc.UpdateContact(ctx, cmd)
}
