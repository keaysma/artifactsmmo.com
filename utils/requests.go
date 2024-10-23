package utils

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const URL_BASE = "https://api.artifactsmmo.com"

var settings = GetSettings()

var BaseHeaders = http.Header{
	"Content-Type":  []string{"application/json"},
	"Accept":        []string{"application/json"},
	"Authorization": []string{fmt.Sprintf("Bearer %s", settings.Api_token)},
}

const CLIENT_TIMEOUT_SECONDS = 15

var client = http.Client{
	Timeout: CLIENT_TIMEOUT_SECONDS * time.Second,
}

func AddQueryParams(r *http.Request, params *map[string]string) *http.Request {
	var q = r.URL.Query()

	for key, value := range *params {
		q.Add(key, value)
	}

	r.URL.RawQuery = q.Encode()

	return r
}

func AddHeaders(r *http.Request, headers map[string]string) *http.Request {

	for key, value := range headers {
		r.Header.Add(key, value)
	}

	for key, value := range BaseHeaders {
		r.Header.Set(key, value[0])
	}

	return r
}

func HttpGet(url string, headers map[string]string, params *map[string]string) (*http.Response, error) {
	var r, error = http.NewRequest(
		"GET",
		fmt.Sprintf("%s/%s", URL_BASE, url),
		nil,
	)

	if error != nil {
		return nil, error
	}

	if params != nil {
		r = AddQueryParams(r, params)
	}
	r = AddHeaders(r, headers)

	client.CloseIdleConnections()

	return client.Do(r)
}

func HttpPost(url string, headers map[string]string, body io.Reader) (*http.Response, error) {
	var r, err = http.NewRequest(
		"POST",
		fmt.Sprintf("%s/%s", URL_BASE, url),
		body,
	)

	if err != nil {
		return nil, err
	}

	r = AddHeaders(r, headers)

	client.CloseIdleConnections()

	return client.Do(r)
}
