// Package external binds Sales' local ports to the services other modules publish.
//
// This is the ONLY package in Sales that may import another module. Everything else depends on
// the interfaces in interfaces/external, so splitting a module into its own process changes this
// file and nothing else — the bindings become REST or CQRS clients and every caller is unaffected.
package external

// InitExternal binds every port Sales consumes, and registers what Sales offers back.
//
// It binds nothing yet. The ports Sales needs — the payment method reader onto paymentinvoice, the
// fulfilment port onto inventory, UoM and currency onto essential, the party read onto contacts —
// are deliberately not guessed here: SALES-002 surveys those contracts first, because three
// planning decisions (D-35, D-38, and the shape of the fulfilment port) rest on what those modules
// actually expose rather than on what their folder structure suggests.
//
// The function exists now so that Init()'s ordering is fixed from the start: ports bind before
// engines are created, because a derived service resolves its ports at construction time. Adding
// the first binding then changes one line here instead of reordering Init.
func InitExternal() error {
	return nil
}
