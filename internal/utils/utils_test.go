package utils

import "testing"

func TestFilenameResolver(t *testing.T) {
	paths := map[string]string{
		// Paths.
		"/User/john/somefile.json": "/User/john/somefile.json",
		"somefile.json":            "somefile.json",

		// URIs.
		"file:///path/to/file.json":              "/path/to/file.json",
		"file://remotehost/path/to/file.json":    "/path/to/file.json",
		"file://localhost/path/to/file.json":     "/path/to/file.json",
		"file:///c:/WINDOWS/clock.json":          "c:/WINDOWS/clock.json",
		"file://localhost/c:/WINDOWS/clock.avi":  "c:/WINDOWS/clock.avi",
		"file:////remotehost/share/dir/file.txt": "//remotehost/share/dir/file.txt",
	}

	for path, expected := range paths {
		filename, err := FilenameResolver(path)
		if err != nil {
			t.Error(err)
		}

		if filename != expected {
			t.Errorf("Expected %s, got %s", expected, filename)
		}
	}
}
