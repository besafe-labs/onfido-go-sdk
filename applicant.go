package onfido

import (
	"context"
	"time"

	"github.com/besafe-labs/onfido-go-sdk/internal/httpclient"
)

// ------------------------------------------------------------------
//                              APPLICANT
// ------------------------------------------------------------------

// Applicant represents an applicant in the Onfido API
type Applicant struct {
	ID          string     `json:"id,omitempty"`
	Email       string     `json:"email,omitempty"`
	Dob         string     `json:"dob,omitempty"`
	IdNumbers   []IdNumber `json:"id_numbers,omitempty"`
	PhoneNumber string     `json:"phone_number,omitempty"`
	FirstName   string     `json:"first_name,omitempty"`
	LastName    string     `json:"last_name,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Href        string     `json:"href,omitempty"`
	Sandbox     bool       `json:"sandbox,omitempty"`
	Address     *Address   `json:"address,omitempty"`
	Location    *Location  `json:"location,omitempty"`
}

type CreateApplicantPayload struct {
	FirstName   string     `json:"first_name,omitempty"`
	LastName    string     `json:"last_name,omitempty"`
	Email       string     `json:"email,omitempty"`
	Dob         time.Time  `json:"dob,omitempty"`
	IdNumbers   []IdNumber `json:"id_numbers,omitempty"`
	PhoneNumber string     `json:"phone_number,omitempty"`
	Consents    []Consent  `json:"consents,omitempty"`
	Address     *Address   `json:"address,omitempty"`
	Location    *Location  `json:"location,omitempty"`
}

type IdNumber struct {
	Type      string `json:"type,omitempty"`
	Value     string `json:"value,omitempty"`
	StateCode string `json:"state_code,omitempty"`
}

type Consent struct {
	Granted bool   `json:"granted,omitempty"`
	Name    string `json:"name,omitempty"`
}

type Location struct {
	IpAddress          string `json:"ip_address,omitempty"`
	CountryOfResidence string `json:"country_of_residence,omitempty"`
}

type Address struct {
	Country        string `json:"country,omitempty"`
	Postcode       string `json:"postcode,omitempty"`
	FlatNumber     string `json:"flat_number,omitempty"`
	BuildingNumber string `json:"building_number,omitempty"`
	BuildingName   string `json:"building_name,omitempty"`
	Street         string `json:"street,omitempty"`
	SubStreet      string `json:"sub_street,omitempty"`
	Town           string `json:"town,omitempty"`
	State          string `json:"state,omitempty"`
	Line1          string `json:"line1,omitempty"`
	Line2          string `json:"line2,omitempty"`
	Line3          string `json:"line3,omitempty"`
}

// ------------------------------------------------------------------
//                           OPTIONS
// ------------------------------------------------------------------

type IsListApplicantOption interface {
	isListApplicantOption()
}

type ListApplicantsOption func(*listApplicantsOptions)

func (ListApplicantsOption) isListApplicantOption() {}

type listApplicantsOptions struct {
	*paginationOption `json:",inline"`
	IncludeDeleted    bool `json:"include_deleted,omitempty"`
}

func WithIncludeDeletedApplicants() ListApplicantsOption {
	return func(o *listApplicantsOptions) {
		o.IncludeDeleted = true
	}
}

// CreateApplicant creates a new applicant in the Onfido API
func (c *Client) CreateApplicant(ctx context.Context, payload CreateApplicantPayload) (*Applicant, error) {
	var applicant Applicant

	req := func() error {
		resp, err := c.client.Post(ctx, "/applicants", payload)
		if err != nil {
			return err
		}

		return c.getResponseOrError(resp, &applicant)
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return &applicant, nil
}

// UpdateApplicant updates an existing applicant in the Onfido API
func (c *Client) UpdateApplicant(ctx context.Context, applicantId string, payload CreateApplicantPayload) (*Applicant, error) {
	var applicant Applicant

	req := func() error {
		resp, err := c.client.Put(ctx, "/applicants/"+applicantId, payload, c.getHttpRequestOptions())
		if err != nil {
			return err
		}

		return c.getResponseOrError(resp, &applicant)
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return &applicant, nil
}

// RetrieveApplicant retrieves an applicant from the Onfido API
func (c *Client) RetrieveApplicant(ctx context.Context, applicantId string) (*Applicant, error) {
	var applicant Applicant

	req := func() error {
		resp, err := c.client.Get(ctx, "/applicants/"+applicantId, c.getHttpRequestOptions())
		if err != nil {
			return err
		}

		return c.getResponseOrError(resp, &applicant)
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return &applicant, nil
}

// ListApplicants retrieves all applicants from the Onfido API
func (c *Client) ListApplicants(ctx context.Context, opts ...IsListApplicantOption) ([]Applicant, *PageDetails, error) {
	var applicants []Applicant
	var pageDetails PageDetails

	req := func() error {
		params := c.getListApplicantParams(opts...)

		resp, err := c.client.Get(ctx, "/applicants", httpclient.WithHttpQueryParams(params), c.getHttpRequestOptions())
		if err != nil {
			return err
		}

		var list struct {
			Applicants []Applicant `json:"applicants"`
		}
		if err := c.getResponseOrError(resp, &list); err != nil {
			return err
		}

		applicants = list.Applicants
		pageDetails = c.extractPageDetails(resp.Headers)
		return nil
	}

	if err := c.do(ctx, req); err != nil {
		return nil, nil, err
	}

	return applicants, &pageDetails, nil
}

// DeleteApplicant deletes an applicant from the Onfido API
func (c *Client) DeleteApplicant(ctx context.Context, applicantId string) error {
	req := func() error {
		resp, err := c.client.Delete(ctx, "/applicants/"+applicantId, c.getHttpRequestOptions())
		if err != nil {
			return err
		}

		return c.getResponseOrError(resp, nil)
	}

	if err := c.do(ctx, req); err != nil {
		return err
	}

	return nil
}

// RestoreApplicant restores a deleted applicant in the Onfido API
func (c *Client) RestoreApplicant(ctx context.Context, applicantId string) error {
	req := func() error {
		resp, err := c.client.Post(ctx, "/applicants/"+applicantId+"/restore", nil, c.getHttpRequestOptions())
		if err != nil {
			return err
		}

		return c.getResponseOrError(resp, nil)
	}

	if err := c.do(ctx, req); err != nil {
		return err
	}

	return nil
}

func (c Client) getListApplicantParams(opts ...IsListApplicantOption) (params map[string]string) {
	options := &listApplicantsOptions{paginationOption: &paginationOption{}}

	for _, opt := range opts {
		switch opt := opt.(type) {
		case ListApplicantsOption:
			opt(options)
		case PaginationOption:
			opt(options.paginationOption)
		}
	}

	params = c.getPaginationOptions(options.paginationOption)

	if options.IncludeDeleted {
		params["include_deleted"] = "true"
	}

	return
}
