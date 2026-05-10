package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/vendor"
)

func NewVendorDomainServiceImpl(repo it.VendorRepository) it.VendorDomainService {
	return &VendorDomainServiceImpl{repo: repo}
}

type VendorDomainServiceImpl struct{ repo it.VendorRepository }

func (this *VendorDomainServiceImpl) CreateVendor(ctx corectx.Context, cmd it.CreateVendorCommand) (*it.CreateVendorResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.Vendor, *models.Vendor]{
		Action: "create vendor", BaseRepoGetter: this.repo, Data: cmd,
	})
}
func (this *VendorDomainServiceImpl) DeleteVendor(ctx corectx.Context, cmd it.DeleteVendorCommand) (*it.DeleteVendorResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete vendor", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}
func (this *VendorDomainServiceImpl) VendorExists(ctx corectx.Context, query it.VendorExistsQuery) (*it.VendorExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if vendors exist", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}
func (this *VendorDomainServiceImpl) GetVendor(ctx corectx.Context, query it.GetVendorQuery) (*it.GetVendorResult, error) {
	return corecrud.GetOne[models.Vendor](ctx, corecrud.GetOneParam{Action: "get vendor", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}
func (this *VendorDomainServiceImpl) SearchVendors(ctx corectx.Context, query it.SearchVendorsQuery) (*it.SearchVendorsResult, error) {
	return corecrud.Search[models.Vendor](ctx, corecrud.SearchParam{Action: "search vendors", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}
func (this *VendorDomainServiceImpl) SetVendorIsArchived(ctx corectx.Context, cmd it.SetVendorIsArchivedCommand) (*it.SetVendorIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
func (this *VendorDomainServiceImpl) UpdateVendor(ctx corectx.Context, cmd it.UpdateVendorCommand) (*it.UpdateVendorResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.Vendor, *models.Vendor]{
		Action: "update vendor", DbRepoGetter: this.repo, Data: cmd,
	})
}
