package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Request struct {
	ID      int
	Method  string
	Host    string
	Scheme  string
	Path    string
	Headers string
	Body    string
	Params  string
}

func ConvertFromHttpRequest(r http.Request) (Request, error) {
	encodedHeaders, err := json.Marshal(r.Header)
	if err != nil {
		return Request{}, err
	}

	encodedParams, err := json.Marshal(r.URL.Query())
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
		Params:  string(encodedParams),
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

	var params url.Values
	err = json.Unmarshal([]byte(r.Params), &params)
	if err != nil {
		return http.Request{}, err
	}

	query := httpRequest.URL.Query()
	for key, values := range params {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	httpRequest.URL.RawQuery = query.Encode()

	httpRequest.Header = http.Header{}
	for key, values := range headers {
		for _, value := range values {
			httpRequest.Header.Add(key, value)
		}
	}

	return httpRequest, nil
}

func (r Request) StringFromRequest() string {
	result := ""
	result += strconv.Itoa(int(r.ID))
	result += ": "
	result += r.Method
	result += " "
	result += r.Scheme
	result += "://"
	result += r.Host
	result += r.Path
	result += r.strFromEncodedParams()

	return result
}

func (r Request) strFromEncodedParams() string {
	var params url.Values
	err := json.Unmarshal([]byte(r.Params), &params)
	if err != nil {
		return ""
	}

	result := ""
	for key, values := range params {
		result += fmt.Sprintf("%s=%s", key, strings.Join(values, ","))
	}

	if result != "" {
		return "?" + result
	}

	return result
}
