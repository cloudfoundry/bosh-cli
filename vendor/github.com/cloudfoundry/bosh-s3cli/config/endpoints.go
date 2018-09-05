package config

import (
	"regexp"
)

var (
	providerRegex = map[string]*regexp.Regexp{
		"aws":      regexp.MustCompile(`(^$|s3[-.]?(.*)\.amazonaws\.com(\.cn)?$)`),
		"alicloud": regexp.MustCompile(`^oss-([a-z]+-[a-z]+(-[1-9])?)(-internal)?.aliyuncs.com$`),
		"google":   regexp.MustCompile(`^storage.googleapis.com$`),
	}
)

func AWSHostToRegion(host string) string {
	regexMatches := providerRegex["aws"].FindStringSubmatch(host)

	region := "us-east-1"

	if len(regexMatches) == 4 && regexMatches[2] != "" && regexMatches[2] != "external-1" {
		region = regexMatches[2]
	}

	return region
}

func AlicloudHostToRegion(host string) string {
	regexMatches := providerRegex["alicloud"].FindStringSubmatch(host)

	if len(regexMatches) == 4 {
		return regexMatches[1]
	}

	return ""
}
