// Copyright 2021 The pacman Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pacman_test

import (
	"testing"

	"github.com/saucelabs/pacman"
)

func TestParseProxy(t *testing.T) {
	proxy := "PROXY 127.0.0.1:8080; SOCKs 127.0.0.1:1080; Direct"

	proxies := pacman.ParseProxy(proxy)

	if len(proxies) != 3 {
		t.Error("Parse failed")
		return
	}

	if proxies[1].Type != "SOCKS" {
		t.Error("Should be SOCKS5")
	}

	if !proxies[1].IsSOCKS() {
		t.Error("Should be SOCKS5")
	}

	if !proxies[2].IsDirect() {
		t.Error("Should be direct")
	}
}

func TestParseSOCKS(t *testing.T) {
	proxy := "SOCKS5 127.0.0.1:1080"

	proxies := pacman.ParseProxy(proxy)

	if len(proxies) != 1 {
		t.Error("Parse failed")
		return
	}

	if !proxies[0].IsSOCKS() {
		t.Error("Should be SOCKS5")
	}

	if proxies[0].IsDirect() {
		t.Error("Should be direct")
	}
}
