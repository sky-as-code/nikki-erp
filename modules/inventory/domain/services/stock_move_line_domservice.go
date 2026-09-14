package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// NewStockMoveLineDomainService closes the move line's write surface. A move line is an
// allocation the reservation engine creates and validate stamps; a client-written line would be a
// claim on a balance the balance does not know about, leaving the two irreconcilable. Editing an
// allocation by hand needs the release-and-re-reserve flow, which the transfer's own reserve and
// unreserve actions provide.
//
// The actions are refused rather than removed, so a caller gets a 400 naming the reason instead
// of a 404 that reads as a wrong URL. The reservation engine itself writes lines through the
// repository, which this does not touch.
func NewStockMoveLineDomainService(base composable.CrudDomainService) itStock.StockMoveLineDomainService {
	return &StockMoveLineDomainServiceImpl{CrudDomainService: base}
}

type StockMoveLineDomainServiceImpl struct {
	composable.CrudDomainService
}
