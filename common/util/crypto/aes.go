package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"strings"
)

const EMPTY_STRING = ""

// GenerateSecret generates a cryptographically safe random string,
// used for password, salt, token etc.
func GenerateSecret(size uint) (string, error) {
	bytes := make([]byte, size)

	if _, err := rand.Read(bytes); err != nil {
		return EMPTY_STRING, err
	}

	return hex.EncodeToString(bytes), nil
}

// EncodeBase64 returns the hexadecimal encoding of src.
func EncodeBase64(src []byte) string {
	return hex.EncodeToString(src)
}

// DecodeBase64 returns the bytes represented by the hexadecimal string src.
// DecodeBase64 expects that src contains only hexadecimal characters and that src has even length.
// If the input is malformed, DecodeBase64 returns the bytes decoded before the error.
func DecodeBase64(src string) ([]byte, error) {
	return hex.DecodeString(src)
}

func EncryptString(value string, secret string) (string, error) {
	if len(value) == 0 {
		return value, nil
	}

	secretBytes, err := DecodeBase64(secret)
	if err != nil {
		return EMPTY_STRING, err
	}

	data := []byte(value)

	data, err = Encrypt(data, secretBytes)
	if err != nil {
		return EMPTY_STRING, err
	}

	return EncodeBase64(data), nil
}

func DecryptString(value string, secret string) (string, error) {
	if len(value) == 0 {
		return value, nil
	}

	secretBytes, err := DecodeBase64(secret)
	if err != nil {
		return EMPTY_STRING, err
	}

	data, err := DecodeBase64(value)
	if err != nil {
		return EMPTY_STRING, err
	}

	data, err = Decrypt(data, secretBytes)
	if err != nil {
		return EMPTY_STRING, err
	}

	return string(data), nil
}

func Encrypt(data []byte, secretBytes []byte) ([]byte, error) {
	// no data
	if len(data) == 0 {
		return data, nil
	}

	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(secretBytes)
	if err != nil {
		return nil, err
	}

	// Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	// Encrypt the data using aesGCM.Seal
	// nonce added as a prefix to the encrypted data
	cipherBytes := aesGCM.Seal(nonce, nonce, data, nil)
	return cipherBytes, nil
}

func Decrypt(data []byte, secretBytes []byte) ([]byte, error) {
	// no data
	if len(data) == 0 {
		return data, nil
	}

	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(secretBytes)
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce size
	nonceSize := aesGCM.NonceSize()

	// check data length
	if len(data) < nonceSize {
		return []byte{}, nil
	}

	// Extract the nonce from the encrypted data
	nonce, cipherBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt the data
	return aesGCM.Open(nil, nonce, cipherBytes, nil)
}

func EncryptObj(src interface{}, secret string) ([]byte, error) {
	d, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	secrectBytes, _ := DecodeBase64(secret)
	return Encrypt(d, secrectBytes)
}

func DecryptObj(src []byte, secret string, destPtr interface{}) error {
	secretBytes, _ := DecodeBase64(secret)

	src, err := Decrypt(src, secretBytes)
	if err != nil {
		return err
	}

	return json.Unmarshal(src, destPtr)
}

func ToRsaPrivateKey(privateKey string, passphrase ...string) (*rsa.PrivateKey, error) {
	privPassphrase := ""
	if len(passphrase) > 0 {
		privPassphrase = passphrase[0]
	}
	privBytes := []byte(privateKey)
	privPem, _ := pem.Decode(privBytes)

	if privPem.Type != "RSA PRIVATE KEY" {
		return nil, errors.New(fmt.Sprintf("RSA private key is of the wrong type: %s", privPem.Type))
	}

	if x509.IsEncryptedPEMBlock(privPem) && privPassphrase == "" {
		return nil, errors.New("Passphrase is required to open private pem file")
	}

	var err error
	if privPassphrase != "" {
		privBytes, err = x509.DecryptPEMBlock(privPem, []byte(privPassphrase))
	} else {
		privBytes = privPem.Bytes
	}

	var parsedKey interface{}
	//PKCS1
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privBytes); err != nil {
		//If what you are sitting on is a PKCS#8 encoded key
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privBytes); err != nil { // note this returns type `interface{}`
			return nil, errors.New("Unable to parse RSA private key")
		}
	}

	var result *rsa.PrivateKey
	var ok bool
	result, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("Unable to cast the result to RSA private key")
	}
	return result, nil
}

func ToRsaPublicKey(publicKey string) (*rsa.PublicKey, error) {
	pubBytes := []byte(publicKey)
	pubPem, _ := pem.Decode(pubBytes)

	if pubPem == nil {
		return nil, errors.New("RSA public key not in pem format")
	}

	if !strings.Contains(pubPem.Type, "PUBLIC KEY") {
		return nil, errors.New(fmt.Sprintf("RSA public key is of the wrong type: %s", pubPem.Type))
	}

	var parsedKey interface{}
	var err error
	if parsedKey, err = x509.ParsePKIXPublicKey(pubPem.Bytes); err != nil {
		return nil, errors.New("Unable to parse RSA public key")
	}

	var result *rsa.PublicKey
	var ok bool
	if result, ok = parsedKey.(*rsa.PublicKey); !ok {
		return nil, errors.New("Unable to parse RSA public key")
	}

	return result, nil
}
