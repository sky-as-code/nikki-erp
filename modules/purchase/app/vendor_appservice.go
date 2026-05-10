package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/vendor"
)

func NewVendorApplicationServiceImpl(vendorSvc it.VendorDomainService) it.VendorAppService {
	return &VendorApplicationServiceImpl{vendorSvc: vendorSvc}
}

type VendorApplicationServiceImpl struct {
	vendorSvc it.VendorDomainService
}

func (this *VendorApplicationServiceImpl) CreateVendor(ctx corectx.Context, cmd it.CreateVendorCommand) (*it.CreateVendorResult, error) {
	return this.vendorSvc.CreateVendor(ctx, cmd)
}

func (this *VendorApplicationServiceImpl) DeleteVendor(ctx corectx.Context, cmd it.DeleteVendorCommand) (*it.DeleteVendorResult, error) {
	return this.vendorSvc.DeleteVendor(ctx, cmd)
}

func (this *VendorApplicationServiceImpl) VendorExists(ctx corectx.Context, query it.VendorExistsQuery) (*it.VendorExistsResult, error) {
	return this.vendorSvc.VendorExists(ctx, query)
}

func (this *VendorApplicationServiceImpl) GetVendor(ctx corectx.Context, query it.GetVendorQuery) (*it.GetVendorResult, error) {
	return this.vendorSvc.GetVendor(ctx, query)
}

func (this *VendorApplicationServiceImpl) SearchVendors(ctx corectx.Context, query it.SearchVendorsQuery) (*it.SearchVendorsResult, error) {
	return this.vendorSvc.SearchVendors(ctx, query)
}

func (this *VendorApplicationServiceImpl) SetVendorIsArchived(ctx corectx.Context, cmd it.SetVendorIsArchivedCommand) (*it.SetVendorIsArchivedResult, error) {
	return this.vendorSvc.SetVendorIsArchived(ctx, cmd)
}

func (this *VendorApplicationServiceImpl) UpdateVendor(ctx corectx.Context, cmd it.UpdateVendorCommand) (*it.UpdateVendorResult, error) {
	return this.vendorSvc.UpdateVendor(ctx, cmd)
}
