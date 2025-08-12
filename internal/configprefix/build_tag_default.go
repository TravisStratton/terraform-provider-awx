//go:build !repoAWX && !repoAAP
// +build !repoAWX,!repoAAP

package configprefix

import "strings"

const Prefix = "awx"

func ReplaceText(input string) string {
	return strings.ReplaceAll(input, "aap_", "awx_")
}
