package util

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// CleanUTF8String sanitizes a string in a single pass by:
//   - Dropping invalid UTF-8 sequences (replacement character U+FFFD)
//   - Replacing unicode whitespace (non-breaking space U+00A0, em space U+2003,
//     thin space U+2009, etc.) with standard ASCII spaces (U+0020)
func CleanUTF8String(s string) string {
	return strings.Map(func(r rune) rune {
		if r == utf8.RuneError {
			return -1
		}
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)
}

func TrimCleanStr(s string) string {
	s = CleanUTF8String(s)
	for _, chars := range []string{"\xe2", "\x80", "\xaf"} {
		s = strings.ReplaceAll(s, chars, " ")
	}
	ret := ""
	for _, word := range strings.Split(s, " ") {
		ret = fmt.Sprintf("%s %s", ret, word)
	}
	return strings.TrimSpace(ret)
}

// cleans string and converts to snake case via spaces
func SnakeCase(s string) string {
	s = TrimCleanStr(s)
	return strings.Join(strings.Split(s, " "), "_")
}

func CleanoutHTML(s string) string {
	s = strings.ReplaceAll(s, "<br>", "/n")
	s = strings.ReplaceAll(s, "<span>", "")
	s = strings.ReplaceAll(s, "</span>", "")
	return s
}

var commonTypos = strings.NewReplacer(
	"Wedneday", "Wednesday",
)

// FixCommonTypos corrects known typos found in scraped source data.
func FixCommonTypos(s string) string {
	return commonTypos.Replace(s)
}
