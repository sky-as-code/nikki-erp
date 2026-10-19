// Package composable is the dynamic-resource engine: schema-driven CRUD, org scoping,
// permission, computed fields, the built-in REST route table and its param binders, split into
// the onion layers a feature module already has: a repository, a domain service, an
// application service and a REST handler, each with a default implementation a module embeds
// and extends.
//
// The split is what lets a resource expose an action as REST plus service API, or as service
// API only: a custom action is a method on the module's application service, and a REST route
// for it is an optional AddRoute call.
package composable
