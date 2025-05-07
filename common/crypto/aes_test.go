package crypto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//nolint

const SECRET_LENGTH = 32

var value = "3h0quYFDEscxF1FQFWTFgQ=="

func TestAes(t *testing.T) {
	secret, err := GenerateSecret(SECRET_LENGTH)
	t.Logf("Secrect: %s", secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, secret)

	data := []byte(value)

	secretBytes, err := DecodeBase64(secret)
	assert.NoError(t, err)

	cipherBytes, err := Encrypt(data, secretBytes)
	assert.NoError(t, err)
	assert.NotEmpty(t, cipherBytes)

	data, err = Decrypt(cipherBytes, secretBytes)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	assert.Equal(t, value, string(data))
}

func TestAesString(t *testing.T) {
	secret, err := GenerateSecret(SECRET_LENGTH)
	assert.NoError(t, err)
	assert.NotEmpty(t, secret)

	cipherValue, err := EncryptString(value, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, cipherValue)

	valueOut, err := DecryptString(cipherValue, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, valueOut)

	assert.Equal(t, value, valueOut)

	t.Logf("Value: %s - Cipher: %s", valueOut, cipherValue)
}

// func TestAesParams(t *testing.T) {
// 	secret, err := GenerateSecret()
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, secret)
// 	println(secret)

// 	params := packet.Params{"password": value}
// 	data, err := params.MarshalBinary()
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, data)

// 	secretBytes, err := ToBytes(secret)
// 	assert.NoError(t, err)

// 	cipherBytes, err := Encrypt(data, secretBytes)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, cipherBytes)

// 	data, err = Decrypt(cipherBytes, secretBytes)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, data)

// 	err = params.UnmarshalBinary(data)
// 	assert.NoError(t, err)
// 	assert.Equal(t, value, params.String("password"))
// }

// func BenchmarkAesEncrypt(b *testing.B) {
// 	secret, _ := GenerateSecret()
// 	secretBytes, _ := ToBytes(secret)

// 	params := packet.Params{"password": value}
// 	data, _ := params.MarshalBinary()

// 	for i := 0; i < b.N; i++ {
// 		Encrypt(data, secretBytes)
// 	}
// }

// func BenchmarkAesDecrypt(b *testing.B) {
// 	secret, _ := GenerateSecret()
// 	secretBytes, _ := ToBytes(secret)

// 	params := packet.Params{"password": secretBytes}
// 	data, _ := params.MarshalBinary()
// 	cipherData, _ := Encrypt(data, secretBytes)

// 	for i := 0; i < b.N; i++ {
// 		Decrypt(cipherData, secretBytes)
// 	}
// }

func TestEncrypt(t *testing.T) {
	secrect, _ := GenerateSecret(SECRET_LENGTH)
	password := "3" // .api,webhook,notification"
	expiry := time.Now().Add(10).UTC().Round(time.Hour * 24)

	passwordHash, _ := EncryptString(password, secrect)
	decrypt, _ := DecryptString(passwordHash, secrect)

	t.Logf("Password hash: %s:  %s", string(password), string(passwordHash))
	t.Logf("Decrypt: %s - Expiry: %s - %s", string(decrypt), expiry, time.Unix(expiry.Unix(), 0).UTC())

	//nolint
	{
		// b5274bbc6653d2c1db048315fe7b18dcd7c6ebb906751b60bfbec186c2faa966a951f3086c3cb822d251c5a4d435f5d0e09bfb948be39c0259f80bac965e89398ca1804629415d7b1cac6da5ce69477864f9e7167bff2560e912ccc0d2b85eb886a36e
		// 1bfd522ea75a82b65f7e43bb84c5cee76518061b43724753e0c81618e84d663a344661cc2749274726c49e81b1ef0a0f28831ccb1ad364bff43c4f8f6142294c7650d77f0a13d440e3c6
		// e9d0238e89a7262e159377a393b11c0129a1d375d7cacd36cfe77c9656ae5290722e4ec5cd35b64bf04646a912126fd812472f6afd7da7fd0ed6d57d292c6c4be14e342aca6dad6948f3d552b0
	}
}

func BenchmarkAesEncryptString(b *testing.B) {
	secret, _ := GenerateSecret(SECRET_LENGTH)

	for i := 0; i < b.N; i++ {
		EncryptString(value, secret)
	}
}

func BenchmarkAesDecryptString(b *testing.B) {
	secret, _ := GenerateSecret(SECRET_LENGTH)
	cipherValue, _ := EncryptString(value, secret)

	for i := 0; i < b.N; i++ {
		DecryptString(cipherValue, secret)
	}
}
