package onfido_test

import (
	"strings"
	"testing"

	"github.com/besafe-labs/onfido-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestApplicant(t *testing.T) {
	run := setupTestRun(t)
	defer run.teardown()

	testApplicant := &onfido.Applicant{}

	t.Run("CreateApplicant", testCreateApplicant(run, testApplicant))
	t.Run("RetrieveApplicant", testRetrieveApplicant(run, testApplicant.ID))
	t.Run("UpdateApplicant", testUpdateApplicant(run, testApplicant.ID))
	t.Run("DeleteApplicant", testDeleteApplicant(run, testApplicant.ID))
	t.Run("RestoreApplicant", testRestoreApplicant(run, testApplicant.ID))
	t.Run("ListApplicants", testListApplicants(run))
}

func testCreateApplicant(run *testRun, setTestApplicant *onfido.Applicant) func(*testing.T) {
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
				if err != nil {
					t.Fatalf("error creating applicant: %v", err)
				}

				// Set the test applicant for later tests
				*setTestApplicant = *applicant

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, applicant, "expected applicant to be created")
				assert.NotEmpty(t, applicant.ID, "expected applicant ID to be set")
				assert.NotEmpty(t, applicant.Href, "expected applicant href to be set")
				assert.Equalf(t, tt.input.FirstName, applicant.FirstName, "expected first name to be %s. got %s", tt.input.FirstName, applicant.FirstName)
				assert.Equalf(t, tt.input.LastName, applicant.LastName, "expected last name to be %s. got %s", tt.input.LastName, applicant.LastName)
				assert.NotNil(t, applicant.CreatedAt, "expected created at to be set")
			})
		}
	}
}

func testRetrieveApplicant(run *testRun, applicantId string) func(*testing.T) {
	tests := []testCase[string]{
		{
			name:  "RetrieveWithoutErrors",
			input: applicantId,
		},
		{
			name:    "ReturnErrorOnInvalidID",
			input:   "invalid-id",
			wantErr: true,
			errMsg:  "resource_not_found",
		},
		{
			name:    "ReturnErrorOnEmptyID",
			input:   "",
			wantErr: true,
			errMsg:  "validation_error",
		},
	}

	return func(t *testing.T) {
		sleep(t, 5)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fetchedApplicant, err := run.client.RetrieveApplicant(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, fetchedApplicant, "expected applicant to be fetched")
				assert.Equalf(t, tt.input, fetchedApplicant.ID, "expected applicant ID to be %s, got %s", tt.input, fetchedApplicant.ID)
			})
		}
	}
}

// testUpdateApplicant tests the update applicant functionality
func testUpdateApplicant(run *testRun, applicantId string) func(*testing.T) {
	tests := []testCase[struct {
		id      string
		payload onfido.CreateApplicantPayload
	}]{
		{
			name: "UpdateWithoutErrors",
			input: struct {
				id      string
				payload onfido.CreateApplicantPayload
			}{
				id: applicantId,
				payload: onfido.CreateApplicantPayload{
					FirstName: "Alice",
					LastName:  "Bob",
				},
			},
		},
		{
			name: "ReturnErrorOnInvalidPayload",
			input: struct {
				id      string
				payload onfido.CreateApplicantPayload
			}{
				id: applicantId,
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
		{
			name: "ReturnErrorOnEmptyID",
			input: struct {
				id      string
				payload onfido.CreateApplicantPayload
			}{},
			wantErr: true,
			errMsg:  "validation_error",
		},
	}

	return func(t *testing.T) {
		sleep(t, 5)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				updatedApplicant, err := run.client.UpdateApplicant(run.ctx, tt.input.id, tt.input.payload)
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
func testDeleteApplicant(run *testRun, applicantId string) func(*testing.T) {
	tests := []testCase[string]{
		{
			name:  "DeleteWithoutErrors",
			input: applicantId,
		},
		{
			name:    "ReturnErrorOnInvalidID",
			input:   "invalid-id",
			wantErr: true,
			errMsg:  "resource_not_found",
		},
		{
			name:    "ReturnErrorOnEmptyID",
			input:   "",
			wantErr: true,
			errMsg:  "validation_error",
		},
	}

	return func(t *testing.T) {
		sleep(t, 5)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := run.client.DeleteApplicant(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)

				// Verify deletion
				applicant, err := run.client.RetrieveApplicant(run.ctx, tt.input)
				assert.Errorf(t, err, expectedError, tt.name, err)
				assert.Containsf(t, err.Error(), tt.errMsg, errorContains, "scheduled for deletion", err.Error())
				assert.Nil(t, applicant, "expected applicant to be deleted")
			})
		}
	}
}

// testRestoreApplicant tests the restore applicant functionality
func testRestoreApplicant(run *testRun, applicantId string) func(*testing.T) {
	tests := []testCase[string]{
		{
			name:  "RestoreWithoutErrors",
			input: applicantId,
		},
		{
			name:    "ReturnErrorOnInvalidID",
			input:   "invalid-id",
			wantErr: true,
			errMsg:  "resource_not_found",
		},
		{
			name:    "ReturnErrorOnEmptyID",
			input:   "",
			wantErr: true,
			errMsg:  "validation_error",
		},
	}

	return func(t *testing.T) {
		sleep(t, 5)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := run.client.RestoreApplicant(run.ctx, tt.input)
				if tt.wantErr {
					assert.Errorf(t, err, expectedError, tt.name, err)
					assert.Containsf(t, err.Error(), tt.errMsg, errorContains, tt.errMsg, err.Error())
					return
				}

				assert.NoErrorf(t, err, expectedNoError, tt.name, err)

				// Verify restoration
				applicant, err := run.client.RetrieveApplicant(run.ctx, tt.input)
				assert.NoErrorf(t, err, expectedNoError, tt.name, err)
				assert.NotNil(t, applicant, "expected applicant to be restored")
			})
		}
	}
}

func testListApplicants(run *testRun) func(*testing.T) {
	tests := []testCase[interface{}]{
		{
			name: "ListWithoutPagination",
		},
		{
			name: "ListWithPaginationNoLimit",
		},
		{
			name: "ListWithPaginationAndLimit",
		},
		{
			name: "ListWithIncludeDeleted",
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
		if _, err := createTestApplicants(run); err != nil {
			t.Fatalf("error creating test applicants: %v", err)
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				isPagination, isIncludeDeleted := strings.Contains(tt.name, "Pagination"), strings.Contains(tt.name, "IncludeDeleted")
				switch true {
				case isPagination && tt.name == "ListWithoutPagination":
					applicants, _, err := run.client.ListApplicants(run.ctx)
					assert.NoErrorf(t, err, expectedNoError, tt.name, err)
					assert.NotNilf(t, applicants, "expected applicants to be fetched. got %v", applicants)
					assert.NotEmptyf(t, applicants, "expected applicants to be fetched. got %v", applicants)
				case isPagination && strings.Contains(tt.name, "Limit"):
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
					assertPaginationApplicantFirstPage(t, applicants, page, withLimit)

					if withLimit {
						// Second page
						applicants, page, err = run.client.ListApplicants(run.ctx,
							onfido.WithPage(*page.NextPage),
							onfido.WithPageLimit(*page.Limit))
						assert.NoErrorf(t, err, expectedNoError, tt.name, err)
						assertPaginationApplicantSecondPage(t, applicants, page)

						// Last page
						applicants, page, err = run.client.ListApplicants(run.ctx,
							onfido.WithPage(*page.NextPage),
							onfido.WithPageLimit(*page.Limit))
						assert.NoErrorf(t, err, expectedNoError, tt.name, err)
						assertPaginationApplicantLastPage(t, applicants, page)
					}
				case isIncludeDeleted:
					// Cleanup applicants
					if err := cleanupApplicants(run.ctx, run.client); err != nil {
						t.Fatalf("error cleaning up applicants: %v", err)
					}

					opts := []onfido.IsListApplicantOption{
						onfido.WithPage(1),
						onfido.WithPageLimit(6),
						onfido.WithIncludeDeletedApplicants(),
					}

					applicants, _, err := run.client.ListApplicants(run.ctx, opts...)
					if err != nil {
						t.Fatalf("error listing applicants: %v", err)
					}

					assert.NotNil(t, applicants, "expected deleted applicants to be fetched")
					assert.Len(t, applicants, 6, "expected 6 deleted applicants to be fetched")

				}
			})
		}
	}
}

func assertPaginationApplicantFirstPage[T any](t *testing.T, data []T, page *onfido.PageDetails, withLimit bool) {
	if page == nil {
		t.Fatalf("expected page to be set")
	}

	if page.Total == nil {
		t.Fatalf("expected page.Total to be set")
	}
	assert.Equalf(t, 6, *page.Total, "expected total to be 6. got %v", page.Total)

	assert.Nilf(t, page.FirstPage, "expected first page to be nil. got %v", page.FirstPage)
	assert.Nilf(t, page.PrevPage, "expected prev page to be nil. got %v", page.PrevPage)

	if withLimit {
		assert.Equalf(t, 2, len(data), "expected data length to be 2. got %v", len(data))
		if page.Limit == nil {
			t.Fatalf("expected page.Limit to be set")
		}
		assert.Equalf(t, 2, *page.Limit, "expected limit to be 2. got %v", *page.Limit)

		if page.NextPage == nil {
			t.Fatalf("expected page.NextPage to be set")
		}
		assert.Equalf(t, 2, *page.NextPage, "expected next page to be 2. got %v", *page.NextPage)

		if page.LastPage == nil {
			t.Fatalf("expected page.LastPage to be set")
		}
		assert.Equalf(t, 3, *page.LastPage, "expected last page to be 3. got %v", *page.LastPage)
	} else {
		assert.Equalf(t, 6, len(data), "expected data length to be 6. got %v", len(data))
		assert.Nilf(t, page.Limit, "expected limit to be nil. got %v", page.Limit)
		assert.Nilf(t, page.FirstPage, "expected first page to be nil. got %v", page.FirstPage)
		assert.Nilf(t, page.LastPage, "expected last page to be nil. got %v", page.LastPage)
		assert.Nilf(t, page.NextPage, "expected next page to be nil. got %v", page.NextPage)
		assert.Nilf(t, page.PrevPage, "expected prev page to be nil. got %v", page.PrevPage)
	}
}

func assertPaginationApplicantSecondPage[T any](t *testing.T, data []T, page *onfido.PageDetails) {
	assert.Equalf(t, len(data), 2, "expected data length to be 2. got %v", len(data))
	if page == nil {
		t.Fatalf("expected page to be set")
	}

	if page.Total == nil {
		t.Fatalf("expected page.Total to be set")
	}
	assert.Equalf(t, 6, *page.Total, "expected total to be 6. got %v", page.Total)

	if page.Limit == nil {
		t.Fatalf("expected page.Limit to be set")
	}
	assert.Equalf(t, 2, *page.Limit, "expected limit to be 2. got %v", *page.Limit)

	if page.NextPage == nil {
		t.Fatalf("expected page.NextPage to be set")
	}
	assert.Equalf(t, 3, *page.NextPage, "expected next page to be 3. got %v", *page.NextPage)

	if page.PrevPage == nil {
		t.Fatalf("expected page.PrevPage to be set")
	}
	assert.Equalf(t, 1, *page.PrevPage, "expected prev page to be 1. got %v", *page.PrevPage)

	if page.FirstPage == nil {
		t.Fatalf("expected page.FirstPage to be set")
	}
	assert.Equalf(t, 1, *page.FirstPage, "expected first page to be 1. got %v", *page.FirstPage)

	if page.LastPage == nil {
		t.Fatalf("expected page.LastPage to be set")
	}
	assert.Equalf(t, 3, *page.LastPage, "expected last page to be 3. got %v", *page.LastPage)
}

func assertPaginationApplicantLastPage[T any](t *testing.T, data []T, page *onfido.PageDetails) {
	assert.Equalf(t, len(data), 2, "expected data length to be 2. got %v", len(data))
	if page == nil {
		t.Fatalf("expected page to be set")
	}

	if page.Total == nil {
		t.Fatalf("expected page.Total to be set")
	}
	assert.Equalf(t, 6, *page.Total, "expected total to be 6. got %v", page.Total)

	if page.Limit == nil {
		t.Fatalf("expected page.Limit to be set")
	}
	assert.Equalf(t, 2, *page.Limit, "expected limit to be 2. got %v", *page.Limit)

	assert.Nilf(t, page.LastPage, "expected last page to be nil. got %v", page.LastPage)
	assert.Nilf(t, page.NextPage, "expected next page to be nil. got %v", page.NextPage)
	assert.NotNilf(t, page.FirstPage, "expected first page to be set. got %v", page.FirstPage)
	assert.NotNilf(t, page.PrevPage, "expected prev page to be set. got %v", page.PrevPage)
}
