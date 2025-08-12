//go:build repoAWX
// +build repoAWX

package configprefix

import "strings"

const Prefix = "awx"

func ReplaceText(input string) string {
	return strings.ReplaceAll(input, "aap_", "awx_")
}
