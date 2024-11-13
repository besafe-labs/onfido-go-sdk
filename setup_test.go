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

func createTestApplicants(run *testRun) error {
	createPayload := []onfido.CreateApplicantPayload{
		{FirstName: "John", LastName: "Doe"},
		{FirstName: "Alice", LastName: "Bob"},
		{FirstName: "Jane", LastName: "Doe"},
		{FirstName: "Bob", LastName: "Alice"},
		{FirstName: "Doe", LastName: "John"},
		{FirstName: "Doe", LastName: "Jane"},
	}

	for _, payload := range createPayload {
		_, err := run.client.CreateApplicant(run.ctx, payload)
		if err != nil {
			return err
		}
	}
	return nil
}

func assertPaginationFirstPage[T any](t *testing.T, data []T, page *onfido.PageDetails, withLimit bool) {
	assert.NotNil(t, page, "expected page to be set")
	assert.Equalf(t, 6, page.Total, "expected total to be 6. got %v", page.Total)
	assert.Nilf(t, page.FirstPage, "expected first page to be nil. got %v", page.FirstPage)
	assert.Nil(t, page.PrevPage, "expected prev page to be nil. got %v", page.PrevPage)
	if withLimit {
		assert.Lenf(t, data, 2, "expected data length to be 2. got %v", len(data))
		assert.NotNil(t, page.Limit, "expected limit to be set")
		assert.Equalf(t, 2, *page.Limit, "expected limit to be 2. got %v", *page.Limit)
		assert.Equalf(t, 3, *page.LastPage, "expected last page to be 3. got %v", *page.LastPage)
		assert.Equalf(t, 2, *page.NextPage, "expected next page to be 2. got %v", *page.NextPage)
	} else {
		assert.Lenf(t, data, 6, "expected data length to be 6. got %v", len(data))
		assert.Nilf(t, page.Limit, "expected limit to be nil. got %v", page.Limit)
		assert.Nilf(t, page.FirstPage, "expected first page to be nil. got %v", page.FirstPage)
		assert.Nilf(t, page.LastPage, "expected last page to be nil. got %v", page.LastPage)
		assert.Nilf(t, page.NextPage, "expected next page to be nil. got %v", page.NextPage)
		assert.Nilf(t, page.PrevPage, "expected prev page to be nil. got %v", page.PrevPage)

	}
}

func assertPaginationSecondPage[T any](t *testing.T, data []T, page *onfido.PageDetails) {
	assert.Lenf(t, data, 2, "expected data length to be 2. got %v", len(data))
	assert.NotNil(t, page, "expected page to be set")
	assert.Equalf(t, 6, page.Total, "expected total to be 6. got %v", page.Total)
	assert.NotNil(t, page.Limit, "expected limit to be set")
	assert.Equalf(t, 2, *page.Limit, "expected limit to be 2. got %v", *page.Limit)
	assert.Equalf(t, 3, *page.LastPage, "expected last page to be 3. got %v", *page.LastPage)
	assert.Equalf(t, 3, *page.NextPage, "expected next page to be 3. got %v", *page.NextPage)
	assert.Equalf(t, 1, *page.FirstPage, "expected first page to be 1. got %v", *page.FirstPage)
	assert.Equalf(t, 1, *page.PrevPage, "expected prev page to be 1. got %v", *page.PrevPage)
}

func assertPaginationLastPage[T any](t *testing.T, data []T, page *onfido.PageDetails) {
	assert.Lenf(t, data, 2, "expected data length to be 2. got %v", len(data))
	assert.NotNil(t, page, "expected page to be set")
	assert.Equalf(t, 6, page.Total, "expected total to be 6. got %v", page.Total)
	assert.NotNil(t, page.Limit, "expected limit to be set")
	assert.Equalf(t, 2, *page.Limit, "expected limit to be 2. got %v", *page.Limit)
	assert.Nilf(t, page.LastPage, "expected last page to be nil. got %v", page.LastPage)
	assert.Nilf(t, page.NextPage, "expected next page to be nil. got %v", page.NextPage)
	assert.NotNilf(t, page.FirstPage, "expected first page to be set. got %v", page.FirstPage)
	assert.NotNilf(t, page.PrevPage, "expected prev page to be set. got %v", page.PrevPage)
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
