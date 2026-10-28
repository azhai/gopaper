package common

import (
	"regexp"
	"strings"
)

var slugRe = regexp.MustCompile(`[^a-z0-9-]`)
var multiHyphenRe = regexp.MustCompile(`-{2,}`)

func GenerateSlug(filename string) string {
	name := strings.TrimSuffix(filename, ".md")
	name = strings.ToLower(name)
	name = slugRe.ReplaceAllString(name, "-")
	name = multiHyphenRe.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	return name
}

func ValidateSlug(slug string) bool {
	if slug == "" {
		return false
	}
	if len(slug) > 100 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
	return matched
}
