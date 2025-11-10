package semver

import (
	"go.bryk.io/pkg/errors"
	"golang.org/x/mod/semver"

	"github.com/sky-as-code/nikki-erp/common/util"
)

func ParseSemVer(version string) (*SemVer, error) {
	if !semver.IsValid(version) {
		return nil, errors.New("invalid semver format")
	}
	return util.ToPtr(SemVer(version)), nil
}

func MustParseSemVer(version string) *SemVer {
	if !semver.IsValid(version) {
		panic(errors.New("invalid semver format"))
	}
	return util.ToPtr(SemVer(version))
}

// SemVer represents a semantic version number.
type SemVer string

func (this SemVer) Build() string {
	return semver.Build(string(this))
}

func (this SemVer) Compare(other *SemVer) int {
	return semver.Compare(string(this), string(*other))
}

func (this SemVer) Canonical() string {
	return semver.Canonical(string(this))
}

func (this SemVer) Major() string {
	return semver.Major(string(this))
}

func (this SemVer) MajorMinor() string {
	return semver.MajorMinor(string(this))
}

func (this SemVer) Prerelease() string {
	return semver.Prerelease(string(this))
}

// String returns the raw version that was used to create the SemVer instance.
func (this SemVer) String() string {
	return string(this)
}

func (this SemVer) IsValid() bool {
	return semver.IsValid(string(this))
}
