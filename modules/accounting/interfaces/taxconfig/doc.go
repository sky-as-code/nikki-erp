// Package taxconfig declares the layers of the thirteen tax configuration resources Accounting
// serves through the composable resource engine at /v1/accounting/{schema_name}: a repository, a
// domain service carrying the lifecycle and integrity rules, and the authorized application service.
//
// The calculation port other modules consume stays in interfaces/tax.
package taxconfig
