package udns

import (
	"strings"
)

const (
	TCP_MSG = "udns_service_online"
)

func trimDot(s string) string {
	return strings.Trim(s, ".")
}
