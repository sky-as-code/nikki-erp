package services

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// order_code and order_id are contracts with parties outside this codebase: order_code is the key
// every gateway callback arrives under, and order_id is what the ordering system and support were
// given. Orders taken by the NestJS service this module supersedes are still being reconciled, so
// both formats have to keep agreeing with what that service produced.
//
// The expectations below are computed from the original TypeScript
// (src/common/utils/util-function.ts) over fixed inputs. Do not "fix" a failing one by pasting in
// whatever the Go now produces — that turns a contract test into a change detector.

func TestDateEncodingMatchesReference(t *testing.T) {
	// 17 August 2026: year 26 → "Q" padded to "0Q", month 8 → "8", day 17 → "H".
	assert.Equal(t, "0Q8H", EncodeDateToBase36(time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)))
}

// Month and day are one base-36 character each, which is the whole reason the prefix fits in four
// characters. December is "C" and the 31st is "V"; rendering either in base 10 would push the
// random part down to eight characters and change every code.
func TestTwoDigitMonthAndDayStayOneCharacter(t *testing.T) {
	encoded := EncodeDateToBase36(time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC))

	assert.Equal(t, "0QCV", encoded)
	assert.Len(t, encoded, 4)
}

// A year below 10 pads to two characters, so the prefix is the same width every day of every year.
func TestSingleDigitYearIsPadded(t *testing.T) {
	assert.Equal(t, "098H", EncodeDateToBase36(time.Date(2009, time.August, 17, 0, 0, 0, 0, time.UTC)))
}

func TestMethodCodesMatchReference(t *testing.T) {
	assert.Equal(t, "MOM", MethodSimpleCode("momo"))
	assert.Equal(t, "VQR", MethodSimpleCode("vietqr"))
	assert.Equal(t, "MPS", MethodSimpleCode("mpos"))
}

// A method with no three-letter code of its own falls back to UNK rather than to an empty segment,
// which would shorten every order_id it appears in and break reading them back by position.
func TestUnknownMethodFallsBackToUnk(t *testing.T) {
	assert.Equal(t, "UNK", MethodSimpleCode("some_new_wallet"))
	assert.Equal(t, "UNK", MethodSimpleCode(""))
}

// The six methods the old service knew but never had an adapter for are kept, so that a deployment
// adding one produces the order_id the old service would have rather than UNK.
func TestMethodsWithoutAdaptersKeepTheirCodes(t *testing.T) {
	assert.Equal(t, "MBB", MethodSimpleCode("mbbank"))
	assert.Equal(t, "ZLP", MethodSimpleCode("zalopay"))
	assert.Equal(t, "VNP", MethodSimpleCode("vnpay"))
}

func TestServiceCodeMatchesReference(t *testing.T) {
	assert.Equal(t, "VDMC", ServiceCode("VDMC"))
	assert.Equal(t, "VEND", ServiceCode("vending"))
}

// A short name is padded rather than left short, so every order_id is the same length and the
// three segments can be read back out by position.
func TestShortServiceNamesArePadded(t *testing.T) {
	assert.Equal(t, "POS0", ServiceCode("pos"))
	assert.Equal(t, "A000", ServiceCode("a"))
}

// An empty or punctuation-only name falls back to the default rather than producing "0000", which
// would name no system at all.
func TestEmptyServiceNameFallsBackToDefault(t *testing.T) {
	assert.Equal(t, defaultSourceCode, ServiceCode(""))
	assert.Equal(t, defaultSourceCode, ServiceCode("--- ---"))
}

// Diacritics are stripped rather than dropped, so "Bán hàng" and "Ban hang" produce one segment
// instead of two that differ invisibly.
func TestDiacriticsAreStrippedNotDropped(t *testing.T) {
	assert.Equal(t, "BANH", ServiceCode("Bán hàng"))
	assert.Equal(t, "DIEN", ServiceCode("Điện máy"))
}

func TestOrderIdIsTheThreeSegmentsInOrder(t *testing.T) {
	orderId := BuildOrderId("VDMC", "momo", "0Q8HABCDEFGH")

	assert.Equal(t, "VDMCMOM0Q8HABCDEFGH", orderId)
	assert.Equal(t, "VDMC", orderId[:4])
	assert.Equal(t, "MOM", orderId[4:7])
	assert.Equal(t, "0Q8HABCDEFGH", orderId[7:])
}

// Every order_id is 19 characters — 4 + 3 + 12 — which is what lets support read one apart by
// position, and what the order_id column's bounds were chosen for.
func TestOrderIdLengthIsFixed(t *testing.T) {
	assert.Len(t, BuildOrderId("VDMC", "momo", "0Q8HABCDEFGH"), 19)
	assert.Len(t, BuildOrderId("a", "unknown_method", "0Q8HABCDEFGH"), 19)
}

func TestRandomCodeUsesOnlyTheDeclaredAlphabet(t *testing.T) {
	allowed := regexp.MustCompile(`^[A-Z0-9]+$`)

	for range 50 {
		code, err := randomCode(9)
		require.NoError(t, err)

		assert.Len(t, code, 9)
		assert.Regexp(t, allowed, code,
			"lower case is excluded because some gateways fold case in their callbacks")
	}
}

// The random part is what stops two orders created in the same second from colliding, so it must
// actually vary.
func TestRandomCodeVaries(t *testing.T) {
	seen := map[string]bool{}

	for range 100 {
		code, err := randomCode(9)
		require.NoError(t, err)
		seen[code] = true
	}

	assert.Greater(t, len(seen), 90, "the random suffix is barely varying")
}
