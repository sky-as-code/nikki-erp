package util

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type ServerTokenPayload struct {
	UserId   string   `json:"userid"`
	SystemId string   `json:"systemid"`
	Roles    []string `json:"roles"`
}
type ServerCustomClaims struct {
	SystemId string   `json:"systemid"`
	UserId   string   `json:"userid"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateServerToken(password, systemid, userid, issuer string,
	roles []string, expire_seconds int64) (string, error) {
	return GenerateServerTokenWithTime(password, systemid, userid, issuer, roles, expire_seconds, time.Now())
}

func GenerateServerTokenWithTime(password string, systemid string, userid string, issuer string, roles []string, expire_seconds int64, current time.Time) (string, error) {
	signingKey := []byte(password)

	claims := ServerCustomClaims{
		systemid,
		userid,
		roles,
		jwt.RegisteredClaims{
			IssuedAt:  &jwt.NumericDate{Time: current},
			ExpiresAt: &jwt.NumericDate{Time: current.Add(time.Duration(expire_seconds) * time.Second)},
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(signingKey)
}

type GJWTokenPayload struct {
	UserId string   `json:"userid,omitempty"`
	DId    string   `json:"did,omitempty"`
	Roles  []string `json:"roles"`
}

type GJWCustomClaims struct {
	UserId string   `json:"userid,omitempty"`
	DId    string   `json:"did,omitempty"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateGJWToken(password, deviceid, userid, issuer string,
	roles []string, expire_seconds int64) (string, error) {
	return GenerateGJWTokenWithTime(password, deviceid, userid, issuer, roles, expire_seconds, time.Now())
}

func GenerateGJWTokenWithTime(password string, deviceid string, userid string, issuer string, roles []string, expire_seconds int64, current time.Time) (string, error) {
	signingKey := []byte(password)

	claims := GJWCustomClaims{
		userid,
		deviceid,
		roles,
		jwt.RegisteredClaims{
			IssuedAt:  &jwt.NumericDate{Time: current},
			ExpiresAt: &jwt.NumericDate{Time: current.Add(time.Duration(expire_seconds) * time.Second)},
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(signingKey)
}
