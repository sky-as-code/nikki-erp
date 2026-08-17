package services

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// The identifiers an order is known by, ported from the NestJS service this module supersedes
// (src/common/utils/util-function.ts and src/modules/orders/orders.service.ts).
//
// Both are contracts with parties outside this codebase and neither may be reshaped to suit us:
// order_code is the key every gateway callback arrives under, and order_id is what is quoted to
// the ordering system and to support. Orders taken by the old service are still being reconciled,
// so the two formats have to keep agreeing with each other.

const (
	// orderCodeLength is fixed at 12: the date prefix plus random padding. Every gateway accepts
	// a reference of this length, which is why it was chosen and why it stays.
	orderCodeLength = 12

	// orderCodeAlphabet is upper-case letters and digits only. Lower case is excluded because
	// some gateways fold case in their callbacks, which would make two codes collide.
	orderCodeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// orderCodeMaxAttempts bounds the search for an unused code. The random part is 9 characters
	// of a 36-character alphabet, so a collision is already remote; retrying is for the day it
	// is not.
	orderCodeMaxAttempts = 5

	// serviceCodeLength is how many characters of the calling system's name go into order_id.
	serviceCodeLength = 4

	// defaultSourceCode names the vending machines, which were the only caller of the service
	// this module supersedes.
	defaultSourceCode = "VDMC"

	// unknownMethodCode is the method segment for a method with no three-letter code of its own.
	unknownMethodCode = "UNK"
)

// methodSimpleCodes is the three-character segment each payment method contributes to an order_id,
// copied verbatim from PaymentMethodCode in src/common/utils/util-function.ts.
//
// It is keyed by the payment method's code — the row's `code`, not its adapter_code — because
// that is what the old service keyed it by, and order_ids already issued have to keep parsing the
// same way. A method with no entry here gets UNK, exactly as before.
//
// The map still carries the six methods the old service knew but never had an adapter for
// (mbbank, zalopay, ...). They are kept so that a deployment adding one of those methods produces
// the same order_id the old service would have, rather than UNK.
var methodSimpleCodes = map[string]string{
	"momo":      "MOM",
	"vietqr":    "VQR",
	"mpos":      "MPS",
	"mbbank":    "MBB",
	"zalopay":   "ZLP",
	"vnpay":     "VNP",
	"shopeepay": "SPY",
	"applePay":  "APL",
	"googlePay": "GGL",
}

// MethodSimpleCode returns the three-character order_id segment for a payment method code.
func MethodSimpleCode(methodCode string) string {
	if code, exists := methodSimpleCodes[methodCode]; exists {
		return code
	}
	return unknownMethodCode
}

// EncodeDateToBase36 renders a date as the prefix of an order code.
//
// The format is the old service's and looks arbitrary because it is: a two-character base-36
// year, then a single base-36 month and a single base-36 day, upper-cased. Month and day are one
// character each precisely because base 36 has enough digits for both — December is "C" and the
// 31st is "V" — which is what keeps the prefix to three characters.
//
// The year is taken modulo 100, so 2026 encodes as "1Q". That wraps in 2100, which the old
// service also did; a code is a within-the-year identifier and the uniqueness check is what
// actually prevents a clash.
func EncodeDateToBase36(at time.Time) string {
	year := at.Year() % 100
	month := int(at.Month())
	day := at.Day()

	encoded := padStart(strconv.FormatInt(int64(year), 36), 2) +
		strconv.FormatInt(int64(month), 36) +
		strconv.FormatInt(int64(day), 36)
	return strings.ToUpper(encoded)
}

func padStart(value string, length int) string {
	if len(value) >= length {
		return value
	}
	return strings.Repeat("0", length-len(value)) + value
}

// randomCode returns n characters drawn from orderCodeAlphabet.
//
// It draws from crypto/rand rather than math/rand, which is a deliberate departure from the old
// service's Math.random(). An order code is quoted in a payment request and accepted back in a
// callback, so a predictable one lets someone guess an order that is not theirs. The format is
// unchanged; only where the bytes come from is.
func randomCode(length int) (string, error) {
	alphabetSize := big.NewInt(int64(len(orderCodeAlphabet)))
	builder := strings.Builder{}
	builder.Grow(length)

	for range length {
		index, err := rand.Int(rand.Reader, alphabetSize)
		if err != nil {
			return "", err
		}
		builder.WriteByte(orderCodeAlphabet[index.Int64()])
	}
	return builder.String(), nil
}

// ServiceCode renders the calling system's name as the first segment of an order_id.
//
// Diacritics are stripped rather than dropped, so that "Bán hàng" and "Ban hang" produce the same
// segment instead of two that differ invisibly. Short names are padded with zeros to keep every
// order_id the same length, which is what lets the segments be read back out by position.
func ServiceCode(service string) string {
	cleaned := stripDiacritics(service)
	cleaned = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return -1
	}, cleaned)

	if cleaned == "" {
		cleaned = defaultSourceCode
	}
	if len(cleaned) > serviceCodeLength {
		cleaned = cleaned[:serviceCodeLength]
	}
	for len(cleaned) < serviceCodeLength {
		cleaned += "0"
	}
	return strings.ToUpper(cleaned)
}

// stripDiacritics decomposes each rune and drops the combining marks, so "ê" becomes "e".
//
// Đ and đ are handled separately: they are not a base letter plus a mark, so decomposition leaves
// them untouched and they would be dropped by the letter filter above. The old service special-cased
// them for the same reason.
func stripDiacritics(value string) string {
	replaced := strings.NewReplacer("đ", "d", "Đ", "D").Replace(value)

	chain := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(chain, replaced)
	if err != nil {
		// Only a malformed encoding reaches here, and the un-normalized text still yields a
		// usable segment once the filter above has run over it.
		return replaced
	}
	return result
}

// BuildOrderId assembles the quoted identifier from its three parts.
//
// The parts are fixed-width by construction — 4 + 3 + 12 — so the result can be taken apart again
// by position, which is how support reads one.
func BuildOrderId(service string, methodCode string, orderCode string) string {
	return ServiceCode(service) + MethodSimpleCode(methodCode) + orderCode
}
