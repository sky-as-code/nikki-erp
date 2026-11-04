package sequence

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

const (
	fingerprintVersion = "sequence-fingerprint-v1"
	// ServerDefaultMarker stands in for an effective_at the caller did not send, so that a retry relying on the
	// server default fingerprints the same as the first attempt.
	ServerDefaultMarker = "SERVER_DEFAULT"
)

// FingerprintInput is the canonical identity of one NextValue request (BR SEQ-IDEM-02). Trace ids, tokens and
// retry times are deliberately absent.
type FingerprintInput struct {
	DefinitionId  string
	TenantId      string
	OrgId         string
	OwnerModule   string
	CallerModule  string
	ScopeObjectId string
	OperationKey  string
	EffectiveAt   *time.Time
}

// Fingerprint returns the hex SHA-256 of a length-prefixed encoding of in, so no field value can shift into its
// neighbour. EffectiveAt is normalized to UTC so the same instant in another zone fingerprints the same.
func Fingerprint(in FingerprintInput) string {
	effectiveAt := ServerDefaultMarker
	if in.EffectiveAt != nil {
		effectiveAt = in.EffectiveAt.UTC().Format(time.RFC3339Nano)
	}
	fields := []string{
		fingerprintVersion, in.DefinitionId, in.TenantId, in.OrgId, in.OwnerModule,
		in.CallerModule, in.ScopeObjectId, in.OperationKey, effectiveAt,
	}
	hasher := sha256.New()
	for _, field := range fields {
		hasher.Write([]byte(strconv.Itoa(len(field))))
		hasher.Write([]byte{':'})
		hasher.Write([]byte(field))
		hasher.Write([]byte{';'})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
