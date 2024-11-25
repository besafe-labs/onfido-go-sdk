# Onfido Go SDK

![CodeQL](https://github.com/besafe-labs/onfido-go-sdk/actions/workflows/codeql.yml/badge.svg)
![Build & Tests](https://github.com/besafe-labs/onfido-go-sdk/actions/workflows/build.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/besafe-labs/onfido-go-sdk)](https://goreportcard.com/report/github.com/besafe-labs/onfido-go-sdk)
[![Go Reference](https://pkg.go.dev/badge/github.com/besafe-labs/onfido-go-sdk.svg)](https://pkg.go.dev/github.com/besafe-labs/onfido-go-sdk)

An unofficial Go SDK for interacting with the Onfido API. This SDK was created because Onfido does not provide an official Go SDK for developers.

## Installation

```bash
go get -u github.com/besafe-labs/onfido-go-sdk
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

- All endpoints related to applicants

### Workflow Runs

- All endpoints related to workflow runs

### Documents

- All endpoints related to documents

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
