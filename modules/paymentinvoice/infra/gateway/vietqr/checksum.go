package vietqr

import (
	"crypto/md5"
	"encoding/hex"
)

// VietQR authenticates a request body with an MD5 checksum over a fixed concatenation of values —
// no separators, no key ordering, just the fields run together in the order the gateway specifies.
//
// MD5 is broken for any purpose that needs collision resistance, and this construction is not an
// HMAC: the secret is simply prefixed to the message. Neither is a choice made here. It is what
// the gateway implements, and a checksum computed any other way is rejected. What follows from
// that is operational rather than cryptographic: the secret key is the only thing standing behind
// these checksums, so it must never be logged or written to a committed file.
//
// The two concatenations below are transcribed from the service this module supersedes
// (src/modules/vietqr/vietqr.service.ts). The order of the parts is load-bearing and must not be
// rearranged into something that reads more naturally.

// refundChecksum is md5(secretKey + referenceNumber + amount + bankNumber).
//
// The amount is the decimal string the request carries, not a number: rendering it differently —
// "50000.00" for "50000" — changes the checksum and the refund is refused.
func refundChecksum(secretKey string, referenceNumber string, amount string, bankNumber string) string {
	return md5Hex(secretKey + referenceNumber + amount + bankNumber)
}

// checkOrderChecksum is md5(bankNumber + username).
//
// Note it does not involve the secret key or the order being asked about: it identifies the
// merchant, not the request. That is the gateway's design, and it is why the check-order call is
// no more privileged than the bearer token already makes it.
func checkOrderChecksum(bankNumber string, username string) string {
	return md5Hex(bankNumber + username)
}

func md5Hex(data string) string {
	sum := md5.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
}
