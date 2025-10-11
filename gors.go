package gors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
)

type Request struct {
	baseURL string
	Method  string
	Path    string
	Query   map[string]string
	Body    []byte
	Headers map[string]string
}

type Client struct {
	BaseURL        string
	DefaultHeaders map[string]string
}

func (c *Client) SetDefaultHeaders(h map[string]string) {
	c.DefaultHeaders = h
}

func (c *Client) AddDefaultHeader(key string, value interface{}) {
	if c.DefaultHeaders == nil {
		c.DefaultHeaders = make(map[string]string)
	}

	c.DefaultHeaders[key] = fmt.Sprintf("%v", value)
}

func (c Client) NewRequest(method string, path string) *Request {
	request := Request{
		baseURL: c.BaseURL,
		Method:  method, Path: path,
		Query:   make(map[string]string),
		Headers: make(map[string]string),
	}

	for k, v := range c.DefaultHeaders {
		request.SetHeader(k, v)
	}

	return &request
}

func (r *Request) SetHeader(key string, value interface{}) {
	r.Headers[key] = fmt.Sprintf("%v", value)
}

func (r *Request) SetQuery(key string, value interface{}) {
	r.Query[key] = fmt.Sprintf("%v", value)
}

func (r *Request) SetBody(body []byte) {
	r.Body = body
}

func (r *Request) SetJSONBody(v interface{}) error {
	j, err := json.Marshal(v)

	if err != nil {
		return err
	}

	r.Body = j
	r.SetHeader("Content-Type", "application/json")

	return nil
}

func (r *Request) Send() (*http.Response, error) {
	apiURL, _ := url.Parse(r.baseURL)
	apiURL.Path = path.Join(apiURL.Path, r.Path)

	if strings.HasSuffix(r.Path, "/") {
		apiURL.Path = fmt.Sprintf("%s/", apiURL.Path)
	}

	payloadBuffer := bytes.NewBuffer(r.Body)

	client := http.Client{Timeout: time.Duration(10 * time.Second)}
	req, err := http.NewRequest(r.Method, apiURL.String(), payloadBuffer)

	if err != nil {
		return nil, err
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	q := req.URL.Query()

	for k, v := range r.Query {
		q.Add(k, v)
	}

	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return res, nil
}

// Unfortunately Go does not support generics with struct methods :-(
// so we need to pass the request as a function parameter.
func SendWithJSONResponse[T any](r *Request) (T, *http.Response, error) {
	res, err := r.Send()

	var j T

	if err != nil {
		return j, res, err
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return j, res, err
	}

	err = json.Unmarshal(body, &j)

	if err != nil {
		return j, res, err
	}

	return j, res, nil
}

func NewClient(baseUrl string) Client {
	return Client{BaseURL: baseUrl}
}
