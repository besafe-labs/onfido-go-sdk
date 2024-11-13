package onfido_test

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/besafe-labs/onfido-go-sdk"
	"github.com/besafe-labs/onfido-go-sdk/internal/utils"
	"github.com/stretchr/testify/assert"
)

const (
	expectedError   = "expected error for %s. got %v"
	expectedNoError = "expected no error for %s. got %v"
	errorContains   = "expected error to have %s. got %v"
)

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
type paginatedResponse[T any] struct {
	data []T
	page *onfido.PageDetails
}

func setupTestRun(t *testing.T) *testRun {
	utils.LoadEnv(".env")
	client, teardown, err := setup(os.Getenv("ONFIDO_API_TOKEN"), defaultRetries)
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

func TestApplicant(t *testing.T) {
	run := setupTestRun(t)
	defer run.teardown()

	t.Run("CreateApplicant", testCreateApplicant(run))
	t.Run("RetrieveApplicant", testRetrieveApplicant(run))
	t.Run("UpdateApplicant", testUpdateApplicant(run))
	t.Run("DeleteApplicant", testDeleteApplicant(run))
	t.Run("RestoreApplicant", testRestoreApplicant(run))
	t.Run("ListApplicants", testListApplicants(run))
}

func testCreateApplicant(run *testRun) func(*testing.T) {
	tests := []testCase[onfido.CreateApplicantPayload]{
		{
			name: "CreateWithoutErrors",
			input: onfido.CreateApplicantPayload{
				FirstName: "John",
				LastName:  "Doe",
			},
		},
		{
			name:    "ReturnErrorOnEmptyPayload",
			input:   onfido.CreateApplicantPayload{},
			wantErr: true,
			errMsg:  "validation_error",
		},
		{
			name: "ReturnErrorOnMissingLastName",
			input: onfido.CreateApplicantPayload{
				FirstName: "John",
			},
			wantErr: true,
			errMsg:  "last_name",
		},
		{
			name: "ReturnErrorOnMissingFirstName",
			input: onfido.CreateApplicantPayload{
				LastName: "Doe",
			},
			wantErr: true,
			errMsg:  "first_name",
		},
	}

	return func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				applicant, err := run.client.CreateApplicant(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, applicant, "expected applicant to be created")
				assert.NotEmpty(t, applicant.ID, "expected applicant ID to be set")
				assert.NotEmpty(t, applicant.Href, "expected applicant href to be set")
				assert.Equal(t, tt.input.FirstName, applicant.FirstName, "expected first name to be equal to input")
				assert.Equal(t, tt.input.LastName, applicant.LastName, "expected last name to be equal to input")
				assert.NotNil(t, applicant.CreatedAt, "expected created at to be set")
			})
		}
	}
}

func testRetrieveApplicant(run *testRun) func(*testing.T) {
	tests := []testCase[string]{
		{
			name: "RetrieveWithoutErrors",
			setup: func(ctx context.Context, client *onfido.Client) (interface{}, error) {
				return client.CreateApplicant(ctx, onfido.CreateApplicantPayload{
					FirstName: "John",
					LastName:  "Doe",
				})
			},
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
				var applicantID string
				if tt.setup != nil {
					applicant, err := tt.setup(run.ctx, run.client)
					if err != nil {
						t.Fatalf("error setting up test: %v", err)
					}
					applicantID = applicant.(*onfido.Applicant).ID
				} else {
					applicantID = tt.input
				}

				fetchedApplicant, err := run.client.RetrieveApplicant(run.ctx, applicantID)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoError(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, fetchedApplicant, "expected applicant to be fetched")
				assert.Equal(t, applicantID, fetchedApplicant.ID, "expected applicant.ID to be equal to applicantID")
			})
		}
	}
}

// testUpdateApplicant tests the update applicant functionality
func testUpdateApplicant(run *testRun) func(*testing.T) {
	tests := []testCase[struct {
		id      string
		payload onfido.CreateApplicantPayload
	}]{
		{
			name: "UpdateWithoutErrors",
			setup: func(ctx context.Context, client *onfido.Client) (interface{}, error) {
				return client.CreateApplicant(ctx, onfido.CreateApplicantPayload{
					FirstName: "John",
					LastName:  "Doe",
				})
			},
			input: struct {
				id      string
				payload onfido.CreateApplicantPayload
			}{
				payload: onfido.CreateApplicantPayload{
					FirstName: "Alice",
					LastName:  "Bob",
				},
			},
		},
		{
			name: "ReturnErrorOnInvalidEmail",
			setup: func(ctx context.Context, client *onfido.Client) (interface{}, error) {
				return client.CreateApplicant(ctx, onfido.CreateApplicantPayload{
					FirstName: "John",
					LastName:  "Doe",
				})
			},
			input: struct {
				id      string
				payload onfido.CreateApplicantPayload
			}{
				payload: onfido.CreateApplicantPayload{
					Email: "invalid-email",
				},
			},
			wantErr: true,
			errMsg:  "validation_error",
		},
		{
			name: "ReturnErrorOnInvalidID",
			input: struct {
				id      string
				payload onfido.CreateApplicantPayload
			}{
				id:      "invalid-id",
				payload: onfido.CreateApplicantPayload{},
			},
			wantErr: true,
			errMsg:  "resource_not_found",
		},
	}

	return func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var applicantID string
				if tt.setup != nil {
					applicant, err := tt.setup(run.ctx, run.client)
					if err != nil {
						t.Fatalf("error setting up test: %v", err)
					}
					applicantID = applicant.(*onfido.Applicant).ID
				} else {
					applicantID = tt.input.id
				}

				updatedApplicant, err := run.client.UpdateApplicant(run.ctx, applicantID, tt.input.payload)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, updatedApplicant, "expected applicant to be updated")
				if tt.input.payload.FirstName != "" {
					assert.Equal(t, tt.input.payload.FirstName, updatedApplicant.FirstName, "expected first name to be updated")
				}
				if tt.input.payload.LastName != "" {
					assert.Equal(t, tt.input.payload.LastName, updatedApplicant.LastName, "expected last name to be updated")
				}
			})
		}
	}
}

// testDeleteApplicant tests the delete applicant functionality
func testDeleteApplicant(run *testRun) func(*testing.T) {
	tests := []testCase[string]{
		{
			name: "DeleteWithoutErrors",
			setup: func(ctx context.Context, client *onfido.Client) (interface{}, error) {
				return client.CreateApplicant(ctx, onfido.CreateApplicantPayload{
					FirstName: "John",
					LastName:  "Doe",
				})
			},
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
				var applicantID string
				if tt.setup != nil {
					applicant, err := tt.setup(run.ctx, run.client)
					if err != nil {
						t.Fatalf("error setting up test: %v", err)
					}
					applicantID = applicant.(*onfido.Applicant).ID
				} else {
					applicantID = tt.input
				}

				err := run.client.DeleteApplicant(run.ctx, applicantID)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)

				// Verify deletion
				applicant, err := run.client.RetrieveApplicant(run.ctx, applicantID)
				assert.Errorf(t, err, expectedError, tt.name, err)
				assert.Containsf(t, err.Error(), tt.errMsg, errorContains, "scheduled for deletion", err.Error())
				assert.Nil(t, applicant, "expected applicant to be deleted")
			})
		}
	}
}

// testRestoreApplicant tests the restore applicant functionality
func testRestoreApplicant(run *testRun) func(*testing.T) {
	tests := []testCase[string]{
		{
			name: "RestoreWithoutErrors",
			setup: func(ctx context.Context, client *onfido.Client) (interface{}, error) {
				applicant, err := client.CreateApplicant(ctx, onfido.CreateApplicantPayload{
					FirstName: "John",
					LastName:  "Doe",
				})
				if err != nil {
					return nil, err
				}

				if err = client.DeleteApplicant(ctx, applicant.ID); err != nil {
					return nil, err
				}

				return applicant, nil
			},
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
				var applicantID string
				if tt.setup != nil {
					applicant, err := tt.setup(run.ctx, run.client)
					if err != nil {
						t.Fatalf("error setting up test: %v", err)
					}
					applicantID = applicant.(*onfido.Applicant).ID
				} else {
					applicantID = tt.input
				}

				err := run.client.RestoreApplicant(run.ctx, applicantID)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)

				// Verify restoration
				applicant, err := run.client.RetrieveApplicant(run.ctx, applicantID)
				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, applicant, "expected applicant to be restored")
			})
		}
	}
}

func testListApplicants(run *testRun) func(*testing.T) {
	if err := cleanupApplicants(run.ctx, run.client); err != nil {
		log.Fatalf("error cleaning up applicants: %v", err)
	}
	if err := createTestApplicants(run); err != nil {
		log.Fatalf("error creating test applicants: %v", err)
	}

	tests := []testCase[string]{
		{
			name: "ListWithoutPagination",
			setup: func(ctx context.Context, client *onfido.Client) (interface{}, error) {
				return nil, nil
			},
		},
		{
			name: "ListWithPaginationNoLimit",
			setup: func(ctx context.Context, client *onfido.Client) (interface{}, error) {
				return nil, nil
			},
		},
		{
			name: "ListWithPaginationAndLimit",
			setup: func(ctx context.Context, client *onfido.Client) (interface{}, error) {
				return nil, nil
			},
		},
	}
	return func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := tt.setup(run.ctx, run.client)
				if err != nil {
					t.Fatalf("error setting up test: %v", err)
				}

				if tt.name == "ListWithoutPagination" {
					applicants, _, err := run.client.ListApplicants(run.ctx)
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assert.NotNilf(t, applicants, "expected applicants to be fetched. got %v", applicants)
					assert.NotEmptyf(t, applicants, "expected applicants to be fetched. got %v", applicants)
				} else {

					withLimit := !strings.Contains(tt.name, "NoLimit")

					// Test pagination
					opts := []onfido.IsListApplicantOption{
						onfido.WithPage(1),
					}
					if withLimit {
						opts = append(opts, onfido.WithPageLimit(2))
					}
					// First page
					applicants, page, err := run.client.ListApplicants(run.ctx, opts...)
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assertPaginationFirstPage(t, applicants, page, withLimit)

					if withLimit {
						// Second page
						applicants, page, err = run.client.ListApplicants(run.ctx,
							onfido.WithPage(*page.NextPage),
							onfido.WithPageLimit(*page.Limit))
						assert.NoErrorf(t, err, expectedNoError, tt.name, err)
						assertPaginationSecondPage(t, applicants, page)

						// Last page
						applicants, page, err = run.client.ListApplicants(run.ctx,
							onfido.WithPage(*page.NextPage),
							onfido.WithPageLimit(*page.Limit))
						assert.NoErrorf(t, err, expectedNoError, tt.name, err)
						assertPaginationLastPage(t, applicants, page)
					}
				}
			})
		}
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
