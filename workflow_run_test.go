package onfido_test

import (
	"os"
	"strings"
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

	testWorkflowRun := &onfido.WorkflowRun{}
	linkExpiry := time.Now().Add(30 * time.Minute).Truncate(time.Second).UTC()

	t.Run("CreateWorkflowRun", testCreateWorkflowRun(run, applicant, linkExpiry, testWorkflowRun))
	if testWorkflowRun.ID == "" {
		t.Fatalf("workflow run ID is empty")
	}
	t.Run("RetrieveWorkflowRun", testRetrieveWorkflowRun(run, testWorkflowRun.ID))
	t.Run("RetrieveWorkflowRunEvidenceSummaryFile", testRetrieveWorkflowRunEvidenceSummaryFile(run, testWorkflowRun.ID))
	linkExpiry = time.Now().Add(30 * time.Minute).Truncate(time.Second).UTC()
	t.Run("ListWorkflowRuns", testListWorkflowRuns(run, linkExpiry))
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

func testListWorkflowRuns(run *testRun, linkExpiry time.Time) func(*testing.T) {
	tests := []testCase[interface{}]{
		{
			name: "ListWithPagination",
		},
		{
			name: "ListWithTags",
		},
		{
			name: "ListWithStatus",
		},
		{
			name: "ListWithDateRange",
		},
		{
			name: "ListWithSortAsc",
		},
		{
			name: "ListWithSortDesc",
		},
		{
			name: "ListWithMultipleFilters",
		},
	}

	return func(t *testing.T) {
		sleep(t, 10)
		t.Log("Cleaning up applicants")
		if err := cleanupApplicants(run.ctx, run.client); err != nil {
			t.Fatalf("error cleaning up applicants: %v", err)
		}
		t.Log("Creating test applicants")
		sleep(t, 5)
		applicants, err := createTestApplicants(run)
		if err != nil {
			t.Fatalf("error creating test applicants: %v", err)
		}
		t.Log("Creating test workflow runs")
		sleep(t, 10)
		if _, err := createTestWorkflowRuns(run, applicants, linkExpiry); err != nil {
			t.Fatalf("error creating test workflow runs: %v", err)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				isPagination := strings.Contains(tt.name, "Pagination")
				isWithTags := strings.Contains(tt.name, "Tags")
				isWithStatus := strings.Contains(tt.name, "Status")
				isWithDateRange := strings.Contains(tt.name, "DateRange")
				isWithSort := strings.Contains(tt.name, "Sort")
				isWithMultipleFilters := strings.Contains(tt.name, "MultipleFilters")

				switch {
				case isPagination:
					workflowRuns, page, err := run.client.ListWorkflowRuns(run.ctx, onfido.WithPage(1))
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assert.NotNilf(t, workflowRuns, "expected workflow runs to be fetched. got %v", workflowRuns)
					assert.Equalf(t, 6, len(workflowRuns), "expected workflow runs to be 6. got %v", workflowRuns)
					if page == nil {
						t.Fatalf("expected page details to be fetched")
					}
					assert.Equalf(t, 1, *page.FirstPage, "expected first page to be 1. got %v", page.FirstPage)
					assert.Equalf(t, 1, *page.LastPage, "expected last page to be 1. got %v", page.LastPage)

					workflowRuns, page, err = run.client.ListWorkflowRuns(run.ctx, onfido.WithPage(2))
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assert.Equalf(t, 0, len(workflowRuns), "expected no workflow runs to be fetched. got %v", workflowRuns)
					assert.NotNilf(t, page, "expected page details to be fetched")
					if page == nil {
						t.Fatalf("expected page details to be fetched")
					}
					assert.Equalf(t, 1, *page.FirstPage, "expected first page to be 1. got %v", page.FirstPage)
					assert.Equalf(t, 1, *page.LastPage, "expected last page to be 1. got %v", page.LastPage)
					assert.Equalf(t, 1, *page.PrevPage, "expected prev page to be 1. got %v", page.PrevPage)

				case isWithTags:
					workflowRuns, _, err := run.client.ListWorkflowRuns(run.ctx,
						onfido.WithWorkflowRunTags("test"))
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assert.NotNilf(t, workflowRuns, "expected workflow runs to be fetched")
					assert.GreaterOrEqual(t, len(workflowRuns), 1, "expected at least one workflow run to be fetched")
					for _, run := range workflowRuns {
						assert.Contains(t, run.Tags, "test", "expected workflow run to have test tag")
					}

				case isWithStatus:
					workflowRuns, _, err := run.client.ListWorkflowRuns(run.ctx,
						onfido.WithWorkflowRunStatus(onfido.WorkflowRunStatusProcessing))
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assert.NotNilf(t, workflowRuns, "expected workflow runs to be fetched")
					assert.GreaterOrEqual(t, len(workflowRuns), 1, "expected at least one workflow run to be fetched")
					for _, run := range workflowRuns {
						assert.Equal(t, onfido.WorkflowRunStatusProcessing, run.Status,
							"expected workflow run to have processing status")
					}

				case isWithDateRange:
					// Test date range filtering
					now := time.Now()
					yesterday := now.AddDate(0, 0, -1)
					tomorrow := now.AddDate(0, 0, 1)

					workflowRuns, _, err := run.client.ListWorkflowRuns(
						run.ctx,
						onfido.WithWorkflowRunCreatedAfter(yesterday),
						onfido.WithWorkflowRunCreatedBefore(tomorrow),
					)
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assert.NotNilf(t, workflowRuns, "expected workflow runs to be fetched")
					assert.GreaterOrEqual(t, len(workflowRuns), 1, "expected at least one workflow run to be fetched")
					for _, run := range workflowRuns {
						assert.True(t, run.CreatedAt.After(yesterday),
							"expected workflow run to be created after yesterday")
						assert.True(t, run.CreatedAt.Before(tomorrow),
							"expected workflow run to be created before tomorrow")
					}

				case isWithSort && strings.Contains(tt.name, "Asc"):
					workflowRuns, _, err := run.client.ListWorkflowRuns(run.ctx,
						onfido.WithWorkflowRunSort(onfido.SortAsc))
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assert.NotNilf(t, workflowRuns, "expected workflow runs to be fetched")
					assert.GreaterOrEqual(t, len(workflowRuns), 2, "expected at least 2 workflow runs to be fetched")
					// Verify ascending order
					if len(workflowRuns) > 1 {
						for i := 1; i < len(workflowRuns); i++ {
							assert.True(t, workflowRuns[i].CreatedAt.After(*workflowRuns[i-1].CreatedAt) ||
								workflowRuns[i].CreatedAt.Equal(*workflowRuns[i-1].CreatedAt),
								"expected workflow runs to be sorted in ascending order")
						}
					}

				case isWithSort && strings.Contains(tt.name, "Desc"):
					workflowRuns, _, err := run.client.ListWorkflowRuns(run.ctx,
						onfido.WithWorkflowRunSort(onfido.SortDesc))
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assert.NotNilf(t, workflowRuns, "expected workflow runs to be fetched")
					assert.GreaterOrEqual(t, len(workflowRuns), 2, "expected at least 2 workflow runs to be fetched")
					// Verify descending order
					if len(workflowRuns) > 1 {
						for i := 1; i < len(workflowRuns); i++ {
							assert.True(t, workflowRuns[i].CreatedAt.Before(*workflowRuns[i-1].CreatedAt) ||
								workflowRuns[i].CreatedAt.Equal(*workflowRuns[i-1].CreatedAt),
								"expected workflow runs to be sorted in descending order")
						}
					}

				case isWithMultipleFilters:
					// Test combining multiple filters
					now := time.Now()
					yesterday := now.AddDate(0, 0, -1)
					workflowRuns, _, err := run.client.ListWorkflowRuns(run.ctx,
						onfido.WithWorkflowRunTags("test"),
						onfido.WithWorkflowRunStatus(onfido.WorkflowRunStatusProcessing),
						onfido.WithWorkflowRunCreatedAfter(yesterday),
						onfido.WithWorkflowRunSort(onfido.SortDesc),
					)
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assert.NotNilf(t, workflowRuns, "expected workflow runs to be fetched")
					assert.GreaterOrEqual(t, len(workflowRuns), 2, "expected at least 2 workflow run to be fetched")
					// Verify filters are applied
					for _, run := range workflowRuns {
						assert.Contains(t, run.Tags, "test", "expected workflow run to have test tag")
						assert.Equal(t, onfido.WorkflowRunStatusProcessing, run.Status,
							"expected workflow run to have processing status")
						assert.True(t, run.CreatedAt.After(yesterday),
							"expected workflow run to be created after yesterday")
					}
					// Verify descending order
					if len(workflowRuns) > 1 {
						assert.True(t, workflowRuns[1].CreatedAt.Before(*workflowRuns[0].CreatedAt) ||
							workflowRuns[1].CreatedAt.Equal(*workflowRuns[0].CreatedAt),
							"expected workflow runs to be sorted in descending order")
					}
				}
			})
		}
	}
}
