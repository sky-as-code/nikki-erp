// Package composable is the successor of package engine. It offers the same dynamic-resource
// capabilities (schema-driven CRUD, org scoping, permission, computed fields, the built-in REST
// route table and its param binders) but split into the onion layers a feature module already
// has: a repository, a domain service, an application service and a REST handler, each with a
// default implementation a module embeds and extends.
//
// The split is what lets a resource expose an action as REST plus service API, or as service
// API only: a custom action is a method on the module's application service, and a REST route
// for it is an optional AddRoute call. Package engine forced every action through one
// DynamicActionDefinition that carried both concerns.
//
// Nothing here imports package engine or package interfaces: the code was duplicated on
// purpose so that engine can be decommissioned once every module has migrated.
package composable
