//go:build windows
// +build windows

package ghost

import "os"

var HostsFile = os.Getenv("SystemRoot") + `\System32\drivers\etc\hosts`
