package crypto

import (
	"crypto/md5"
	"math/big"
	"regexp"
	"time"

	"github.com/gofrs/uuid"
)

const version = "1"

var BASE32 = "0123456789abcdefghijklmnopqrstuv"
var MAP_BASE32 = map[byte]int{}
var seeds = "GJW-465c8dc1-d4bd-8c10-98dd-342df17e4645"

func SetSeeds(seed string) {
	seeds = seed
}

type GjwIdPurpose int

func NewGjwId(randomLength uint, suffix string, timestamp ...int64) string {

	var timestampMicro int64
	if len(timestamp) > 0 {
		timestampMicro = timestamp[0]
	} else {
		timestampMicro = time.Now().UnixMicro()
	}

	datestr := ""
	for i := 0; i < 11; i++ { // Why 11?
		datestr = string(BASE32[timestampMicro&0x1F]) + datestr
		timestampMicro >>= 5
	}

	randstr := getRandom62(randomLength)
	hashstr := getHashValue(datestr + randstr + version + suffix)
	return datestr + randstr + version + hashstr + suffix
}

func ExtractTime(gjwId string) time.Time {
	timestamp := int64(0)
	for i := 0; i < 11; i++ {
		timestamp = (timestamp << 5) + int64(MAP_BASE32[gjwId[i]])
	}
	return time.Unix(timestamp/1_000_000, (timestamp%1_000_000)*1_000)
}

var re = regexp.MustCompile("^[a-zA-Z0-9]{23,33}$")

func ValidateGjwId(gjwId string) bool {
	if !re.MatchString(gjwId) {
		return false
	}

	length := len(gjwId)
	return getHashValue(gjwId[:length-4]+gjwId[length-2:]) == gjwId[length-4:length-2]
}

func getHashValue(text string) string {
	h := md5.Sum([]byte(text + seeds))
	return string(BASE32[h[9]&0x1F]) + string(BASE32[h[11]&0x1F])
}

func getRandom62(ilen uint) string {

	uid1, _ := uuid.NewV1()
	uid4, _ := uuid.NewV4()

	hash := md5.New()
	defer hash.Reset()
	hash.Write(uid1.Bytes())
	hash.Write(uid4.Bytes())
	h := hash.Sum(nil)

	var i big.Int
	i.SetBytes(h[:])
	return i.Text(62)[:ilen]
}
