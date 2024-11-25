package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/besafe-labs/onfido-go-sdk/internal/utils"
)

type HttpClient struct {
	baseURL string
	client  *http.Client
	headers http.Header
}

// Create a new HTTP client
func NewHttpClient(baseURL string, opts ...ClientOption) *HttpClient {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	c := &HttpClient{
		baseURL: baseURL,
		client:  client,
		headers: make(http.Header),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Client options
type ClientOption func(*HttpClient)

func WithHttpTimeout(timeout time.Duration) ClientOption {
	return func(c *HttpClient) {
		c.client.Timeout = timeout
	}
}

func WithHttpHeaders(headers http.Header) ClientOption {
	return func(c *HttpClient) {
		if c.headers == nil {
			c.headers = make(http.Header)
		}
		for k, v := range headers {
			c.headers[k] = v
		}
	}
}

// Request options

type requestOptions struct {
	headers     http.Header
	queryParams url.Values
	formData    map[string]formDataEntry
	timeout     time.Duration
	retries     int
	retryWait   time.Duration
}

type formDataEntry struct {
	value    string
	filename string
	content  []byte
}

type RequestOption func(*requestOptions)

func WithHttpQueryParams(params map[string]string) RequestOption {
	return func(o *requestOptions) {
		if o.queryParams == nil {
			o.queryParams = make(url.Values)
		}
		for k, v := range params {
			o.queryParams.Add(k, v)
		}
	}
}

func WithHttpRetries(retries int, wait time.Duration) RequestOption {
	return func(o *requestOptions) {
		o.retries = retries
		o.retryWait = wait
	}
}

func WithRequestHttpHeaders(headers http.Header) RequestOption {
	return func(o *requestOptions) {
		if o.headers == nil {
			o.headers = make(http.Header)
		}
		for k, v := range headers {
			o.headers[k] = v
		}
	}
}

type isHttpBody interface {
	isHttpBody()
}

type MultipartBody struct {
	*multipart.Writer

	body *bytes.Buffer
}

func NewMultipartBody() *MultipartBody {
	var buf bytes.Buffer
	return &MultipartBody{multipart.NewWriter(&buf), &buf}
}

func (MultipartBody) isHttpBody() {}

type UrlEncodedBody struct {
	url.Values
}

func NewUrlEncodedBody() *UrlEncodedBody {
	return &UrlEncodedBody{url.Values{}}
}

func (UrlEncodedBody) isHttpBody() {}

type JsonBody map[string]interface{}

func (JsonBody) isHttpBody() {}

// Request methods
func (c *HttpClient) Get(ctx context.Context, path string, opts ...RequestOption) (*HttpResponse, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, opts...)
}

func (c *HttpClient) Post(ctx context.Context, path string, body isHttpBody, opts ...RequestOption) (*HttpResponse, error) {
	return c.doRequest(ctx, http.MethodPost, path, body, opts...)
}

func (c *HttpClient) Put(ctx context.Context, path string, body isHttpBody, opts ...RequestOption) (*HttpResponse, error) {
	return c.doRequest(ctx, http.MethodPut, path, body, opts...)
}

func (c *HttpClient) Patch(ctx context.Context, path string, body isHttpBody, opts ...RequestOption) (*HttpResponse, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body, opts...)
}

func (c *HttpClient) Delete(ctx context.Context, path string, opts ...RequestOption) (*HttpResponse, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil, opts...)
}

// Close closes the idle connections of the underlying HTTP client.
//
// The client can be reused after closing as per the [http.Client] documentation.
func (c *HttpClient) Close() {
	c.client.CloseIdleConnections()
}

func (c *HttpClient) doRequest(ctx context.Context, method, path string, body isHttpBody, opts ...RequestOption) (*HttpResponse, error) {
	options := &requestOptions{
		headers: make(http.Header),
	}

	for _, opt := range opts {
		opt(options)
	}

	// Set default retry wait time if retries are enabled but no wait time is set
	if options.retries > 0 && options.retryWait == 0 {
		options.retryWait = 2 * time.Second
	}

	reqURL, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if options.queryParams != nil {
		reqURL.RawQuery = options.queryParams.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case *MultipartBody:
			// Handle multipart form data
			if err := v.Close(); err != nil {
				return nil, fmt.Errorf("failed to close multipart writer: %w", err)
			}
			reqBody = v.body
			options.headers.Set("Content-Type", v.FormDataContentType())
		case *UrlEncodedBody:
			// Handle URL-encoded form data
			reqBody = strings.NewReader(v.Encode())
			options.headers.Set("Content-Type", "application/x-www-form-urlencoded")
		default:
			// Handle JSON body
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			reqBody = bytes.NewReader(jsonData)
			options.headers.Set("Content-Type", "application/json")
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range c.headers {
		req.Header[k] = v
	}
	for k, v := range options.headers {
		req.Header[k] = v
	}

	// Execute request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= options.retries; attempt++ {
		// if attempt is not first trial, wait for retryWait time
		if attempt > 0 {
			// For 429, try to use Retry-After header if available
			if resp != nil && resp.StatusCode == 429 {
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, err := strconv.Atoi(retryAfter); err == nil {
						time.Sleep(time.Duration(seconds) * time.Second)
						continue
					}
				}
			}
			time.Sleep(options.retryWait)
		}

		resp, lastErr = c.client.Do(req)
		// if request is not successful and retries are not enabled or max retries reached, break the loop
		if !shouldRetry(resp, lastErr) || attempt >= options.retries {
			break
		}

		if utils.IsTestRun() {
			log.Printf("\033[33m retrying request %s %s, attempt %d\033[0m\n", method, reqURL.String(), attempt+1)
		}

		// Close the response body if the request is going to be retried
		if lastErr == nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", options.retries, lastErr)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	response := &HttpResponse{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
		Request:    resp.Request,
	}

	return response, nil
}

type HttpResponse struct {
	Status     string        `json:"status"`
	StatusCode int           `json:"status_code"`
	Headers    http.Header   `json:"headers"`
	Body       []byte        `json:"body"`
	Request    *http.Request `json:"request"`
}

func (r *HttpResponse) DecodeJSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

func (r *HttpResponse) String() string {
	return string(r.Body)
}

func shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError
}
