// Package itstorage owns every uploaded file: the storage space tree, the file records, and the
// objects in S3. Files are served only through its APIs, never through presigned URLs, so every
// read passes a permission check.
//
// Other modules keep a `<name>_file_id` (metadata.file_field) and drive the file's lifecycle
// through the core/storage.StorageConsumer port, which this module implements and apptrait
// registers. They never import this module for it, so itstorage must never depend on a module
// that stores such an id (iam, inventory, vendingmachine).
package itstorage

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/app"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/dynamicengines"
	external "github.com/sky-as-code/nikki-erp/modules/itstorage/infra/external"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/transport"
)

// ModuleSingleton is typed DynamicModule so that dropping RegisterModels fails the build.
var ModuleSingleton modules.DynamicModule = &ItStorageModule{}

type ItStorageModule struct {
}

func (*ItStorageModule) LabelKey() string {
	return "itstorage.moduleLabel"
}

func (*ItStorageModule) Name() string {
	return c.ItStorageModuleName
}

func (*ItStorageModule) ModelPrefix() string {
	return c.ItStorageModuleName
}

func (*ItStorageModule) HasUi() bool {
	return false
}

func (*ItStorageModule) Category() modules.ModuleCategory {
	return modules.ModuleCategoryItInfra
}

// Deps implements InCodeModule. iam is deliberately absent and must stay so: iam stores avatar
// paths and depends on itstorage, so the reverse edge would close a cycle and abort startup.
func (*ItStorageModule) Deps() []string {
	return []string{
		"dynamicresource",
		// The upload reads the tenant's storage quota from Settings.
		"settings",
	}
}

func (*ItStorageModule) IsInternal() bool {
	return false
}

func (*ItStorageModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init builds the engines before anything binds to them, and transport last.
func (*ItStorageModule) Init() error {
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	return stdErr.Join(
		external.InitExternalServices(),
		app.InitApplicationServices(),
		transport.InitTransport(),
	)
}

// RegisterModels registers the space first: the file's edge points at it.
func (*ItStorageModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(models.SpaceSchemaBuilder()),
		dmodel.RegisterSchemaB(models.FileSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ConsumptionSchemaBuilder()),
	)
}
