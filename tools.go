package packago

import (
	"strings"
)

func GetVarName(n string) string {
	n = strings.ReplaceAll(n, ".", "_")
	n = strings.ReplaceAll(n, "-", "_")
	n = strings.Title(n)
	return n
}
