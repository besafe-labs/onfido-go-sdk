# Onfido Go SDK

An unofficial Go SDK for interacting with the Onfido API. This SDK was created because Onfido does not provide an official Go SDK for developers.

## Installation

```bash
go get github.com/besafe-labs/onfido-go-sdk
```

## Usage

```go

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

```

## Available Resources

### Applicants

- Create applicant
- Retrieve applicant
- List applicants
- Update applicant
- Delete applicant
- Restore deleted applicant

### Workflow Runs

- Create workflow run
- Retrieve workflow run
- List workflow runs
- Retrieve workflow run evidence summary file

## Features

- Automatic retries with configurable retry count and wait time
- Region-specific endpoints (EU, US, CA)
- Pagination support
- Comprehensive error handling
- Context support for cancellation and timeouts

## Configuration Options

```go
// Configure region
client, err := onfido.NewClient(token, onfido.WithRegion(onfido.API_REGION_US))

// Configure retries
client, err := onfido.NewClient(token, onfido.WithRetries(3, 5*time.Second))
```

## Error Handling

The SDK provides detailed error information through the `OnfidoError` struct:

```go
type OnfidoError struct {
    Type    string
    Message string
    Fields  map[string]any
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Tests

This SDK directly tests with the onfido API. To run the tests, you need to provide your API token as an environment variable:

```bash
# Set environment variables
export ONFIDO_API_TOKEN=your-api-token
export ONFIDO_WORKFLOW_ID=your-workflow-id

# Run tests
go test -v
```

## Related Projects

- [go-onfido](https://github.com/uw-labs/go-onfido) - Another version of go SDK for Onfido API. This sdk was found after bootstrapping this SDK project.

## License

MIT License
