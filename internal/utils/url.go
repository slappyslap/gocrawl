package utils

import (
	"GoCrawl/internal/log"
	"net/url"
	"regexp"
	"strings"
)

var InvalidHostnames = []string{"localhost", ""}
var InvalidHostnamesMapping = make(map[string]bool)

func init() {
	for _, hostname := range InvalidHostnames {
		InvalidHostnamesMapping[hostname] = true
	}
}

func UrlCleaner(url string) string {
	url = strings.TrimRight(url, "/")
	if strings.Contains(url, "?") {
		url = url[:strings.Index(url, "?")]
	}

	return url
}

func GetHostname(fullUrl string) string {
	if len(fullUrl) == 0 {
		return ""
	}

	u, err := url.ParseRequestURI(fullUrl)
	if err != nil {
		log.Error("illegal url %s, err:%s", fullUrl, err.Error())
		return ""
	}

	return u.Hostname()
}

func IsValidHostname(hostname string) bool {
	if len(hostname) == 0 {
		return false
	}

	if InvalidHostnamesMapping[hostname] {
		return false
	}

	var validHostnamePattern = `^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(validHostnamePattern, hostname)

	if err != nil {
		log.Error("regex match error: %s", err.Error())
		return false
	}

	return matched
}
