package contact

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type ContactService interface {
	CreateContact(ctx corectx.Context, cmd CreateContactCommand) (*CreateContactResult, error)
	DeleteContact(ctx corectx.Context, cmd DeleteContactCommand) (*DeleteContactResult, error)
	ContactExists(ctx corectx.Context, query ContactExistsQuery) (*ContactExistsResult, error)
	GetContact(ctx corectx.Context, query GetContactQuery) (*GetContactResult, error)
	SearchContacts(ctx corectx.Context, query SearchContactsQuery) (*SearchContactsResult, error)
	UpdateContact(ctx corectx.Context, cmd UpdateContactCommand) (*UpdateContactResult, error)
}
