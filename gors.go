// Package gors provides a small, opinionated HTTP client helper
// that simplifies building and sending requests with JSON support
// and convenient default headers/timeout handling.
package gors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// HTTP method string constants to use with Request.Method.
const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
)

// Request represents an HTTP request to be sent by the client.
//
// Fields:
//   - baseURL: internal field set from Client.BaseURL when created
//   - Method: HTTP verb (use provided constants)
//   - Path: path component to append to baseURL
//   - Query: query string key/value pairs
//   - Body: raw request body bytes
//   - Headers: request headers
//   - Timeout: per-request timeout (default 10s)
type Request struct {
	baseURL string
	Method  string
	Path    string
	Query   map[string]string
	Body    []byte
	Headers map[string]string
	Timeout time.Duration // Default is 10 seconds
}

// Client holds configuration for creating requests, primarily the
// BaseURL that will be prepended to request paths and any
// DefaultHeaders that are copied into new requests.
type Client struct {
	BaseURL        string
	DefaultHeaders map[string]string
}

// SetDefaultHeaders replaces the client's default headers map.
// These headers are copied into every Request created by NewRequest.
func (c *Client) SetDefaultHeaders(h map[string]string) {
	c.DefaultHeaders = h
}

// AddDefaultHeader adds or updates a single default header on the Client.
// The value is formatted using fmt.Sprintf("%v").
func (c *Client) AddDefaultHeader(key string, value interface{}) {
	if c.DefaultHeaders == nil {
		c.DefaultHeaders = make(map[string]string)
	}

	c.DefaultHeaders[key] = fmt.Sprintf("%v", value)
}

// NewRequest creates a new Request associated with this Client.
// The returned Request will have default headers copied from the Client
// and a default timeout of 10 seconds.
func (c Client) NewRequest(method string, path string) *Request {
	request := Request{
		baseURL: c.BaseURL,
		Method:  method, Path: path,
		Query:   make(map[string]string),
		Headers: make(map[string]string),
		Timeout: 10 * time.Second,
	}

	for k, v := range c.DefaultHeaders {
		request.SetHeader(k, v)
	}

	return &request
}

// SetTimeout sets the per-request timeout used by Send().
func (r *Request) SetTimeout(d time.Duration) {
	r.Timeout = d
}

// SetHeader sets a header on the Request. The value is formatted with
// fmt.Sprintf("%v").
func (r *Request) SetHeader(key string, value interface{}) {
	r.Headers[key] = fmt.Sprintf("%v", value)
}

// SetQuery sets a URL query parameter for the Request.
func (r *Request) SetQuery(key string, value interface{}) {
	r.Query[key] = fmt.Sprintf("%v", value)
}

// SetBody assigns raw bytes to the request body.
func (r *Request) SetBody(body []byte) {
	r.Body = body
}

// SetJSONBody marshals v to JSON and sets it as the request body.
// It also sets the Content-Type header to application/json.
func (r *Request) SetJSONBody(v interface{}) error {
	j, err := json.Marshal(v)

	if err != nil {
		return err
	}

	r.Body = j
	r.SetHeader("Content-Type", "application/json")

	return nil
}

// SendWithCtx builds and sends the HTTP request using the provided
// context. It constructs the full URL from Request.baseURL + Request.Path,
// applies headers and query parameters, and returns the raw *http.Response.
func (r *Request) SendWithCtx(ctx context.Context) (*http.Response, error) {
	apiURL, _ := url.Parse(r.baseURL)
	apiURL.Path = path.Join(apiURL.Path, r.Path)

	if strings.HasSuffix(r.Path, "/") {
		apiURL.Path = fmt.Sprintf("%s/", apiURL.Path)
	}

	payloadBuffer := bytes.NewBuffer(r.Body)

	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, r.Method, apiURL.String(), payloadBuffer)

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

// Send sends the request using a context with the Request.Timeout value.
// It is a convenience wrapper around SendWithCtx.
func (r *Request) Send() (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	return r.SendWithCtx(ctx)
}

// Unfortunately Go does not support generics with struct methods :-(
// so we need to pass the request as a function parameter.
// SendWithJSONResponse executes the request and attempts to unmarshal the
// response body into the generic type T. It returns the unmarshaled value,
// the raw *http.Response, and an error if any step fails.
//
// Note: the function reads the entire response body into memory before
// unmarshalling, so use with caution for very large responses.
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

// NewClient constructs a Client preconfigured with the provided base URL.
func NewClient(baseUrl string) Client {
	return Client{BaseURL: baseUrl}
}
