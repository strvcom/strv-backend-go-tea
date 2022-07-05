package spec

import (
	"net/url"
	"path/filepath"
)

// absPath makes a file path absolute and compatible with a URI path component.
//
// The parameter must be a path, not an URI.
func absPath(in string) string {
	anchored, err := filepath.Abs(in)
	if err != nil {
		return in
	}
	return anchored
}

func repairURI(in string) (*url.URL, string) {
	u, _ := url.Parse("")
	return u, ""
}

func fixWindowsURI(u *url.URL, in string) {
}
