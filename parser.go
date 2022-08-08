// Copyright 2021 The pacman Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package pacman

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/saucelabs/customerror"
	"github.com/saucelabs/pacman/internal/credential"
	"github.com/saucelabs/pacman/internal/utils"
	"github.com/saucelabs/pacman/internal/validation"
	"github.com/saucelabs/sypl"
	"github.com/saucelabs/sypl/fields"
	"github.com/saucelabs/sypl/level"
	"github.com/saucelabs/sypl/options"
)

const defaultRequestTimeout = 3

var l *sypl.Sypl

// Localhost regex. Based on Golang `net.Listen` and `net.LookupStaticHost` tests.
var localhostRegex = regexp.MustCompile(`(?mi)0\.0\.0\.0|127\.0\.0\.1|localhost`)

type ProxiesCredentials map[string]*credential.BasicAuth

//////
// Helpers.
//////

// Checks if the given URI is "localhost".
func IsLocalhost(uri *url.URL) bool {
	return localhostRegex.MatchString(uri.Host)
}

// Compares if two given URIs are "localhost".
func EqualLocalhost(uri1, uri2 *url.URL) bool {
	return IsLocalhost(uri1) && IsLocalhost(uri2)
}

// Loads, validate credential from env var, and set URI's user.
func setCredentialFromEnvVar(envVar string, uri *url.URL, req *http.Request) error {
	credentialFromEnvVar := os.Getenv(envVar)

	if credentialFromEnvVar != "" {
		if err := validation.Get().Var(credentialFromEnvVar, "basicAuth"); err != nil {
			errMsg := fmt.Sprintf("env var (%s)", envVar)

			return customerror.NewInvalidError(errMsg, customerror.WithError(err))
		}

		cred := strings.Split(credentialFromEnvVar, ":")

		uri.User = url.UserPassword(cred[0], cred[1])

		req.SetBasicAuth(cred[0], cred[1])
	}

	return nil
}

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
			return nil, customerror.NewInvalidError("PAC proxy URI", customerror.WithError(err))
		}

		parsedProxyURI, err := url.ParseRequestURI(proxyURI)
		if err != nil {
			return nil, customerror.New(
				"Failed to parse PAC proxy URI",
				customerror.WithStatusCode(http.StatusBadRequest),
				customerror.WithError(err),
			)
		}

		// Should be a valid credential.
		if err := validation.Get().Var(parsedProxyURI.User.String(), "basicAuth"); err != nil {
			return nil, customerror.NewInvalidError("PAC proxy URI", customerror.WithError(err))
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
		return nil, customerror.NewInvalidError("params", customerror.WithError(err))
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

	proxiesURIsEnvVar := os.Getenv("PACMAN_PROXIES_AUTH")
	if proxiesURIsEnvVar != "" {
		proxiesURIsFromEnvVar := strings.Split(proxiesURIsEnvVar, ",")

		if len(proxiesURIsFromEnvVar) > 0 {
			proxiesURIs = proxiesURIsFromEnvVar
		}
	}

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

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return initialize(source, string(buf), proxiesURIs...)
}

// File loader. Optionally, receives a list of proxies URIs which will be
// used to map each proxy to its credential.
//
// NOTE:
// - Absolute, and relative paths are supported.
// - `file://` scheme is supported. IT SHOULD BE AN ABSOLUTE PATH:
//   - SEE: https://datatracker.ietf.org/doc/html/rfc1738#section-3.10
//   - SEE: https://datatracker.ietf.org/doc/html/draft-ietf-appsawg-file-scheme-03#section-2
func fromFile(filename string, proxiesURIs ...string) (*Parser, error) {
	resolvedFilename, err := utils.FilenameResolver(filename)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(resolvedFilename)
	if err != nil {
		return nil, err
	}

	return fromReader(filename, f, proxiesURIs...)
}

// Remote loader (http/https). Optionally, receives a list of proxies URIs which
// will be used to map each proxy to its credential.
func fromURL(uri string, proxiesURIs ...string) (*Parser, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if u.User != nil {
		if err := validation.Get().Var(u.User.String(), "basicAuth"); err != nil {
			return nil, customerror.NewInvalidError("PAC URI credential", customerror.WithError(err))
		}

		password, _ := u.User.Password()

		req.SetBasicAuth(u.User.Username(), password)
	}

	if err := setCredentialFromEnvVar("PACMAN_AUTH", u, req); err != nil {
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

		errMsg := fmt.Sprintf("failed to retrieve PAC content from %s", u.Redacted())

		resp.Body.Close()

		return nil, customerror.New(errMsg, customerror.WithStatusCode(statusCode), customerror.WithError(err))
	}

	// Should only read body if request succeeded.
	return fromReader(uri, resp.Body, proxiesURIs...)
}

// Direct text loader. Optionally, receives a list of proxies URIs which will be
// used to map each proxy to its credential.
func fromText(text string, proxiesURIs ...string) (*Parser, error) {
	return fromReader("text", io.NopCloser(strings.NewReader(text)), proxiesURIs...)
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
			customerror.WithError(err),
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
	// TODO: May be move it to `ParseProxy`.
	for _, parsedProxy := range parsedProxies {
		parsedProxyURI := parsedProxy.GetURI()

		if parsedProxyURI != nil {
			cred, ok := p.proxiesCredentials[parsedProxyURI.Host]
			if ok {
				parsedProxy.uri.User = url.UserPassword(cred.Username, cred.Password)
			} else {
				// NOTE: It deals with localhost synonyms (`127.0.0.1`, and `0.0.0.0`).
				parsedProxyURIPort := parsedProxyURI.Port()

				if IsLocalhost(parsedProxyURI) {
					var lCred *credential.BasicAuth

					if cred, ok := p.proxiesCredentials[fmt.Sprintf("127.0.0.1:%s", parsedProxyURIPort)]; ok {
						lCred = cred
					}

					if cred, ok := p.proxiesCredentials[fmt.Sprintf("localhost:%s", parsedProxyURIPort)]; lCred == nil && ok {
						lCred = cred
					}

					if cred, ok := p.proxiesCredentials[fmt.Sprintf("0.0.0.0:%s", parsedProxyURIPort)]; lCred == nil && ok {
						lCred = cred
					}

					if lCred != nil {
						parsedProxy.uri.User = url.UserPassword(lCred.Username, lCred.Password)
					}
				}
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
//   - `file://` scheme is supported. IT SHOULD BE AN ABSOLUTE PATH:
//   - SEE: https://datatracker.ietf.org/doc/html/rfc1738#section-3.10
//   - SEE: https://datatracker.ietf.org/doc/html/draft-ietf-appsawg-file-scheme-03#section-2
//
// Notes:
//   - Optionally, credentials for each/any proxy specified in the PAC content can
//     be set (`proxiesURIs`) using standard URI format. These credentials will be
//     automatically set when `FindProxy` is called.
//   - URI is: scheme://credential@host/path` where:
//   - `credential` is `username:password`, and is optional
//   - `host` is `hostname:port`, and is optional.
func New(textOrURI string, proxiesURIs ...string) (*Parser, error) {
	l = sypl.NewDefault("pacman", level.Info)

	if err := validation.Get().Var(textOrURI, "pacTextOrURI"); err != nil {
		return nil, customerror.NewInvalidError("params", customerror.WithError(err))
	}

	// Remote loading.
	if strings.HasPrefix(textOrURI, "http://") ||
		strings.HasPrefix(textOrURI, "https://") {
		return fromURL(textOrURI, proxiesURIs...)
	}

	// Directly loading.
	if strings.Contains(textOrURI, "FindProxyForURL") {
		return fromText(textOrURI, proxiesURIs...)
	}

	// File loading.
	return fromFile(textOrURI, proxiesURIs...)
}
