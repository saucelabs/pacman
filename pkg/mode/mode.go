package mode

import (
	"regexp"
	"strings"
)

var validModes = regexp.MustCompile(`(?m)SOCKS5|SOCKS|DIRECT|PROXY`)

// Mode is the type of the proxy as specified in the PAC content. Called `Mode`
// because `type` is a Golang reserved word.
type Mode string

// String is the Stringer interface implementation.
func (m Mode) String() string {
	return string(m)
}

// List of possible modes.
const (
	Direct Mode = "DIRECT"
	Proxy  Mode = "PROXY"
	Socks  Mode = "SOCKS"
	Socks5 Mode = "SOCKS5"
)

// IsMode returns if `s` is a valid `Mode`.
//
// Note: It will automatically uppercase `s`.
func IsMode(s string) bool {
	return validModes.MatchString(strings.ToUpper(s))
}
