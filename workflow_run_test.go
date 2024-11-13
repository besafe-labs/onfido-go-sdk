package onfido_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/besafe-labs/onfido-go-sdk"
	"github.com/besafe-labs/onfido-go-sdk/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowRun(t *testing.T) {
	utils.LoadEnv(".env")
	client, teardown, err := setup(os.Getenv("ONFIDO_API_TOKEN"), defaultRetries)
	if err != nil || client == nil {
		t.Fatalf("error setting up client: %v", err)
	}
	defer teardown()

	ctx := context.Background()

	var testWorkflowRun *onfido.WorkflowRun

	t.Run("CreateWorkflowRun", func(t *testing.T) {
		applicant, err := client.CreateApplicant(ctx, onfido.CreateApplicantPayload{
			FirstName: "John",
			LastName:  "WorkflowTest",
		})
		if err != nil {
			t.Fatalf("error creating applicant: %v", err)
		}

		linkExpiresAt := time.Now().Add(5 * time.Minute)
		tests := []struct {
			name    string
			payload onfido.CreateWorkflowRunPayload
			wantErr bool
		}{
			{
				name: "Successfully create workflow run",
				payload: onfido.CreateWorkflowRunPayload{
					ApplicantID: applicant.ID,
					WorkflowID:  os.Getenv("ONFIDO_WORKFLOW_ID"), // Set this in your .env
					Tags:        []string{"test", "integration"},
					Link: &onfido.CreateWorkflowRunLink{
						ExpiresAt: &linkExpiresAt,
					},
				},
				wantErr: false,
			},
			{
				name: "Fail with invalid workflow ID",
				payload: onfido.CreateWorkflowRunPayload{
					ApplicantID: applicant.ID,
					WorkflowID:  "invalid-workflow-id",
				},
				wantErr: true,
			},
			{
				name: "Fail with invalid applicant ID",
				payload: onfido.CreateWorkflowRunPayload{
					ApplicantID: "invalid-applicant-id",
					WorkflowID:  os.Getenv("ONFIDO_WORKFLOW_ID"),
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				workflowRun, err := client.CreateWorkflowRun(ctx, tt.payload)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				assert.NoError(t, err)
				assert.NotNil(t, workflowRun)
				assert.NotEmpty(t, workflowRun.ID)
				assert.Equal(t, tt.payload.ApplicantID, workflowRun.ApplicantID)
				assert.Equal(t, tt.payload.WorkflowID, workflowRun.WorkflowID)
				assert.NotNil(t, workflowRun.CreatedAt)
				assert.NotEmpty(t, workflowRun.Status)

				if tt.name == "Successfully create workflow run" {
					testWorkflowRun = workflowRun
				}
			})
		}
	})

	t.Run("RetrieveWorkflowRun", func(t *testing.T) {
		if testWorkflowRun == nil {
			t.Fatal("test workflow run not created")
		}

		tests := []struct {
			name         string
			workflowID   string
			wantErr      bool
			errorMessage string
		}{
			{
				name:       "Successfully retrieve workflow run",
				workflowID: testWorkflowRun.ID,
				wantErr:    false,
			},
			{
				name:         "Fail with invalid workflow run ID",
				workflowID:   "invalid-id",
				wantErr:      true,
				errorMessage: "resource_not_found",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				workflowRun, err := client.RetrieveWorkflowRun(ctx, tt.workflowID)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errorMessage)
					return
				}

				assert.NoError(t, err)
				assert.NotNil(t, workflowRun)
				assert.Equal(t, testWorkflowRun.ID, workflowRun.ID)
				assert.Equal(t, testWorkflowRun.ApplicantID, workflowRun.ApplicantID)
				assert.Equal(t, testWorkflowRun.WorkflowID, workflowRun.WorkflowID)
			})
		}
	})

	t.Run("ListWorkflowRuns", func(t *testing.T) {
		// Test pagination
		tests := []struct {
			name      string
			opts      []onfido.IsListWorkflowRunOption
			wantCount int
			wantErr   bool
		}{
			{
				name:      "List without pagination",
				opts:      nil,
				wantCount: 1, // At least one from our creation test
				wantErr:   false,
			},
			{
				name: "List with pagination",
				opts: []onfido.IsListWorkflowRunOption{
					onfido.WithPage(1),
					onfido.WithPageLimit(2),
				},
				wantCount: 1,
				wantErr:   false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				workflowRuns, pageDetails, err := client.ListWorkflowRuns(ctx, tt.opts...)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				assert.NoError(t, err)
				assert.NotNil(t, workflowRuns)
				assert.GreaterOrEqual(t, len(workflowRuns), tt.wantCount)

				if len(tt.opts) > 0 {
					assert.NotNil(t, pageDetails)
					assert.NotZero(t, pageDetails.Total)
				}
			})
		}
	})

	t.Run("RetrieveWorkflowRunEvidenceSummaryFile", func(t *testing.T) {
		if testWorkflowRun == nil {
			t.Fatal("test workflow run not created")
		}

		// Wait for workflow run to complete to ensure evidence file is available
		maxAttempts := 10
		for i := 0; i < maxAttempts; i++ {
			workflowRun, err := client.RetrieveWorkflowRun(ctx, testWorkflowRun.ID)
			if err != nil {
				t.Fatalf("error retrieving workflow run: %v", err)
			}

			if workflowRun.Status != onfido.WorkflowRunStatusProcessing &&
				workflowRun.Status != onfido.WorkflowRunStatusAwaitingInput {
				break
			}

			if i == maxAttempts-1 {
				t.Fatalf("workflow run did not complete in time")
			}

			time.Sleep(2 * time.Second)
		}

		tests := []struct {
			name         string
			workflowID   string
			wantErr      bool
			errorMessage string
		}{
			{
				name:       "Successfully retrieve evidence summary",
				workflowID: testWorkflowRun.ID,
				wantErr:    false,
			},
			{
				name:         "Fail with invalid workflow run ID",
				workflowID:   "invalid-id",
				wantErr:      true,
				errorMessage: "resource_not_found",
			},
			{
				name:         "Fail with empty workflow run ID",
				workflowID:   "",
				wantErr:      true,
				errorMessage: "workflow run ID is required",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				evidenceSummary, err := client.RetrieveWorkflowRunEvidenceSummaryFile(ctx, tt.workflowID)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errorMessage)
					return
				}

				assert.NoError(t, err)
				assert.NotNil(t, evidenceSummary)
				assert.NotEmpty(t, evidenceSummary.Content)
				assert.Equal(t, "application/pdf", evidenceSummary.ContentType)

				// Optionally save the PDF for manual inspection during testing
				if os.Getenv("SAVE_TEST_FILES") == "true" {
					err = os.WriteFile("test_evidence.pdf", evidenceSummary.Content, 0o644)
					assert.NoError(t, err)
				}
			})
		}
	})
}
