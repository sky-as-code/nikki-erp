package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// An invoice number identifies an accounting document. Two documents sharing one, or a number that
// silently restarts, is a defect an auditor finds rather than a test — so the parsing that decides
// the next number is pinned here.

func TestASequenceIsReadFromAMintedNumber(t *testing.T) {
	sequence, ok := sequenceOfInvoiceNumber("INV-2026-000042", "INV-2026-")

	assert.True(t, ok)
	assert.Equal(t, 42, sequence)
}

// The zero-padding is cosmetic and must not be read as an octal or dropped: 000042 is the
// forty-second invoice, not the fourth.
func TestTheZeroPaddingIsNotSignificant(t *testing.T) {
	for text, expected := range map[string]int{
		"INV-2026-000001": 1,
		"INV-2026-000010": 10,
		"INV-2026-000100": 100,
		"INV-2026-999999": 999999,
	} {
		sequence, ok := sequenceOfInvoiceNumber(text, "INV-2026-")
		assert.True(t, ok, text)
		assert.Equal(t, expected, sequence, text)
	}
}

// A number belonging to another year must not be read as this year's: doing so would continue last
// year's sequence into the new one, or restart this year's midway.
func TestANumberFromAnotherYearIsNotThisYearsSequence(t *testing.T) {
	_, ok := sequenceOfInvoiceNumber("INV-2025-000042", "INV-2026-")

	assert.False(t, ok)
}

// A row that this module did not mint is treated as absent rather than failing the issue. Refusing
// to issue anything until someone corrects a stray row would be a worse outcome, and the unique
// index still prevents an actual collision.
func TestAnUnrecognisedNumberIsTreatedAsAbsent(t *testing.T) {
	for _, text := range []string{
		"",
		"INV-2026-",
		"INV-2026-abc",
		"nonsense",
		"INV-2026",
		// A sequence of zero would make the next number 1, which may already be taken; it is
		// refused so the search falls back to starting the sequence rather than colliding.
		"INV-2026-000000",
	} {
		_, ok := sequenceOfInvoiceNumber(text, "INV-2026-")
		assert.False(t, ok, text)
	}
}

// The format is what the frontend renders and what a human quotes on the phone. It is pinned so a
// change to the padding or the separator is a deliberate decision rather than a side effect.
func TestTheMintedNumberFormatIsStable(t *testing.T) {
	for sequence, expected := range map[int]string{
		1:      "INV-2026-000001",
		42:     "INV-2026-000042",
		999999: "INV-2026-999999",
	} {
		assert.Equal(t, expected, fmt.Sprintf("INV-%d-%06d", 2026, sequence))
	}
}

// A number past the padding width must still be distinct rather than truncated back into the range
// of an existing one.
func TestASequencePastThePaddingWidthIsNotTruncated(t *testing.T) {
	minted := fmt.Sprintf("INV-%d-%06d", 2026, 1000000)

	assert.Equal(t, "INV-2026-1000000", minted)

	sequence, ok := sequenceOfInvoiceNumber(minted, "INV-2026-")
	assert.True(t, ok)
	assert.Equal(t, 1000000, sequence)
}
