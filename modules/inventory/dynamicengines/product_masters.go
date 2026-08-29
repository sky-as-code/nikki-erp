package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The master and junction resources below are plain CRUD and define no actions of their own.
// Product Category is the exception, and lives in its own file with the rules it needs.

func productTypeEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductTypeSchemaName,
	}
}

func brandEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.BrandSchemaName,
	}
}

func productAttributeEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductAttributeSchemaName,
	}
}

func productAttributeValueEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductAttributeValueSchemaName,
	}
}

func productTemplateAttributeEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductTemplateAttributeSchemaName,
	}
}

func productTemplateAttributeValueEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductTemplateAttributeValueSchemaName,
	}
}

func productVariantAttributeValueEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.ProductVariantAttributeValueSchemaName,
	}
}
