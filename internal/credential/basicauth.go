// Copyright 2021 The pacman Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package credential

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/saucelabs/customerror"
	"github.com/saucelabs/pacman/internal/validation"
)

var (
	redactorRegex = regexp.MustCompile(`(?mi).`)

	ErrMissingCredential        = customerror.NewMissingError("credential")
	ErrUsernamePasswordRequired = customerror.NewRequiredError("username, and password are")
)

// BasicAuth is the basic authentication credential definition.
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication#basic_authentication_scheme
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ToBase64 converts a basic auth credential to the base64 format.
func (bC *BasicAuth) ToBase64() string {
	return base64.
		StdEncoding.
		EncodeToString([]byte(bC.Username + ":" + bC.Password))
}

// String is the Stringer interface implementation. Password will be redacted.
func (bC BasicAuth) String() string {
	return fmt.Sprintf("%s:%s",
		bC.Username,
		redactorRegex.ReplaceAllString(bC.Password, "*"),
	)
}

//////
// Factory.
//////

// NewBasicAuthFromText is a BasicAuth factory that automatically parses
// `credential` from text.
func NewBasicAuthFromText(cred string) (*BasicAuth, error) {
	if err := validation.Get().Var(cred, "basicAuth"); err != nil {
		return nil, customerror.NewInvalidError("credential", customerror.WithError(err))
	}

	c := strings.Split(cred, ":")

	return NewBasicAuth(c[0], c[1])
}

// NewBasicAuth is the BasicAuth factory.
func NewBasicAuth(username, password string) (*BasicAuth, error) {
	bC := &BasicAuth{
		Username: username,
		Password: password,
	}

	if err := validation.Get().Var(bC.String(), "basicAuth"); err != nil {
		return nil, customerror.NewInvalidError("credential", customerror.WithError(err))
	}

	return bC, nil
}
