// Copyright 2021 The pacman Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pacman_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/saucelabs/pacman"
	"github.com/saucelabs/pacman/pkg/mode"
)

// Creates a mocked HTTP server. Any error will throw a fatal error. Don't
// forget to defer close it.
func createMockedHTTPServer(statusCode int, body string) *httptest.Server {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(statusCode)

		if _, err := res.Write([]byte(body)); err != nil {
			log.Fatalln("Failed to write body.", err)
		}
	}))

	// Give enough time to start, and be ready.
	time.Sleep(1 * time.Second)

	return testServer
}

func findProxy(t *testing.T, pac *pacman.Parser) {
	t.Helper()

	proxies, err := pac.FindProxy("http://www.example.com/")
	if err != nil {
		t.Fatal(err)
	}

	if len(proxies) != 6 {
		t.Fatalf("Expected 6 got %d proxies", len(proxies))
	}

	expectedProxies := []string{
		"http://4.5.6.7:8080",
		"https://4.5.6.7:8081",
		"socks://4.5.6.7:8082",
		"socks5://4.5.6.7:8083",
		"quic://4.5.6.7:8084",
		"http://4.5.6.7:8085",
	}

	for i, p := range proxies {
		uri := p.GetURI()
		if uri == nil {
			t.Fatalf("`GetURI` expected not nil %+v", uri)
		}

		if uri.String() != expectedProxies[i] {
			t.Fatalf("`expectedProxies` to be %s got %s", expectedProxies[i], uri.String())
		}

		proxyPrefixedWithProxy := fmt.Sprintf("PROXY %s", uri.String())
		if p.String() != proxyPrefixedWithProxy {
			t.Errorf("expected %s to be prefixed with PROXY, got %s", p.String(), proxyPrefixedWithProxy)
		}
	}
}

func TestParser_New_fromfile(t *testing.T) {
	pacFromFile, err := pacman.New("resources/data.pac")
	if err != nil {
		t.Fatal(err)
	}

	findProxy(t, pacFromFile)
}

func TestParser_New_fromweb(t *testing.T) {
	pacData, err := os.Open("resources/data.pac")
	if err != nil {
		t.Fatal(err)
	}

	defer pacData.Close()

	data, err := ioutil.ReadAll(pacData)
	if err != nil {
		t.Fatal(err)
	}

	pacServer := createMockedHTTPServer(http.StatusOK, string(data))

	defer pacServer.Close()

	pacFromWeb, err := pacman.New(pacServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	findProxy(t, pacFromWeb)
}

func TestParser_New_fromweb_non2xx(t *testing.T) {
	pacData, err := os.Open("resources/data.pac")
	if err != nil {
		t.Fatal(err)
	}

	defer pacData.Close()

	data, err := ioutil.ReadAll(pacData)
	if err != nil {
		t.Fatal(err)
	}

	pacServer := createMockedHTTPServer(http.StatusNotFound, string(data))

	defer pacServer.Close()

	_, err = pacman.New(pacServer.URL)
	if err == nil {
		t.Fatal("`From` expected error, got nil")
	}
}

func TestParser_New_fromweb_noBody(t *testing.T) {
	pacServer := createMockedHTTPServer(http.StatusOK, "")

	defer pacServer.Close()

	_, err := pacman.New(pacServer.URL)

	if err == nil {
		t.Fatalf("`From` expected no error, got %+v", err)
	}

	if !strings.Contains(err.Error(), "invalid params") {
		t.Fatalf("`From` expected error content, got %+v", err)
	}
}

func TestParser_New_fromweb_invalidBody(t *testing.T) {
	pacServer := createMockedHTTPServer(http.StatusOK, "invalid content")

	defer pacServer.Close()

	_, err := pacman.New(pacServer.URL)

	if err == nil {
		t.Fatalf("`From` expected no error, got %+v", err)
	}

	if !strings.Contains(err.Error(), "invalid params") {
		t.Fatalf("`From` expected error content, got %+v", err)
	}
}

func TestParser_New_text(t *testing.T) {
	pacData, err := os.Open("resources/data.pac")
	if err != nil {
		t.Fatal(err)
	}

	defer pacData.Close()

	data, err := ioutil.ReadAll(pacData)
	if err != nil {
		t.Fatal(err)
	}

	pacFromText, err := pacman.New(string(data))
	if err != nil {
		t.Fatal(err)
	}

	findProxy(t, pacFromText)

	if pacFromText.Source() != "text" {
		if err != nil {
			t.Fatalf("pacFromText.Source() expected text, got %s", pacFromText.Source())
		}
	}

	if !strings.Contains(pacFromText.Content(), "FindProxyForURL") {
		if err != nil {
			t.Fatalf("pacFromText.Content() expected FindProxyForURL, got %s", pacFromText.Content())
		}
	}
}

func TestFindProxy_direct(t *testing.T) {
	dsts := []string{
		"http://localhost/",
		"https://intranet.domain.com",
		"http://abcdomain.com",
		"http://www.abcdomain.com",
		"ftp://example.com.com",
	}

	pac, err := pacman.New("resources/data.pac")
	if err != nil {
		t.Fatal(err)
	}

	for _, dst := range dsts {
		proxies, err := pac.FindProxy(dst)
		if err != nil {
			t.Fatal(err)
		}

		if len(proxies) != 1 {
			t.Fatalf("Find proxy failed for %s", dst)
		}

		p := proxies[0]
		isDirect := p.GetMode() == mode.Direct

		if !isDirect {
			t.Fatalf("`IsDirect()` expected to be DIRECT got %v", isDirect)
		}

		if p.GetURI() != nil {
			t.Errorf("Expected `URI` to be nil %+v", p.GetURI())
		}
	}
}

func TestParser_New_noTextOrURI(t *testing.T) {
	_, err := pacman.New("")

	if err == nil {
		t.Fatalf("`From` expected no error, got %+v", err)
	}

	if !strings.Contains(err.Error(), "invalid params") {
		t.Fatalf("`From` expected error content, got %+v", err)
	}
}

func TestParser_New_invalidTextOrURI(t *testing.T) {
	_, err := pacman.New("invalid content")

	if err == nil {
		t.Fatalf("`From` expected no error, got %+v", err)
	}

	if !strings.Contains(err.Error(), "invalid params") {
		t.Fatalf("`From` expected error content, got %+v", err)
	}
}

func TestParser_New_withProxyCredentials(t *testing.T) {
	os.Setenv("PACMAN_LOG_LEVEL", "debug")
	defer os.Unsetenv("PACMAN_LOG_LEVEL")

	pacFromFile, err := pacman.New("resources/data.pac", "http://user:pass@4.5.6.7:8080")
	if err != nil {
		t.Fatal(err)
	}

	proxies, err := pacFromFile.FindProxy("http://www.example.com/")
	if err != nil {
		t.Fatal(err)
	}

	uriWithCredential := proxies[0].GetURI().String()
	if uriWithCredential != "http://user:pass@4.5.6.7:8080" {
		t.Fatalf("Expected proxy URI with creds, got %s", uriWithCredential)
	}
}

func BenchmarkFind(b *testing.B) {
	pacf, _ := os.Open("resources/data.pac")
	defer pacf.Close()

	data, _ := ioutil.ReadAll(pacf)
	pac, _ := pacman.New(string(data))

	for n := 0; n < b.N; n++ {
		_, _ = pac.FindProxyForURL("http://www.example.com/")
		_, _ = pac.FindProxyForURL("http://localhost/")
		_, _ = pac.FindProxyForURL("http://192.168.1.1/")
	}
}

func Example() {
	pacf, _ := os.Open("resources/data.pac")
	defer pacf.Close()

	data, _ := ioutil.ReadAll(pacf)
	pac, _ := pacman.New(string(data))

	r, _ := pac.FindProxyForURL("http://www.example.com/")

	fmt.Println(r)

	// Output:
	// PROXY http://4.5.6.7:8080; PROXY https://4.5.6.7:8081; PROXY socks://4.5.6.7:8082; PROXY socks5://4.5.6.7:8083; PROXY quic://4.5.6.7:8084; PROXY 4.5.6.7:8085
}
