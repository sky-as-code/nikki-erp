package crud

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
)

type GetOneQuery interface {
	schema.SchemaGetter
	GetIncludeArchived() bool
	GetColumns() []string
}

type GetOneQueryBase struct {
	IncludeArchived bool     `json:"include_archived" query:"include_archived"`
	Columns         []string `json:"columns" query:"columns"`
}

// Implements GetOneQuery interface
func (this GetOneQueryBase) GetIncludeArchived() bool {
	return this.IncludeArchived
}

// Implements GetOneQuery interface
func (this GetOneQueryBase) GetColumns() []string {
	return this.Columns
}
