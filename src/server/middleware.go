package server

import "strings"

type ClientType string

const (
	ClientTypeHTML ClientType = "html"
	ClientTypeText ClientType = "text"
	ClientTypeJSON ClientType = "json"
)

func NormalizeURLPath(inputPath string) (string, bool) {
	if inputPath == "" || inputPath == "/" {
		return "/", false
	}

	if strings.HasSuffix(inputPath, "/") && !isFileLikePath(inputPath) {
		return strings.TrimSuffix(inputPath, "/"), true
	}

	return inputPath, false
}

func DetectClientType(acceptHeader string, userAgent string) ClientType {
	switch {
	case strings.Contains(acceptHeader, "text/html"):
		return ClientTypeHTML
	case strings.Contains(acceptHeader, "text/plain"):
		return ClientTypeText
	case strings.Contains(acceptHeader, "application/json"):
		return ClientTypeJSON
	}

	browserUserAgents := []string{
		"Mozilla/", "Chrome/", "Safari/", "Edge/", "Firefox/",
		"Opera/", "MSIE", "Trident/",
	}
	for _, browserUserAgent := range browserUserAgents {
		if strings.Contains(userAgent, browserUserAgent) {
			return ClientTypeHTML
		}
	}

	cliUserAgents := []string{
		"curl/", "Wget/", "HTTPie/", "python-requests/",
		"Go-http-client/", "node-fetch/",
	}
	for _, cliUserAgent := range cliUserAgents {
		if strings.Contains(userAgent, cliUserAgent) {
			return ClientTypeText
		}
	}

	if userAgent == "" {
		return ClientTypeText
	}

	return ClientTypeHTML
}

func DetectAPIResponseFormat(requestPath string, acceptHeader string, userAgent string) ClientType {
	switch {
	case strings.HasSuffix(requestPath, ".txt"):
		return ClientTypeText
	case strings.Contains(acceptHeader, "text/plain"):
		return ClientTypeText
	case strings.Contains(acceptHeader, "application/json"):
		return ClientTypeJSON
	}

	cliUserAgents := []string{
		"curl/", "Wget/", "HTTPie/", "python-requests/",
		"Go-http-client/", "node-fetch/",
	}
	for _, cliUserAgent := range cliUserAgents {
		if strings.Contains(userAgent, cliUserAgent) {
			return ClientTypeText
		}
	}

	if userAgent == "" {
		return ClientTypeText
	}

	return ClientTypeJSON
}

func isFileLikePath(inputPath string) bool {
	lastSlashIndex := strings.LastIndex(inputPath, "/")
	if lastSlashIndex == -1 {
		return strings.Contains(inputPath, ".")
	}
	return strings.Contains(inputPath[lastSlashIndex:], ".")
}
