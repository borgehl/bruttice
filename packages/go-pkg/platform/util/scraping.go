package util

import (
	"fmt"
	"strings"
)

func TagClassSelector(tag string, classes string) string {
	return fmt.Sprintf("%s.%s", tag, strings.ReplaceAll(classes, " ", "."))
}

func AttributeSelector(tag string, attribute string, value string) string {
	ret := fmt.Sprintf("%s[%s", tag, attribute)
	if value != "" {
		ret = fmt.Sprintf("%s='%s'", ret, value)
	}
	return fmt.Sprintf("%s]", ret)
}

func IsFree(cost string) bool {
	cost = strings.ToLower(cost)
	return cost == "$0" ||
		cost == "0" ||
		strings.Contains(cost, "free") ||
		strings.Contains(cost, "donations") ||
		strings.Contains(cost, "no cover") ||
		strings.Contains(cost, "pwyc") ||
		strings.Contains(cost, "notaflof") ||
		strings.Contains(cost, "no one turned away")
}
