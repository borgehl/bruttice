package util

import "regexp"

func GetFirstSubmatch(re *regexp.Regexp, s string) string {
	found := re.FindStringSubmatch(s)
	if len(found) < 2 {
		return ""
	}
	return found[1]
}
