package system

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed dev.portnado.anchor.in
var pfAnchorTemplate string

const pfIncludeLine = `rdr-anchor "dev.portnado"`

func RenderPFAnchor(proxyPort int) (string, error) {
	if proxyPort < 1 || proxyPort > 65535 {
		return "", fmt.Errorf("proxy port must be between 1 and 65535")
	}
	return strings.ReplaceAll(pfAnchorTemplate, "{{PORT}}", fmt.Sprintf("%d", proxyPort)), nil
}

func ManagedPFIncludeLine() string {
	return pfIncludeLine
}
