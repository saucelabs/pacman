// Copyright 2021 The pacman Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pacman

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	NonDirectTypeAddsLen = 2
	SOCKS5Number         = 5
)

var re = regexp.MustCompile(`(?m)http|https|socks|socks5|quic`)

//////
// Helpers.
//////

func split(source string, s rune) []string {
	return strings.FieldsFunc(source, func(r rune) bool {
		return r == s
	})
}

//////
// Exported.
//////

// Proxy definition.
type Proxy struct {
	Address  string
	Password string
	Type     string
	Username string
}

// IsDirect tests whether it is using direct connection.
func (p *Proxy) IsDirect() bool {
	return p.Type == "DIRECT"
}

// IsSOCKS test whether it is a socks proxy.
func (p *Proxy) IsSOCKS() bool {
	if len(p.Type) >= SOCKS5Number {
		return p.Type[:SOCKS5Number] == "SOCKS"
	}

	return false
}

// URL returns a url representation.
func (p *Proxy) URL() string {
	switch p.Type {
	case "DIRECT":
		return ""
	case "PROXY":
		if !re.MatchString(p.Address) {
			p.Address = fmt.Sprintf("http://%s", p.Address)
		}

		return p.Address
	default:
		return fmt.Sprintf("%s://%s", strings.ToLower(p.Type), p.Address)
	}
}

func (p *Proxy) String() string {
	if p.IsDirect() {
		return p.Type
	}

	return fmt.Sprintf("%s %s", p.Type, p.Address)
}

// ParseProxy parses proxy string returned by `FindProxyForURL` and returns a
// slice of proxies.
func ParseProxy(pstr string) []Proxy {
	var proxies []Proxy

	for _, p := range split(pstr, ';') {
		typeAddr := strings.Fields(p)

		if len(typeAddr) == NonDirectTypeAddsLen {
			typ := strings.ToUpper(typeAddr[0])

			addr := typeAddr[1]

			var user, pass string

			if at := strings.Index(addr, "@"); at > 0 {
				auth := split(addr[:at], ':')

				if len(auth) == NonDirectTypeAddsLen {
					user = auth[0]
					pass = auth[1]
				}

				addr = addr[at+1:]
			}

			proxies = append(proxies, Proxy{
				Type:     typ,
				Address:  addr,
				Username: user,
				Password: pass,
			})
		} else if len(typeAddr) == 1 {
			proxies = append(proxies, Proxy{
				Type: strings.ToUpper(typeAddr[0]),
			})
		}
	}

	return proxies
}
