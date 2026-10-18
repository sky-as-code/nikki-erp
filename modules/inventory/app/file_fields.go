package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable/filefield"
)

type fileBackedApplicationService = filefield.ApplicationService

var newFileBackedApplicationService = filefield.NewApplicationService
