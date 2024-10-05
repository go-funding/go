package utils

import (
	"net/url"
	"slices"
	"strings"
)

func UrlDirname(urlData *url.URL) string {
	return HostDirname(urlData.Host)
}
func HostDirname(host string) string {
	result := strings.Split(host, ".")
	slices.Reverse(result)
	return strings.Join(result, ".")
}
func DirnameHost(urlData string) string {
	hostDirName := urlData
	result := strings.Split(hostDirName, ".")
	slices.Reverse(result)
	return strings.Join(result, ".")
}

func UrlParse(rawUrl string) (*url.URL, error) {
	return url.Parse(rawUrl)
}

func MustUrlParse(rawUrl string) *url.URL {
	u, err := url.Parse(rawUrl)
	if err != nil {
		panic(err)
	}
	return u
}
