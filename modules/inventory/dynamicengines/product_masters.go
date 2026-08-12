package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The master and junction resources below are plain CRUD: everything they need is already
// expressed by their schema, so they define no actions of their own. Product Category is the
// exception, and lives in its own file with the rules it needs.

func productTypeEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductTypeSchemaName,
		DefaultFields: []string{
			models.ProductTypeFieldCode,
			models.ProductTypeFieldName,
			models.ProductTypeFieldSupportsStock,
			models.ProductTypeFieldSupportsSale,
			models.ProductTypeFieldSupportsPurchase,
			models.ProductTypeFieldSupportsManufacturing,
		},
	}
}

func brandEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.BrandSchemaName,
		DefaultFields: []string{
			models.BrandFieldCode,
			models.BrandFieldName,
			models.BrandFieldWebsite,
		},
	}
}

func productAttributeEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductAttributeSchemaName,
		DefaultFields: []string{
			models.ProductAttributeFieldCode,
			models.ProductAttributeFieldName,
			models.ProductAttributeFieldDataType,
			models.ProductAttributeFieldVariantCreationMode,
			models.ProductAttributeFieldSequence,
		},
	}
}

func productAttributeValueEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductAttributeValueSchemaName,
		DefaultFields: []string{
			models.ProductAttributeValueFieldAttributeId,
			models.ProductAttributeValueFieldCode,
			models.ProductAttributeValueFieldName,
			models.ProductAttributeValueFieldSequence,
			models.ProductAttributeValueFieldPriceExtra,
		},
	}
}

func productTemplateAttributeEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductTemplateAttributeSchemaName,
		DefaultFields: []string{
			models.ProductTemplateAttributeFieldProductTemplateId,
			models.ProductTemplateAttributeFieldAttributeId,
			models.ProductTemplateAttributeFieldSequence,
		},
	}
}

func productTemplateAttributeValueEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductTemplateAttributeValueSchemaName,
		DefaultFields: []string{
			models.ProductTemplateAttributeValueFieldTemplateAttributeId,
			models.ProductTemplateAttributeValueFieldAttributeValueId,
			models.ProductTemplateAttributeValueFieldSequence,
		},
	}
}

func productVariantAttributeValueEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductVariantAttributeValueSchemaName,
		DefaultFields: []string{
			models.ProductVariantAttributeValueFieldProductVariantId,
			models.ProductVariantAttributeValueFieldTemplateAttributeValueId,
		},
	}
}
