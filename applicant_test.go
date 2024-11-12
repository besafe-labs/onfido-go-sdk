package onfido_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/besafe-labs/onfido-go-sdk"
	"github.com/besafe-labs/onfido-go-sdk/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateApplicant(t *testing.T) {
	utils.LoadEnv(".env")
	client, teardown, err := setup(os.Getenv("ONFIDO_API_TOKEN"), defaultRetries)
	if err != nil || client == nil {
		t.Fatalf("error setting up client: %v", err)
	}
	defer teardown()

	ctx := context.Background()
	defer cleanupApplicants(ctx, client)

	t.Run("CreateWithoutErrors", func(t *testing.T) {
		payload := onfido.CreateApplicantPayload{
			FirstName: "John",
			LastName:  "Doe",
		}

		applicant, err := client.CreateApplicant(ctx, payload)
		if err != nil {
			assert.FailNowf(t, "error creating applicant", "%v", err)
		}
		assert.NotNil(t, applicant, "applicant should not be nil")
		assert.NotEmpty(t, applicant.ID, "applicant ID should not be empty")
		assert.NotEmpty(t, applicant.Href, "applicant href should not be empty")
		assert.Equal(t, "John", applicant.FirstName, "applicant first name should be John")
		assert.Equal(t, "Doe", applicant.LastName, "applicant last name should be Doe")
		assert.NotNil(t, applicant.CreatedAt, "applicant created at should not be nil")
		assert.Nil(t, applicant.DeleteAt, "applicant delete at should be nil")
	})

	t.Run("ReturnErrorOnInvalidPayload", func(t *testing.T) {
		_, err := client.CreateApplicant(ctx, onfido.CreateApplicantPayload{})
		assert.Error(t, err, "error should not be nil")
		assert.Containsf(t, err.Error(), "validation_error", "error should contain validation_error. got %v", err)

		_, err = client.CreateApplicant(ctx, onfido.CreateApplicantPayload{FirstName: "John"})
		assert.Error(t, err, "error should not be nil")
		assert.Containsf(t, err.Error(), "validation_error", "error should contain validation_error. got %v", err)
		assert.Containsf(t, err.Error(), "last_name", "error should contain last_name. got %v", err)

		_, err = client.CreateApplicant(ctx, onfido.CreateApplicantPayload{LastName: "Doe"})
		assert.Error(t, err, "error should not be nil")
		assert.Containsf(t, err.Error(), "validation_error", "error should contain validation_error. got %v", err)
		assert.Containsf(t, err.Error(), "first_name", "error should contain first_name. got %v", err)
	})
}

func TestRetrieveApplicant(t *testing.T) {
	utils.LoadEnv(".env")
	client, teardown, err := setup(os.Getenv("ONFIDO_API_TOKEN"), defaultRetries)
	if err != nil || client == nil {
		t.Fatalf("error setting up client: %v", err)
	}
	defer teardown()

	ctx := context.Background()
	defer cleanupApplicants(ctx, client)

	t.Run("RetrieveWithoutErrors", func(t *testing.T) {
		payload := onfido.CreateApplicantPayload{
			FirstName: "John",
			LastName:  "Doe",
		}

		applicant, err := client.CreateApplicant(ctx, payload)
		if err != nil {
			assert.FailNowf(t, "error creating applicant", "%v", err)
		}

		fetchedApplicant, err := client.RetrieveApplicant(ctx, applicant.ID)
		if err != nil {
			assert.FailNowf(t, "error fetching applicant", "%v", err)
		}

		assert.NotNil(t, fetchedApplicant, "fetched applicant should not be nil")
		assert.Equal(t, applicant.ID, fetchedApplicant.ID, "fetched applicant ID should be the same as created applicant")
	})

	t.Run("ReturnErrorOnInvalidApplicantID", func(t *testing.T) {
		_, err := client.RetrieveApplicant(ctx, "invalid-id")
		assert.Error(t, err, "error should not be nil")
		assert.Containsf(t, err.Error(), "resource_not_found", "error should contain not_found. got %v", err)
	})
}

func TestListApplicants(t *testing.T) {
	utils.LoadEnv(".env")
	client, teardown, err := setup(os.Getenv("ONFIDO_API_TOKEN"), defaultRetries)
	if err != nil || client == nil {
		t.Fatalf("error setting up client: %v", err)
	}
	defer teardown()

	ctx := context.Background()

	cleanupApplicants(ctx, client)
	defer cleanupApplicants(ctx, client)

	t.Run("ListWithoutErrors", func(t *testing.T) {
		payload := onfido.CreateApplicantPayload{
			FirstName: "John",
			LastName:  "Doe",
		}

		_, err := client.CreateApplicant(ctx, payload)
		if err != nil {
			assert.FailNowf(t, "error creating applicant", "%v", err)
		}

		applicants, _, err := client.ListApplicants(ctx)
		if err != nil {
			assert.FailNowf(t, "error fetching applicants", "%v", err)
		}

		assert.NotNil(t, applicants, "applicants should not be nil")
		assert.NotEmpty(t, applicants, "applicants should not be empty")
	})

	t.Run("ListWithPagination", func(t *testing.T) {
		// Create 6 applicants, assuming that there are no other applicants in the account
		createPayload := []onfido.CreateApplicantPayload{
			{FirstName: "John", LastName: "Doe"},
			{FirstName: "Alice", LastName: "Bob"},
			{FirstName: "Jane", LastName: "Doe"},
			{FirstName: "Bob", LastName: "Alice"},
			{FirstName: "Doe", LastName: "John"},
			{FirstName: "Doe", LastName: "Jane"},
		}

		for _, payload := range createPayload {
			_, err := client.CreateApplicant(ctx, payload)
			if err != nil {
				assert.FailNowf(t, "error creating applicant", "%v", err)
			}
		}

		opts := []onfido.IsListApplicantOption{
			onfido.WithPage(1),
			onfido.WithPageLimit(1),
		}

		applicants, _, err := client.ListApplicants(ctx, opts...)
		if err != nil {
			assert.FailNowf(t, "error fetching applicants", "%v", err)
		}

		assert.NotNil(t, applicants, "applicants should not be nil")
		assert.NotEmpty(t, applicants, "applicants should not be empty")
		assert.Equal(t, 1, len(applicants), "applicants length should be 1")
	})

	// t.Run("ListWithInvalidPagination", func(t *testing.T) {
	// 	opts := []onfido.OnfidoPaginationOption{
	// 		onfido.WithPage(0),
	// 		onfido.WithPerPage(0),
	// 	}
	//
	// 	_, err := client.ListApplicants(ctx, opts...)
	// 	assert.Error(t, err, "error should not be nil")
	// 	assert.Contains(t, err.Error(), "validation_error", "error should contain validation_error")
	// })
}

func TestUpdateApplicant(t *testing.T) {
	utils.LoadEnv(".env")
	client, teardown, err := setup(os.Getenv("ONFIDO_API_TOKEN"), defaultRetries)
	if err != nil || client == nil {
		t.Fatalf("error setting up client: %v", err)
	}
	defer teardown()

	ctx := context.Background()
	defer cleanupApplicants(ctx, client)

	t.Run("UpdateWithoutErrors", func(t *testing.T) {
		payload := onfido.CreateApplicantPayload{
			FirstName: "John",
			LastName:  "Doe",
		}

		applicant, err := client.CreateApplicant(ctx, payload)
		if err != nil {
			assert.FailNowf(t, "error creating applicant", "%v", err)
		}

		updatedPayload := onfido.CreateApplicantPayload{
			FirstName: "Alice",
			LastName:  "Bob",
		}

		updatedApplicant, err := client.UpdateApplicant(ctx, applicant.ID, updatedPayload)
		if err != nil {
			assert.FailNowf(t, "error updating applicant", "%v", err)
		}
		assert.NotNil(t, updatedApplicant, "updated applicant should not be nil")
		assert.Equal(t, "Alice", updatedApplicant.FirstName, "updated applicant first name should be Alice")
		assert.Equal(t, "Bob", updatedApplicant.LastName, "updated applicant last name should be Bob")
	})

	t.Run("ReturnErrorOnInvalidPayload", func(t *testing.T) {
		payload := onfido.CreateApplicantPayload{
			FirstName: "John",
			LastName:  "Doe",
		}

		applicant, err := client.CreateApplicant(ctx, payload)
		if err != nil {
			assert.FailNowf(t, "error creating applicant", "%v", err)
		}

		_, err = client.UpdateApplicant(ctx, applicant.ID, onfido.CreateApplicantPayload{Email: "invalid-email"})
		assert.Error(t, err, "error should not be nil")
		assert.Containsf(t, err.Error(), "validation_error", "error should contain validation_error. got %v", err)
		assert.Containsf(t, err.Error(), "email", "error should contain email. got %v", err)
	})

	t.Run("ReturnErrorOnInvalidApplicantID", func(t *testing.T) {
		_, err := client.UpdateApplicant(ctx, "invalid-id", onfido.CreateApplicantPayload{})
		fmt.Println(err)
		assert.Error(t, err, "error should not be nil")
		assert.Contains(t, err.Error(), "resource_not_found", "error should contain not_found")
	})
}

func TestDeleteApplicant(t *testing.T) {
	utils.LoadEnv(".env")
	client, teardown, err := setup(os.Getenv("ONFIDO_API_TOKEN"), defaultRetries)
	if err != nil || client == nil {
		t.Fatalf("error setting up client: %v", err)
	}
	defer teardown()

	ctx := context.Background()
	defer cleanupApplicants(ctx, client)

	t.Run("DeleteWithoutErrors", func(t *testing.T) {
		payload := onfido.CreateApplicantPayload{
			FirstName: "John",
			LastName:  "Doe",
		}

		applicant, err := client.CreateApplicant(ctx, payload)
		if err != nil {
			assert.FailNowf(t, "error creating applicant", "%v", err)
		}

		err = client.DeleteApplicant(ctx, applicant.ID)
		if err != nil {
			assert.FailNowf(t, "error deleting applicant", "%v", err)
		}

		applicant, err = client.RetrieveApplicant(ctx, applicant.ID)
		assert.Error(t, err, "error should not be nil")
		assert.Containsf(t, err.Error(), "scheduled for deletion", "error should contain not_found. got %v", err)
		assert.Nil(t, applicant, "applicant should be nil")
	})

	t.Run("ReturnErrorOnInvalidApplicantID", func(t *testing.T) {
		err := client.DeleteApplicant(ctx, "invalid-id")
		assert.Error(t, err, "error should not be nil")
		assert.Contains(t, err.Error(), "resource_not_found", "error should contain not_found")
	})
}

func TestRestoreApplicant(t *testing.T) {
	utils.LoadEnv(".env")
	client, teardown, err := setup(os.Getenv("ONFIDO_API_TOKEN"), defaultRetries)
	if err != nil || client == nil {
		t.Fatalf("error setting up client: %v", err)
	}
	defer teardown()

	ctx := context.Background()
	defer cleanupApplicants(ctx, client)

	t.Run("RestoreWithoutErrors", func(t *testing.T) {
		payload := onfido.CreateApplicantPayload{
			FirstName: "John",
			LastName:  "Doe",
		}

		applicant, err := client.CreateApplicant(ctx, payload)
		if err != nil {
			assert.FailNowf(t, "error creating applicant", "%v", err)
		}

		err = client.DeleteApplicant(ctx, applicant.ID)
		if err != nil {
			assert.FailNowf(t, "error deleting applicant", "%v", err)
		}

		err = client.RestoreApplicant(ctx, applicant.ID)
		if err != nil {
			assert.FailNowf(t, "error restoring applicant", "%v", err)
		}

		applicant, err = client.RetrieveApplicant(ctx, applicant.ID)
		if err != nil {
			assert.FailNowf(t, "error fetching applicant", "%v", err)
		}
		assert.NotNil(t, applicant, "applicant should not be nil")
		assert.Nil(t, applicant.DeleteAt, "applicant delete at should be nil")
	})

	t.Run("ReturnErrorOnInvalidApplicantID", func(t *testing.T) {
		err := client.RestoreApplicant(ctx, "invalid-id")
		assert.Error(t, err, "error should not be nil")
		assert.Contains(t, err.Error(), "resource_not_found", "error should contain not_found")
	})
}

func cleanupApplicants(ctx context.Context, client *onfido.Client) {
	fmt.Println("------- fetching applicants for cleanup -------")
	applicants, page, err := client.ListApplicants(ctx, onfido.WithPageLimit(500))
	if err != nil {
		log.Fatalf("error fetching applicants: %v\n", err)
		return
	}

	fmt.Printf("----- found %d applicants. deleting... -----\n", len(applicants))

	for _, applicant := range applicants {
		err = client.DeleteApplicant(ctx, applicant.ID)
		if err != nil {
			log.Printf("error deleting applicant: %v\n", err)
		} else {
			fmt.Printf("deleted applicant: %s\n", applicant.ID)
		}
	}

	if page.NextPage != nil {
		fmt.Println("----- fetching next page -----")
		cleanupApplicants(ctx, client)
	}

	fmt.Println("----- applicants cleanup done -----")
}
