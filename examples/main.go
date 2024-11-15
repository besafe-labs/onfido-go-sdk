package main

import (
	"context"
	"log"

	"github.com/besafe-labs/onfido-go-sdk"
)

func main() {
	client, err := onfido.NewClient("your-api-token")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Create an applicant
	applicant, err := client.CreateApplicant(context.Background(), onfido.CreateApplicantPayload{
		FirstName: "John",
		LastName:  "Doe",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a workflow run
	workflowRun, err := client.CreateWorkflowRun(context.Background(), onfido.CreateWorkflowRunPayload{
		ApplicantID: applicant.ID,
		WorkflowID:  "your-workflow-id",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Workflow run created: %s", workflowRun.ID)
}
