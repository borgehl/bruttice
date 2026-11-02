package util

import "fmt"

func PrefixErrorName(name string, err error) error {
	return fmt.Errorf("(%s): %s", name, err)
}
