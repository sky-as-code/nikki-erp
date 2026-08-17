package momo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// MoMo signs a request by concatenating a fixed set of fields as "key=value", joined with "&", in
// **alphabetical order of the key**, and taking the HMAC-SHA256 of that string under the merchant
// secret, hex-encoded lower case.
//
// Every detail here is MoMo's, not ours, and none of it may be tidied: the field set differs per
// operation and is exact — a field too many or too few produces a signature MoMo rejects, and the
// error it returns says only that the signature is wrong. The ordering is alphabetical by key
// rather than the order the fields are documented in, which is why this builds a map and sorts it
// rather than writing the string out by hand.
//
// The four field sets below are transcribed from the service this module supersedes
// (src/modules/momo/momo.service.ts), which is the only description of them we have that is known
// to work against MoMo in production.

// signingFields is one operation's field set, as key → value already rendered to a string.
type signingFields map[string]string

// rawSignature renders the fields into MoMo's canonical "k=v&k=v" form, keys ascending.
func (this signingFields) rawSignature() string {
	keys := make([]string, 0, len(this))
	for key := range this {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+this[key])
	}
	return strings.Join(parts, "&")
}

// sign returns the hex-encoded HMAC-SHA256 of the canonical form.
func (this signingFields) sign(secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(this.rawSignature()))
	return hex.EncodeToString(mac.Sum(nil))
}

// verify reports whether an expected signature matches, comparing in constant time.
//
// MoMo sends the signature lower case, but the comparison folds case anyway: the service this
// replaces did so, and a callback that differed only in case would otherwise be discarded as
// forged and the payment left unsettled.
func (this signingFields) verify(secretKey string, expected string) bool {
	computed := this.sign(secretKey)
	return hmac.Equal([]byte(computed), []byte(strings.ToLower(expected)))
}
