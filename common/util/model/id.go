package model

import (
	"go.bryk.io/pkg/ulid"
)

func MustNewULID() string {
	newUlid, err := ulid.New()
	if err != nil {
		panic(err)
	}
	return newUlid.String()
}
