package services

import (
	"testing"

	"github.com/shopspring/decimal"
)

// The guard of the estimate: it is recorded, compared and never charged. A divergence is a
// diagnostic signal about a client's pricing, so switching the check on must not start refusing
// sales, and leaving it off must not start emitting noise.

func amount(value string) decimal.Decimal {
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		panic(err)
	}
	return parsed
}

// The default. An organization that configured no tolerance has not asked to be told, and every
// order would otherwise diverge by rounding alone.
func TestNoToleranceMeansNoCheck(t *testing.T) {
	estimated := amount("15000")
	if exceedsTolerance(&estimated, amount("14500"), decimal.Zero) {
		t.Fatal("a zero tolerance must switch the check off, not make every order divergent")
	}
}

// A client that sent nothing has made no claim, so there is nothing to disagree with.
func TestAMissingEstimateIsNotADivergence(t *testing.T) {
	if exceedsTolerance(nil, amount("14500"), amount("100")) {
		t.Fatal("an absent estimate must not be read as a claim that the price was zero")
	}
}

// The CR's own example: client 15000, Sales 14500. Within tolerance it is silence, beyond it a
// signal — and either way the sale stands, which is why this returns a bool and never an error.
func TestDivergenceIsReportedOnlyBeyondTheTolerance(t *testing.T) {
	estimated := amount("15000")

	if exceedsTolerance(&estimated, amount("14500"), amount("500")) {
		t.Fatal("a difference equal to the tolerance is within it and must stay quiet")
	}
	if !exceedsTolerance(&estimated, amount("14500"), amount("499")) {
		t.Fatal("a difference beyond the tolerance must be reported")
	}
}

// Divergence is symmetric: a client quoting too little is as wrong as one quoting too much, and only
// the absolute difference decides.
func TestDivergenceIgnoresWhichSideIsHigher(t *testing.T) {
	low := amount("14000")
	high := amount("15000")

	if !exceedsTolerance(&low, amount("15000"), amount("500")) {
		t.Fatal("an estimate below the charged price must diverge too")
	}
	if !exceedsTolerance(&high, amount("14000"), amount("500")) {
		t.Fatal("an estimate above the charged price must diverge too")
	}
}
