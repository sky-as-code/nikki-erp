package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
)

var chars = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789") // no O, 0, I, 1 for clarity

// Generate OTP Recovery Code. Format as XXXX-XXXX-XXXX-XXXX
func GenerateRecoveryCode() (string, error) {
	code := make([]rune, 16)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		code[i] = chars[n.Int64()]
	}

	return fmt.Sprintf("%s-%s-%s-%s", string(code[0:4]), string(code[4:8]), string(code[8:12]), string(code[12:16])), nil
}

var recoveryCodeRegexp = regexp.MustCompile(`^[A-Z2-9]{4}-[A-Z2-9]{4}-[A-Z2-9]{4}-[A-Z2-9]{4}$`)

func IsRecoveryCodeFormat(s string) bool {
	return recoveryCodeRegexp.MatchString(s)
}
