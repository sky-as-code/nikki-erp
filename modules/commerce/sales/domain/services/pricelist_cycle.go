package services

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Cycle detection for FORMULA rules that price against another list.
//
// A rule whose base_price_source is OTHER_PRICELIST reads its base from a second list, whose rules
// may do the same, so the derivations form a graph that may loop. Resolving a price inside a loop
// never terminates, so the loop is refused when the rule is written rather than discovered as a hung
// request when a price is resolved.

// MaxPricelistDerivationDepth bounds the walk, keeping a single price resolution bounded whatever
// the data says. It is a second line of defence: the cycle check below already terminates on any
// loop, and this catches a chain that is merely absurd.
const MaxPricelistDerivationDepth = 16

// PricelistBaseReader answers which list a given list derives from, returning the empty string when
// a list derives from nothing, which is what ends the walk normally.
type PricelistBaseReader interface {
	BasePricelistOf(pricelistId string) (string, error)
}

// AssertNoPricelistCycle refuses a rule that would make pricelist derivation circular. It walks
// forward from the proposed base — the rule does not exist yet, so the graph must be traversed as it
// will be once it does — and arriving back at the rule's own list means the rule closes a loop.
func AssertNoPricelistCycle(
	pricelistId string, proposedBaseId string, reader PricelistBaseReader,
) error {
	if proposedBaseId == "" {
		return nil // Derives from nothing. No edge, no cycle.
	}
	if proposedBaseId == pricelistId {
		return errors.New("a pricelist cannot derive its prices from itself")
	}

	// Visited guards against a loop that does not include the starting list — B → C → B, reached
	// while checking A → B — which would otherwise spin until the depth bound and be reported as a
	// depth error.
	visited := map[string]bool{pricelistId: true, proposedBaseId: true}
	current := proposedBaseId

	for depth := 0; depth < MaxPricelistDerivationDepth; depth++ {
		next, err := reader.BasePricelistOf(current)
		if err != nil {
			return err
		}
		if next == "" {
			return nil // Reached a list that derives from nothing. The chain terminates.
		}
		if next == pricelistId {
			return errors.New(
				"this pricelist is already reachable from the one it would derive from, " +
					"so pricing would never resolve")
		}
		if visited[next] {
			return errors.New(
				"the pricelists this one would derive from are circular among themselves")
		}
		visited[next] = true
		current = next
	}

	return errors.Errorf(
		"pricelist derivation is more than %d lists deep; shorten the chain",
		MaxPricelistDerivationDepth)
}

// repoPricelistBaseReader answers BasePricelistOf from the pricelist item table. A list derives from
// another when any of its FORMULA rules names one as a base; the first found is enough, since the
// walk is looking for reachability.
type repoPricelistBaseReader struct {
	ctx  corectx.Context
	repo models.SalesSearcher
}

func newRepoPricelistBaseReader(ctx corectx.Context, repo models.SalesSearcher) *repoPricelistBaseReader {
	return &repoPricelistBaseReader{ctx: ctx, repo: repo}
}

func (this *repoPricelistBaseReader) BasePricelistOf(pricelistId string) (string, error) {
	found, err := models.FindPricelistBaseOf(this.ctx, this.repo, pricelistId)
	if err != nil {
		return "", err
	}
	if len(found) == 0 {
		return "", nil
	}
	return stringOf(found[0], models.SalesPricelistItemFieldBasePricelistId), nil
}
