package util

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func GetUUID() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

func GetStreamUUID() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	u[14] = 'C'
	u[15] = 'D'
	sha := sha256.Sum256(u[:16])
	u[14] = sha[11]
	u[15] = sha[13]

	return u.String(), nil
}

func IsValidStreamUUID(id string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	if !r.MatchString(id) {
		return false
	}

	u, err := uuid.Parse(id)
	if err != nil {
		return false
	}

	valid := make([]byte, 2)
	copy(valid, u[14:])
	u[14] = 'C'
	u[15] = 'D'
	sha := sha256.Sum256(u[:16])

	if sha[11] != valid[0] || sha[13] != valid[1] {
		return false
	}

	return true
}

const HASH_KEY = "CDKCloud."
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const alphaNumBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringLetters(n int) string {
	b := make([]byte, n)
	max := len(letterBytes)
	for i := range b {
		b[i] = letterBytes[rand.Intn(max)]
	}
	return string(b)
}

func RandStringAlphaNums(n int) string {
	b := make([]byte, n)
	max := len(alphaNumBytes)
	for i := range b {
		b[i] = alphaNumBytes[rand.Intn(max)]
	}
	return string(b)
}

/*
func GetBinaryBySHA256WithKey(msg, key string) ([]byte, error) {
	mac := hmac.New(sha256.New, GetBinaryBySHA256(key))
	_, err := mac.Write([]byte(msg))
	return mac.Sum(nil), err
}
func GetBinaryBySHA256(s string) []byte {
	r := sha256.Sum256([]byte(s))
	return r[:]
}
*/

func GetStringSignature(src string, key string, length int) string {

	sha := sha256.Sum256([]byte(src + key + HASH_KEY))
	dst := make([]byte, hex.EncodedLen(len(sha)))
	hex.Encode(dst, sha[:])
	sig := dst[:length]

	return string(sig)
}

/*
func CheckStringSignature(source string, key string, length int) bool {

	if len(source) < 6 {
		return false
	}

	src := source[:len(source)-length]
	sig := source[len(source)-length:]

	return GetStringSignature(src, key, length) == sig
}

func SignString(src string, key string) string {

	return src + GetStringSignature(src, key, 4)
}
*/

func IsValidURL(urlstr string) bool {

	// Don't allow long URL
	if len(urlstr) > 256 {
		return false
	}

	// Don't allow ? or space.
	if strings.Contains(urlstr, "?") || strings.Contains(urlstr, " ") {
		return false
	}
	_, err := url.ParseRequestURI(urlstr)
	return err == nil
}

func IsValidID(id string) bool {
	r := regexp.MustCompile(`^[a-zA-Z0-9_.\-]*$`)
	return r.MatchString(id) && len(id) < 50
}

func IsValidRole(role string) bool {
	r := regexp.MustCompile(`^[a-zA-Z0-9_.:\-]*$`)
	return r.MatchString(role) && len(role) < 100
}
