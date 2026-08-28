package services

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Cycle detection for FORMULA rules that price against another list (section 14).
//
// A rule whose base_price_source is OTHER_PRICELIST reads its base from a second list, whose own
// rules may do the same. That is a graph, and a graph may contain a loop: A derives from B, B from
// C, C from A. Resolving a price in such a loop never terminates — there is no base to reach — and
// the failure surfaces as a hung request rather than as anything a user could act on.
//
// So the loop is refused when the rule is written rather than discovered when a price is resolved.
// The cost is one traversal per rule saved; the alternative is an outage nobody can diagnose from
// the order that triggered it.

// MaxPricelistDerivationDepth bounds the walk.
//
// A chain longer than this is not a legitimate pricing policy — it is a mistake or an attempt to
// make resolution expensive — and refusing it keeps a single price resolution bounded no matter
// what the data says. The bound is a second line of defence: the cycle check below already
// terminates on any loop, and this catches a chain that is merely absurd.
const MaxPricelistDerivationDepth = 16

// PricelistBaseReader answers which list a given list derives from, for the walk below.
//
// An interface rather than the repository, so the traversal can be tested without a database. It
// returns the empty string when a list derives from nothing, which is what ends the walk normally.
type PricelistBaseReader interface {
	BasePricelistOf(pricelistId string) (string, error)
}

// AssertNoPricelistCycle refuses a rule that would make pricelist derivation circular.
//
// It walks forward from the proposed base: if the walk ever arrives back at the list the rule
// belongs to, adding the rule would close a loop. Walking forward rather than backward matters —
// the rule does not exist yet, so the graph must be traversed as it will be once it does.
//
// Self-reference is the degenerate case and is checked first, because a list deriving from itself
// reads as a loop of length zero and would otherwise need the walk to notice it.
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
	// while checking A → B. Without it the walk would spin between them until the depth bound,
	// reporting a depth error for what is really a cycle.
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

// repoPricelistBaseReader answers BasePricelistOf from the pricelist item table.
//
// A list "derives from" another when any of its FORMULA rules names one as a base. Several rules
// may name several different bases and the first found is enough: the walk is looking for
// reachability, and a cycle through any one of them is still a cycle.
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
