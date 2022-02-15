package utils

import (
	"net/url"
	"regexp"
	"strings"
)

const fileURIScheme = "file://"

var windowsVolumeRegex = regexp.MustCompile(`\w{1}:/`)

// FilenameResolver resolves filenames, with or without URI, *nix or Windows,
// local or remote.
//
// NOTE:
// - Absolute, and relative paths are supported.
// - `file://` scheme is supported. IT SHOULD BE AN ABSOLUTE PATH:
//   - SEE: https://datatracker.ietf.org/doc/html/rfc1738#section-3.10
//   - SEE: https://datatracker.ietf.org/doc/html/draft-ietf-appsawg-file-scheme-03#section-2
func FilenameResolver(filename string) (string, error) {
	finalFilename := filename

	if strings.Contains(finalFilename, fileURIScheme) {
		parsedFilename, err := url.ParseRequestURI(finalFilename)
		if err != nil {
			return "", err
		}

		finalFilename = parsedFilename.Path

		// Deals with Windows path spec.
		if windowsVolumeRegex.MatchString(finalFilename) {
			finalFilename = strings.TrimPrefix(finalFilename, "/")
		}
	}

	return finalFilename, nil
}
