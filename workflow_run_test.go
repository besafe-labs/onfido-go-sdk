package onfido_test

import (
	"os"
	"testing"
	"time"

	"github.com/besafe-labs/onfido-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowRun(t *testing.T) {
	run := setupTestRun(t)
	defer run.teardown()

	applicant, err := run.client.CreateApplicant(run.ctx, onfido.CreateApplicantPayload{
		FirstName: "John",
		LastName:  "WorkflowTest",
	})
	if err != nil {
		t.Fatalf("error creating applicant: %v", err)
	}

	linkExpiresAt := time.Now().Add(30 * time.Minute).Truncate(time.Second).UTC()
	testWorkflowRun := &onfido.WorkflowRun{}

	t.Run("CreateWorkflowRun", testCreateWorkflowRun(run, applicant, linkExpiresAt, testWorkflowRun))
	if testWorkflowRun.ID == "" {
		t.Fatalf("workflow run ID is empty")
	}
	t.Run("RetrieveWorkflowRun", testRetrieveWorkflowRun(run, testWorkflowRun.ID))
	t.Run("RetrieveWorkflowRunEvidenceSummaryFile", testRetrieveWorkflowRunEvidenceSummaryFile(run, testWorkflowRun.ID))
	t.Run("ListWorkflowRuns", testListWorkflowRuns(run))

	// t.Run("ListWorkflowRuns", func(t *testing.T) {
	// 	// Test pagination
	// 	tests := []struct {
	// 		name      string
	// 		opts      []onfido.IsListWorkflowRunOption
	// 		wantCount int
	// 		wantErr   bool
	// 	}{
	// 		{
	// 			name:      "List without pagination",
	// 			opts:      nil,
	// 			wantCount: 1, // At least one from our creation test
	// 			wantErr:   false,
	// 		},
	// 		{
	// 			name: "List with pagination",
	// 			opts: []onfido.IsListWorkflowRunOption{
	// 				onfido.WithPage(1),
	// 				onfido.WithPageLimit(2),
	// 			},
	// 			wantCount: 1,
	// 			wantErr:   false,
	// 		},
	// 	}
	//
	// 	for _, tt := range tests {
	// 		t.Run(tt.name, func(t *testing.T) {
	// 			workflowRuns, pageDetails, err := client.ListWorkflowRuns(ctx, tt.opts...)
	// 			if tt.wantErr {
	// 				assert.Error(t, err)
	// 				return
	// 			}
	//
	// 			assert.NoError(t, err)
	// 			assert.NotNil(t, workflowRuns)
	// 			assert.GreaterOrEqual(t, len(workflowRuns), tt.wantCount)
	//
	// 			if len(tt.opts) > 0 {
	// 				assert.NotNil(t, pageDetails)
	// 				assert.NotZero(t, pageDetails.Total)
	// 			}
	// 		})
	// 	}
	// })
}

func testCreateWorkflowRun(run *testRun, applicant *onfido.Applicant, expiry time.Time, setWorkflow *onfido.WorkflowRun) func(*testing.T) {
	// read license file

	tests := []testCase[onfido.CreateWorkflowRunPayload]{
		{
			name: "CreateWithoutErrors",
			input: onfido.CreateWorkflowRunPayload{
				ApplicantID:    applicant.ID,
				WorkflowID:     os.Getenv("ONFIDO_WORKFLOW_ID"),
				Tags:           []string{"test", "integration"},
				CustomerUserID: "customer-user-id",
				Link: &onfido.CreateWorkflowRunLink{
					ExpiresAt: &expiry,
					Language:  "en_GB",
				},
				CustomData: map[string]any{
					"document_id": []any{},
				},
			},
		},
		{
			name:    "ReturnErrorOnEmptyPayload",
			input:   onfido.CreateWorkflowRunPayload{},
			wantErr: true,
			errMsg:  "validation_error",
		},
		{
			name: "ReturnErrorOnMissingLastName",
			input: onfido.CreateWorkflowRunPayload{
				ApplicantID: applicant.ID,
			},
			wantErr: true,
			errMsg:  "workflow_id",
		},
		{
			name: "ReturnErrorOnMissingFirstName",
			input: onfido.CreateWorkflowRunPayload{
				WorkflowID: os.Getenv("ONFIDO_WORKFLOW_ID"),
			},
			wantErr: true,
			errMsg:  "applicant_id",
		},
	}

	return func(t *testing.T) {
		sleep(t, 5)

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				workflow, err := run.client.CreateWorkflowRun(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				if err != nil {
					t.Fatalf(expectedNoError, tt.name, err)
				}

				// Set the workflow run for later tests
				*setWorkflow = *workflow

				assert.NotNil(t, workflow, "expected work to be created")
				assert.Equalf(t, tt.input.ApplicantID, workflow.ApplicantID, "expected workflow applicant_id to be %s, got %s", tt.input.ApplicantID, workflow.ApplicantID)
				assert.Equalf(t, tt.input.WorkflowID, workflow.WorkflowID, "expected workflow_id to be %s, got %s", tt.input.WorkflowID, workflow.WorkflowID)
				assert.Equalf(t, tt.input.Tags, workflow.Tags, "expected workflow tags to be %v, got %v", tt.input.Tags, workflow.Tags)
				assert.Equalf(t, tt.input.CustomerUserID, workflow.CustomerUserID, "expected workflow customer_user_id to be %s, got %s", tt.input.CustomerUserID, workflow.CustomerUserID)
				// expect to be equal stripping milliseconds
				assert.Equalf(t, tt.input.Link.ExpiresAt, workflow.Link.ExpiresAt, "expected workflow link expiry to be %v, got %v", tt.input.Link.ExpiresAt, workflow.Link.ExpiresAt)
				assert.Equalf(t, tt.input.Link.Language, workflow.Link.Language, "expected workflow link language to be %s, got %s", tt.input.Link.Language, workflow.Link.Language)
				assert.NotNil(t, workflow.CreatedAt, "expected created at to be set")
			})
		}
	}
}

func testRetrieveWorkflowRun(run *testRun, workflowId string) func(*testing.T) {
	tests := []testCase[string]{
		{
			name:  "RetrieveWithoutErrors",
			input: workflowId,
		},
		{
			name:    "ReturnErrorOnInvalidID",
			input:   "invalid-id",
			wantErr: true,
			errMsg:  "resource_not_found",
		},
	}
	return func(t *testing.T) {
		sleep(t, 5)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fetchedWorkflowRun, err := run.client.RetrieveWorkflowRun(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				if err != nil {
					t.Fatalf(expectedNoError, tt.name, err)
				}

				assert.NotNil(t, fetchedWorkflowRun, "expected workflowRun to be fetched")
				assert.Equal(t, tt.input, fetchedWorkflowRun.ID, "expected workflowRun ID to be %s, got %s", tt.input, fetchedWorkflowRun.ID)
			})
		}
	}
}

func testRetrieveWorkflowRunEvidenceSummaryFile(run *testRun, workflowRunId string) func(*testing.T) {
	tests := []testCase[string]{
		{
			name:  "RetrieveWithoutErrors",
			input: workflowRunId,
		},
		{
			name:    "ReturnErrorOnInvalidID",
			input:   "invalid-id",
			wantErr: true,
			errMsg:  "resource_not_found",
		},
	}

	return func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				evidenceSummary, err := run.client.RetrieveWorkflowRunEvidenceSummaryFile(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}
				if err != nil {
					t.Fatalf("error retrieving evidence summary: %v", err)
				}

				assert.NotNil(t, evidenceSummary, "expected evidenceSummary to be fetched")
				assert.NotEmptyf(t, evidenceSummary.URL, "expected evidenceSummary url to not be empty")
			})
		}
	}
}

func testListWorkflowRuns(*testRun) func(*testing.T) {
	return func(t *testing.T) {
		t.Skip("Not implemented")
	}
}
