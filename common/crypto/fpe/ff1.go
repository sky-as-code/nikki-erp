// Package fpe implements format-preserving encryption.
//
// Ff1 is the FF1 mode of NIST SP 800-38G (Rev. 1) over AES, written against the standard so that no
// third-party dependency is needed. It supports radix 2..65536, AES-128/192/256 keys and tweaks of any length.
package fpe

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"math/big"

	"go.bryk.io/pkg/errors"
)

const (
	MinRadix = 2
	MaxRadix = 65536

	ff1Rounds  = 10
	blockSize  = aes.BlockSize
	minDomain  = 1_000_000
	maxTweakLn = 1<<32 - 1
)

// Ff1 holds the AES block cipher and radix of one FF1 instance. It is safe for concurrent use.
type Ff1 struct {
	block cipher.Block
	radix int
	// maxLen is 2*floor(log_radix(2^96)), the longest input the standard allows.
	maxLen int
}

func NewFf1(key []byte, radix int) (*Ff1, error) {
	if radix < MinRadix || radix > MaxRadix {
		return nil, errors.Errorf("fpe: radix %d out of range [%d, %d]", radix, MinRadix, MaxRadix)
	}
	switch len(key) {
	case 16, 24, 32:
	default:
		return nil, errors.Errorf("fpe: invalid AES key length %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "fpe: create AES cipher")
	}
	return &Ff1{block: block, radix: radix, maxLen: computeMaxLen(radix)}, nil
}

func computeMaxLen(radix int) int {
	limit := new(big.Int).Lsh(big.NewInt(1), 96)
	bigRadix := big.NewInt(int64(radix))
	power := big.NewInt(1)
	exponent := 0
	for {
		power.Mul(power, bigRadix)
		if power.Cmp(limit) > 0 {
			return 2 * exponent
		}
		exponent++
	}
}

// Encrypt returns the FF1 ciphertext of numerals (each < radix) under tweak.
func (this *Ff1) Encrypt(tweak []byte, numerals []uint16) ([]uint16, error) {
	return this.run(tweak, numerals, true)
}

// Decrypt reverses Encrypt.
func (this *Ff1) Decrypt(tweak []byte, numerals []uint16) ([]uint16, error) {
	return this.run(tweak, numerals, false)
}

// ff1Round carries the values that stay fixed across the ten Feistel rounds.
type ff1Round struct {
	tweak     []byte
	prefix    []byte // P || T || zero padding, i.e. everything before [i] in P || Q
	numBytes  int    // b
	outBytes  int    // d
	lenA      int    // u
	lenB      int    // v
	modulusA  *big.Int
	modulusB  *big.Int
	bigRadix  *big.Int
	inputSize int
}

func (this *Ff1) run(tweak []byte, numerals []uint16, encrypt bool) ([]uint16, error) {
	if err := this.validate(tweak, numerals); err != nil {
		return nil, err
	}
	round := this.newRound(tweak, len(numerals))
	numA := this.num(numerals[:round.lenA])
	numB := this.num(numerals[round.lenA:])
	if encrypt {
		for i := 0; i < ff1Rounds; i++ {
			numA, numB = numB, this.feistelAdd(round, i, numA, numB)
		}
	} else {
		for i := ff1Rounds - 1; i >= 0; i-- {
			numA, numB = this.feistelSub(round, i, numA, numB), numA
		}
	}
	result := make([]uint16, 0, len(numerals))
	result = append(result, this.str(numA, round.lenA)...)
	return append(result, this.str(numB, round.lenB)...), nil
}

func (this *Ff1) validate(tweak []byte, numerals []uint16) error {
	length := len(numerals)
	if length < 2 || length > this.maxLen {
		return errors.Errorf("fpe: input length %d out of range [2, %d]", length, this.maxLen)
	}
	domain := new(big.Int).Exp(big.NewInt(int64(this.radix)), big.NewInt(int64(length)), nil)
	if domain.Cmp(big.NewInt(minDomain)) < 0 {
		return errors.Errorf("fpe: domain radix^%d is below %d", length, minDomain)
	}
	if uint64(len(tweak)) > maxTweakLn {
		return errors.New("fpe: tweak too long")
	}
	for i, numeral := range numerals {
		if int(numeral) >= this.radix {
			return errors.Errorf("fpe: numeral %d at index %d is not below radix %d", numeral, i, this.radix)
		}
	}
	return nil
}

func (this *Ff1) newRound(tweak []byte, length int) *ff1Round {
	round := &ff1Round{tweak: tweak, lenA: length / 2, inputSize: length}
	round.lenB = length - round.lenA
	round.bigRadix = big.NewInt(int64(this.radix))
	round.modulusA = new(big.Int).Exp(round.bigRadix, big.NewInt(int64(round.lenA)), nil)
	round.modulusB = new(big.Int).Exp(round.bigRadix, big.NewInt(int64(round.lenB)), nil)
	maxB := new(big.Int).Sub(round.modulusB, big.NewInt(1))
	round.numBytes = (maxB.BitLen() + 7) / 8
	if round.numBytes == 0 {
		round.numBytes = 1
	}
	round.outBytes = 4*((round.numBytes+3)/4) + 4
	round.prefix = this.buildPrefix(round)
	return round
}

// buildPrefix returns P || T || [0]^((-t-b-1) mod 16).
func (this *Ff1) buildPrefix(round *ff1Round) []byte {
	header := make([]byte, blockSize)
	header[0], header[1], header[2] = 1, 2, 1
	header[3] = byte(this.radix >> 16)
	header[4] = byte(this.radix >> 8)
	header[5] = byte(this.radix)
	header[6] = 10
	header[7] = byte(round.lenA % 256)
	binary.BigEndian.PutUint32(header[8:12], uint32(round.inputSize))
	binary.BigEndian.PutUint32(header[12:16], uint32(len(round.tweak)))
	padding := mod(-len(round.tweak)-round.numBytes-1, blockSize)
	prefix := make([]byte, 0, blockSize+len(round.tweak)+padding)
	prefix = append(prefix, header...)
	prefix = append(prefix, round.tweak...)
	return append(prefix, make([]byte, padding)...)
}

func mod(value, modulus int) int {
	result := value % modulus
	if result < 0 {
		result += modulus
	}
	return result
}

// roundValue returns y = NUM(S) for round i, where the round function input is numeral value half.
func (this *Ff1) roundValue(round *ff1Round, i int, half *big.Int) *big.Int {
	message := make([]byte, 0, len(round.prefix)+1+round.numBytes)
	message = append(message, round.prefix...)
	message = append(message, byte(i))
	message = append(message, half.FillBytes(make([]byte, round.numBytes))...)
	mac := this.cbcMac(message)
	return new(big.Int).SetBytes(this.expand(mac, round.outBytes))
}

func (this *Ff1) cbcMac(message []byte) []byte {
	state := make([]byte, blockSize)
	for offset := 0; offset < len(message); offset += blockSize {
		for j := 0; j < blockSize; j++ {
			state[j] ^= message[offset+j]
		}
		this.block.Encrypt(state, state)
	}
	return state
}

// expand returns the first outBytes of R || CIPH(R xor [1]^16) || CIPH(R xor [2]^16) || ...
func (this *Ff1) expand(mac []byte, outBytes int) []byte {
	output := make([]byte, 0, outBytes+blockSize)
	output = append(output, mac...)
	for counter := uint64(1); len(output) < outBytes; counter++ {
		block := make([]byte, blockSize)
		binary.BigEndian.PutUint64(block[8:], counter)
		for j := 0; j < blockSize; j++ {
			block[j] ^= mac[j]
		}
		this.block.Encrypt(block, block)
		output = append(output, block...)
	}
	return output[:outBytes]
}

func (this *Ff1) modulusFor(round *ff1Round, i int) *big.Int {
	if i%2 == 0 {
		return round.modulusA
	}
	return round.modulusB
}

func (this *Ff1) feistelAdd(round *ff1Round, i int, numA, numB *big.Int) *big.Int {
	value := this.roundValue(round, i, numB)
	value.Add(value, numA)
	return value.Mod(value, this.modulusFor(round, i))
}

func (this *Ff1) feistelSub(round *ff1Round, i int, numA, numB *big.Int) *big.Int {
	value := this.roundValue(round, i, numA)
	value.Sub(numB, value)
	// big.Int.Mod is Euclidean, so the result is already non-negative.
	return value.Mod(value, this.modulusFor(round, i))
}

func (this *Ff1) num(numerals []uint16) *big.Int {
	bigRadix := big.NewInt(int64(this.radix))
	result := new(big.Int)
	digit := new(big.Int)
	for _, numeral := range numerals {
		result.Mul(result, bigRadix)
		result.Add(result, digit.SetUint64(uint64(numeral)))
	}
	return result
}

func (this *Ff1) str(value *big.Int, length int) []uint16 {
	bigRadix := big.NewInt(int64(this.radix))
	remaining := new(big.Int).Set(value)
	digit := new(big.Int)
	result := make([]uint16, length)
	for i := length - 1; i >= 0; i-- {
		remaining.DivMod(remaining, bigRadix, digit)
		result[i] = uint16(digit.Uint64())
	}
	return result
}
