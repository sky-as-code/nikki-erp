package services

import (
	"strings"
	"testing"

	"go.bryk.io/pkg/errors"
)

// chainReader answers BasePricelistOf from a fixed map; a missing key means the list derives from
// nothing, which is how a chain terminates.
type chainReader map[string]string

func (this chainReader) BasePricelistOf(pricelistId string) (string, error) {
	return this[pricelistId], nil
}

// failingReader stands in for a repository that cannot answer.
type failingReader struct{}

func (failingReader) BasePricelistOf(string) (string, error) {
	return "", errors.New("repository unavailable")
}

func TestNoBaseIsNotACycle(t *testing.T) {
	if err := AssertNoPricelistCycle("A", "", chainReader{}); err != nil {
		t.Fatalf("a rule with no base has no edge to make a cycle from: %v", err)
	}
}

func TestSelfReferenceIsRefused(t *testing.T) {
	if err := AssertNoPricelistCycle("A", "A", chainReader{}); err == nil {
		t.Fatal("a pricelist deriving from itself is the shortest possible cycle")
	}
}

// The ordinary case: a chain that ends.
func TestTerminatingChainIsAllowed(t *testing.T) {
	reader := chainReader{"B": "C"}

	if err := AssertNoPricelistCycle("A", "B", reader); err != nil {
		t.Fatalf("A -> B -> C -> nothing is a legitimate derivation: %v", err)
	}
}

// The case the check exists for: adding A -> B would close A -> B -> C -> A.
func TestCycleBackToTheStartingListIsRefused(t *testing.T) {
	reader := chainReader{"B": "C", "C": "A"}

	err := AssertNoPricelistCycle("A", "B", reader)
	if err == nil {
		t.Fatal("A -> B -> C -> A is a cycle and pricing it would never resolve")
	}
}

// A loop that does not include the list being edited. Without the visited set the walk would spin
// between B and C until the depth bound and blame the depth.
func TestCycleAmongTheBasesIsRefusedAsACycle(t *testing.T) {
	reader := chainReader{"B": "C", "C": "B"}

	err := AssertNoPricelistCycle("A", "B", reader)
	if err == nil {
		t.Fatal("B and C deriving from each other is still a cycle")
	}
	if !strings.Contains(err.Error(), "circular among themselves") {
		t.Fatalf("the error should name the loop among the bases, got: %s", err)
	}
}

// A chain that is merely absurdly long, with no loop in it, is refused by the depth bound.
func TestOverlyDeepChainIsRefused(t *testing.T) {
	reader := chainReader{}
	previous := "B"
	for index := 0; index < MaxPricelistDerivationDepth+5; index++ {
		next := "L" + string(rune('a'+index%26)) + string(rune('a'+index/26))
		reader[previous] = next
		previous = next
	}

	err := AssertNoPricelistCycle("A", "B", reader)
	if err == nil {
		t.Fatal("a chain deeper than the bound must be refused")
	}
	if !strings.Contains(err.Error(), "deep") {
		t.Fatalf("the error should name the depth as the problem, got: %s", err)
	}
}

// A repository failure must surface as itself, not be mistaken for "no base" and silently allow a
// cycle the walk never saw.
func TestReaderErrorIsPropagated(t *testing.T) {
	if err := AssertNoPricelistCycle("A", "B", failingReader{}); err == nil {
		t.Fatal("a repository failure must not read as a terminating chain")
	}
}
