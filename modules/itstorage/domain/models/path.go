package models

import (
	"path"
	"regexp"
	"strings"

	"github.com/sky-as-code/nikki-erp/common/model"
)

const PathSeparator = "/"

// noTenantSegment stands in for the tenant in an object key of a binary without tenants, so
// every key keeps the same three-segment shape.
const noTenantSegment = "_"

var (
	ulidRegex      = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)
	extensionRegex = regexp.MustCompile(`^\.[a-z0-9]{1,16}$`)
)

// JoinPath appends a space name to its parent's path; a root space's path is its own name.
func JoinPath(parentPath string, name string) string {
	if parentPath == "" {
		return name
	}
	return parentPath + PathSeparator + name
}

// IsUnderPath reports whether candidate lies strictly below prefix.
func IsUnderPath(candidate string, prefix string) bool {
	return prefix != "" && strings.HasPrefix(candidate, prefix+PathSeparator)
}

// ExtensionOf keeps an upload's extension on its object key, lowercased. Anything that is not a
// short alphanumeric extension is dropped rather than carried into the key.
func ExtensionOf(fileName string) string {
	ext := strings.ToLower(path.Ext(fileName))
	if !extensionRegex.MatchString(ext) {
		return ""
	}
	return ext
}

// BuildObjectKey is {tenant_id|_}/{space_id}/{file_id}{ext}: immutable, so renaming a space never
// moves an object.
func BuildObjectKey(tenantId string, spaceId model.Id, fileId model.Id, ext string) string {
	if tenantId == "" {
		tenantId = noTenantSegment
	}
	return tenantId + PathSeparator + spaceId + PathSeparator + fileId + ext
}

// BuildFilePath is the path clients store in a `*_file_path` field.
func BuildFilePath(spacePath string, fileId model.Id, ext string) string {
	return spacePath + PathSeparator + fileId + ext
}

// ParsedFilePath is a file path split into the space path it was issued under and its leaf.
type ParsedFilePath struct {
	SpacePath string
	FileId    model.Id
	Leaf      string
}

// ParseFilePath accepts "<space path>/<file id><ext>". The id is what a lookup goes by: the space
// path is only the one current when the path was issued, and a rename leaves it stale.
func ParseFilePath(filePath string) (ParsedFilePath, bool) {
	trimmed := strings.TrimSpace(filePath)
	cut := strings.LastIndex(trimmed, PathSeparator)
	if cut <= 0 || cut == len(trimmed)-1 {
		return ParsedFilePath{}, false
	}
	leaf := trimmed[cut+1:]
	fileId, ext, _ := strings.Cut(leaf, ".")
	if !ulidRegex.MatchString(fileId) {
		return ParsedFilePath{}, false
	}
	if ext != "" && !extensionRegex.MatchString("."+ext) {
		return ParsedFilePath{}, false
	}
	return ParsedFilePath{SpacePath: trimmed[:cut], FileId: fileId, Leaf: leaf}, true
}

// MimeFamilyMatches compares a declared media type with the one sniffed from the content.
//
// Signatures cannot tell every type apart, so only a disagreement the sniffer is sure of fails:
// an unrecognised binary (application/octet-stream) matches anything, and plain text matches any
// textual declaration (CSV, JSON, SVG, XML and the like sniff as text/plain).
func MimeFamilyMatches(declared string, detected string) bool {
	declared = normalizeMime(declared)
	detected = normalizeMime(detected)
	if declared == "" || detected == "" {
		return false
	}
	if detected == "application/octet-stream" || declared == detected {
		return true
	}
	declaredFamily, declaredSub, _ := strings.Cut(declared, PathSeparator)
	detectedFamily, _, _ := strings.Cut(detected, PathSeparator)
	if detected == "text/plain" {
		return declaredFamily == "text" || isTextualApplication(declaredSub)
	}
	return declaredFamily == detectedFamily
}

func normalizeMime(mime string) string {
	base, _, _ := strings.Cut(mime, ";")
	return strings.ToLower(strings.TrimSpace(base))
}

func isTextualApplication(subType string) bool {
	for _, marker := range []string{"json", "xml", "csv", "javascript", "yaml", "sql", "x-sh"} {
		if strings.Contains(subType, marker) {
			return true
		}
	}
	return false
}
