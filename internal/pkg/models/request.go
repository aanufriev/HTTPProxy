package models

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	ID      int64
	Method  string
	Host    string
	Scheme  string
	Path    string
	Headers string
	Body    string
}

func ConvertFromHttpRequest(r http.Request) (Request, error) {
	encodedHeaders, err := json.Marshal(r.Header)
	if err != nil {
		return Request{}, err
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return Request{}, err
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if r.URL.Scheme == "" {
		r.URL.Scheme = "https"
	}

	return Request{
		Method:  r.Method,
		Host:    r.Host,
		Scheme:  r.URL.Scheme,
		Path:    r.URL.Path,
		Headers: string(encodedHeaders),
		Body:    string(body),
	}, nil
}

func ConvertToHttpRequest(r Request) (http.Request, error) {
	httpRequest := http.Request{
		Method: r.Method,
		Host:   r.Host,
		URL: &url.URL{
			Scheme: r.Scheme,
			Host:   r.Host,
			Path:   r.Path,
		},
		Body: ioutil.NopCloser(strings.NewReader(r.Body)),
	}

	var headers http.Header
	err := json.Unmarshal([]byte(r.Headers), &headers)
	if err != nil {
		return http.Request{}, err
	}

	for key, values := range headers {
		for _, value := range values {
			httpRequest.Header.Add(key, value)
		}
	}

	return httpRequest, nil
}
