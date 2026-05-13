package util

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

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
