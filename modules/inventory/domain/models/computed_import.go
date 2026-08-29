package models

// The computed package registers the "computed" JSON-block parser at init time. Schemas here
// declare computed fields, so without this blank import ParseModelJson panics.
import (
	_ "github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
)
