package constants

// CommerceModuleName is the one module inventory, purchase and sales ship as. Each part keeps its
// own schema prefix, REST group and translation namespace; this name is what the loader, the usage
// check, the org lifecycle and the settings module key know it by.
const CommerceModuleName = "commerce"

// ModelPrefixes lists the schema prefixes Commerce owns, one per part.
var ModelPrefixes = []string{"inventory", "purchase", "sales"}
