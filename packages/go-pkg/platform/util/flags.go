package util

import (
	"regexp"
	"strings"
)

const (
	flagTags = "bruttice-tags"
	flagLoc  = "bruttice-loc"
	flagAge  = "bruttice-ages"
)

// A value is letters, digits, "_" and "-" ("live-music", "all-ages"), and for an age also "+". Values
// are separated by ":".
var ageRegexp = regexp.MustCompile(`(bruttice-ages[:|\w+-]+)[:\s]?`)
var locRegexp = regexp.MustCompile(`(bruttice-loc[:|\w-]+)[:\s]?`)
var tagRegexp = regexp.MustCompile(`(bruttice-tags[:|\w-]+)[:\s]?`)

func parseValues(s string, prefix string, reg *regexp.Regexp) []string {
	return strings.Split(
		// remove prefix from the flag line
		strings.ReplaceAll(reg.FindString(s), prefix, ""),
		":",
	)
}

func ParseTagsFromStr(s string) []string {
	tags := parseValues(s, flagTags, tagRegexp)
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
	return parseSingularFlag(s, flagLoc, locRegexp)
}

func ParseAge(s string) string {
	return parseSingularFlag(s, flagAge, ageRegexp)
}
