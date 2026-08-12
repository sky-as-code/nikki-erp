package constants

const EssentialModuleName = "essential"

// IAM resource codes served by this module. The dynamic resource engine asserts permissions
// using the schema name as the resource code, so these must equal the XSchemaName constants
// in domain/models verbatim.
const (
	ResourceEssentialUom    = "essential_uom"
	ResourceEssentialUomCat = "essential_uomcat"
)
