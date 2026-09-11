// Package uomcat declares the UoM Category resource's layers. The category is reference data
// with one rule of its own: its Reference UoM must belong to it (BR-UOM-ESS-004).
package uomcat

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// UomCatRepository reads and writes UoM Category rows.
type UomCatRepository interface {
	composable.CrudRepository
}

// UomCatDomainService is the CRUD of the resource with the reference-UoM rules enforced on
// create and update.
type UomCatDomainService interface {
	composable.CrudDomainService
}

// UomCatApplicationService is the authorized CRUD the REST handler serves.
type UomCatApplicationService interface {
	composable.CrudApplicationService
}
