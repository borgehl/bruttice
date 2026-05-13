package util

import (
	"regexp"
	"strings"
)

const (
	ditTags = "dit-tags"
	ditLoc  = "dit-loc"
	ditAge  = "dit-ages"
)

var ageRegexp = regexp.MustCompile(`(dit-ages[:|\w+]+)[:\s]?`)
var locRegexp = regexp.MustCompile(`(dit-loc[:|\w]+)[:\s]?`)
var tagRegexp = regexp.MustCompile(`(dit-tags[:|\w]+)[:\s]?`)

func parseValues(s string, prefix string, reg *regexp.Regexp) []string {
	return strings.Split(
		strings.ReplaceAll(reg.FindString(s), prefix, ""),
		":",
	)
}

func ParseTagsFromStr(s string) []string {
	tags := parseValues(s, ditTags, tagRegexp)
	ret := make([]string, 0)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		ret = append(ret, tag)
	}
	return ret
}

func parseSingularFlag(s string, prefix string, reg *regexp.Regexp) string {
	filtered := make([]string, 0)
	for _, v := range parseValues(s, prefix, reg) {
		if v = strings.TrimSpace(v); v != "" {
			filtered = append(filtered, v)
		}
	}
	if len(filtered) != 1 {
		return ""
	}
	return filtered[0]
}

func ParseLocationFromStr(s string) string {
	return parseSingularFlag(s, ditLoc, locRegexp)
}

func ParseAge(s string) string {
	return parseSingularFlag(s, ditAge, ageRegexp)
}
