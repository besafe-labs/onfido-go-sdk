package onfido_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/besafe-labs/onfido-go-sdk"
	"github.com/besafe-labs/onfido-go-sdk/internal/utils"
)

const (
	expectedError   = "expected error for %s. got %v"
	expectedNoError = "expected no error for %s. got %v"
	errorContains   = "expected error to have %s. got %v"
)

var defaultRetries = onfido.WithRetries(30, 5*time.Second)

// testCase represents a generic test case structure for applicants
type testCase[T any] struct {
	name    string
	input   T
	setup   func(context.Context, *onfido.Client) (interface{}, error)
	wantErr bool
	errMsg  string
}

// testRun represents a generic test runner
type testRun struct {
	ctx      context.Context
	client   *onfido.Client
	teardown func()
}

// paginatedResponse represents a paginated test runner
// type paginatedResponse[T any] struct {
// 	data []T
// 	page *onfido.PageDetails
// }

func setupClient(token string, opts ...onfido.ClientOption) (client *onfido.Client, teardown func(), err error) {
	client, err = onfido.NewClient(token, opts...)
	if err != nil {
		return
	}

	teardown = func() {
		client.Close()
	}

	return
}

func setupTestRun(t *testing.T) *testRun {
	utils.LoadEnv(".env")
	client, teardown, err := setupClient(os.Getenv("ONFIDO_API_TOKEN"), defaultRetries)
	if err != nil || client == nil {
		t.Fatalf("error setting up client: %v", err)
	}

	ctx := context.Background()
	return &testRun{
		ctx:    ctx,
		client: client,
		teardown: func() {
			cleanupApplicants(ctx, client)
			teardown()
		},
	}
}

func createTestApplicants(run *testRun) ([]*onfido.Applicant, error) {
	createPayload := []onfido.CreateApplicantPayload{
		{FirstName: "John", LastName: "Doe"},
		{FirstName: "Alice", LastName: "Bob"},
		{FirstName: "Jane", LastName: "Doe"},
		{FirstName: "Bob", LastName: "Alice"},
		{FirstName: "Doe", LastName: "John"},
		{FirstName: "Doe", LastName: "Jane"},
	}

	applicants := make([]*onfido.Applicant, len(createPayload))

	for i, payload := range createPayload {
		applicant, err := run.client.CreateApplicant(run.ctx, payload)
		if err != nil {
			return nil, err
		}
		applicants[i] = applicant
	}

	return applicants, nil
}

func createTestWorkflowRuns(run *testRun, applicants []*onfido.Applicant, expiry time.Time) ([]*onfido.WorkflowRun, error) {
	workflowID := os.Getenv("ONFIDO_WORKFLOW_ID")

	workflowRuns := make([]*onfido.WorkflowRun, len(applicants))
	for i, applicant := range applicants {
		workflowRun, err := run.client.CreateWorkflowRun(run.ctx, onfido.CreateWorkflowRunPayload{
			ApplicantID:    applicant.ID,
			WorkflowID:     workflowID,
			Tags:           []string{"test", fmt.Sprintf("workflow-%d", i)},
			CustomerUserID: fmt.Sprintf("customer-user-id-%d", i),
			Link: &onfido.CreateWorkflowRunLink{
				ExpiresAt: &expiry,
				Language:  "en_GB",
			},
			CustomData: map[string]any{
				"document_id": []any{},
			},
		})
		if err != nil {
			return nil, err
		}
		workflowRuns[i] = workflowRun
	}

	return workflowRuns, nil
}

func cleanupApplicants(ctx context.Context, client *onfido.Client) error {
	applicants, page, err := client.ListApplicants(ctx, onfido.WithPageLimit(500))
	if err != nil {
		return err
	}

	for _, applicant := range applicants {
		if err := client.DeleteApplicant(ctx, applicant.ID); err != nil {
			return err
		}
	}

	if page.NextPage != nil {
		if err := cleanupApplicants(ctx, client); err != nil {
			return err
		}
	}

	return nil
}

func sleep(t *testing.T, secs int) {
	// Sleep for a few seconds to allow for eventual consistency
	// in the Onfido API
	// sleep for 5 seconds to avoid rate limiting
	t.Log("------------------------------------------------------------")
	t.Logf("\t\tSleeping for %d seconds to avoid rate limiting", secs)
	time.Sleep(time.Duration(secs) * time.Second)
	t.Log("\t\t--------------------")
	t.Log("\t\tResuming test")
	t.Log("------------------------------------------------------------")
}
