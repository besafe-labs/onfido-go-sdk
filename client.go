package onfido

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/besafe-labs/onfido-go-sdk/internal/httpclient"
)

const (
	// CURRENT_CLIENT_VERSION is the current version of the Go-Onfido client
	CURRENT_CLIENT_VERSION = "1.0.0"
	// LATEST_API_VERSION is the latest version of the Onfido API
	LATEST_API_VERSION = "v3.6"
	// DEFAULT_API_REGION is the default region for the Onfido API
	DEFAULT_API_REGION = API_REGION_EU
)

// ------------------------------------------------------------------
//                              CLIENT
// ------------------------------------------------------------------

// Client is a client for the Onfido API
type Client struct {
	client *httpclient.HttpClient

	Endpoint  string
	Retries   int
	RetryWait time.Duration
}

// NewClient creates a new Client
func NewClient(apiToken string, opts ...ClientOption) (*Client, error) {
	if apiToken == "" {
		return nil, fmt.Errorf("apiToken is required")
	}

	options := &clientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	baseURL := fmt.Sprintf("https://api.%s.onfido.com", DEFAULT_API_REGION)
	if options.region != "" {
		baseURL = fmt.Sprintf("https://api.%s.onfido.com", options.region)
	}

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("User-Agent", "Go-Onfido/"+CURRENT_CLIENT_VERSION)
	headers.Set("Authorization", "Token token="+apiToken)

	endpoint := fmt.Sprintf("%s/%s", baseURL, LATEST_API_VERSION)
	client := httpclient.NewHttpClient(endpoint, httpclient.WithHttpHeaders(headers))

	return &Client{client, endpoint, options.retries, options.retryWait}, nil
}

// Close closes the idle connections of the underlying HTTP client.
//
// The client can be reused after closing as per the [http.Client] documentation.
func (c *Client) Close() {
	c.client.Close()
}

func (c *Client) do(ctx context.Context, req func() error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := req(); err != nil {
				return err
			}
			return nil
		}
	}
}

func (c Client) getHttpRequestOptions() httpclient.RequestOption {
	return httpclient.WithHttpRetries(c.Retries, c.RetryWait)
}

func (c Client) getResponseOrError(resp *httpclient.HttpResponse, dest interface{}) error {
	if err := c.getError(resp, false); err != nil {
		return err
	}

	if dest != nil {
		if err := resp.DecodeJSON(dest); err != nil {
			return &OnfidoError{Type: "unknown internal error", Message: err.Error()}
		}
	}

	return nil
}

func (c Client) getError(resp *httpclient.HttpResponse, ingoreFound bool) error {
	// if ingoreFound is true, then we ignore the http.StatusFound status code
	if resp.StatusCode == http.StatusFound && ingoreFound {
		return nil
	}

	// any status code between 200 and 299 is considered a success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var onfidoError struct {
			Error *OnfidoError `json:"error"`
		}
		if err := resp.DecodeJSON(&onfidoError); err != nil {
			return &OnfidoError{Type: "unknown internal error", Message: fmt.Sprintf("OnfidoErrorDecode: %v", err.Error())}
		}
		return onfidoError.Error
	}

	return nil
}

// ------------------------------------------------------------------
//                              OPTIONS
// ------------------------------------------------------------------

type ClientOption func(*clientOptions)

type clientOptions struct {
	retries   int
	retryWait time.Duration
	region    apiRegion
}

func WithRetries(retries int, wait time.Duration) ClientOption {
	return func(c *clientOptions) {
		c.retries = retries
		c.retryWait = wait
	}
}

type apiRegion string

const (
	// API_REGION_EU is the EU region for the Onfido API
	API_REGION_EU apiRegion = "eu"
	// API_REGION_US is the US region for the Onfido API
	API_REGION_US apiRegion = "us"
	// API_REGION_CA is the CA region for the Onfido API
	API_REGION_CA apiRegion = "ca"
)

func WithRegion(region apiRegion) ClientOption {
	return func(c *clientOptions) {
		c.region = region
	}
}

// ------------------------------------------------------------------
//                              PAGINATION
// ------------------------------------------------------------------

type sortDirection string

const (
	// SortAsc is the ascending sort direction
	SortAsc sortDirection = "asc"
	// SortDesc is the descending sort direction
	SortDesc sortDirection = "desc"
)

type PageDetails struct {
	Total     *int
	Limit     *int
	FirstPage *int
	LastPage  *int
	NextPage  *int
	PrevPage  *int
}

type PaginationOption func(*paginationOption)

func (PaginationOption) isListApplicantOption() {}

func (PaginationOption) isListWorkflowRunOption() {}

type paginationOption struct {
	Page int `json:"page"`
}

func (paginationOption) isPaginationOption() {}

func WithPage(page int) PaginationOption {
	return func(p *paginationOption) {
		p.Page = page
	}
}

type LimitPaginationOption func(*limitPaginationOption)

func (LimitPaginationOption) isListApplicantOption() {}

type limitPaginationOption struct {
	PerPage int `json:"per_page"`
}

func (limitPaginationOption) isPaginationOption() {}

func WithPageLimit(limit int) LimitPaginationOption {
	return func(l *limitPaginationOption) {
		l.PerPage = limit
	}
}

type isPaginationOption interface {
	isPaginationOption()
}

func (c Client) getPaginationOptions(opts ...isPaginationOption) (params map[string]string) {
	params = make(map[string]string)

	for _, opt := range opts {
		switch opt := opt.(type) {
		case paginationOption:
			if opt.Page != 0 {
				params["page"] = fmt.Sprintf("%d", opt.Page)
			}
		case limitPaginationOption:
			if opt.PerPage != 0 {
				params["per_page"] = fmt.Sprintf("%d", opt.PerPage)
			}
		}
	}

	return
}

func (c Client) extractPageDetails(headers http.Header) PageDetails {
	pageResponse := PageDetails{}

	total, _ := strconv.Atoi(headers.Get("X-Total-Count"))
	if total != 0 {
		pageResponse.Total = &total
	}

	links := strings.Split(headers.Get("Link"), ",")
	for _, link := range links {
		splitted := strings.Split(link, ">; rel=")
		if len(splitted) != 2 {
			continue
		}
		main, rel := splitted[0], strings.ReplaceAll(splitted[1], "\"", "")

		page, per_page := 0, 0

		// Check if the main string contains "per_page=" and extract the value
		if strings.Contains(main, "per_page=") {
			splittedMain := strings.Split(main, "&per_page=")
			if len(splittedMain) != 2 {
				continue
			}
			per_page, _ = strconv.Atoi(splittedMain[1])
			main = splittedMain[0]
		}

		if per_page != 0 {
			pageResponse.Limit = &per_page
		}

		// extract the page number
		splittedMain := strings.Split(main, "page=")
		if len(splittedMain) != 2 {
			continue
		}
		page, _ = strconv.Atoi(splittedMain[1])

		if page != 0 {
			switch rel {
			case "first":
				pageResponse.FirstPage = &page
			case "last":
				pageResponse.LastPage = &page
			case "next":
				pageResponse.NextPage = &page
			case "prev":
				pageResponse.PrevPage = &page
			}
		}

	}

	return pageResponse
}
