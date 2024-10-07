package uurls

import (
	"errors"
	"net/url"
	"strings"
)

var EmptyStringError = errors.New("empty string")

func ParseHost(urlString string) (string, error) {
	simpleURL := strings.Trim(urlString, " \n\t")
	if simpleURL == "" {
		return "", EmptyStringError
	}

	u, err := url.Parse(simpleURL)
	if err != nil {
		return "", err
	}

	return u.Host, nil
}
