// Copyright 2021 The pacman Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pacman_test

import (
	"testing"

	"github.com/saucelabs/pacman"
	"github.com/saucelabs/pacman/pkg/mode"
)

func TestParseProxy(t *testing.T) {
	proxy := "PROXY 127.0.0.1:8080; SOCKS5 127.0.0.1:1080; SOCKS 127.0.0.1:1080; Direct"

	proxies, err := pacman.ParseProxy(proxy)
	if err != nil {
		t.Fatal("Parse failed", err)
	}

	if len(proxies) != 4 {
		t.Fatal("Expected 4 proxies")
	}

	if proxies[0].GetURI().String() != "http://127.0.0.1:8080" {
		t.Fatalf("Expected %s to be http://127.0.0.1:8080", proxies[0].GetURI())
	}

	if proxies[0].GetAddress() != "http://127.0.0.1:8080" {
		t.Fatalf("Expected %s to be http://127.0.0.1:8080", proxies[0].GetAddress())
	}

	if proxies[1].GetMode() != mode.Socks5 {
		t.Fatal("Should be SOCKS5")
	}

	if proxies[2].GetMode() != mode.Socks {
		t.Fatal("Should be SOCKS")
	}

	if proxies[3].GetMode() != mode.Direct {
		t.Fatal("Should be direct")
	}

	if proxies[3].String() != "DIRECT" {
		t.Fatalf("Expected String to be DIRECT, got %s", proxies[2].String())
	}
}
