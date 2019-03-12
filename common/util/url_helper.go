package util

import "regexp"

var scrubUserinfoRegex = regexp.MustCompile("(https?://).*:.*@")

func RedactBasicAuth(url string) string {
	redactedUrl := scrubUserinfoRegex.ReplaceAllString(url, "$1<redacted>:<redacted>@")
	return redactedUrl
}
