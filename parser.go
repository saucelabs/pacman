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
	"github.com/saucelabs/pacman/internal/credential"
	"github.com/saucelabs/pacman/internal/validation"
	"github.com/saucelabs/sypl"
	"github.com/saucelabs/sypl/fields"
	"github.com/saucelabs/sypl/level"
	"github.com/saucelabs/sypl/options"
)

const defaultRequestTimeout = 3

type ProxiesCredentials map[string]*credential.BasicAuth

var l *sypl.Sypl

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

// Associates a proxy - specified in PAC, with its credential.
//
// `proxiesURIs` is a list of URIs (`scheme://credential@host`) where:
// - `credential` is `username:password`
// - `host` is `hostname:port`.
//
// TODO: Store in `*url.URL` format instead of Credential.
func processProxiesCredentials(proxiesURIs ...string) (ProxiesCredentials, error) {
	// Initialize proxy's credential map
	proxiesCredentials := make(ProxiesCredentials)

	for _, proxyURI := range proxiesURIs {
		// Should be a valid proxy URI.
		if err := validation.Get().Var(proxyURI, "proxyURI"); err != nil {
			return nil, customerror.NewInvalidError("PAC proxy URI", "", err)
		}

		parsedProxyURI, err := url.ParseRequestURI(proxyURI)
		if err != nil {
			return nil, customerror.New(
				"Failed to parse PAC proxy URI",
				"",
				http.StatusBadRequest,
				err,
			)
		}

		// Should be a valid credential.
		if err := validation.Get().Var(parsedProxyURI.User.String(), "basicAuth"); err != nil {
			return nil, customerror.NewInvalidError("PAC proxy URI", "", err)
		}

		c, err := credential.NewBasicAuthFromText(parsedProxyURI.User.String())
		if err != nil {
			return nil, err
		}

		// Map host to credential.
		proxiesCredentials[parsedProxyURI.Host] = c
	}

	return proxiesCredentials, nil
}

// Initializes Goja, parse PAC content, and process proxies credentials.
func initialize(source, content string, proxiesURIs ...string) (*Parser, error) {
	if err := validation.Get().Var(content, "pacTextOrURI"); err != nil {
		return nil, customerror.NewInvalidError("params", "", err)
	}

	vm := goja.New()

	if err := registerBuiltinNatives(vm); err != nil {
		return nil, err
	}

	if err := registerBuiltinJS(vm); err != nil {
		return nil, err
	}

	if _, err := vm.RunString(content); err != nil {
		return nil, err
	}

	l.PrintlnWithOptions(&options.Options{
		Fields: fields.Fields{
			"content": "\n" + content,
		},
	}, level.Trace, "PAC")

	// Associates a proxy - specified in PAC, with its credential - if any.
	var proxiesCredentials ProxiesCredentials

	if proxiesURIs != nil {
		pC, err := processProxiesCredentials(proxiesURIs...)
		if err != nil {
			return nil, err
		}

		proxiesCredentials = pC
	}

	p := &Parser{
		content:            content,
		source:             source,
		vm:                 vm,
		proxiesCredentials: proxiesCredentials,
	}

	l.PrintlnWithOptions(&options.Options{
		Fields: fields.Fields{
			"source":     source,
			"credential": proxiesCredentials,
		},
	}, level.Debug, "Parser created")

	return p, nil
}

// Centralized PAC content reading. Optionally, receives a list of proxies URIs
// which will be used to map each proxy to its credential.
func fromReader(source string, r io.ReadCloser, proxiesURIs ...string) (*Parser, error) {
	defer r.Close()

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return initialize(source, string(buf), proxiesURIs...)
}

// File loader. Optionally, receives a list of proxies URIs which will be
// used to map each proxy to its credential.
func fromFile(filename string, proxiesURIs ...string) (*Parser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return fromReader(filename, f, proxiesURIs...)
}

// Remote loader. Optionally, receives a list of proxies URIs which will be
// used to map each proxy to its credential.
func fromURL(uri string, proxiesURIs ...string) (*Parser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
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

		errMsg := fmt.Sprintf("failed to retrieve PAC content from %s", uri)

		resp.Body.Close()

		return nil, customerror.New(errMsg, "", statusCode, err)
	}

	// Should only read body if request succeeded.
	return fromReader(uri, resp.Body, proxiesURIs...)
}

// Direct text loader. Optionally, receives a list of proxies URIs which will be
// used to map each proxy to its credential.
func fromText(text string, proxiesURIs ...string) (*Parser, error) {
	return fromReader("text", ioutil.NopCloser(strings.NewReader(text)), proxiesURIs...)
}

//////
// Exported.
//////

// Parser definition.
type Parser struct {
	sync.Mutex

	content            string
	proxiesCredentials ProxiesCredentials
	source             string
	vm                 *goja.Runtime
}

// Source of the PAC content.
func (p *Parser) Source() string {
	return p.source
}

// Content returns the PAC content.
func (p *Parser) Content() string {
	return p.content
}

// FindProxyForURL for the given `url`, returning as string, example:
// "PROXY 4.5.6.7:8080; PROXY 7.8.9.10:8080; DIRECT".
func (p *Parser) FindProxyForURL(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	f := fmt.Sprintf("FindProxyForURL('%s', '%s')", uri, u.Hostname())

	// Go routine safe.
	p.Lock()

	r, err := p.vm.RunString(f)

	p.Unlock()

	if err != nil {
		return "", customerror.NewFailedToError(
			"call `FindProxyForURL`. Is that defined?",
			"",
			err,
		)
	}

	return r.String(), nil
}

// FindProxy for the given `url`, returning a list of `Proxies`.
//
// Note: If the returned proxies requires credentials, and it was set when the
// Parser was created (`proxiesURIs`), it will be automatically added to the
// `Proxy`.
func (p *Parser) FindProxy(uri string) ([]Proxy, error) {
	proxiesAsString, err := p.FindProxyForURL(uri)
	if err != nil {
		return nil, err
	}

	parsedProxies, err := ParseProxy(proxiesAsString)
	if err != nil {
		return nil, err
	}

	// Adds credential - if any.
	//
	// TODO: Move it to `ParseProxy`.
	for _, parsedProxy := range parsedProxies {
		if parsedProxy.GetURI() != nil {
			if credential, ok := p.proxiesCredentials[parsedProxy.GetURI().Host]; ok {
				parsedProxy.uri.User = url.UserPassword(credential.Username, credential.Password)
			}
		}

		l.PrintlnfWithOptions(&options.Options{
			Fields: fields.Fields{
				"proxy": parsedProxy.GetURI(),
			},
		}, level.Debug, "Proxy found for %s", uri)
	}

	return parsedProxies, nil
}

//////
// Factory.
//////

// New is able to load PAC from many sources:
// - Direct: `textOrURI` is the PAC content
// - Remote: `textOrURI` is an HTTP/HTTPS URI
// - File: `textOrURI` points to a file:
//   - As per PAC spec, PAC file should have the `.pac` extension
//   - Absolute and relative paths are supported
//   - `file://` scheme is supported. It should be an absolute path.
//
// Notes:
// - Optionally, credentials for each/any proxy specified in the PAC content can
//   be set (`proxiesURIs`) using standard URI format. These credentials will be
//   automatically set when `FindProxy` is called.
// - URI is: scheme://credential@host/path` where:
//   - `credential` is `username:password`, and is optional
//   - `host` is `hostname:port`, and is optional.
func New(textOrURI string, proxiesURIs ...string) (*Parser, error) {
	l = sypl.NewDefault("pacman", level.Info).New("parser")

	logLevelEnvVar := os.Getenv("PACMAN_LOG_LEVEL")

	if logLevelEnvVar != "" {
		l.GetOutput("console").SetMaxLevel(level.MustFromString(logLevelEnvVar))
	}

	if err := validation.Get().Var(textOrURI, "pacTextOrURI"); err != nil {
		return nil, customerror.NewInvalidError("params", "", err)
	}

	// Remote loading.
	if strings.HasPrefix(textOrURI, "http://") ||
		strings.HasPrefix(textOrURI, "https://") {
		return fromURL(textOrURI, proxiesURIs...)
	}

	// File loading.
	if strings.Contains(textOrURI, ".pac") {
		return fromFile(textOrURI, proxiesURIs...)
	}

	// Directly loading.
	return fromText(textOrURI, proxiesURIs...)
}
