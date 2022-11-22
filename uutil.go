package udns

import "strings"

func trimDot(s string) string {
	return strings.Trim(s, ".")
}
