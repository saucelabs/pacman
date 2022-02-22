// Copyright 2021 The pacman Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pacman

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/saucelabs/customerror"
	"github.com/saucelabs/pacman/internal/validation"
	"github.com/saucelabs/pacman/pkg/mode"
)

const nonDirectTypeAddsLen = 2

var validProxySchemesRegex = regexp.MustCompile(`(?m)http|https|socks5|socks|quic`)

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
	// The original address parsed from PAC content. If type (`m`) is DIRECT,
	// address is empty.
	address string

	// The type of the proxy from the parsed PAC content.
	mode mode.Mode

	// The parsed proxy URI from the parsed PAC content.
	uri *url.URL
}

// GetAddress returns the original address parsed from PAC content. If type
// is `DIRECT`, returns empty.
func (p *Proxy) GetAddress() string {
	return p.address
}

// GetType returns the mode of the proxy.
func (p *Proxy) GetMode() mode.Mode {
	return p.mode
}

// GetURI returns the proxy URI. If proxy mode is `DIRECT`, returns `nil`.
func (p *Proxy) GetURI() *url.URL {
	if p.mode == "DIRECT" {
		return nil
	}

	return p.uri
}

// String is the Stringer interface implementation. Returns mode is `DIRECT`,
// otherwise returns `{MODE} {URI}`.
func (p *Proxy) String() string {
	if p.GetMode() == mode.Direct {
		return p.GetMode().String()
	}

	return fmt.Sprintf("%s %s", p.GetMode(), p.GetURI())
}

// ParseProxy parses proxy string returned by `FindProxyForURL`, and returns a
// list of proxies.
func ParseProxy(pstr string) ([]Proxy, error) {
	const errMsgPrefix = "parse PAC proxy URI"

	var proxies []Proxy

	for _, p := range split(pstr, ';') {
		modeAndAddress := strings.Fields(p)

		if len(modeAndAddress) == nonDirectTypeAddsLen {
			m := strings.ToUpper(modeAndAddress[0]) // mode.

			address := modeAndAddress[1]

			if address == "" {
				return nil, customerror.NewInvalidError(errMsgPrefix + ", empty")
			}

			// Deal with cases where the address has no scheme.
			if !validProxySchemesRegex.MatchString(address) {
				switch strings.ToUpper(m) {
				case mode.Socks5.String():
					address = fmt.Sprintf("socks5://%s", address)
				case mode.Socks.String():
					address = fmt.Sprintf("socks://%s", address)
				case mode.Proxy.String():
					address = fmt.Sprintf("http://%s", address)
				}
			}

			// Should be a valid URI.
			if err := validation.Get().Var(address, "proxyURI"); err != nil {
				return nil, customerror.NewFailedToError(errMsgPrefix+", invalid", customerror.WithError(err))
			}

			parsedProxyURI, err := url.ParseRequestURI(address)
			if err != nil {
				return nil, customerror.NewFailedToError(errMsgPrefix, customerror.WithError(err))
			}

			proxies = append(proxies, Proxy{
				address: address,
				mode:    mode.Mode(m),

				uri: parsedProxyURI,
			})
		} else if len(modeAndAddress) == 1 { // DIRECT case.
			proxies = append(proxies, Proxy{
				mode: mode.Mode(strings.ToUpper(modeAndAddress[0])),
			})
		}
	}

	return proxies, nil
}
