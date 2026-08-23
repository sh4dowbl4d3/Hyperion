package xss

import "strings"

// bodyForFeed is the single point where the vulnerable and safe rendering
// paths diverge. Vulnerable comments carry no Rendered field, so their raw
// stored body goes into the HTML verbatim — that is the XSS. Safe comments
// always have a server-escaped Rendered field.
func bodyForFeed(cm Comment) string {
	if cm.Rendered != nil {
		return *cm.Rendered
	}
	return cm.Body
}

// HasRawScript reports whether a body would execute as markup (i.e. it still
// contains an unescaped script tag). Used by tests to assert mitigation.
func HasRawScript(body string) bool {
	return strings.Contains(strings.ToLower(body), "<script")
}
