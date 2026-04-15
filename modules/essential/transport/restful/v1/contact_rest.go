package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/contact"
)

type contactRestParams struct {
	dig.In

	ContactSvc it.ContactService
}

func NewContactRest(params contactRestParams) *ContactRest {
	return &ContactRest{contactSvc: params.ContactSvc}
}

type ContactRest struct {
	contactSvc it.ContactService
}

func (this ContactRest) CreateContact(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST create contact"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.contactSvc.CreateContact,
		func(request CreateContactRequest) it.CreateContactCommand {
			cmd := it.CreateContactCommand{}
			cmd.SetFieldData(request.DynamicFields)
			return cmd
		},
		func(data domain.Contact) CreateContactResponse {
			return *httpserver.NewRestCreateResponseDyn(data.GetFieldData())
		},
		httpserver.JsonCreated,
	)
}

func (this ContactRest) DeleteContact(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST delete contact"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.contactSvc.DeleteContact,
		func(request DeleteContactRequest) it.DeleteContactCommand {
			return it.DeleteContactCommand(request)
		},
		func(data dyn.MutateResultData) DeleteContactResponse {
			return httpserver.NewRestDeleteResponse2(data)
		},
		httpserver.JsonOk,
	)
}

func (this ContactRest) GetContact(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST get contact"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.contactSvc.GetContact,
		func(request GetContactRequest) it.GetContactQuery {
			return it.GetContactQuery(request)
		},
		func(data domain.Contact) GetContactResponse {
			return data.GetFieldData()
		},
		httpserver.JsonOk,
	)
}

func (this ContactRest) ContactExists(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST contact exists"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.contactSvc.ContactExists,
		func(request ContactExistsRequest) it.ContactExistsQuery {
			return it.ContactExistsQuery(request)
		},
		func(data dyn.ExistsResultData) ContactExistsResponse {
			return ContactExistsResponse(data)
		},
		httpserver.JsonOk,
	)
}

func (this ContactRest) SearchContacts(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST search contacts"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.contactSvc.SearchContacts,
		func(request SearchContactsRequest) it.SearchContactsQuery {
			return it.SearchContactsQuery(request)
		},
		func(data it.SearchContactsResultData) SearchContactsResponse {
			return httpserver.NewSearchResponseDyn(data)
		},
		httpserver.JsonOk,
		true,
	)
}

func (this ContactRest) UpdateContact(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST update contact"); e != nil {
			err = e
		}
	}()

	return httpserver.ServeRequest2(
		echoCtx,
		this.contactSvc.UpdateContact,
		func(request UpdateContactRequest) it.UpdateContactCommand {
			cmd := it.UpdateContactCommand{}
			cmd.SetFieldData(request.DynamicFields)
			cmd.SetId(util.ToPtr(model.Id(request.ContactId)))
			return cmd
		},
		httpserver.NewRestMutateResponse,
		httpserver.JsonOk,
	)
}
