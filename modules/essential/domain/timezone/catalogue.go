// Package timezone is the hardcoded IANA catalogue a timezone setting is chosen from.
package timezone

import (
	"strings"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 500
)

var known = func() map[string]bool {
	set := make(map[string]bool, len(Ids))
	for _, id := range Ids {
		set[id] = true
	}
	return set
}()

// IsKnown reports whether id is a zone of the catalogue.
func IsKnown(id string) bool {
	return known[id]
}

type Zone struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

// Search returns one page of the zones whose id or translated label contains keyword. The match
// ignores case and treats "_" as a space, so "ho chi" finds Asia/Ho_Chi_Minh and "hồ chí" its vi label.
// labels maps an id to its label; an id without one is labelled by itself.
func Search(keyword string, labels map[string]string, page int, size int) (items []Zone, total int) {
	needle := normalize(keyword)
	matched := make([]Zone, 0, len(Ids))
	for _, id := range Ids {
		label := labels[id]
		if label == "" {
			label = id
		}
		if needle == "" || strings.Contains(normalize(id), needle) || strings.Contains(normalize(label), needle) {
			matched = append(matched, Zone{Code: id, Label: label})
		}
	}
	start, end := pageBounds(len(matched), page, size)
	return matched[start:end], len(matched)
}

func normalize(text string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(text), "_", " "))
}

func pageBounds(count int, page int, size int) (int, int) {
	if size <= 0 {
		size = DefaultPageSize
	}
	size = min(size, MaxPageSize)
	start := min(max(page, 0)*size, count)
	return start, min(start+size, count)
}
