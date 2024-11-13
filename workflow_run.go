package onfido

import (
	"context"
	"fmt"
	"time"

	"github.com/besafe-labs/onfido-go-sdk/internal/httpclient"
)

// ------------------------------------------------------------------
//                              WORKFLOW RUN
// ------------------------------------------------------------------

type WorkflowRun struct {
	ID                string            `json:"id,omitempty"`
	ApplicantID       string            `json:"applicant_id,omitempty"`
	WorkflowID        string            `json:"workflow_id,omitempty"`
	WorkflowVersionID uint              `json:"workflow_version_id"`
	DashboardURL      string            `json:"dashboard_url,omitempty"`
	Status            WorkflowRunStatus `json:"status,omitempty"`
	Tags              []string          `json:"tags,omitempty"`
	CustomerUserID    string            `json:"customer_user_id,omitempty"`
	Output            map[string]any    `json:"output,omitempty"`
	Reasons           []string          `json:"reasons,omitempty"`
	Error             *OnfidoError      `json:"error,omitempty"`
	SDKToken          string            `json:"sdk_token,omitempty"`
	Link              *WorkflowRunLink  `json:"link,omitempty"`
	CreatedAt         *time.Time        `json:"created_at,omitempty"`
	UpdatedAt         *time.Time        `json:"updated_at,omitempty"`
}

type WorkflowRunLink struct {
	URL                   string `json:"url,omitempty"`
	CreateWorkflowRunLink `json:",inline"`
}

type WorkflowRunStatus string

const (
	WorkflowRunStatusProcessing    WorkflowRunStatus = "processing"
	WorkflowRunStatusAwaitingInput WorkflowRunStatus = "awaiting_input"
	WorkflowRunStatusApproved      WorkflowRunStatus = "approved"
	WorkflowRunStatusDeclined      WorkflowRunStatus = "declined"
	WorkflowRunStatusReview        WorkflowRunStatus = "review"
	WorkflowRunStatusAbandoned     WorkflowRunStatus = "abandoned"
	WorkflowRunStatusError         WorkflowRunStatus = "error"
)

type CreateWorkflowRunPayload struct {
	ApplicantID    string                 `json:"applicant_id,omitempty"`
	WorkflowID     string                 `json:"workflow_id,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	CustomerUserID string                 `json:"customer_user_id,omitempty"`
	Link           *CreateWorkflowRunLink `json:"link,omitempty"`
	CustomData     map[string]any         `json:"custom_data,omitempty"`
}

type CreateWorkflowRunLink struct {
	CompletedRedirectURL string     `json:"completed_redirect_url,omitempty"`
	ExpiredRedirectURL   string     `json:"expired_redirect_url,omitempty"`
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
	Language             string     `json:"language,omitempty"`
}

// WorkflowRunEvidenceSummary represents the evidence summary file response
type WorkflowRunEvidenceSummary struct {
	Content     []byte // Raw PDF content
	ContentType string // Content type of the file
}

// ------------------------------------------------------------------
//                              OPTIONS
// ------------------------------------------------------------------

type IsListWorkflowRunOption interface {
	isListWorkflowRunOption()
}

type ListWorkflowRunOption func(*listWorkflowRunOptions)

func (ListWorkflowRunOption) isListWorkflowRunOption() {}

type listWorkflowRunOptions struct {
	*paginationOption
}

func (c *Client) CreateWorkflowRun(ctx context.Context, payload CreateWorkflowRunPayload) (*WorkflowRun, error) {
	var workflowRun WorkflowRun

	req := func() error {
		resp, err := c.client.Post(ctx, "/workflow_runs", payload)
		if err != nil {
			return err
		}

		return c.getResponseOrError(resp, &workflowRun)
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return &workflowRun, nil
}

func (c *Client) RetrieveWorkflowRun(ctx context.Context, workflowRunID string) (*WorkflowRun, error) {
	var workflowRun WorkflowRun

	req := func() error {
		resp, err := c.client.Get(ctx, "/workflow_runs/"+workflowRunID, c.getHttpRequestOptions())
		if err != nil {
			return err
		}

		return c.getResponseOrError(resp, &workflowRun)
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return &workflowRun, nil
}

func (c *Client) ListWorkflowRuns(ctx context.Context, opts ...IsListWorkflowRunOption) ([]WorkflowRun, *PageDetails, error) {
	var workflowRuns []WorkflowRun
	var pageDetails PageDetails

	req := func() error {
		params := c.getListWorkflowRunParams(opts...)

		resp, err := c.client.Get(ctx, "/workflow_runs", httpclient.WithHttpQueryParams(params), c.getHttpRequestOptions())
		if err != nil {
			return err
		}

		pageDetails = c.extractPageDetails(resp.Headers)
		return c.getResponseOrError(resp, &workflowRuns)
	}

	if err := c.do(ctx, req); err != nil {
		return nil, nil, err
	}

	return workflowRuns, &pageDetails, nil
}

// RetrieveWorkflowRunEvidenceSummaryFile retrieves the signed evidence file for a workflow run
// The file is returned as a PDF document
func (c *Client) RetrieveWorkflowRunEvidenceSummaryFile(ctx context.Context, workflowRunID string) (*WorkflowRunEvidenceSummary, error) {
	var evidenceSummary WorkflowRunEvidenceSummary

	req := func() error {
		resp, err := c.client.Get(ctx, "/workflow_runs/"+workflowRunID+"/signed_evidence_file", c.getHttpRequestOptions())
		if err != nil {
			return fmt.Errorf("failed to retrieve evidence summary: %w", err)
		}

		contentType := resp.Headers.Get("Content-Type")
		if contentType != "application/pdf" {
			return fmt.Errorf("unexpected content type: %s", contentType)
		}

		evidenceSummary = WorkflowRunEvidenceSummary{
			Content:     resp.Body,
			ContentType: contentType,
		}

		return nil
	}

	if err := c.do(ctx, req); err != nil {
		return nil, err
	}

	return &evidenceSummary, nil
}

func (c Client) getListWorkflowRunParams(opts ...IsListWorkflowRunOption) (params map[string]string) {
	options := &listWorkflowRunOptions{paginationOption: &paginationOption{}}

	for _, opt := range opts {
		switch opt := opt.(type) {
		case ListWorkflowRunOption:
			opt(options)
		case PaginationOption:
			opt(options.paginationOption)
		}
	}

	params = c.getPaginationOptions(options.paginationOption)

	return
}
