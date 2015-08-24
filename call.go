package main

import (
	"crypto/tls"

	"net/http"
)

var (
	// this http client will verify TLS if used
	httpClientTLSVerify = &http.Client{}

	// this http client will not verify TLS
	httpClientSkipTLSVerify = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
)

func Call(r *http.Request, InsecureSkipVerify bool) (*http.Response, error) {
	if InsecureSkipVerify {
		return httpClientSkipTLSVerify.Do(r)
	} else {
		return httpClientTLSVerify.Do(r)
	}
}
