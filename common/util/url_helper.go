package util

import "regexp"

var scrubUserinfoRegex = regexp.MustCompile(`(?i)(https?://)[^/\s?#]*@`)

func RedactBasicAuth(url string) string {
	redactedUrl := scrubUserinfoRegex.ReplaceAllString(url, "$1<redacted>:<redacted>@")
	return redactedUrl
}
