package models

// The computed package registers the "computed" JSON-block parser into the dynamic-model JSON
// builder at its init time (see common/dynamicmodel/model/computed_hooks.go). Schemas in this
// package declare computed fields, so the package must be linked in — without this import,
// ParseModelJson panics with "the computed package is not linked in".
import (
	_ "github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
)
