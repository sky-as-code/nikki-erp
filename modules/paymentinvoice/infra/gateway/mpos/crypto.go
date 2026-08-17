package mpos

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"

	"go.bryk.io/pkg/errors"
)

// mPOS encrypts every request body with AES-128 in ECB mode under the merchant secret, PKCS#7
// padded, base64 encoded, and decrypts its replies the same way.
//
// ECB is a weak mode: identical plaintext blocks encrypt to identical ciphertext blocks, so it
// leaks structure. That is not a choice made here — it is what the gateway implements, and the
// payloads have to match byte for byte or nothing is accepted. The secret is therefore a
// credential like any other and must never be logged.
//
// Go's standard library deliberately ships no ECB mode for exactly the reason above, so the block
// loop is written out below. This is the whole of it: split into 16-byte blocks and encrypt each
// independently, with no IV and no chaining. Do not "improve" it to CBC; that would silently stop
// mPOS from being able to read anything we send.

// SecretKeyLength is how many bytes the merchant secret must be. It is used directly as the
// AES-128 key, so a secret of any other length is refused when the adapter is built rather than
// on the first payment, where it would present as an opaque cipher failure.
const SecretKeyLength = 16

// encrypt renders a payload as JSON, encrypts it and returns the base64 form mPOS expects.
func encrypt(payload any, secretKey string) (string, error) {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return "", errors.Wrap(err, "mpos payload could not be encoded")
	}

	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		// AES-128 needs a 16-byte key. A secret of any other length is a misconfiguration, and
		// saying so plainly here beats every request failing with a decrypt error at the gateway.
		return "", errors.Wrap(err, "mpos secret key is not a valid AES key (it must be 16 bytes)")
	}

	padded := padPkcs7(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(padded))
	for start := 0; start < len(padded); start += block.BlockSize() {
		block.Encrypt(ciphertext[start:start+block.BlockSize()], padded[start:start+block.BlockSize()])
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt reverses encrypt and decodes the JSON into out.
func decrypt(encoded string, secretKey string, out any) error {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return errors.Wrap(err, "mpos payload is not valid base64")
	}

	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return errors.Wrap(err, "mpos secret key is not a valid AES key (it must be 16 bytes)")
	}

	if len(ciphertext) == 0 || len(ciphertext)%block.BlockSize() != 0 {
		return errors.New("mpos payload is not a whole number of cipher blocks")
	}

	plaintext := make([]byte, len(ciphertext))
	for start := 0; start < len(ciphertext); start += block.BlockSize() {
		block.Decrypt(plaintext[start:start+block.BlockSize()], ciphertext[start:start+block.BlockSize()])
	}

	unpadded, err := unpadPkcs7(plaintext, block.BlockSize())
	if err != nil {
		return err
	}

	return errors.Wrap(json.Unmarshal(unpadded, out), "mpos payload could not be decoded")
}

// padPkcs7 appends n bytes of value n, where n is what it takes to reach the next block boundary.
// A plaintext that already ends on a boundary gains a whole block of padding, which is what makes
// the padding unambiguous to strip.
func padPkcs7(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

// unpadPkcs7 removes the padding, rejecting anything that does not look like it.
//
// The checks matter beyond tidiness: this runs on the webhook path, where the ciphertext arrives
// from outside. Decrypting attacker-chosen bytes yields plaintext whose final byte could claim
// any padding length, and trusting it would slice outside the buffer.
func unpadPkcs7(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("mpos payload decrypted to nothing")
	}

	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, errors.New("mpos payload has invalid padding")
	}

	for _, b := range data[len(data)-padding:] {
		if int(b) != padding {
			return nil, errors.New("mpos payload has invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}
