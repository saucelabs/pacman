// Copyright 2021 The pacman Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pacman

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/saucelabs/customerror"
)

const defaultRequestTimeout = 3

//////
// Helpers.
//////

func registerBuiltinNatives(vm *goja.Runtime) error {
	for name, function := range builtinNatives {
		if err := vm.Set(name, function(vm)); err != nil {
			return err
		}
	}

	return nil
}

func registerBuiltinJS(vm *goja.Runtime) error {
	_, err := vm.RunString(builtinJS)

	return err
}

func fromReader(r io.ReadCloser) (*Parser, error) {
	defer r.Close()

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return New(string(buf))
}

//////
// Exported.
//////

// Parser definition.
type Parser struct {
	vm  *goja.Runtime
	src string // the FindProxyForURL source code

	sync.Mutex
}

// FindProxyForURL finding proxy for url returns string like:
// "PROXY 4.5.6.7:8080; PROXY 7.8.9.10:8080; DIRECT".
func (p *Parser) FindProxyForURL(urlstr string) (string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return "", err
	}

	f := fmt.Sprintf("FindProxyForURL('%s', '%s')", urlstr, u.Hostname())

	// Go routine safe.
	p.Lock()

	r, err := p.vm.RunString(f)

	p.Unlock()

	if err != nil {
		return "", customerror.NewFailedToError("call `FindProxyForURL`. Is that defined?", "", err)
	}

	return r.String(), nil
}

// FindProxy find the proxy in pac and return a list of Proxy.
func (p *Parser) FindProxy(urlstr string) ([]Proxy, error) {
	ps, err := p.FindProxyForURL(urlstr)
	if err != nil {
		return nil, err
	}

	return ParseProxy(ps), nil
}

//////
// Factory.
//////

// New create a parser from text content. You may want to call some of the
// loaders (`FromFile`, `FromURL`).
func New(text string) (*Parser, error) {
	const errMsgPrefix = "PAC content"

	if text == "" {
		return nil, customerror.NewMissingError(errMsgPrefix, "", nil)
	}

	if !strings.Contains(text, "FindProxyForURL") {
		return nil, customerror.NewInvalidError(
			errMsgPrefix+". Missing `FindProxyForURL`",
			"",
			nil,
		)
	}

	vm := goja.New()

	if err := registerBuiltinNatives(vm); err != nil {
		return nil, err
	}

	if err := registerBuiltinJS(vm); err != nil {
		return nil, err
	}

	if _, err := vm.RunString(text); err != nil {
		return nil, err
	}

	return &Parser{vm: vm, src: text}, nil
}

//////
// Loaders.
//////

// FromFile load pac from file.
func FromFile(filename string) (*Parser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return fromReader(f)
}

// FromURL load pac from url.
func FromURL(urlstr string) (*Parser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlstr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	statusCode := resp.StatusCode

	// Checks if request succeeded.
	if statusCode != http.StatusOK {
		// Check if status code is valid, otherwise fallback to `500`.
		if statusCode < http.StatusContinue ||
			statusCode > http.StatusNetworkAuthenticationRequired {
			statusCode = http.StatusInternalServerError
		}

		errMsg := fmt.Sprintf("failed to retrieve PAC content from %s", urlstr)

		resp.Body.Close()

		return nil, customerror.New(errMsg, "", statusCode, err)
	}

	// Should only read body if request succeeded.
	return fromReader(resp.Body)
}

// From load pac from file or url.
func From(dst string) (*Parser, error) {
	if strings.HasPrefix(dst, "http://") ||
		strings.HasPrefix(dst, "https://") {
		return FromURL(dst)
	}

	return FromFile(dst)
}
